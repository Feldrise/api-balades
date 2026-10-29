package testharness

import (
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
)

// CreateUser inserts a user with the given roles ("admin", "user", "guide").
func (h *Harness) CreateUser(email string, roles ...string) *dbmodel.User {
	h.t.Helper()

	user := &dbmodel.User{
		Email:       email,
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
		h.t.Fatalf("creating user %q: %v", email, err)
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
