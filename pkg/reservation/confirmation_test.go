package reservation_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	reservations "feldrise.com/balade/pkg/reservation"
	"feldrise.com/balade/pkg/testharness"
)

// sendConfirmationRequests runs the scheduler's confirmation requests at the clock's time.
func sendConfirmationRequests(t *testing.T, h *testharness.Harness) {
	t.Helper()

	if err := reservations.New(h.Config).SendConfirmationRequests(h.Clock.Now()); err != nil {
		t.Fatalf("sending confirmation requests: %v", err)
	}
}

func TestSchedulerSendsASingleConfirmationRequestToTheBookerThreeDaysBefore(t *testing.T) {
	h := testharness.New(t)
	soon := h.CreateRamble(testharness.RambleOptions{Title: "Balade bientôt"})
	later := h.CreateRamble(testharness.RambleOptions{Title: "Balade plus tard", Date: timePtr(h.Clock.Now().AddDate(0, 0, 20))})
	created := reserve(t, h, reservationPayload{RambleID: soon.ID, ContactEmail: "alice@example.com", Participants: people(3, "")})
	reserve(t, h, reservationPayload{RambleID: later.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})

	h.Clock.Set(soon.Date.Add(-3*24*time.Hour + time.Hour))
	sendConfirmationRequests(t, h)
	h.Clock.Advance(time.Hour)
	sendConfirmationRequests(t, h)

	sent := h.Emails.SentWithTemplate("event_confirmation_request")
	if len(sent) != 1 {
		t.Fatalf("expected a single confirmation request, got %d", len(sent))
	}
	if sent[0].To != "alice@example.com" {
		t.Errorf("expected the request to go to the booker, got %q", sent[0].To)
	}
	link := fmt.Sprintf("https://app.balade.test/mes-reservations/%d", created.ID)
	for _, expected := range []string{link, "Balade bientôt"} {
		if !strings.Contains(sent[0].Body, expected) {
			t.Errorf("expected the request to contain %q", expected)
		}
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}

// cancelUnconfirmedReservations runs the scheduler's cancellation of unconfirmed reservations at the clock's time.
func cancelUnconfirmedReservations(t *testing.T, h *testharness.Harness) {
	t.Helper()

	if err := reservations.New(h.Config).CancelUnconfirmedReservations(h.Clock.Now()); err != nil {
		t.Fatalf("cancelling unconfirmed reservations: %v", err)
	}
}

func TestSchedulerCancelsUnconfirmedReservationsAtTheDeadlineAndAdmitsTheWaitingList(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	unconfirmed := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(2, "")})
	alice := h.Login("alice@example.com")
	bob := h.Login("bob@example.com")

	h.Clock.Set(ramble.Date.Add(-13 * time.Hour))
	cancelUnconfirmedReservations(t, h)
	if got := getReservation(t, h, unconfirmed.ID, alice); got.Status != "pending" {
		t.Fatalf("expected the reservation to stay pending before the deadline, got %q", got.Status)
	}

	h.Clock.Set(ramble.Date.Add(-11 * time.Hour))
	cancelUnconfirmedReservations(t, h)

	if got := getReservation(t, h, unconfirmed.ID, alice); got.Status != "cancelled" {
		t.Errorf("expected the unconfirmed reservation to be cancelled, got %q", got.Status)
	}
	if got := getReservation(t, h, waiting.ID, bob); got.Status != "confirmed" || len(got.Participants) != 2 {
		t.Errorf("expected the whole waiting reservation to be admitted, got %q with %d participants", got.Status, len(got.Participants))
	}
	sent := h.Emails.SentWithTemplate("reservation_confirmation")
	if last := sent[len(sent)-1]; last.To != "bob@example.com" || !strings.Contains(last.Subject, "Place disponible") {
		t.Errorf("expected bob to be told a place is available, got %q to %q", last.Subject, last.To)
	}
}

func confirmReservation(h *testharness.Harness, id uint, token string) *testharness.Response {
	return h.Do(http.MethodPut, fmt.Sprintf("/reservations/%d/confirm", id), nil, token)
}

func TestBookerConfirmsAReservationWithoutOnlinePayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")

	response := confirmReservation(h, created.ID, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var body struct {
		Status  string `json:"status"`
		Payment any    `json:"payment"`
	}
	response.JSON(t, &body)
	if body.Status != "confirmed" || body.Payment != nil {
		t.Errorf("expected a confirmed reservation without payment to make, got %s", response.Body)
	}
	if got := getReservation(t, h, created.ID, token); got.Status != "confirmed" || got.ConfirmationDate == nil {
		t.Errorf("expected the reservation to be confirmed with its date, got %+v", got)
	}
	if intents := h.Stripe.PaymentIntents(); len(intents) != 0 {
		t.Errorf("expected no payment without online payment, got %+v", intents)
	}
}

func TestReservationWithAMissingPriceCannotBeConfirmed(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(familyRamble())
	group := l.group(ramble, "parent@example.com", "pending")
	l.registration(ramble, "Paul", "parent@example.com", "pending", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "pending", &group.ID)
	h.MigrateAtStartup()
	migrated := myReservations(t, h, "parent@example.com")[0]
	token := h.Login("parent@example.com")

	response := confirmReservation(h, migrated.ID, token)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "missing_prices" {
		t.Errorf("expected reason missing_prices, got %s", response.Body)
	}
	if got := getReservation(t, h, migrated.ID, token); got.Status != "pending" {
		t.Errorf("expected the reservation to stay pending, got %q", got.Status)
	}
}

