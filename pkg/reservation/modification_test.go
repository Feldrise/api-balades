package reservation_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/testharness"
)

type addParticipantsPayload struct {
	Participants []participantPayload `json:"participants"`
}

func getReservation(t *testing.T, h *testharness.Harness, id uint, token string) reservation {
	t.Helper()

	response := h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", id), nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /reservations/%d: expected 200, got %d: %s", id, response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	return got
}

func TestBookerAddsParticipantsToAPendingReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")

	response := h.Do(http.MethodPost, fmt.Sprintf("/reservations/%d/participants", created.ID), addParticipantsPayload{
		Participants: []participantPayload{{FirstName: "Zoé", LastName: "Martin", PriceLabel: "Enfant"}},
	}, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	got := getReservation(t, h, created.ID, token)
	if got.Status != "pending" || len(got.Participants) != 3 || got.EstimatedAmount != 3100 {
		t.Fatalf("expected a pending reservation of 3 participants for 31,00 €, got %+v", got)
	}
	if last := got.Participants[2]; last.FirstName != "Zoé" || last.PriceLabel != "Enfant" {
		t.Errorf("expected Zoé to be added as Enfant, got %+v", last)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 7 {
		t.Errorf("expected 7 places left, got %s", describe(left))
	}

	sent := h.Emails.SentWithTemplate("reservation_confirmation")
	if len(sent) != 2 || sent[1].To != "alice@example.com" {
		t.Fatalf("expected a second summary email to the booker, got %+v", sent)
	}
	for _, expected := range []string{"Zoé Martin", "31,00 €", "Participants ajoutés"} {
		if !strings.Contains(sent[1].Body, expected) {
			t.Errorf("expected the summary email to contain %q", expected)
		}
	}
}

func addParticipants(h *testharness.Harness, id uint, participants []participantPayload, token string) *testharness.Response {
	return h.Do(http.MethodPost, fmt.Sprintf("/reservations/%d/participants", id), addParticipantsPayload{Participants: participants}, token)
}

func TestAddingParticipantsThatDoNotFitInThePlacesLeftIsRefused(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(5)})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")})
	token := h.Login("alice@example.com")

	response := addParticipants(h, created.ID, people(3, ""), token)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "not_enough_places" {
		t.Errorf("expected reason not_enough_places, got %s", response.Body)
	}
	if got := getReservation(t, h, created.ID, token); got.Status != "pending" || len(got.Participants) != 2 {
		t.Errorf("expected the reservation untouched (pending, 2 participants), got %q with %d", got.Status, len(got.Participants))
	}
	if sent := h.Emails.SentWithTemplate("reservation_confirmation"); len(sent) != 2 {
		t.Errorf("expected no summary email for a refused addition, got %d emails", len(sent))
	}

	if fits := addParticipants(h, created.ID, people(2, ""), token); fits.Code != http.StatusOK {
		t.Fatalf("expected the 2 participants fitting the places left to be added, got %d: %s", fits.Code, fits.Body)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 0 {
		t.Errorf("expected no place left, got %s", describe(left))
	}
}

func TestParticipantsAddedToAWaitingReservationWaitWithIt(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "first@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	token := h.Login("alice@example.com")

	response := addParticipants(h, waiting.ID, people(1, ""), token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if got := getReservation(t, h, waiting.ID, token); got.Status != "waiting_list" || len(got.Participants) != 3 {
		t.Errorf("expected the whole reservation of 3 participants in waiting list, got %q with %d", got.Status, len(got.Participants))
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 1 {
		t.Errorf("expected the waiting list not to take the place left, got %s", describe(left))
	}
}

func TestAddingParticipantsCannotExceedTheMaximumPerReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(19, "")})

	response := addParticipants(h, created.ID, people(2, ""), h.Login("alice@example.com"))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "too_many_participants" {
		t.Errorf("expected reason too_many_participants, got %s", response.Body)
	}
}

