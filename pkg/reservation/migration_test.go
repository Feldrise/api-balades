package reservation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/testharness"
	"gorm.io/gorm"
)

// Legacy registrations and groups are inserted directly: their routes are
// replaced by reservations, and only their migrated form is observed, through the API.

type legacy struct {
	t *testing.T
	h *testharness.Harness
}

func (l legacy) registration(ramble *dbmodel.Ramble, firstName string, email string, status string, groupID *uint) *dbmodel.RambleRegistration {
	l.t.Helper()

	user := &dbmodel.User{Email: email, UserProfile: dbmodel.UserProfile{FirstName: firstName}}
	if err := l.h.DB.Where("email = ?", email).FirstOrCreate(user).Error; err != nil {
		l.t.Fatalf("creating legacy user: %v", err)
	}

	registration := &dbmodel.RambleRegistration{
		FirstName:        firstName,
		LastName:         "Martin",
		Email:            email,
		Status:           status,
		RegistrationDate: l.h.Clock.Now().Add(-48 * time.Hour),
		RambleID:         ramble.ID,
		UserID:           &user.ID,
		GroupID:          groupID,
	}
	if status == "confirmed" {
		confirmedAt := l.h.Clock.Now().Add(-time.Hour)
		registration.ConfirmationDate = &confirmedAt
	}
	if status == "cancelled" {
		cancelledAt := l.h.Clock.Now().Add(-time.Hour)
		reason := "Empêchement"
		registration.CancellationDate = &cancelledAt
		registration.CancellationReason = &reason
	}
	if err := l.h.DB.Create(registration).Error; err != nil {
		l.t.Fatalf("creating legacy registration: %v", err)
	}
	return registration
}

func (l legacy) group(ramble *dbmodel.Ramble, primaryEmail string, status string) *dbmodel.RambleRegistrationGroup {
	l.t.Helper()

	group := &dbmodel.RambleRegistrationGroup{
		PrimaryEmail:     primaryEmail,
		Status:           status,
		RegistrationDate: l.h.Clock.Now().Add(-48 * time.Hour),
		RambleID:         ramble.ID,
	}
	if err := l.h.DB.Create(group).Error; err != nil {
		l.t.Fatalf("creating legacy group: %v", err)
	}
	return group
}

func (l legacy) payment(guide *dbmodel.Guide, amount int64, registrationID *uint, groupID *uint) {
	l.t.Helper()

	paidAt := l.h.Clock.Now().Add(-time.Hour)
	payment := &dbmodel.Payment{
		StripePaymentIntentID: "pi_legacy_" + time.Now().Format("150405.000000000"),
		Amount:                amount,
		Status:                "succeeded",
		RegistrationID:        registrationID,
		GroupID:               groupID,
		PayerEmail:            "payer@example.com",
		PayerName:             "Payer",
		GuideID:               guide.ID,
		PaidAt:                &paidAt,
	}
	if err := l.h.DB.Create(payment).Error; err != nil {
		l.t.Fatalf("creating legacy payment: %v", err)
	}
}

func myReservations(t *testing.T, h *testharness.Harness, email string) []reservation {
	t.Helper()

	response := h.Do(http.MethodGet, "/reservations", nil, h.Login(email))
	if response.Code != http.StatusOK {
		t.Fatalf("GET /reservations: expected 200, got %d: %s", response.Code, response.Body)
	}
	var mine []reservation
	response.JSON(t, &mine)
	return mine
}

func names(r reservation) []string {
	result := make([]string, len(r.Participants))
	for i, p := range r.Participants {
		result[i] = p.FirstName
	}
	return result
}

func TestStartupMigratesIndividualRegistrationsWithTheirPayments(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(testharness.RambleOptions{
		MaxParticipants: testharness.Int(10),
		Prices:          []testharness.Price{{Label: "Plein tarif", Amount: 12}},
	})
	guide := h.CreateGuide(h.CreateUser("guide@example.com", "guide"), ramble)
	paid := l.registration(ramble, "Alice", "alice@example.com", "confirmed", nil)
	l.payment(guide, 1200, &paid.ID, nil)
	l.registration(ramble, "Bruno", "bruno@example.com", "cancelled", nil)

	h.MigrateAtStartup()

	alice := myReservations(t, h, "alice@example.com")
	if len(alice) != 1 {
		t.Fatalf("expected alice's registration as one reservation, got %d", len(alice))
	}
	got := alice[0]
	if got.Status != "confirmed" || got.ConfirmationDate == nil || got.ContactEmail != "alice@example.com" {
		t.Errorf("expected a confirmed reservation for alice@example.com, got %+v", got)
	}
	if len(got.Participants) != 1 || got.Participants[0].FirstName != "Alice" || !got.Participants[0].IsBooker {
		t.Errorf("expected alice as the single participant and booker, got %+v", got.Participants)
	}
	if got.EstimatedAmount != 1200 {
		t.Errorf("expected the single price of the ramble to apply (1200 cents), got %d", got.EstimatedAmount)
	}
	if len(got.Payments) != 1 || got.Payments[0].Amount != 1200 || got.Payments[0].Status != "succeeded" {
		t.Errorf("expected the succeeded 12 € payment attached to the reservation, got %+v", got.Payments)
	}

	bruno := myReservations(t, h, "bruno@example.com")
	if len(bruno) != 1 || bruno[0].Status != "cancelled" {
		t.Errorf("expected bruno's cancelled reservation, got %+v", bruno)
	}

	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 9 {
		t.Errorf("expected 9 places left after migration, got %s", describe(left))
	}
}

