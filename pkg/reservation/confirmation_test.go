package reservation_test

import (
	"fmt"
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
