package reservation_test

import (
	"fmt"
	"net/http"
	"strings"
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
	PriceAmount  float64 `json:"price_amount"`
	PriceMissing bool    `json:"price_missing"`
	IsBooker     bool    `json:"is_booker"`
}

type payment struct {
	ID     uint   `json:"id"`
	Amount int64  `json:"amount"`
	Status string `json:"status"`
}

type reservation struct {
	ID               uint          `json:"id"`
	RambleID         uint          `json:"ramble_id"`
	UserID           uint          `json:"user_id"`
	ContactEmail     string        `json:"contact_email"`
	Status           string        `json:"status"`
	ConfirmationDate *time.Time    `json:"confirmation_date"`
	Participants     []participant `json:"participants"`
	EstimatedAmount  int64         `json:"estimated_amount"`
	HasMissingPrices bool          `json:"has_missing_prices"`
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
	if got.ContactEmail != "alice@example.com" || got.Status != "pending" || got.EstimatedAmount != 3100 || got.HasMissingPrices {
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

type rambleView struct {
	PlacesLeft                    *int `json:"places_left"`
	MaxParticipantsPerReservation int  `json:"max_participants_per_reservation"`
}

func getRamble(t *testing.T, h *testharness.Harness, id uint) rambleView {
	t.Helper()

	response := h.Do(http.MethodGet, fmt.Sprintf("/rambles/%d", id), nil, "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /rambles/%d: expected 200, got %d: %s", id, response.Code, response.Body)
	}
	var view rambleView
	response.JSON(t, &view)
	return view
}

func describe(places *int) string {
	if places == nil {
		return "unlimited"
	}
	return fmt.Sprint(*places)
}

func TestReservationThatDoesNotFitGoesEntirelyToWaitingList(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(5)})

	first := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "a@example.com", Participants: people(3, "")})
	tooBig := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "b@example.com", Participants: people(3, "")})

	if first.Status != "pending" {
		t.Errorf("expected the first reservation pending, got %q", first.Status)
	}
	if tooBig.Status != "waiting_list" || len(tooBig.Participants) != 3 {
		t.Errorf("expected the whole second reservation (3 participants) in waiting list, got %q with %d participants", tooBig.Status, len(tooBig.Participants))
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 2 {
		t.Fatalf("expected 2 places left (waiting list not counted), got %s", describe(left))
	}

	fits := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "c@example.com", Participants: people(2, "")})

	if fits.Status != "pending" {
		t.Errorf("expected a reservation fitting the 2 places left to be pending, got %q", fits.Status)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 0 {
		t.Errorf("expected no place left, got %s", describe(left))
	}
}

func TestReservationCannotExceedRambleCapacity(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(4)})

	response := h.Do(http.MethodPost, "/reservations", reservationPayload{
		RambleID: ramble.ID, ContactEmail: "a@example.com", Participants: people(5, ""),
	}, "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "too_many_participants" {
		t.Errorf("expected reason too_many_participants, got %s", response.Body)
	}
	if max := getRamble(t, h, ramble.ID).MaxParticipantsPerReservation; max != 4 {
		t.Errorf("expected the ramble to expose a maximum of 4 participants per reservation, got %d", max)
	}

	if created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "a@example.com", Participants: people(4, "")}); created.Status != "pending" {
		t.Errorf("expected a reservation of the whole capacity to be pending, got %q", created.Status)
	}
}

func TestUnlimitedRambleCapsReservationsAtTwentyParticipants(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})

	tooMany := h.Do(http.MethodPost, "/reservations", reservationPayload{
		RambleID: ramble.ID, ContactEmail: "a@example.com", Participants: people(21, ""),
	}, "")
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "b@example.com", Participants: people(20, "")})

	if tooMany.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for 21 participants, got %d: %s", tooMany.Code, tooMany.Body)
	}
	if created.Status != "pending" || len(created.Participants) != 20 {
		t.Errorf("expected a pending reservation of 20 participants, got %q with %d", created.Status, len(created.Participants))
	}
	view := getRamble(t, h, ramble.ID)
	if view.MaxParticipantsPerReservation != 20 || view.PlacesLeft != nil {
		t.Errorf("expected unlimited places and a maximum of 20 per reservation, got %+v", view)
	}
}