func TestStartupMigratesGroupsAsOneReservationOfTheirContact(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(familyRamble())
	guide := h.CreateGuide(h.CreateUser("guide@example.com", "guide"), ramble)
	group := l.group(ramble, "parent@example.com", "pending")
	l.registration(ramble, "Paul", "parent@example.com", "pending", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "pending", &group.ID)
	l.registration(ramble, "Hugo", "hugo@example.com", "pending", &group.ID)
	l.payment(guide, 3100, nil, &group.ID)

	h.MigrateAtStartup()

	mine := myReservations(t, h, "parent@example.com")
	if len(mine) != 1 {
		t.Fatalf("expected the group as one reservation of its contact, got %d", len(mine))
	}
	got := mine[0]
	if got.Status != "pending" || got.ContactEmail != "parent@example.com" {
		t.Errorf("expected a pending reservation for parent@example.com, got %+v", got)
	}
	if len(got.Participants) != 3 || !got.Participants[0].IsBooker || got.Participants[1].IsBooker || got.Participants[2].IsBooker {
		t.Errorf("expected 3 participants with the contact as booker, got %+v", got.Participants)
	}
	if len(got.Payments) != 1 || got.Payments[0].Amount != 3100 {
		t.Errorf("expected the group payment attached, got %+v", got.Payments)
	}
	if others := myReservations(t, h, "lea@example.com"); len(others) != 0 {
		t.Errorf("expected group members other than the contact to have no reservation, got %d", len(others))
	}
}

func TestMigratedParticipantsOfARambleWithSeveralPricesHaveAMissingPrice(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(familyRamble())
	group := l.group(ramble, "parent@example.com", "pending")
	l.registration(ramble, "Paul", "parent@example.com", "pending", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "pending", &group.ID)

	h.MigrateAtStartup()

	mine := myReservations(t, h, "parent@example.com")
	if len(mine) != 1 {
		t.Fatalf("expected the group as one reservation, got %d", len(mine))
	}
	got := mine[0]
	if !got.HasMissingPrices {
		t.Errorf("expected the reservation to have missing prices, got %+v", got)
	}
	for _, p := range got.Participants {
		if !p.PriceMissing || p.PriceLabel != "" || p.PriceAmount != 0 {
			t.Errorf("expected participant %s to have a missing price, got %+v", p.FirstName, p)
		}
	}
	if got.EstimatedAmount != 0 {
		t.Errorf("expected the estimated amount to sum the known prices only (0), got %d", got.EstimatedAmount)
	}
}

func TestStartupMigratesSplitGroupsWithTheirMostAdvancedStatus(t *testing.T) {
	cases := []struct {
		name     string
		statuses []string
		status   string
		kept     []string
	}{
		{"confirmed, waiting list and cancelled", []string{"confirmed", "waiting_list", "cancelled"}, "confirmed", []string{"P1", "P2"}},
		{"pending and waiting list", []string{"waiting_list", "pending"}, "pending", []string{"P1", "P2"}},
		{"all cancelled", []string{"cancelled", "cancelled"}, "cancelled", []string{"P1", "P2"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := testharness.New(t)
			l := legacy{t, h}
			ramble := h.CreateRamble(familyRamble())
			group := l.group(ramble, "contact@example.com", "pending")
			for i, status := range c.statuses {
				email := "contact@example.com"
				if i > 0 {
					email = "member" + string(rune('0'+i)) + "@example.com"
				}
				l.registration(ramble, "P"+string(rune('1'+i)), email, status, &group.ID)
			}

			h.MigrateAtStartup()

			mine := myReservations(t, h, "contact@example.com")
			if len(mine) != 1 {
				t.Fatalf("expected one reservation, got %d", len(mine))
			}
			if mine[0].Status != c.status {
				t.Errorf("expected status %q, got %q", c.status, mine[0].Status)
			}
			if got := names(mine[0]); len(got) != len(c.kept) || got[0] != c.kept[0] || got[1] != c.kept[1] {
				t.Errorf("expected participants %v, got %v", c.kept, got)
			}
		})
	}
}