func TestWaitingReservationCannotBeConfirmed(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(1)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "first@example.com", Participants: people(1, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	token := h.Login("alice@example.com")

	response := confirmReservation(h, waiting.ID, token)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "reservation_waiting_list" {
		t.Errorf("expected reason reservation_waiting_list, got %s", response.Body)
	}
	if got := getReservation(t, h, waiting.ID, token); got.Status != "waiting_list" {
		t.Errorf("expected the reservation to keep waiting, got %q", got.Status)
	}
}

func TestOnlyTheBookerCanConfirmAReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	admin := h.CreateUser("admin@example.com", "admin")

	for _, c := range []struct {
		name  string
		token string
		code  int
	}{
		{"anonymous visitor", "", http.StatusUnauthorized},
		{"another user", h.Login("mallory@example.com"), http.StatusForbidden},
		{"admin", h.Token(admin), http.StatusForbidden},
	} {
		if response := confirmReservation(h, created.ID, c.token); response.Code != c.code {
			t.Errorf("%s: expected %d, got %d: %s", c.name, c.code, response.Code, response.Body)
		}
	}
	if got := getReservation(t, h, created.ID, h.Login("alice@example.com")); got.Status != "pending" {
		t.Errorf("expected the reservation to stay pending, got %q", got.Status)
	}
}

type paymentToMake struct {
	Payment struct {
		ID            uint   `json:"id"`
		Amount        int64  `json:"amount"`
		Status        string `json:"status"`
		ReservationID *uint  `json:"reservation_id"`
	} `json:"payment"`
	ClientSecret   string `json:"client_secret"`
	PublishableKey string `json:"publishable_key"`
}

type reservationWithPayment struct {
	reservation
	Payment *paymentToMake `json:"payment"`
}

func payingFamilyRamble() testharness.RambleOptions {
	options := familyRamble()
	options.PaymentEnabled = true
	return options
}

func TestConfirmationCreatesThePaymentOfTheParticipantsPrices(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: append(people(2, "Adulte"), people(1, "Enfant")...)})
	token := h.Login("alice@example.com")

	response := confirmReservation(h, created.ID, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var body reservationWithPayment
	response.JSON(t, &body)
	if body.Status != "confirmed" || body.Payment == nil {
		t.Fatalf("expected a confirmed reservation with a payment to make, got %s", response.Body)
	}
	if p := body.Payment; p.Payment.Amount != 3100 || p.ClientSecret == "" || p.PublishableKey != testharness.PaymentGuidePublicKey ||
		p.Payment.ReservationID == nil || *p.Payment.ReservationID != created.ID {
		t.Errorf("expected a payment of 31,00 € for the reservation to the guide's account, got %+v", p)
	}

	intents := h.Stripe.PaymentIntents()
	if len(intents) != 1 || intents[0].Amount != 3100 || intents[0].Credentials.SecretKey != "sk_test_guide" {
		t.Fatalf("expected a single Stripe payment of 31,00 € on the guide's account, got %+v", intents)
	}
	if got := getReservation(t, h, created.ID, token); len(got.Payments) != 1 || got.Payments[0].Status != "pending" {
		t.Errorf("expected the reservation to have a pending payment, got %+v", got.Payments)
	}
}

func TestConfirmedReservationWithAMissingPriceCannotBePaid(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(payingFamilyRamble())
	group := l.group(ramble, "parent@example.com", "confirmed")
	l.registration(ramble, "Paul", "parent@example.com", "confirmed", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "confirmed", &group.ID)
	h.MigrateAtStartup()
	migrated := myReservations(t, h, "parent@example.com")[0]
	token := h.Login("parent@example.com")

	response := confirmReservation(h, migrated.ID, token)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "missing_prices" {
		t.Errorf("expected reason missing_prices, got %s", response.Body)
	}
	if intents := h.Stripe.PaymentIntents(); len(intents) != 0 {
		t.Errorf("expected no payment of an incomplete amount, got %+v", intents)
	}
}

func TestConfirmingAnUnpaidConfirmedReservationGivesItsPaymentOnceOnly(t *testing.T) {
	h := testharness.New(t)
	options := payingFamilyRamble()
	ramble := h.CreateRamble(options)
	h.Clock.Set(ramble.Date.Add(-2 * 24 * time.Hour))
	confirmed := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	if confirmed.Status != "confirmed" {
		t.Fatalf("expected the reservation confirmed on creation within 3 days, got %q", confirmed.Status)
	}
	token := h.Login("alice@example.com")

	var unpaid reservationWithPayment
	confirmReservation(h, confirmed.ID, token).JSON(t, &unpaid)
	if unpaid.Payment == nil || unpaid.Payment.Payment.Amount != 2500 {
		t.Fatalf("expected the payment of 25,00 € of the unpaid reservation, got %+v", unpaid.Payment)
	}

	h.PayReservation(confirmed.ID, 2500)
	var paid reservationWithPayment
	confirmReservation(h, confirmed.ID, token).JSON(t, &paid)
	if paid.Status != "confirmed" || paid.Payment != nil {
		t.Errorf("expected nothing more to pay once paid, got %+v", paid.Payment)
	}
}

func TestConfirmedButUnpaidReservationKeepsItsPlacesAtTheDeadline(t *testing.T) {
	h := testharness.New(t)
	options := payingFamilyRamble()
	options.PaymentRequired = true
	ramble := h.CreateRamble(options)
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	if response := confirmReservation(h, created.ID, token); response.Code != http.StatusOK {
		t.Fatalf("confirming: expected 200, got %d: %s", response.Code, response.Body)
	}

	h.Clock.Set(ramble.Date.Add(-time.Hour))
	cancelUnconfirmedReservations(t, h)

	if got := getReservation(t, h, created.ID, token); got.Status != "confirmed" {
		t.Errorf("expected the confirmed reservation to keep its places without payment, got %q", got.Status)
	}
}
