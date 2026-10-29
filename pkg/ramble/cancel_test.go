package ramble_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"feldrise.com/balade/pkg/testharness"
)

type participant struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsBooker  bool   `json:"is_booker,omitempty"`
}

func reserveFor(t *testing.T, h *testharness.Harness, rambleID uint, address string, firstName string) uint {
	t.Helper()

	response := h.Do(http.MethodPost, "/reservations", map[string]any{
		"ramble_id":     rambleID,
		"contact_email": address,
		"participants":  []participant{{FirstName: firstName, LastName: "Martin", IsBooker: true}, {FirstName: "Invité", LastName: "Martin"}},
	}, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("POST /reservations: expected 201, got %d: %s", response.Code, response.Body)
	}
	var created struct {
		ID uint `json:"id"`
	}
	response.JSON(t, &created)
	return created.ID
}

func TestCancellingARambleTellsTheBookersOfItsActiveReservations(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{Title: "Balade des oiseaux", MaxParticipants: testharness.Int(2)})
	reserveFor(t, h, ramble.ID, "pending@example.com", "Alice")
	reserveFor(t, h, ramble.ID, "waiting@example.com", "Bob")
	cancelled := reserveFor(t, h, ramble.ID, "cancelled@example.com", "Carol")
	h.Do(http.MethodPut, fmt.Sprintf("/reservations/%d/cancel", cancelled), nil, h.Login("cancelled@example.com"))
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := h.Do(http.MethodPut, fmt.Sprintf("/rambles/%d/cancel", ramble.ID), map[string]string{"reason": "Orage annoncé"}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	// The emails are sent in the background.
	deadline := time.Now().Add(3 * time.Second)
	for len(h.Emails.SentWithTemplate("ramble_cancellation")) < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)

	sent := h.Emails.SentWithTemplate("ramble_cancellation")
	if len(sent) != 2 {
		t.Fatalf("expected the 2 bookers of active reservations to be told, got %d emails", len(sent))
	}
	got := map[string]string{}
	for _, email := range sent {
		got[email.To] = email.Body
	}
	for address, name := range map[string]string{"pending@example.com": "Alice", "waiting@example.com": "Bob"} {
		body, ok := got[address]
		if !ok {
			t.Errorf("expected %s to be told, got %v", address, got)
			continue
		}
		for _, expected := range []string{"Bonjour " + name, "Balade des oiseaux", "Orage annoncé"} {
			if !strings.Contains(body, expected) {
				t.Errorf("expected the email to %s to contain %q", address, expected)
			}
		}
	}
}