func TestParticipantsAddedToAPaidReservationAreConfirmedWithAComplementaryPayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	if response := confirmReservation(h, created.ID, token); response.Code != http.StatusOK {
		t.Fatalf("confirming: expected 200, got %d: %s", response.Code, response.Body)
	}
	h.PayReservation(created.ID, 2500)

	response := addParticipants(h, created.ID, []participantPayload{{FirstName: "Zoé", LastName: "Martin", PriceLabel: "Enfant"}}, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var body reservationWithPayment
	response.JSON(t, &body)
	if body.Status != "confirmed" || len(body.Participants) != 3 {
		t.Errorf("expected the participant confirmed with the reservation, got %q with %d participants", body.Status, len(body.Participants))
	}
	if body.Payment == nil || body.Payment.Payment.Amount != 600 || body.Payment.ClientSecret == "" {
		t.Fatalf("expected a complementary payment of 6,00 € for the added participant, got %s", response.Body)
	}
	if intents := h.Stripe.PaymentIntents(); len(intents) != 2 || intents[1].Amount != 600 {
		t.Errorf("expected a second Stripe payment of 6,00 €, got %+v", intents)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 7 {
		t.Errorf("expected 7 places left, got %s", describe(left))
	}
}

func removeParticipant(h *testharness.Harness, id uint, participantID uint, token string) *testharness.Response {
	return h.Do(http.MethodDelete, fmt.Sprintf("/reservations/%d/participants/%d", id, participantID), nil, token)
}

func cancelReservation(h *testharness.Harness, id uint, token string) *testharness.Response {
	return h.Do(http.MethodPut, fmt.Sprintf("/reservations/%d/cancel", id), nil, token)
}

func TestRemovingAParticipantReleasesItsPlaceAndTellsTheBooker(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{
		RambleID:     ramble.ID,
		ContactEmail: "alice@example.com",
		Participants: []participantPayload{
			{FirstName: "Alice", LastName: "Martin", PriceLabel: "Adulte", IsBooker: true},
			{FirstName: "Bruno", LastName: "Martin", PriceLabel: "Adulte"},
		},
	})
	token := h.Login("alice@example.com")

	response := removeParticipant(h, created.ID, created.Participants[1].ID, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	if got.Status != "pending" || len(got.Participants) != 1 || got.Participants[0].FirstName != "Alice" || got.EstimatedAmount != 1250 {
		t.Errorf("expected Alice alone in the pending reservation for 12,50 €, got %s", response.Body)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 9 {
		t.Errorf("expected 9 places left, got %s", describe(left))
	}
	sent := h.Emails.SentWithTemplate("reservation_confirmation")
	if len(sent) != 2 || sent[1].To != "alice@example.com" || !strings.Contains(sent[1].Body, "Participants retirés") || strings.Contains(sent[1].Body, "Bruno") {
		t.Errorf("expected a summary email to the booker without Bruno, got %+v", sent)
	}
}

func TestRemovingTheLastParticipantCancelsTheReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(4)})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	token := h.Login("alice@example.com")

	response := removeParticipant(h, created.ID, created.Participants[0].ID, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if got := getReservation(t, h, created.ID, token); got.Status != "cancelled" {
		t.Errorf("expected the reservation cancelled, got %q", got.Status)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 4 {
		t.Errorf("expected all 4 places left, got %s", describe(left))
	}
	if again := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")}); again.Status != "pending" {
		t.Errorf("expected the booker to be able to reserve again, got %q", again.Status)
	}
}

