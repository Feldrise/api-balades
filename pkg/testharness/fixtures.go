package testharness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"feldrise.com/balade/database"
	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/notifications/email"
	"feldrise.com/balade/pkg/security"
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
	// PaymentEnabled enables online payment, to a paying guide created with
	// Stripe credentials (see PaymentGuidePublicKey).
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

	if options.PaymentEnabled {
		guide := h.createPayingGuide()
		ramble.PaymentGuideID = &guide.ID
	}

	if err := h.DB.Create(ramble).Error; err != nil {
		h.t.Fatalf("creating ramble: %v", err)
	}

	return ramble
}

// PaymentGuidePublicKey is the Stripe public key of the paying guides of rambles with online payment.
const PaymentGuidePublicKey = "pk_test_guide"

// PaymentGuideWebhookSecret is the Stripe webhook secret of the paying guides.
const PaymentGuideWebhookSecret = "whsec_test_guide"

func (h *Harness) createPayingGuide() *dbmodel.Guide {
	h.t.Helper()

	encryption := security.NewEncryptionService(jwtSecret)
	secretKey, err := encryption.Encrypt("sk_test_guide")
	if err != nil {
		h.t.Fatalf("encrypting Stripe secret key: %v", err)
	}
	webhookSecret, err := encryption.Encrypt(PaymentGuideWebhookSecret)
	if err != nil {
		h.t.Fatalf("encrypting Stripe webhook secret: %v", err)
	}
	publicKey := PaymentGuidePublicKey

	var count int64
	h.DB.Model(&dbmodel.Guide{}).Count(&count)
	guide := &dbmodel.Guide{
		FirstName:           "Guide",
		LastName:            "Payant",
		Email:               fmt.Sprintf("paying-guide-%d@example.com", count+1),
		StripePublicKey:     &publicKey,
		StripeSecretKey:     &secretKey,
		StripeWebhookSecret: &webhookSecret,
		PaymentEnabled:      true,
	}
	if err := h.DB.Create(guide).Error; err != nil {
		h.t.Fatalf("creating paying guide: %v", err)
	}
	return guide
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

// PayReservation records a succeeded payment of amount cents for the reservation,
// as if the booker had paid it.
func (h *Harness) PayReservation(reservationID uint, amount int64) *dbmodel.Payment {
	h.t.Helper()

	var reservation dbmodel.Reservation
	if err := h.DB.First(&reservation, reservationID).Error; err != nil {
		h.t.Fatalf("finding reservation %d: %v", reservationID, err)
	}

	guide := h.createPayingGuide()

	var count int64
	h.DB.Model(&dbmodel.Payment{}).Count(&count)
	paidAt := h.Clock.Now()
	payment := &dbmodel.Payment{
		StripePaymentIntentID: fmt.Sprintf("pi_paid_%d", count+1),
		Amount:                amount,
		Status:                "succeeded",
		ReservationID:         &reservationID,
		PayerEmail:            reservation.ContactEmail,
		PayerName:             "Payer",
		GuideID:               guide.ID,
		PaidAt:                &paidAt,
	}
	if err := h.DB.Create(payment).Error; err != nil {
		h.t.Fatalf("creating payment: %v", err)
	}

	return payment
}

// CreateUserWithPermissions inserts a user whose single role grants exactly the
// given permissions, to test what each permission allows on its own.
func (h *Harness) CreateUserWithPermissions(address string, permissions ...string) *dbmodel.User {
	h.t.Helper()

	role := dbmodel.Role{Name: "only-" + address}
	for _, name := range permissions {
		var permission dbmodel.Permission
		if err := h.DB.Where("name = ?", name).First(&permission).Error; err != nil {
			h.t.Fatalf("finding permission %q: %v", name, err)
		}
		role.Permissions = append(role.Permissions, permission)
	}
	if err := h.DB.Create(&role).Error; err != nil {
		h.t.Fatalf("creating role for %q: %v", address, err)
	}

	user := &dbmodel.User{
		Email:       address,
		UserProfile: dbmodel.UserProfile{FirstName: "Test"},
		Roles:       []dbmodel.Role{role},
	}
	if err := h.DB.Create(user).Error; err != nil {
		h.t.Fatalf("creating user %q: %v", address, err)
	}

	return user
}

// SendStripeWebhook sends the paying guide's webhook a Stripe event of the
// given type (payment_intent.succeeded, payment_intent.payment_failed...) on
// the payment intent, signed as Stripe would.
func (h *Harness) SendStripeWebhook(guideID uint, eventType string, paymentIntentID string) *Response {
	h.t.Helper()

	event := map[string]any{
		"id":     fmt.Sprintf("evt_%s_%s", eventType, paymentIntentID),
		"object": "event",
		"type":   eventType,
		"data": map[string]any{
			"object": map[string]any{"id": paymentIntentID, "object": "payment_intent"},
		},
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		h.t.Fatalf("encoding webhook event: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/payments/webhook?guide_id=%d", guideID), bytes.NewReader(encoded))
	request.Header.Set("Stripe-Signature", FakeWebhookSignature(PaymentGuideWebhookSecret))
	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, request)

	return &Response{Code: recorder.Code, Body: recorder.Body.Bytes()}
}