func TestReservationWithinThreeDaysOfRambleIsConfirmedAutomatically(t *testing.T) {
	h := testharness.New(t)
	rambleDate := time.Date(2030, time.June, 15, 9, 0, 0, 0, time.UTC)
	soon := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate, MaxParticipants: testharness.Int(3)})
	h.Clock.Set(rambleDate.Add(-72 * time.Hour))

	confirmed := reserve(t, h, reservationPayload{RambleID: soon.ID, ContactEmail: "a@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: soon.ID, ContactEmail: "b@example.com", Participants: people(2, "")})

	if confirmed.Status != "confirmed" {
		t.Errorf("expected a reservation 3 days before the ramble to be confirmed, got %q", confirmed.Status)
	}
	if confirmed.ConfirmationDate == nil || !confirmed.ConfirmationDate.Equal(h.Clock.Now()) {
		t.Errorf("expected the confirmation date to be now (%s), got %v", h.Clock.Now(), confirmed.ConfirmationDate)
	}
	if waiting.Status != "waiting_list" {
		t.Errorf("expected a reservation that does not fit to stay in waiting list, got %q", waiting.Status)
	}

	h.Clock.Set(rambleDate.Add(-73 * time.Hour))
	later := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	pending := reserve(t, h, reservationPayload{RambleID: later.ID, ContactEmail: "a@example.com", Participants: people(1, "")})
	if pending.Status != "pending" {
		t.Errorf("expected a reservation more than 3 days before the ramble to be pending, got %q", pending.Status)
	}
}

func TestOnlyTheBookerOrAnAuthorizedStaffCanReadAReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	otherRamble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	path := fmt.Sprintf("/reservations/%d", created.ID)

	admin := h.CreateUser("admin@example.com", "admin")
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)
	stranger := h.CreateUser("bob@example.com", "user")

	cases := []struct {
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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := h.Do(http.MethodGet, path, nil, c.token)

			if response.Code != c.code {
				t.Fatalf("expected %d, got %d: %s", c.code, response.Code, response.Body)
			}
			if c.code == http.StatusOK {
				var got reservation
				response.JSON(t, &got)
				if got.ID != created.ID || len(got.Participants) != 2 || got.EstimatedAmount != 2500 {
					t.Errorf("unexpected reservation: %s", response.Body)
				}
			}
		})
	}
}

func TestReadingAnUnknownReservationIsNotFound(t *testing.T) {
	h := testharness.New(t)
	admin := h.CreateUser("admin@example.com", "admin")

	response := h.Do(http.MethodGet, "/reservations/999", nil, h.Token(admin))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", response.Code, response.Body)
	}
}

func TestAnonymousVisitorCannotListReservations(t *testing.T) {
	h := testharness.New(t)

	response := h.Do(http.MethodGet, "/reservations", nil, "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", response.Code, response.Body)
	}
}

func TestBookerAloneReceivesAReservationSummaryEmail(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())

	reserve(t, h, reservationPayload{
		RambleID:     ramble.ID,
		ContactEmail: "alice@example.com",
		Participants: []participantPayload{
			{FirstName: "Bruno", LastName: "Martin", PriceLabel: "Adulte"},
			{FirstName: "Chloé", LastName: "Martin", PriceLabel: "Enfant"},
		},
	})

	sent := h.Emails.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected a single email, got %d", len(sent))
	}
	email := sent[0]
	if email.To != "alice@example.com" {
		t.Errorf("expected the email to go to the booker, got %q", email.To)
	}
	for _, expected := range []string{
		ramble.Title,
		"Bruno Martin", "Chloé Martin",
		"Adulte", "12,50 €", "Enfant", "6,00 €",
		"18,50 €",
		"confirmation et paiement 3 jours avant la balade",
	} {
		if !strings.Contains(email.Body, expected) {
			t.Errorf("expected the email to contain %q", expected)
		}
	}
}

func TestWaitingListReservationEmailTellsTheReservationWaits(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(2)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "first@example.com", Participants: people(1, "")})

	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})

	sent := h.Emails.Sent()
	if len(sent) != 2 || sent[1].To != "alice@example.com" || !strings.Contains(sent[1].Body, "liste d'attente") {
		t.Fatalf("expected an email to alice mentioning the waiting list, got %+v", sent)
	}
}

func TestCancelledOrUnpublishedRambleCannotBeReserved(t *testing.T) {
	h := testharness.New(t)
	unpublished := h.CreateRamble(testharness.RambleOptions{Unpublished: true})
	cancelled := h.CreateRamble(testharness.RambleOptions{})
	if err := h.DB.Model(cancelled).Update("is_cancelled", true).Error; err != nil {
		t.Fatalf("cancelling ramble: %v", err)
	}

	for name, c := range map[string]struct {
		rambleID uint
		code     int
	}{
		"unpublished": {unpublished.ID, http.StatusNotFound},
		"cancelled":   {cancelled.ID, http.StatusBadRequest},
		"unknown":     {999, http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			response := h.Do(http.MethodPost, "/reservations", reservationPayload{
				RambleID: c.rambleID, ContactEmail: "alice@example.com", Participants: people(1, ""),
			}, "")

			if response.Code != c.code {
				t.Fatalf("expected %d, got %d: %s", c.code, response.Code, response.Body)
			}
		})
	}
	if sent := h.Emails.Sent(); len(sent) != 0 {
		t.Errorf("expected no email, got %d", len(sent))
	}
}
