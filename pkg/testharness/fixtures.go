package testharness

import (
	"time"

	"feldrise.com/balade/database"
	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/notifications/email"
)

// CreateUser inserts a user with the given roles ("admin", "user", "guide").
func (h *Harness) CreateUser(address string, roles ...string) *dbmodel.User {
	h.t.Helper()

	user := &dbmodel.User{
		Email:       address,
		UserProfile: dbmodel.UserProfile{FirstName: "Test"},
	}

	for _, roleName := range roles {
		var role dbmodel.Role
		if err := h.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
			h.t.Fatalf("finding role %q: %v", roleName, err)
		}
		user.Roles = append(user.Roles, role)
	}

	if err := h.DB.Create(user).Error; err != nil {
		h.t.Fatalf("creating user %q: %v", address, err)
	}

	return user
}

// Token returns a valid JWT for the user.
func (h *Harness) Token(user *dbmodel.User) string {
	h.t.Helper()

	token, err := authentication.GenerateToken(jwtSecret, user.ID)
	if err != nil {
		h.t.Fatalf("generating token: %v", err)
	}

	return token
}

// Price is a ramble price for CreateRamble.
type Price struct {
	Label  string
	Amount float64
}

// RambleOptions describes a ramble created by CreateRamble.
type RambleOptions struct {
	Title string
	// Date defaults to 10 days after the clock's current time.
	Date            *time.Time
	MaxParticipants *int
	Prices          []Price
	Unpublished     bool
	PaymentEnabled  bool
	PaymentRequired bool
}

// CreateRamble inserts a ramble, published by default.
func (h *Harness) CreateRamble(options RambleOptions) *dbmodel.Ramble {
	h.t.Helper()

	date := h.Clock.Now().AddDate(0, 0, 10)
	if options.Date != nil {
		date = *options.Date
	}

	title := options.Title
	if title == "" {
		title = "Balade des oiseaux"
	}
	location := "Forêt de Brocéliande"

	ramble := &dbmodel.Ramble{
		Title:           title,
		Date:            &date,
		Location:        &location,
		MaxParticipants: options.MaxParticipants,
		PaymentEnabled:  options.PaymentEnabled,
		PaymentRequired: options.PaymentRequired,
	}
	if !options.Unpublished {
		publishedAt := time.Now().Add(-time.Hour)
		ramble.PublishedAt = &publishedAt
	}
	for _, price := range options.Prices {
		ramble.Prices = append(ramble.Prices, dbmodel.RamblePrice{Label: price.Label, Amount: price.Amount})
	}

	if err := h.DB.Create(ramble).Error; err != nil {
		h.t.Fatalf("creating ramble: %v", err)
	}

	return ramble
}

// Int returns a pointer to value.
func Int(value int) *int {
	return &value
}

// Login goes through the email verification login flow and returns the token.
func (h *Harness) Login(address string) string {
	h.t.Helper()

	firstName := "Test"
	response := h.Do("POST", "/authentication/authenticate", map[string]any{
		"email":      address,
		"first_name": firstName,
	}, "")
	if response.Code != 200 {
		h.t.Fatalf("authenticate %q: expected 200, got %d: %s", address, response.Code, response.Body)
	}

	sent := h.Emails.SentWithTemplate("authentication")
	if len(sent) == 0 {
		h.t.Fatalf("authenticate %q: no authentication email sent", address)
	}
	data, ok := sent[len(sent)-1].Data.(email.AuthenticationData)
	if !ok {
		h.t.Fatalf("authenticate %q: unexpected authentication email data %T", address, sent[len(sent)-1].Data)
	}

	response = h.Do("POST", "/authentication/verify-authentication", map[string]any{
		"email": address,
		"code":  data.Code,
	}, "")
	if response.Code != 200 {
		h.t.Fatalf("verify authentication %q: expected 200, got %d: %s", address, response.Code, response.Body)
	}

	var body struct {
		Token string `json:"token"`
	}
	response.JSON(h.t, &body)

	return body.Token
}

// CreateGuide inserts a guide profile linked to user and assigned to the rambles.
func (h *Harness) CreateGuide(user *dbmodel.User, rambles ...*dbmodel.Ramble) *dbmodel.Guide {
	h.t.Helper()

	guide := &dbmodel.Guide{
		UserID:    &user.ID,
		FirstName: "Guide",
		LastName:  "Nature",
		Email:     user.Email,
	}
	if err := h.DB.Create(guide).Error; err != nil {
		h.t.Fatalf("creating guide: %v", err)
	}

	for _, ramble := range rambles {
		if err := h.DB.Create(&dbmodel.RambleGuide{RambleID: ramble.ID, GuideID: guide.ID}).Error; err != nil {
			h.t.Fatalf("assigning guide to ramble %d: %v", ramble.ID, err)
		}
	}

	return guide
}

// MigrateAtStartup runs the database migration the API runs when it starts.
func (h *Harness) MigrateAtStartup() {
	h.t.Helper()

	database.Migrate(h.DB)
}
