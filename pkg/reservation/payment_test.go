package reservation_test

import (
	"fmt"
	"net/http"
	"testing"

	"feldrise.com/balade/pkg/testharness"
)

func TestOnlyTheBookerOrAnAuthorizedStaffCanReadAPayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	otherRamble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	paid := h.PayReservation(created.ID, 2500)

	admin := h.CreateUser("admin@example.com", "admin")
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)
	stranger := h.CreateUser("bob@example.com", "user")

	for _, c := range []struct {
		name  string
		token string
		code  int
	}{
		{"booker", h.Login("alice@example.com"), http.StatusOK},
		{"admin", h.Token(admin), http.StatusOK},
		{"guide of the ramble", h.Token(owningGuide), http.StatusOK},
		{"guide of another ramble", h.Token(otherGuide), http.StatusForbidden},
		{"another user", h.Token(stranger), http.StatusForbidden},
		{"anonymous visitor", "", http.StatusUnauthorized},
	} {
		response := h.Do(http.MethodGet, fmt.Sprintf("/payments/%d", paid.ID), nil, c.token)
		if response.Code != c.code {
			t.Errorf("%s: expected %d, got %d: %s", c.name, c.code, response.Code, response.Body)
		}
	}
}

func refundPayment(h *testharness.Harness, id uint, amount *int64, token string) *testharness.Response {
	return h.Do(http.MethodPost, fmt.Sprintf("/payments/%d/refund", id), map[string]any{"amount": amount}, token)
}

func TestOnlyStaffAllowedToRefundCanRefundAPayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	paid := h.PayReservation(created.ID, 2500)
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	admin := h.CreateUser("admin@example.com", "admin")

	for _, c := range []struct {
		name  string
		token string
		code  int
	}{
		{"anonymous visitor", "", http.StatusUnauthorized},
		{"booker", h.Login("alice@example.com"), http.StatusForbidden},
		{"guide of the ramble", h.Token(owningGuide), http.StatusForbidden},
	} {
		if response := refundPayment(h, paid.ID, nil, c.token); response.Code != c.code {
			t.Errorf("%s: expected %d, got %d: %s", c.name, c.code, response.Code, response.Body)
		}
	}
	if refunds := h.Stripe.Refunds(); len(refunds) != 0 {
		t.Fatalf("expected no refund by unauthorized users, got %+v", refunds)
	}

	if response := refundPayment(h, paid.ID, nil, h.Token(admin)); response.Code != http.StatusOK {
		t.Errorf("admin: expected 200, got %d: %s", response.Code, response.Body)
	}
}
