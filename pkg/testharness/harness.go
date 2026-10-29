// Package testharness runs the complete API router against a Postgres test
// database, with fake email, Stripe and clock, so that tests exercise the API
// only through HTTP.
package testharness

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"feldrise.com/balade/config"
	"feldrise.com/balade/database"
	"feldrise.com/balade/pkg/router"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// DatabaseURLEnv names the environment variable holding the test database URL.
	DatabaseURLEnv = "TEST_DATABASE_URL"

	jwtSecret      = "test-jwt-secret"
	applicationURL = "https://app.balade.test"

	// databaseLockID serializes tests of every package on the shared test database.
	databaseLockID = 742001
)

// Tables kept between tests: reference data created by the seeds.
var preservedTables = map[string]bool{
	"seeds":            true,
	"permissions":      true,
	"roles":            true,
	"role_permissions": true,
}

var (
	openOnce   sync.Once
	sharedDB   *gorm.DB
	openErr    error
	migrateMux sync.Once
)

type Harness struct {
	t *testing.T

	DB     *gorm.DB
	Config *config.Config
	Router http.Handler

	Emails *FakeEmailService
	Stripe *FakeStripeService
	Clock  *FakeClock
}

// New returns a harness on an emptied test database. The test is skipped when
// TEST_DATABASE_URL is not set.
func New(t *testing.T) *Harness {
	t.Helper()

	db := openDatabase(t)
	lockDatabase(t, db)

	migrateMux.Do(func() {
		database.Migrate(db)
	})

	resetDatabase(t, db)

	emails := NewFakeEmailService(TemplatesPath())
	stripe := NewFakeStripeService()
	clock := NewFakeClock()

	constants := config.Constants{
		JWTSecret:      jwtSecret,
		DataPath:       t.TempDir(),
		ApplicationURL: applicationURL,
	}
	configuration := config.Build(constants, db, config.Services{
		EmailService:  emails,
		StripeService: stripe,
		Clock:         clock,
	})

	return &Harness{
		t:      t,
		DB:     db,
		Config: configuration,
		Router: router.New(configuration),
		Emails: emails,
		Stripe: stripe,
		Clock:  clock,
	}
}

// TemplatesPath is the directory of the real email templates of the repository.
func TemplatesPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "emails")
}

func openDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	url := os.Getenv(DatabaseURLEnv)
	if url == "" {
		t.Skipf("%s is not set: skipping test that needs the Postgres test database", DatabaseURLEnv)
	}

	openOnce.Do(func() {
		sharedDB, openErr = gorm.Open(postgres.Open(url), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	})
	if openErr != nil {
		t.Fatalf("opening test database: %v", openErr)
	}

	return sharedDB
}

// lockDatabase holds a Postgres advisory lock for the duration of the test, so
// that test packages run by `go test ./...` in parallel never share the database.
func lockDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("getting sql database: %v", err)
	}

	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("getting database connection: %v", err)
	}

	if _, err := conn.ExecContext(context.Background(), "SELECT pg_advisory_lock($1)", databaseLockID); err != nil {
		conn.Close()
		t.Fatalf("locking test database: %v", err)
	}

	t.Cleanup(func() {
		unlock(conn)
	})
}

func unlock(conn *sql.Conn) {
	conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", databaseLockID)
	conn.Close()
}

func resetDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	var tables []string
	if err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables).Error; err != nil {
		t.Fatalf("listing tables: %v", err)
	}

	toTruncate := make([]string, 0, len(tables))
	for _, table := range tables {
		if !preservedTables[table] {
			toTruncate = append(toTruncate, `"`+table+`"`)
		}
	}
	if len(toTruncate) == 0 {
		return
	}

	query := "TRUNCATE TABLE "
	for i, table := range toTruncate {
		if i > 0 {
			query += ", "
		}
		query += table
	}
	query += " RESTART IDENTITY CASCADE"

	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("resetting test database: %v", err)
	}
}

// Response is a recorded HTTP response.
type Response struct {
	Code int
	Body []byte
}

// JSON decodes the response body into target, failing the test on error.
func (r *Response) JSON(t *testing.T, target any) {
	t.Helper()

	if err := json.Unmarshal(r.Body, target); err != nil {
		t.Fatalf("decoding response body %q: %v", string(r.Body), err)
	}
}

// Do sends a request to the API router. path is relative to /api/v1, body is
// encoded as JSON when not nil, and token authenticates the request when not empty.
func (h *Harness) Do(method string, path string, body any, token string) *Response {
	h.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("encoding request body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}

	request := httptest.NewRequest(method, "/api/v1"+path, reader)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, request)

	return &Response{Code: recorder.Code, Body: recorder.Body.Bytes()}
}