func TestStartupMigrationIsIdempotent(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(familyRamble())
	l.registration(ramble, "Alice", "alice@example.com", "pending", nil)
	group := l.group(ramble, "parent@example.com", "pending")
	l.registration(ramble, "Paul", "parent@example.com", "pending", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "pending", &group.ID)

	h.MigrateAtStartup()
	h.MigrateAtStartup()

	if alice := myReservations(t, h, "alice@example.com"); len(alice) != 1 || len(alice[0].Participants) != 1 {
		t.Errorf("expected alice's single reservation of 1 participant, got %+v", alice)
	}
	if parent := myReservations(t, h, "parent@example.com"); len(parent) != 1 || len(parent[0].Participants) != 2 {
		t.Errorf("expected the group's single reservation of 2 participants, got %+v", parent)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 7 {
		t.Errorf("expected 7 places left, got %s", describe(left))
	}
}

func TestMissingPricesOfAPaidMigratedReservationAreSpecifiedThenTheBalanceIsPayable(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(payingFamilyRamble())
	group := l.group(ramble, "parent@example.com", "confirmed")
	l.registration(ramble, "Paul", "parent@example.com", "confirmed", &group.ID)
	l.registration(ramble, "Léa", "lea@example.com", "confirmed", &group.ID)
	l.payment(&dbmodel.Guide{Model: gorm.Model{ID: *ramble.PaymentGuideID}}, 1250, nil, &group.ID)
	h.MigrateAtStartup()
	migrated := myReservations(t, h, "parent@example.com")[0]
	token := h.Login("parent@example.com")
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	paul, lea := migrated.Participants[0], migrated.Participants[1]

	if response := updateParticipant(h, migrated.ID, paul.ID, participantUpdatePayload{FirstName: paul.FirstName, LastName: paul.LastName, PriceLabel: "Adulte"}, token); response.Code != http.StatusOK {
		t.Fatalf("booker specifying a missing price: expected 200, got %d: %s", response.Code, response.Body)
	}
	if response := adminUpdateParticipant(h, migrated.ID, lea.ID, adminParticipantPayload{PriceLabel: stringPtr("Enfant")}, admin); response.Code != http.StatusOK {
		t.Fatalf("admin specifying a missing price: expected 200, got %d: %s", response.Code, response.Body)
	}
	response := updateParticipant(h, migrated.ID, paul.ID, participantUpdatePayload{FirstName: paul.FirstName, LastName: paul.LastName, PriceLabel: "Enfant"}, token)
	if reason := confirmationRefusal(t, response); reason != "reservation_paid" {
		t.Errorf("changing a known price once paid: expected reason reservation_paid, got %q", reason)
	}

	var balance reservationWithPayment
	confirmReservation(h, migrated.ID, token).JSON(t, &balance)
	if balance.HasMissingPrices || balance.Payment == nil || balance.Payment.Payment.Amount != 600 {
		t.Errorf("expected the balance of 6,00 € (18,50 € minus 12,50 € paid) payable, got %+v", balance.Payment)
	}
}

func TestStartupMarksTheLegacyPartialRefundsPartiallyRefunded(t *testing.T) {
	h := testharness.New(t)
	l := legacy{t, h}
	ramble := h.CreateRamble(testharness.RambleOptions{
		MaxParticipants: testharness.Int(10),
		Prices:          []testharness.Price{{Label: "Plein tarif", Amount: 25}},
	})
	guide := h.CreateGuide(h.CreateUser("guide@example.com", "guide"), ramble)
	partial := l.registration(ramble, "Alice", "alice@example.com", "confirmed", nil)
	full := l.registration(ramble, "Bruno", "bruno@example.com", "confirmed", nil)
	refunded := func(registrationID uint, refund int64) {
		refundedAt := h.Clock.Now()
		l.payment(guide, 2500, &registrationID, nil)
		// The legacy refunds marked the payment refunded, whatever the amount refunded.
		h.DB.Model(&dbmodel.Payment{}).Where("registration_id = ?", registrationID).
			Updates(map[string]any{"status": "refunded", "refund_amount": refund, "refunded_at": refundedAt})
	}
	refunded(partial.ID, 1000)
	refunded(full.ID, 2500)

	h.MigrateAtStartup()
	h.MigrateAtStartup()

	var alice refundedReservation
	h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", myReservations(t, h, "alice@example.com")[0].ID), nil, h.Login("alice@example.com")).JSON(t, &alice)
	if p := alice.Payments[0]; p.Status != "partially_refunded" || alice.AmountPaid != 1500 || alice.AmountDue != 1000 {
		t.Errorf("expected the payment partially refunded, 15,00 € kept, got %+v, paid %d, due %d", p, alice.AmountPaid, alice.AmountDue)
	}
	var bruno refundedReservation
	h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", myReservations(t, h, "bruno@example.com")[0].ID), nil, h.Login("bruno@example.com")).JSON(t, &bruno)
	if p := bruno.Payments[0]; p.Status != "refunded" || bruno.AmountPaid != 0 {
		t.Errorf("expected the payment entirely refunded, got %+v, paid %d", p, bruno.AmountPaid)
	}
}