func TestCancellingAReservationReleasesAllItsPlaces(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(6)})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(4, "")})
	token := h.Login("alice@example.com")

	response := cancelReservation(h, created.ID, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	if got.Status != "cancelled" {
		t.Errorf("expected the reservation cancelled, got %q", got.Status)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 6 {
		t.Errorf("expected all 6 places left, got %s", describe(left))
	}

	if again := cancelReservation(h, created.ID, token); again.Code != http.StatusConflict {
		t.Errorf("expected cancelling again to be refused with 409, got %d: %s", again.Code, again.Body)
	}
	if add := addParticipants(h, created.ID, people(1, ""), token); add.Code != http.StatusConflict {
		t.Errorf("expected adding to a cancelled reservation to be refused with 409, got %d: %s", add.Code, add.Body)
	}
}

func TestCancellationAdmitsTheWaitingReservationsThatFitEntirely(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(5)})
	leaving := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "leaving@example.com", Participants: people(3, "")})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "staying@example.com", Participants: people(2, "")})
	big := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "big@example.com", Participants: people(4, "")})
	small := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "small@example.com", Participants: people(2, "")})
	single := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "single@example.com", Participants: people(1, "")})
	h.Clock.Advance(time.Minute)

	response := cancelReservation(h, leaving.ID, h.Login("leaving@example.com"))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	for _, c := range []struct {
		name   string
		id     uint
		status string
	}{
		{"the bigger reservation that does not fit", big.ID, "waiting_list"},
		{"the smaller reservation that fits", small.ID, "pending"},
		{"the next reservation that fits the last place", single.ID, "pending"},
	} {
		if got := getReservation(t, h, c.id, admin); got.Status != c.status {
			t.Errorf("expected %s to be %s, got %q", c.name, c.status, got.Status)
		}
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 0 {
		t.Errorf("expected no place left, got %s", describe(left))
	}

	var told []string
	for _, sent := range h.Emails.SentWithTemplate("reservation_confirmation") {
		if strings.Contains(sent.Body, "Une place s'est libérée") {
			told = append(told, sent.To)
		}
	}
	if len(told) != 2 || told[0] != "small@example.com" || told[1] != "single@example.com" {
		t.Errorf("expected the admitted bookers to be told, got %v", told)
	}
}

func TestRemovingAParticipantAdmitsAWaitingReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	full := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(3, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")})

	response := removeParticipant(h, full.ID, full.Participants[2].ID, h.Login("alice@example.com"))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if got := getReservation(t, h, waiting.ID, h.Login("bob@example.com")); got.Status != "pending" {
		t.Errorf("expected the waiting reservation admitted, got %q", got.Status)
	}
}

type participantUpdatePayload struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PriceLabel string `json:"price_label"`
}

func updateParticipant(h *testharness.Harness, id uint, participantID uint, payload participantUpdatePayload, token string) *testharness.Response {
	return h.Do(http.MethodPut, fmt.Sprintf("/reservations/%d/participants/%d", id, participantID), payload, token)
}

func TestBookerChangesTheNameAndPriceOfAParticipant(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")

	response := updateParticipant(h, created.ID, created.Participants[1].ID, participantUpdatePayload{FirstName: " Zoé ", LastName: "Durand", PriceLabel: "Enfant"}, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	got := getReservation(t, h, created.ID, token)
	if p := got.Participants[1]; p.FirstName != "Zoé" || p.LastName != "Durand" || p.PriceLabel != "Enfant" || p.PriceAmount != 6 {
		t.Errorf("expected Zoé Durand as Enfant, got %+v", p)
	}
	if got.EstimatedAmount != 1850 {
		t.Errorf("expected the estimated amount to follow the new price (18,50 €), got %d", got.EstimatedAmount)
	}

	unknown := updateParticipant(h, created.ID, created.Participants[1].ID, participantUpdatePayload{FirstName: "Zoé", LastName: "Durand", PriceLabel: "Senior"}, token)
	if unknown.Code != http.StatusBadRequest {
		t.Errorf("expected an unknown price to be refused with 400, got %d: %s", unknown.Code, unknown.Body)
	}
}

func TestBookerSpecifiesTheMissingPriceOfAMigratedParticipant(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(familyRamble())
	group := l.group(ramble, "parent@example.com", "pending")
	l.registration(ramble, "Paul", "parent@example.com", "pending", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "pending", &group.ID)
	h.MigrateAtStartup()
	migrated := myReservations(t, h, "parent@example.com")[0]
	token := h.Login("parent@example.com")

	for i, price := range []string{"Adulte", "Enfant"} {
		p := migrated.Participants[i]
		response := updateParticipant(h, migrated.ID, p.ID, participantUpdatePayload{FirstName: p.FirstName, LastName: p.LastName, PriceLabel: price}, token)
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
		}
	}

	got := getReservation(t, h, migrated.ID, token)
	if got.HasMissingPrices || got.EstimatedAmount != 1850 {
		t.Errorf("expected all prices known and 18,50 €, got missing=%v amount=%d", got.HasMissingPrices, got.EstimatedAmount)
	}
}

