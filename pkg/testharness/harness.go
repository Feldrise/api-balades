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
	"feldrise.com/balade/database/dbmodel"
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

	// referenceSchema holds a copy of the reference data, restored before each test.
	referenceSchema = "testharness_reference"
)

// Tables of the reference data created by the seeds, restored before each test.
var referenceTables = []string{"seeds", "permissions", "roles", "role_permissions"}

var (
	openOnce    sync.Once
	sharedDB    *gorm.DB
	openErr     error
	prepareOnce sync.Once
	prepareErr  error
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

// Option changes the API run by the harness.
type Option func(services *config.Services)

// WithPaymentTimeouts shortens the timeouts of the decisions on payments, so
// that a test can outlast them.
func WithPaymentTimeouts(timeouts dbmodel.PaymentTimeouts) Option {
	return func(services *config.Services) {
		services.PaymentTimeouts = timeouts
	}
}

// New returns a harness on an emptied test database. The test is skipped when
// TEST_DATABASE_URL is not set.
func New(t *testing.T, options ...Option) *Harness {
	t.Helper()

	db := openDatabase(t)
	lockDatabase(t, db)

	prepareOnce.Do(func() {
		database.Migrate(db)
		prepareErr = snapshotReferenceData(db)
	})
	if prepareErr != nil {
		t.Fatalf("preparing test database: %v", prepareErr)
	}

	resetDatabase(t, db)

	emails := NewFakeEmailService(TemplatesPath())
	stripe := NewFakeStripeService()
	clock := NewFakeClock()

	constants := config.Constants{
		JWTSecret:      jwtSecret,
		DataPath:       t.TempDir(),
		ApplicationURL: applicationURL,
	}
	services := config.Services{
		EmailService:  emails,
		StripeService: stripe,
		Clock:         clock,
	}
	for _, option := range options {
		option(&services)
	}
	configuration := config.Build(constants, db, services)

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

// snapshotReferenceData rebuilds the reference data from the seeds on an
// emptied database, whatever a previous run left, and copies it aside.
func snapshotReferenceData(db *gorm.DB) error {
	if err := emptyTables(db, nil); err != nil {
		return err
	}
	database.ApplySeeds(db)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DROP SCHEMA IF EXISTS "` + referenceSchema + `" CASCADE`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE SCHEMA "` + referenceSchema + `"`).Error; err != nil {
			return err
		}
		for _, table := range referenceTables {
			if err := tx.Exec(`CREATE TABLE "` + referenceSchema + `"."` + table + `" AS TABLE "` + table + `"`).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// resetDatabase empties every table, restores the reference data and restarts
// identities after it. DELETE is used rather than TRUNCATE, which is much slower
// on small tables.
func resetDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	err := emptyTables(db, func(tx *gorm.DB) error {
		for _, table := range referenceTables {
			if err := tx.Exec(`INSERT INTO "` + table + `" SELECT * FROM "` + referenceSchema + `"."` + table + `"`).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("resetting test database: %v", err)
	}
}

// emptyTables deletes the rows of every table, runs fill in the same
// transaction, and restarts every identity after the remaining rows.
func emptyTables(db *gorm.DB, fill func(tx *gorm.DB) error) error {
	var tables []string
	if err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Skips foreign key checks, so that tables can be emptied in any order.
		if err := tx.Exec("SET LOCAL session_replication_role = replica").Error; err != nil {
			return err
		}

		for _, table := range tables {
			if err := tx.Exec(`DELETE FROM "` + table + `"`).Error; err != nil {
				return err
			}
		}

		if fill != nil {
			if err := fill(tx); err != nil {
				return err
			}
		}

		var sequences []struct {
			Table  string
			Column string
		}
		err := tx.Raw(`
			SELECT table_name AS "table", column_name AS "column"
			FROM information_schema.columns
			WHERE table_schema = 'public' AND column_default LIKE 'nextval%'`).Scan(&sequences).Error
		if err != nil {
			return err
		}
		for _, sequence := range sequences {
			err := tx.Exec(`SELECT setval(pg_get_serial_sequence(?, ?), COALESCE((SELECT MAX("`+sequence.Column+`") FROM "`+sequence.Table+`"), 0) + 1, false)`,
				`"`+sequence.Table+`"`, sequence.Column).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
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
