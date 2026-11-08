package reservation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"feldrise.com/balade/pkg/testharness"
)

type presenceChange struct {
	ParticipantID uint       `json:"participant_id"`
	Present       *bool      `json:"present"`
	MarkedAt      *time.Time `json:"marked_at,omitempty"`
}

type presenceOutcome struct {
	ParticipantID    uint       `json:"participant_id"`
	Outcome          string     `json:"outcome"`
	Present          *bool      `json:"present"`
	PresenceMarkedAt *time.Time `json:"presence_marked_at"`
	PresenceMarkedBy *uint      `json:"presence_marked_by"`
}

type presenceResponse struct {
	Participants []presenceOutcome `json:"participants"`
}

type participantPresence struct {
	ID               uint       `json:"id"`
	Present          *bool      `json:"present"`
	PresenceMarkedAt *time.Time `json:"presence_marked_at"`
	PresenceMarkedBy *uint      `json:"presence_marked_by"`
}

type reservationWithPresence struct {
	ID           uint                  `json:"id"`
	ContactPhone *string               `json:"contact_phone"`
	Participants []participantPresence `json:"participants"`
}

type adminListWithPresence struct {
	Reservations []reservationWithPresence `json:"reservations"`
}

type reservationPayloadWithPhone struct {
	RambleID     uint                 `json:"ramble_id"`
	ContactEmail string               `json:"contact_email"`
	ContactPhone *string              `json:"contact_phone,omitempty"`
	Participants []participantPayload `json:"participants"`
}

func boolPointer(value bool) *bool { return &value }

func markPresence(h *testharness.Harness, rambleID uint, token string, changes ...presenceChange) *testharness.Response {
	return h.Do(http.MethodPut, fmt.Sprintf("/reservations/admin/rambles/%d/presence", rambleID), map[string]any{"changes": changes}, token)
}

func presenceOf(t *testing.T, h *testharness.Harness, rambleID uint, token string) map[uint]participantPresence {
	t.Helper()

	response := h.Do(http.MethodGet, fmt.Sprintf("/reservations/admin?ramble_id=%d", rambleID), nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /reservations/admin: expected 200, got %d: %s", response.Code, response.Body)
	}
	var list adminListWithPresence
	response.JSON(t, &list)

	result := map[uint]participantPresence{}
	for _, reservation := range list.Reservations {
		for _, participant := range reservation.Participants {
			result[participant.ID] = participant
		}
	}
	return result
}

// rambleToday creates a ramble taking place today at 18:00 Paris time, the clock being set at 9:00 the same day.
func rambleToday(h *testharness.Harness) (rambleDate time.Time) {
	paris, _ := time.LoadLocation("Europe/Paris")
	rambleDate = time.Date(2030, time.June, 15, 18, 0, 0, 0, paris)
	h.Clock.Set(time.Date(2030, time.June, 15, 9, 0, 0, 0, paris))
	return rambleDate
}

func TestAdminReservationsGiveThePhoneOfTheBooker(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	phone := "0611223344"
	withPhone := h.Do(http.MethodPost, "/reservations", reservationPayloadWithPhone{RambleID: ramble.ID, ContactEmail: "alice@example.com", ContactPhone: &phone, Participants: people(1, "Adulte")}, "")
	if withPhone.Code != http.StatusCreated {
		t.Fatalf("POST /reservations: expected 201, got %d: %s", withPhone.Code, withPhone.Body)
	}
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(1, "Adulte")})

	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	for name, token := range map[string]string{"admin": admin, "guide of the ramble": h.Token(owningGuide)} {
		response := h.Do(http.MethodGet, fmt.Sprintf("/reservations/admin?ramble_id=%d", ramble.ID), nil, token)
		var list adminListWithPresence
		response.JSON(t, &list)

		phones := map[string]*string{}
		for _, reservation := range list.Reservations {
			key := "without phone"
			if reservation.ContactPhone != nil {
				key = *reservation.ContactPhone
			}
			phones[key] = reservation.ContactPhone
		}
		if _, ok := phones[phone]; !ok {
			t.Errorf("%s: expected the phone %s of the booker who gave one, got %+v", name, phone, list.Reservations)
		}
		if _, ok := phones["without phone"]; !ok {
			t.Errorf("%s: expected no phone for the booker who gave none, got %+v", name, list.Reservations)
		}
	}
}