func TestParticipantsCannotBeChangedOnceTheReservationIsPaid(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	h.PayReservation(created.ID, 2500)
	token := h.Login("alice@example.com")

	response := updateParticipant(h, created.ID, created.Participants[0].ID, participantUpdatePayload{FirstName: "Zoé", LastName: "Durand", PriceLabel: "Enfant"}, token)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "reservation_paid" {
		t.Errorf("expected reason reservation_paid, got %s", response.Body)
	}
}

func TestRemovalOrCancellationTellsWhetherThePaymentSucceeded(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	paid := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	h.PayReservation(paid.ID, 2500)
	unpaid := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(2, "Adulte")})
	alice := h.Login("alice@example.com")
	bob := h.Login("bob@example.com")

	for _, c := range []struct {
		name     string
		response *testharness.Response
		paid     bool
	}{
		{"removal on a paid reservation", removeParticipant(h, paid.ID, paid.Participants[1].ID, alice), true},
		{"cancellation of a paid reservation", cancelReservation(h, paid.ID, alice), true},
		{"removal on an unpaid reservation", removeParticipant(h, unpaid.ID, unpaid.Participants[1].ID, bob), false},
		{"cancellation of an unpaid reservation", cancelReservation(h, unpaid.ID, bob), false},
	} {
		if c.response.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d: %s", c.name, c.response.Code, c.response.Body)
		}
		var body struct {
			HasSuccessfulPayment bool `json:"has_successful_payment"`
		}
		c.response.JSON(t, &body)
		if body.HasSuccessfulPayment != c.paid {
			t.Errorf("%s: expected has_successful_payment %v, got %s", c.name, c.paid, c.response.Body)
		}
	}
	if refunds := h.Stripe.Refunds(); len(refunds) != 0 {
		t.Errorf("expected no automatic refund, got %+v", refunds)
	}
}

func TestOnlyTheBookerOrAnAuthorizedStaffCanModifyAReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	otherRamble := h.CreateRamble(familyRamble())
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)
	stranger := h.CreateUser("bob@example.com", "user")
	admin := h.CreateUser("admin@example.com", "admin")

	cases := []struct {
		name  string
		token string
		code  int
	}{
		{"anonymous visitor", "", http.StatusUnauthorized},
		{"another user", h.Token(stranger), http.StatusForbidden},
		{"guide of another ramble", h.Token(otherGuide), http.StatusForbidden},
		{"guide of the ramble", h.Token(owningGuide), http.StatusOK},
		{"admin", h.Token(admin), http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: fmt.Sprintf("booker-%d@example.com", len(c.name)), Participants: people(2, "Adulte")})
			p := created.Participants[0]

			for _, response := range []*testharness.Response{
				addParticipants(h, created.ID, people(1, "Adulte"), c.token),
				updateParticipant(h, created.ID, p.ID, participantUpdatePayload{FirstName: "Zoé", LastName: "Durand", PriceLabel: "Enfant"}, c.token),
				removeParticipant(h, created.ID, p.ID, c.token),
				cancelReservation(h, created.ID, c.token),
			} {
				if response.Code != c.code {
					t.Errorf("expected %d, got %d: %s", c.code, response.Code, response.Body)
				}
			}
		})
	}
}

func TestParticipantOfAnotherReservationIsNotFound(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	mine := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	theirs := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")

	removal := removeParticipant(h, mine.ID, theirs.Participants[0].ID, token)
	update := updateParticipant(h, mine.ID, theirs.Participants[0].ID, participantUpdatePayload{FirstName: "Zoé", LastName: "Durand", PriceLabel: "Enfant"}, token)

	if removal.Code != http.StatusNotFound || update.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for both, got %d and %d", removal.Code, update.Code)
	}
	if got := getReservation(t, h, theirs.ID, h.Login("bob@example.com")); len(got.Participants) != 2 {
		t.Errorf("expected bob's reservation untouched, got %d participants", len(got.Participants))
	}
}

