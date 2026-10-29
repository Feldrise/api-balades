package testharness_test

import (
	"fmt"
	"net/http"
	"testing"

	"feldrise.com/balade/pkg/testharness"
)

func TestHarnessServesPublishedRamble(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{
		Title:  "Balade du marais",
		Prices: []testharness.Price{{Label: "Adulte", Amount: 12}},
	})

	response := h.Do(http.MethodGet, fmt.Sprintf("/rambles/%d", ramble.ID), nil, "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var body struct {
		Title  string `json:"title"`
		Prices []struct {
			Label string `json:"label"`
		} `json:"prices"`
	}
	response.JSON(t, &body)
	if body.Title != "Balade du marais" || len(body.Prices) != 1 || body.Prices[0].Label != "Adulte" {
		t.Fatalf("unexpected ramble: %s", response.Body)
	}
}

func TestHarnessAuthenticatesWithTestToken(t *testing.T) {
	h := testharness.New(t)
	user := h.CreateUser("alice@example.com", "user")

	authenticated := h.Do(http.MethodGet, "/authentication/me", nil, h.Token(user))
	anonymous := h.Do(http.MethodGet, "/authentication/me", nil, "")

	if authenticated.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", authenticated.Code, authenticated.Body)
	}
	var me struct {
		Email string `json:"email"`
	}
	authenticated.JSON(t, &me)
	if me.Email != "alice@example.com" {
		t.Fatalf("expected alice@example.com, got %q", me.Email)
	}
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous request, got %d", anonymous.Code)
	}
}

func TestHarnessResetsDatabaseBetweenTests(t *testing.T) {
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("run %d", i), func(t *testing.T) {
			h := testharness.New(t)
			// Fails on the unique email constraint if the previous run's data remains.
			user := h.CreateUser("same@example.com")
			if user.ID != 1 {
				t.Fatalf("expected identities to restart at 1, got %d", user.ID)
			}
		})
	}
}
