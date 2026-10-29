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