func TestViewingPermissionsDoNotAllowModifyingAReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	viewer := h.CreateUserWithPermissions("viewer@example.com", "view:all-registrations")
	guideViewer := h.CreateUserWithPermissions("guide-viewer@example.com", "view:registrations:self")
	h.CreateGuide(guideViewer, ramble)
	p := created.Participants[0]

	for _, user := range []*dbmodel.User{viewer, guideViewer} {
		token := h.Token(user)
		if response := h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", created.ID), nil, token); response.Code != http.StatusOK {
			t.Fatalf("%s: expected to read the reservation, got %d", user.Email, response.Code)
		}
		for _, response := range []*testharness.Response{
			addParticipants(h, created.ID, people(1, "Adulte"), token),
			updateParticipant(h, created.ID, p.ID, participantUpdatePayload{FirstName: "Zoé", LastName: "Durand", PriceLabel: "Enfant"}, token),
			removeParticipant(h, created.ID, p.ID, token),
			cancelReservation(h, created.ID, token),
		} {
			if response.Code != http.StatusForbidden {
				t.Errorf("%s: expected 403, got %d: %s", user.Email, response.Code, response.Body)
			}
		}
	}
	if got := getReservation(t, h, created.ID, h.Login("alice@example.com")); got.Status != "pending" || len(got.Participants) != 2 {
		t.Errorf("expected the reservation untouched, got %q with %d participants", got.Status, len(got.Participants))
	}
}

func TestRaisingTheCapacityOfARambleAdmitsTheWaitingList(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(2)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	bob := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(2, "")})
	carol := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "carol@example.com", Participants: people(2, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	updateCapacity := func(maxParticipants any) {
		t.Helper()
		response := h.Do(http.MethodPut, fmt.Sprintf("/rambles/%d", ramble.ID), map[string]any{"max_participants": maxParticipants}, admin)
		if response.Code != http.StatusOK {
			t.Fatalf("updating the ramble: expected 200, got %d: %s", response.Code, response.Body)
		}
	}

	updateCapacity(5)

	if got := getReservation(t, h, bob.ID, h.Login("bob@example.com")); got.Status != "pending" {
		t.Errorf("expected bob's reservation admitted in the places added, got %q", got.Status)
	}
	if got := getReservation(t, h, carol.ID, h.Login("carol@example.com")); got.Status != "waiting_list" {
		t.Errorf("expected carol's reservation to keep waiting, got %q", got.Status)
	}
	sent := h.Emails.SentWithTemplate("reservation_confirmation")
	if last := sent[len(sent)-1]; last.To != "bob@example.com" || !strings.Contains(last.Subject, "Place disponible") {
		t.Errorf("expected bob to be told a place is available, got %q to %q", last.Subject, last.To)
	}

	updateCapacity(nil)

	if got := getReservation(t, h, carol.ID, h.Login("carol@example.com")); got.Status != "pending" {
		t.Errorf("expected carol's reservation admitted once the ramble is unlimited, got %q", got.Status)
	}
}

func TestWaitingListOfACancelledRambleIsNeverPromoted(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(2)})
	leaving := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(2, "")})
	h.DB.Model(&dbmodel.Ramble{}).Where("id = ?", ramble.ID).Update("is_cancelled", true)

	if response := cancelReservation(h, leaving.ID, h.Login("alice@example.com")); response.Code != http.StatusOK {
		t.Fatalf("cancelling: expected 200, got %d: %s", response.Code, response.Body)
	}

	if got := getReservation(t, h, waiting.ID, h.Login("bob@example.com")); got.Status != "waiting_list" {
		t.Errorf("expected bob to stay in waiting list on the cancelled ramble, got %q", got.Status)
	}
	for _, sent := range h.Emails.Sent() {
		if sent.To == "bob@example.com" && strings.HasPrefix(sent.Subject, "Place disponible") {
			t.Errorf("expected bob not to be told a place is available, got %q", sent.Subject)
		}
	}
}