func TestGuideMarksTheParticipantsOfTheirRamblePresentOnTheDay(t *testing.T) {
	h := testharness.New(t)
	rambleDate := rambleToday(h)
	ramble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate, MaxParticipants: testharness.Int(10)})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	guideUser := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(guideUser, ramble)
	guide := h.Token(guideUser)

	first, second := created.Participants[0].ID, created.Participants[1].ID
	response := markPresence(h, ramble.ID, guide,
		presenceChange{ParticipantID: first, Present: boolPointer(true)},
		presenceChange{ParticipantID: second, Present: boolPointer(false)},
	)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var result presenceResponse
	response.JSON(t, &result)
	for _, outcome := range result.Participants {
		if outcome.Outcome != "applied" {
			t.Errorf("participant %d: expected the change to be applied, got %q", outcome.ParticipantID, outcome.Outcome)
		}
	}

	presence := presenceOf(t, h, ramble.ID, guide)
	if p := presence[first]; p.Present == nil || !*p.Present {
		t.Errorf("expected participant %d present, got %+v", first, p)
	}
	if p := presence[second]; p.Present == nil || *p.Present {
		t.Errorf("expected participant %d absent, got %+v", second, p)
	}
	if p := presence[first]; p.PresenceMarkedBy == nil || *p.PresenceMarkedBy != guideUser.ID || p.PresenceMarkedAt == nil {
		t.Errorf("expected the guide %d and the time of the last change, got %+v", guideUser.ID, p)
	}

	h.Clock.Advance(time.Minute)
	markPresence(h, ramble.ID, guide, presenceChange{ParticipantID: first, Present: nil})
	if p := presenceOf(t, h, ramble.ID, guide)[first]; p.Present != nil {
		t.Errorf("expected participant %d no longer marked, got %+v", first, p)
	}
}

func TestPresenceIsRefusedBeforeTheDayOfTheRamble(t *testing.T) {
	h := testharness.New(t)
	paris, _ := time.LoadLocation("Europe/Paris")
	rambleDate := time.Date(2030, time.June, 15, 9, 0, 0, 0, paris)
	h.Clock.Set(time.Date(2030, time.June, 14, 23, 30, 0, 0, paris))
	ramble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := markPresence(h, ramble.ID, admin, presenceChange{ParticipantID: created.Participants[0].ID, Present: boolPointer(true)})
	if response.Code != http.StatusConflict {
		t.Fatalf("the day before the ramble: expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "ramble_not_started" {
		t.Errorf("expected reason ramble_not_started, got %q", body.Reason)
	}

	h.Clock.Set(time.Date(2030, time.June, 15, 0, 5, 0, 0, paris))
	if response := markPresence(h, ramble.ID, admin, presenceChange{ParticipantID: created.Participants[0].ID, Present: boolPointer(true)}); response.Code != http.StatusOK {
		t.Errorf("just after midnight on the day of the ramble: expected 200, got %d: %s", response.Code, response.Body)
	}
}

func TestOnlyTheAdminOrTheGuideOfTheRambleMarksPresence(t *testing.T) {
	h := testharness.New(t)
	rambleDate := rambleToday(h)
	ramble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	otherRamble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	change := presenceChange{ParticipantID: created.Participants[0].ID, Present: boolPointer(true)}

	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)
	booker := h.Login("alice@example.com")

	if response := markPresence(h, ramble.ID, "", change); response.Code != http.StatusUnauthorized {
		t.Errorf("anonymous visitor: expected 401, got %d", response.Code)
	}
	if response := markPresence(h, ramble.ID, booker, change); response.Code != http.StatusForbidden {
		t.Errorf("booker: expected 403, got %d", response.Code)
	}
	if response := markPresence(h, ramble.ID, h.Token(otherGuide), change); response.Code != http.StatusForbidden {
		t.Errorf("guide of another ramble: expected 403, got %d", response.Code)
	}
	if response := markPresence(h, ramble.ID, h.Token(h.CreateUser("admin@example.com", "admin")), change); response.Code != http.StatusOK {
		t.Errorf("admin: expected 200, got %d: %s", response.Code, response.Body)
	}
}

