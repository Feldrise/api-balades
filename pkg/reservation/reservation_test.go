package reservation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"feldrise.com/balade/pkg/testharness"
)

type participantPayload struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PriceLabel string `json:"price_label,omitempty"`
	IsBooker   bool   `json:"is_booker,omitempty"`
}

type reservationPayload struct {
	RambleID     uint                 `json:"ramble_id"`
	ContactEmail string               `json:"contact_email"`
	Participants []participantPayload `json:"participants"`
}

type participant struct {
	ID          uint    `json:"id"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	PriceLabel  string  `json:"price_label"`
	PriceAmount float64 `json:"price_amount"`
	IsBooker    bool    `json:"is_booker"`
}

type payment struct {
	ID     uint   `json:"id"`
	Amount int64  `json:"amount"`
	Status string `json:"status"`
}

type reservation struct {
	ID               uint          `json:"id"`
	RambleID         uint          `json:"ramble_id"`
	ContactEmail     string        `json:"contact_email"`
	Status           string        `json:"status"`
	ConfirmationDate *time.Time    `json:"confirmation_date"`
	Participants     []participant `json:"participants"`
	EstimatedAmount  int64         `json:"estimated_amount"`
	Payments         []payment     `json:"payments"`
	Ramble           *struct {
		Title string `json:"title"`
	} `json:"ramble"`
}

type errorBody struct {
	Reason        string `json:"reason"`
	ReservationID uint   `json:"reservation_id"`
}

func reserve(t *testing.T, h *testharness.Harness, payload reservationPayload) reservation {
	t.Helper()

	response := h.Do(http.MethodPost, "/reservations", payload, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("POST /reservations: expected 201, got %d: %s", response.Code, response.Body)
	}

	var created reservation
	response.JSON(t, &created)
	return created
}

func people(count int, priceLabel string) []participantPayload {
	result := make([]participantPayload, count)
	for i := range result {
		result[i] = participantPayload{FirstName: fmt.Sprintf("Personne%d", i+1), LastName: "Dupont", PriceLabel: priceLabel}
	}
	return result
}

func familyRamble() testharness.RambleOptions {
	return testharness.RambleOptions{
		MaxParticipants: testharness.Int(10),
		Prices: []testharness.Price{
			{Label: "Adulte", Amount: 12.5},
			{Label: "Enfant", Amount: 6},
		},
	}
}

func TestBookerReservesForSeveralParticipantsWithPrices(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())

	created := reserve(t, h, reservationPayload{
		RambleID:     ramble.ID,
		ContactEmail: "alice@example.com",
		Participants: []participantPayload{
			{FirstName: "Alice", LastName: "Martin", PriceLabel: "Adulte", IsBooker: true},
			{FirstName: "Bruno", LastName: "Martin", PriceLabel: "Adulte"},
			{FirstName: "Chloé", LastName: "Martin", PriceLabel: "Enfant"},
		},
	})

	if created.Status != "pending" {
		t.Errorf("expected status pending, got %q", created.Status)
	}
	if created.EstimatedAmount != 3100 {
		t.Errorf("expected estimated amount 3100 cents (2 × 12.50 € + 6 €), got %d", created.EstimatedAmount)
	}

	token := h.Login("alice@example.com")
	response := h.Do(http.MethodGet, "/reservations", nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /reservations: expected 200, got %d: %s", response.Code, response.Body)
	}
	var mine []reservation
	response.JSON(t, &mine)

	if len(mine) != 1 || mine[0].ID != created.ID {
		t.Fatalf("expected the booker to see their reservation %d, got %s", created.ID, response.Body)
	}
	got := mine[0]
	if got.ContactEmail != "alice@example.com" || got.Status != "pending" || got.EstimatedAmount != 3100 {
		t.Errorf("unexpected reservation: %s", response.Body)
	}
	if got.Ramble == nil || got.Ramble.Title != ramble.Title {
		t.Errorf("expected the ramble summary, got %s", response.Body)
	}
	if len(got.Participants) != 3 {
		t.Fatalf("expected 3 participants, got %s", response.Body)
	}
	expected := []participant{
		{FirstName: "Alice", LastName: "Martin", PriceLabel: "Adulte", PriceAmount: 12.5, IsBooker: true},
		{FirstName: "Bruno", LastName: "Martin", PriceLabel: "Adulte", PriceAmount: 12.5},
		{FirstName: "Chloé", LastName: "Martin", PriceLabel: "Enfant", PriceAmount: 6},
	}
	for i, want := range expected {
		p := got.Participants[i]
		p.ID = 0
		if p != want {
			t.Errorf("participant %d: expected %+v, got %+v", i, want, p)
		}
	}
	if got.Payments == nil || len(got.Payments) != 0 {
		t.Errorf("expected an empty payment list, got %s", response.Body)
	}
}

func TestBookerReservesOnlyForOthersWithASingleAccount(t *testing.T) {
	h := testharness.New(t)
	morning := h.CreateRamble(familyRamble())
	evening := h.CreateRamble(familyRamble())

	reserve(t, h, reservationPayload{
		RambleID:     morning.ID,
		ContactEmail: "parent@example.com",
		Participants: []participantPayload{
			{FirstName: "Léa", LastName: "Petit", PriceLabel: "Enfant"},
			{FirstName: "Hugo", LastName: "Petit", PriceLabel: "Enfant"},
		},
	})
	reserve(t, h, reservationPayload{
		RambleID:     evening.ID,
		ContactEmail: "parent@example.com",
		Participants: []participantPayload{{FirstName: "Léa", LastName: "Petit", PriceLabel: "Enfant"}},
	})

	var mine []reservation
	h.Do(http.MethodGet, "/reservations", nil, h.Login("parent@example.com")).JSON(t, &mine)

	if len(mine) != 2 {
		t.Fatalf("expected both reservations on the booker's single account, got %d", len(mine))
	}
	for _, r := range mine {
		for _, p := range r.Participants {
			if p.IsBooker {
				t.Errorf("expected no participant marked as the booker, got %+v", p)
			}
		}
	}
}

func TestReservationParticipantsMustHaveAPriceOfTheRamble(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())

	for name, label := range map[string]string{"unknown price": "Senior", "missing price": ""} {
		t.Run(name, func(t *testing.T) {
			response := h.Do(http.MethodPost, "/reservations", reservationPayload{
				RambleID:     ramble.ID,
				ContactEmail: "alice@example.com",
				Participants: []participantPayload{{FirstName: "Alice", LastName: "Martin", PriceLabel: label}},
			}, "")

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", response.Code, response.Body)
			}
			var body errorBody
			response.JSON(t, &body)
			if body.Reason != "invalid_price" {
				t.Errorf("expected reason invalid_price, got %s", response.Body)
			}
		})
	}
}

func TestSinglePriceIsAppliedWhenParticipantsHaveNoPrice(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{Prices: []testharness.Price{{Label: "Plein tarif", Amount: 8}}})

	created := reserve(t, h, reservationPayload{
		RambleID:     ramble.ID,
		ContactEmail: "alice@example.com",
		Participants: people(2, ""),
	})

	if created.EstimatedAmount != 1600 || created.Participants[0].PriceLabel != "Plein tarif" {
		t.Errorf("expected the single price on every participant (16 €), got %+v", created)
	}
}

func TestFreeRambleReservationHasNoAmount(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})

	created := reserve(t, h, reservationPayload{
		RambleID:     ramble.ID,
		ContactEmail: "alice@example.com",
		Participants: people(3, ""),
	})

	if created.EstimatedAmount != 0 || len(created.Participants) != 3 || created.Participants[0].PriceLabel != "" {
		t.Errorf("expected 3 participants without price, got %+v", created)
	}
}

func TestSecondReservationWithSameEmailIsRefusedWithDedicatedError(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	other := h.CreateRamble(familyRamble())
	first := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "Adulte")})

	response := h.Do(http.MethodPost, "/reservations", reservationPayload{
		RambleID:     ramble.ID,
		ContactEmail: "  Alice@Example.com ",
		Participants: people(1, "Enfant"),
	}, "")

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "reservation_already_exists" || body.ReservationID != first.ID {
		t.Errorf("expected reason reservation_already_exists for reservation %d, got %s", first.ID, response.Body)
	}

	reserve(t, h, reservationPayload{RambleID: other.ID, ContactEmail: "alice@example.com", Participants: people(1, "Adulte")})
}