func TestPresenceKeepsTheLatestChangeAndIgnoresResends(t *testing.T) {
	h := testharness.New(t)
	rambleDate := rambleToday(h)
	ramble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	id := created.Participants[0].ID

	earlier := h.Clock.Now().Add(-10 * time.Minute)
	later := h.Clock.Now().Add(-5 * time.Minute)

	outcomes := func(response *testharness.Response) string {
		t.Helper()
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
		}
		var result presenceResponse
		response.JSON(t, &result)
		return result.Participants[0].Outcome
	}

	if outcome := outcomes(markPresence(h, ramble.ID, admin, presenceChange{ParticipantID: id, Present: boolPointer(true), MarkedAt: &later})); outcome != "applied" {
		t.Errorf("first change: expected applied, got %q", outcome)
	}
	if outcome := outcomes(markPresence(h, ramble.ID, admin, presenceChange{ParticipantID: id, Present: boolPointer(true), MarkedAt: &later})); outcome != "unchanged" {
		t.Errorf("same change sent again: expected unchanged, got %q", outcome)
	}
	if outcome := outcomes(markPresence(h, ramble.ID, admin, presenceChange{ParticipantID: id, Present: boolPointer(false), MarkedAt: &earlier})); outcome != "outdated" {
		t.Errorf("older change arriving late: expected outdated, got %q", outcome)
	}
	if p := presenceOf(t, h, ramble.ID, admin)[id]; p.Present == nil || !*p.Present || !p.PresenceMarkedAt.Equal(later) {
		t.Errorf("expected the latest change (present at %s) to be kept, got %+v", later, p)
	}

	future := h.Clock.Now().Add(2 * time.Hour)
	outcomes(markPresence(h, ramble.ID, admin, presenceChange{ParticipantID: id, Present: boolPointer(false), MarkedAt: &future}))
	if p := presenceOf(t, h, ramble.ID, admin)[id]; p.PresenceMarkedAt == nil || p.PresenceMarkedAt.After(h.Clock.Now()) {
		t.Errorf("expected a change dated in the future to be recorded at the current time, got %+v", p)
	}
}

func TestPresenceIsRefusedForCancelledReservationsAndOtherRambles(t *testing.T) {
	h := testharness.New(t)
	rambleDate := rambleToday(h)
	ramble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	otherRamble := h.CreateRamble(testharness.RambleOptions{Date: &rambleDate})
	cancelled := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	elsewhere := reserve(t, h, reservationPayload{RambleID: otherRamble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")})
	expected := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "carol@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	if response := h.Do(http.MethodPut, fmt.Sprintf("/reservations/admin/%d/status", cancelled.ID), map[string]string{"status": "cancelled"}, admin); response.Code != http.StatusOK {
		t.Fatalf("cancelling: expected 200, got %d: %s", response.Code, response.Body)
	}

	response := markPresence(h, ramble.ID, admin,
		presenceChange{ParticipantID: cancelled.Participants[0].ID, Present: boolPointer(true)},
		presenceChange{ParticipantID: elsewhere.Participants[0].ID, Present: boolPointer(true)},
		presenceChange{ParticipantID: expected.Participants[0].ID, Present: boolPointer(true)},
	)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var result presenceResponse
	response.JSON(t, &result)
	got := map[uint]string{}
	for _, outcome := range result.Participants {
		got[outcome.ParticipantID] = outcome.Outcome
	}
	if got[cancelled.Participants[0].ID] != "reservation_cancelled" {
		t.Errorf("participant of a cancelled reservation: expected reservation_cancelled, got %q", got[cancelled.Participants[0].ID])
	}
	if got[elsewhere.Participants[0].ID] != "unknown_participant" {
		t.Errorf("participant of another ramble: expected unknown_participant, got %q", got[elsewhere.Participants[0].ID])
	}
	if got[expected.Participants[0].ID] != "applied" {
		t.Errorf("expected participant: expected applied, got %q", got[expected.Participants[0].ID])
	}
}
