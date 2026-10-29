package reservation_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"feldrise.com/balade/pkg/testharness"
)

type adminList struct {
	Reservations []reservation `json:"reservations"`
	Total        int64         `json:"total"`
	Page         int           `json:"page"`
	PerPage      int           `json:"per_page"`
	TotalPages   int           `json:"total_pages"`
}

func adminReservations(t *testing.T, h *testharness.Harness, query string, token string) adminList {
	t.Helper()

	response := h.Do(http.MethodGet, "/reservations/admin"+query, nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /reservations/admin%s: expected 200, got %d: %s", query, response.Code, response.Body)
	}
	var list adminList
	response.JSON(t, &list)
	return list
}

func TestAdminListsTheReservationsOfARambleWithTheirParticipants(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	other := h.CreateRamble(familyRamble())
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(3, "Enfant")})
	reserve(t, h, reservationPayload{RambleID: other.ID, ContactEmail: "carol@example.com", Participants: people(1, "Adulte")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	list := adminReservations(t, h, "?ramble_id="+fmt.Sprint(ramble.ID), admin)

	if list.Total != 2 || len(list.Reservations) != 2 {
		t.Fatalf("expected the 2 reservations of the ramble, got total %d with %d reservations", list.Total, len(list.Reservations))
	}
	for _, r := range list.Reservations {
		if r.RambleID != ramble.ID {
			t.Errorf("expected only reservations of ramble %d, got one of ramble %d", ramble.ID, r.RambleID)
		}
	}
	if first := list.Reservations[0]; len(first.Participants) == 0 || first.ContactEmail == "" || first.Ramble == nil {
		t.Errorf("expected reservations with their participants, contact and ramble, got %+v", first)
	}
	if list.Page != 1 || list.PerPage != 50 || list.TotalPages != 1 {
		t.Errorf("expected page 1 of 1 with 50 per page, got page %d of %d with %d per page", list.Page, list.TotalPages, list.PerPage)
	}
}

func TestAdminListIsRestrictedToAuthorizedStaff(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	otherRamble := h.CreateRamble(familyRamble())
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	reserve(t, h, reservationPayload{RambleID: otherRamble.ID, ContactEmail: "bob@example.com", Participants: people(1, "Adulte")})

	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	stranger := h.CreateUser("stranger@example.com", "user")

	if response := h.Do(http.MethodGet, "/reservations/admin", nil, ""); response.Code != http.StatusUnauthorized {
		t.Errorf("anonymous visitor: expected 401, got %d", response.Code)
	}
	if response := h.Do(http.MethodGet, "/reservations/admin", nil, h.Token(stranger)); response.Code != http.StatusForbidden {
		t.Errorf("another user: expected 403, got %d", response.Code)
	}

	guide := h.Token(owningGuide)
	if list := adminReservations(t, h, "", guide); list.Total != 1 || list.Reservations[0].RambleID != ramble.ID {
		t.Errorf("guide without filter: expected only the reservation of their ramble, got %+v", list)
	}
	if list := adminReservations(t, h, "?ramble_id="+fmt.Sprint(ramble.ID), guide); list.Total != 1 {
		t.Errorf("guide of the ramble: expected 1 reservation, got %d", list.Total)
	}
	if response := h.Do(http.MethodGet, "/reservations/admin?ramble_id="+fmt.Sprint(otherRamble.ID), nil, guide); response.Code != http.StatusForbidden {
		t.Errorf("guide of another ramble: expected 403, got %d", response.Code)
	}
}

func TestAdminFiltersTheReservationsByStatus(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(2)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "first@example.com", Participants: people(2, "")})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(1, "")})
	cancelled := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "cancelled@example.com", Participants: people(1, "")})
	cancelReservation(h, cancelled.ID, h.Login("cancelled@example.com"))
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	for _, c := range []struct {
		query string
		want  []string
	}{
		{"?status=pending", []string{"first@example.com"}},
		{"?status=waiting_list", []string{"waiting@example.com"}},
		{"?statuses=waiting_list&statuses=cancelled", []string{"cancelled@example.com", "waiting@example.com"}},
		{"", []string{"cancelled@example.com", "waiting@example.com", "first@example.com"}},
	} {
		list := adminReservations(t, h, c.query, admin)

		var got []string
		for _, r := range list.Reservations {
			got = append(got, r.ContactEmail)
		}
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q: expected %v, got %v", c.query, c.want, got)
		}
	}
}

func TestAdminSearchesTheReservationsByParticipantContactAndRambleTitle(t *testing.T) {
	h := testharness.New(t)
	birds := h.CreateRamble(testharness.RambleOptions{Title: "Balade des oiseaux"})
	mushrooms := h.CreateRamble(testharness.RambleOptions{Title: "Cueillette de champignons"})
	reserve(t, h, reservationPayload{RambleID: birds.ID, ContactEmail: "alice@example.com", Participants: []participantPayload{
		{FirstName: "Alice", LastName: "Martin"}, {FirstName: "Léo", LastName: "Martin"},
	}})
	reserve(t, h, reservationPayload{RambleID: mushrooms.ID, ContactEmail: "bob@sample.org", Participants: []participantPayload{
		{FirstName: "Bob", LastName: "Durand"},
	}})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	for _, c := range []struct {
		name  string
		query string
		want  []string
	}{
		{"a participant's first name, whatever its case", "?search=l%C3%A9o", []string{"alice@example.com"}},
		{"a participant's last name", "?search=durand", []string{"bob@sample.org"}},
		{"the contact email", "?search=sample.org", []string{"bob@sample.org"}},
		{"the exact contact email", "?email=alice@example.com", []string{"alice@example.com"}},
		{"the ramble title", "?ramble_title=champignons", []string{"bob@sample.org"}},
		{"nothing", "?search=zzz", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			list := adminReservations(t, h, c.query, admin)

			var got []string
			for _, r := range list.Reservations {
				got = append(got, r.ContactEmail)
			}
			if fmt.Sprint(got) != fmt.Sprint(c.want) || list.Total != int64(len(c.want)) {
				t.Errorf("expected %v, got %v (total %d)", c.want, got, list.Total)
			}
		})
	}
}

func TestAdminPaginatesAndSortsTheReservations(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	for _, address := range []string{"first@example.com", "second@example.com", "third@example.com"} {
		reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: address, Participants: people(1, "")})
	}
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	list := adminReservations(t, h, "?per_page=2&page=2&sort_order=asc", admin)

	if list.Total != 3 || list.Page != 2 || list.PerPage != 2 || list.TotalPages != 2 {
		t.Errorf("expected page 2 of 2 of 3 reservations, got %+v", list)
	}
	if len(list.Reservations) != 1 || list.Reservations[0].ContactEmail != "third@example.com" {
		t.Errorf("expected the third reservation alone on page 2, got %+v", list.Reservations)
	}

	if first := adminReservations(t, h, "?sort_by=contact_email&sort_order=desc&per_page=1", admin); first.Reservations[0].ContactEmail != "third@example.com" {
		t.Errorf("expected the reservation sorted last by contact email first, got %s", first.Reservations[0].ContactEmail)
	}
	if response := h.Do(http.MethodGet, "/reservations/admin?page=0", nil, admin); response.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for page 0, got %d", response.Code)
	}
}

func TestAdminFiltersTheReservationsByCreationDateAndBooker(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	alice := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	today := time.Now().UTC()
	day := func(offset int) string { return today.AddDate(0, 0, offset).Format("2006-01-02") }

	for _, c := range []struct {
		name  string
		query string
		want  int64
	}{
		{"created since today", "?date_from=" + day(0), 2},
		{"created since tomorrow", "?date_from=" + day(1), 0},
		{"created until today included", "?date_to=" + day(0), 2},
		{"created until yesterday", "?date_to=" + day(-1), 0},
	} {
		if got := adminReservations(t, h, c.query, admin).Total; got != c.want {
			t.Errorf("%s: expected %d reservations, got %d", c.name, c.want, got)
		}
	}

	if list := adminReservations(t, h, fmt.Sprintf("?user_id=%d", alice.UserID), admin); list.Total != 1 || list.Reservations[0].ID != alice.ID {
		t.Errorf("expected only the reservations of the booker, got %+v", list)
	}
	if response := h.Do(http.MethodGet, "/reservations/admin?date_from=yesterday", nil, admin); response.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for an invalid date, got %d", response.Code)
	}
}

type statusPayload struct {
	Status    string  `json:"status"`
	Reason    *string `json:"reason,omitempty"`
	SendEmail bool    `json:"send_email,omitempty"`
}

func adminSetStatus(h *testharness.Harness, id uint, payload statusPayload, token string) *testharness.Response {
	return h.Do(http.MethodPut, fmt.Sprintf("/reservations/admin/%d/status", id), payload, token)
}

func TestAdminConfirmsAPendingReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := adminSetStatus(h, created.ID, statusPayload{Status: "confirmed"}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	if got.Status != "confirmed" || got.ConfirmationDate == nil || len(got.Participants) != 2 {
		t.Errorf("expected the reservation confirmed with its participants, got %+v", got)
	}
	if saved := getReservation(t, h, created.ID, admin); saved.Status != "confirmed" {
		t.Errorf("expected the status to be saved, got %q", saved.Status)
	}
}

func TestAdminCanOnlyAdmitAWaitingReservationThatFitsInThePlacesLeft(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "first@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(2, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := adminSetStatus(h, waiting.ID, statusPayload{Status: "confirmed"}, admin)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "not_enough_places" {
		t.Errorf("expected reason not_enough_places, got %s", response.Body)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "waiting_list" {
		t.Errorf("expected the reservation to stay in waiting list, got %q", got.Status)
	}

	if err := h.DB.Table("rambles").Where("id = ?", ramble.ID).Update("max_participants", 4).Error; err != nil {
		t.Fatalf("raising the capacity: %v", err)
	}
	response = adminSetStatus(h, waiting.ID, statusPayload{Status: "pending"}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 once the 2 participants fit, got %d: %s", response.Code, response.Body)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "pending" || len(got.Participants) != 2 {
		t.Errorf("expected the whole reservation pending, got %q with %d participants", got.Status, len(got.Participants))
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 0 {
		t.Errorf("expected no place left, got %s", describe(left))
	}
}

func TestAdminCancellationReleasesThePlacesAndAdmitsTheWaitingList(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	leaving := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "leaving@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(2, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	reason := "Doublon"

	response := adminSetStatus(h, leaving.ID, statusPayload{Status: "cancelled", Reason: &reason}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	if got.Status != "cancelled" || got.CancellationDate == nil || got.CancellationReason == nil || *got.CancellationReason != reason {
		t.Errorf("expected a cancelled reservation with its date and reason, got %+v", got)
	}
	if admitted := getReservation(t, h, waiting.ID, admin); admitted.Status != "pending" {
		t.Errorf("expected the waiting reservation to be admitted, got %q", admitted.Status)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 1 {
		t.Errorf("expected 1 place left, got %s", describe(left))
	}

	var told []string
	for _, sent := range h.Emails.SentWithTemplate("reservation_confirmation") {
		if strings.Contains(sent.Body, "Une place s'est libérée") {
			told = append(told, sent.To)
		}
	}
	if len(told) != 1 || told[0] != "waiting@example.com" {
		t.Errorf("expected the admitted booker to be told, got %v", told)
	}
}

func TestAdminMovesAReservationToTheWaitingListAndGivesItsPlacesToTheOthers(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(2)})
	moved := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "moved@example.com", Participants: people(1, "")})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "staying@example.com", Participants: people(1, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := adminSetStatus(h, moved.ID, statusPayload{Status: "waiting_list"}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if got := getReservation(t, h, moved.ID, admin); got.Status != "waiting_list" {
		t.Errorf("expected the moved reservation to wait, got %q", got.Status)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "pending" {
		t.Errorf("expected the place to go to the reservation that was waiting, got %q", got.Status)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 0 {
		t.Errorf("expected no place left, got %s", describe(left))
	}
}

func TestAdminReactivatesACancelledReservationThatFits(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "")})
	cancelled := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")})
	cancelReservation(h, cancelled.ID, h.Login("bob@example.com"))
	big := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "carol@example.com", Participants: people(2, "")})
	cancelReservation(h, big.ID, h.Login("carol@example.com"))
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	refused := adminSetStatus(h, big.ID, statusPayload{Status: "pending"}, admin)
	if refused.Code != http.StatusConflict {
		t.Fatalf("expected 409 for 2 participants and 1 place left, got %d: %s", refused.Code, refused.Body)
	}
	var body errorBody
	refused.JSON(t, &body)
	if body.Reason != "not_enough_places" {
		t.Errorf("expected reason not_enough_places, got %s", refused.Body)
	}

	response := adminSetStatus(h, cancelled.ID, statusPayload{Status: "pending"}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	if got.Status != "pending" || got.CancellationDate != nil || got.CancellationReason != nil || len(got.Participants) != 1 {
		t.Errorf("expected the reservation pending again with its participant and no cancellation, got %+v", got)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 0 {
		t.Errorf("expected no place left, got %s", describe(left))
	}
}

func TestAdminCannotReactivateAReservationWhoseBookerReservedAgain(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	first := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	cancelReservation(h, first.ID, h.Login("alice@example.com"))
	again := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := adminSetStatus(h, first.ID, statusPayload{Status: "confirmed"}, admin)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "reservation_already_exists" || body.ReservationID != again.ID {
		t.Errorf("expected reason reservation_already_exists for reservation %d, got %s", again.ID, response.Body)
	}
	if got := getReservation(t, h, first.ID, admin); got.Status != "cancelled" {
		t.Errorf("expected the reservation to stay cancelled, got %q", got.Status)
	}
	if again := adminSetStatus(h, first.ID, statusPayload{Status: "cancelled"}, admin); again.Code != http.StatusOK {
		t.Errorf("expected cancelling a cancelled reservation to change nothing, got %d: %s", again.Code, again.Body)
	}
}

func TestAdminStatusChangeTellsTheBookerOnlyWhenAsked(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	adminSetStatus(h, created.ID, statusPayload{Status: "cancelled"}, admin)
	adminSetStatus(h, created.ID, statusPayload{Status: "pending"}, admin)
	if sent := h.Emails.SentWithTemplate("reservation_confirmation"); len(sent) != 1 {
		t.Fatalf("expected only the email of the reservation without email for the status changes, got %d", len(sent))
	}

	adminSetStatus(h, created.ID, statusPayload{Status: "confirmed", SendEmail: true}, admin)

	sent := h.Emails.SentWithTemplate("reservation_confirmation")
	if len(sent) != 2 || sent[1].To != "alice@example.com" {
		t.Fatalf("expected a summary email to the booker, got %+v", sent)
	}
	for _, expected := range []string{"Réservation mise à jour", "confirmée", "25,00 €"} {
		if !strings.Contains(sent[1].Body, expected) {
			t.Errorf("expected the email to contain %q", expected)
		}
	}
}

func TestAdminStatusChangeIsRestrictedToAuthorizedStaff(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	otherRamble := h.CreateRamble(testharness.RambleOptions{})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)

	for _, c := range []struct {
		name  string
		token string
		code  int
	}{
		{"anonymous visitor", "", http.StatusUnauthorized},
		{"the booker", h.Login("alice@example.com"), http.StatusForbidden},
		{"guide of another ramble", h.Token(otherGuide), http.StatusForbidden},
		{"guide of the ramble", h.Token(owningGuide), http.StatusOK},
	} {
		t.Run(c.name, func(t *testing.T) {
			if response := adminSetStatus(h, created.ID, statusPayload{Status: "confirmed"}, c.token); response.Code != c.code {
				t.Fatalf("expected %d, got %d: %s", c.code, response.Code, response.Body)
			}
		})
	}
}

func TestAdminStatusChangeRejectsUnknownStatusAndReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	if response := adminSetStatus(h, created.ID, statusPayload{Status: "archived"}, admin); response.Code != http.StatusBadRequest {
		t.Errorf("unknown status: expected 400, got %d", response.Code)
	}
	if response := adminSetStatus(h, created.ID, statusPayload{}, admin); response.Code != http.StatusBadRequest {
		t.Errorf("missing status: expected 400, got %d", response.Code)
	}
	if response := adminSetStatus(h, 999, statusPayload{Status: "confirmed"}, admin); response.Code != http.StatusNotFound {
		t.Errorf("unknown reservation: expected 404, got %d", response.Code)
	}
}

func adminDelete(h *testharness.Harness, id uint, token string) *testharness.Response {
	return h.Do(http.MethodDelete, fmt.Sprintf("/reservations/admin/%d", id), nil, token)
}

func TestAdminDeletesAReservationAndAdmitsTheWaitingList(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	deleted := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "deleted@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(2, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	response := adminDelete(h, deleted.ID, admin)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", response.Code, response.Body)
	}
	if gone := h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", deleted.ID), nil, admin); gone.Code != http.StatusNotFound {
		t.Errorf("expected the reservation to be gone, got %d", gone.Code)
	}
	if admitted := getReservation(t, h, waiting.ID, admin); admitted.Status != "pending" {
		t.Errorf("expected the waiting reservation to be admitted, got %q", admitted.Status)
	}
	if list := adminReservations(t, h, "", admin); list.Total != 1 {
		t.Errorf("expected only the remaining reservation to be listed, got %d", list.Total)
	}
	if left := getRamble(t, h, ramble.ID).PlacesLeft; left == nil || *left != 1 {
		t.Errorf("expected 1 place left, got %s", describe(left))
	}
}

func TestAdminDeletesACancelledReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	cancelReservation(h, created.ID, h.Login("alice@example.com"))
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	if response := adminDelete(h, created.ID, admin); response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", response.Code, response.Body)
	}
	if list := adminReservations(t, h, "", admin); list.Total != 0 {
		t.Errorf("expected no reservation left, got %d", list.Total)
	}
}

func TestAdminDeletionIsRestrictedToAuthorizedStaff(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	otherRamble := h.CreateRamble(testharness.RambleOptions{})
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)

	for _, c := range []struct {
		name  string
		id    uint
		token string
		code  int
	}{
		{"anonymous visitor", created.ID, "", http.StatusUnauthorized},
		{"the booker", created.ID, h.Login("alice@example.com"), http.StatusForbidden},
		{"guide of another ramble", created.ID, h.Token(otherGuide), http.StatusForbidden},
		{"unknown reservation", 999, h.Token(owningGuide), http.StatusNotFound},
		{"guide of the ramble", created.ID, h.Token(owningGuide), http.StatusNoContent},
	} {
		t.Run(c.name, func(t *testing.T) {
			if response := adminDelete(h, c.id, c.token); response.Code != c.code {
				t.Fatalf("expected %d, got %d: %s", c.code, response.Code, response.Body)
			}
		})
	}
}

type bulkPayload struct {
	ReservationIDs []uint  `json:"reservation_ids"`
	Action         string  `json:"action"`
	Reason         *string `json:"reason,omitempty"`
	SendEmail      bool    `json:"send_email,omitempty"`
}

type bulkResult struct {
	SuccessCount int `json:"success_count"`
	FailureCount int `json:"failure_count"`
	Errors       []struct {
		ReservationID uint   `json:"reservation_id"`
		Error         string `json:"error"`
	} `json:"errors"`
	Updated []reservation `json:"updated"`
}

func adminBulk(t *testing.T, h *testharness.Harness, payload bulkPayload, token string) bulkResult {
	t.Helper()

	response := h.Do(http.MethodPost, "/reservations/admin/bulk-action", payload, token)
	if response.Code != http.StatusOK {
		t.Fatalf("POST /reservations/admin/bulk-action: expected 200, got %d: %s", response.Code, response.Body)
	}
	var result bulkResult
	response.JSON(t, &result)
	return result
}

func TestAdminBulkActionReportsEachReservationSeparately(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(3)})
	pending := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "pending@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(2, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	result := adminBulk(t, h, bulkPayload{ReservationIDs: []uint{pending.ID, waiting.ID, 999}, Action: "confirm"}, admin)

	if result.SuccessCount != 1 || result.FailureCount != 2 || len(result.Errors) != 2 {
		t.Fatalf("expected 1 success and 2 failures, got %+v", result)
	}
	if len(result.Updated) != 1 || result.Updated[0].ID != pending.ID || result.Updated[0].Status != "confirmed" {
		t.Errorf("expected the first reservation confirmed, got %+v", result.Updated)
	}
	if result.Errors[0].ReservationID != waiting.ID || result.Errors[1].ReservationID != 999 {
		t.Errorf("expected the failures of the waiting and the unknown reservations, got %+v", result.Errors)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "waiting_list" {
		t.Errorf("expected the reservation that does not fit to stay in waiting list, got %q", got.Status)
	}
}

func TestAdminBulkCancellationAndDeletionAdmitTheWaitingList(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(4)})
	first := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "first@example.com", Participants: people(2, "")})
	second := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "second@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(3, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	reason := "Balade reportée"

	cancelled := adminBulk(t, h, bulkPayload{ReservationIDs: []uint{first.ID}, Action: "cancel", Reason: &reason}, admin)

	if cancelled.SuccessCount != 1 || cancelled.Updated[0].Status != "cancelled" || *cancelled.Updated[0].CancellationReason != reason {
		t.Fatalf("expected the reservation cancelled with its reason, got %+v", cancelled)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "waiting_list" {
		t.Errorf("expected 3 participants not to fit in the 2 places released, got %q", got.Status)
	}

	deleted := adminBulk(t, h, bulkPayload{ReservationIDs: []uint{second.ID}, Action: "delete"}, admin)

	if deleted.SuccessCount != 1 || len(deleted.Updated) != 0 {
		t.Fatalf("expected the reservation deleted, got %+v", deleted)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "pending" {
		t.Errorf("expected the waiting reservation to be admitted once 4 places are free, got %q", got.Status)
	}
}

func TestAdminBulkMovesToTheWaitingListUnderTheSameRules(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(2)})
	moved := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "moved@example.com", Participants: people(2, "")})
	waiting := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "waiting@example.com", Participants: people(1, "")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	result := adminBulk(t, h, bulkPayload{ReservationIDs: []uint{moved.ID}, Action: "move_to_waiting"}, admin)

	if result.SuccessCount != 1 || result.Updated[0].Status != "waiting_list" {
		t.Fatalf("expected the reservation in waiting list, got %+v", result)
	}
	if got := getReservation(t, h, waiting.ID, admin); got.Status != "pending" {
		t.Errorf("expected the waiting reservation to take a place released, got %q", got.Status)
	}
}

func TestAdminBulkActionIsRestrictedToAuthorizedStaffAndTheirRambles(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{})
	otherRamble := h.CreateRamble(testharness.RambleOptions{})
	mine := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})
	others := reserve(t, h, reservationPayload{RambleID: otherRamble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")})
	guideUser := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(guideUser, ramble)

	body := bulkPayload{ReservationIDs: []uint{mine.ID}, Action: "confirm"}
	if response := h.Do(http.MethodPost, "/reservations/admin/bulk-action", body, ""); response.Code != http.StatusUnauthorized {
		t.Errorf("anonymous visitor: expected 401, got %d", response.Code)
	}
	if response := h.Do(http.MethodPost, "/reservations/admin/bulk-action", body, h.Login("alice@example.com")); response.Code != http.StatusForbidden {
		t.Errorf("the booker: expected 403, got %d", response.Code)
	}
	if response := h.Do(http.MethodPost, "/reservations/admin/bulk-action", bulkPayload{ReservationIDs: []uint{mine.ID}, Action: "explode"}, h.Token(guideUser)); response.Code != http.StatusBadRequest {
		t.Errorf("unknown action: expected 400, got %d", response.Code)
	}

	result := adminBulk(t, h, bulkPayload{ReservationIDs: []uint{mine.ID, others.ID}, Action: "delete"}, h.Token(guideUser))

	if result.SuccessCount != 1 || result.FailureCount != 1 || result.Errors[0].ReservationID != others.ID {
		t.Errorf("expected the guide to act on their ramble only, got %+v", result)
	}
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	if got := getReservation(t, h, others.ID, admin); got.Status != "pending" {
		t.Errorf("expected the reservation of another ramble to be untouched, got %q", got.Status)
	}
}

type adminParticipantPayload struct {
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	PriceLabel *string `json:"price_label,omitempty"`
}

func adminUpdateParticipant(h *testharness.Harness, id uint, participantID uint, payload adminParticipantPayload, token string) *testharness.Response {
	return h.Do(http.MethodPut, fmt.Sprintf("/reservations/admin/%d/participants/%d", id, participantID), payload, token)
}

func stringPtr(value string) *string {
	return &value
}

func TestAdminChangesTheNameAndPriceOfAParticipantKeepingTheOtherFields(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	target := created.Participants[1]

	response := adminUpdateParticipant(h, created.ID, target.ID, adminParticipantPayload{FirstName: stringPtr("Zoé")}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	var got reservation
	response.JSON(t, &got)
	if p := got.Participants[1]; p.FirstName != "Zoé" || p.LastName != target.LastName || p.PriceLabel != "Adulte" {
		t.Errorf("expected only the first name to change, got %+v", p)
	}

	response = adminUpdateParticipant(h, created.ID, target.ID, adminParticipantPayload{LastName: stringPtr("Durand"), PriceLabel: stringPtr("Enfant")}, admin)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	saved := getReservation(t, h, created.ID, admin)
	if p := saved.Participants[1]; p.FirstName != "Zoé" || p.LastName != "Durand" || p.PriceLabel != "Enfant" || saved.EstimatedAmount != 1850 {
		t.Errorf("expected Zoé Durand as Enfant for a reservation of 18,50 €, got %+v (%d)", p, saved.EstimatedAmount)
	}
}

func TestAdminCanCorrectTheNameOfAPaidParticipantButNotItsPrice(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	h.PayReservation(created.ID, 2500)
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	target := created.Participants[0].ID

	if response := adminUpdateParticipant(h, created.ID, target, adminParticipantPayload{FirstName: stringPtr("Zoé"), PriceLabel: stringPtr("Adulte")}, admin); response.Code != http.StatusOK {
		t.Fatalf("expected the name to be corrected, got %d: %s", response.Code, response.Body)
	}

	response := adminUpdateParticipant(h, created.ID, target, adminParticipantPayload{PriceLabel: stringPtr("Enfant")}, admin)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body)
	}
	var body errorBody
	response.JSON(t, &body)
	if body.Reason != "reservation_paid" {
		t.Errorf("expected reason reservation_paid, got %s", response.Body)
	}
	if got := getReservation(t, h, created.ID, admin); got.Participants[0].FirstName != "Zoé" || got.Participants[0].PriceLabel != "Adulte" {
		t.Errorf("expected Zoé to keep the Adulte price, got %+v", got.Participants[0])
	}
}

func TestAdminParticipantChangeIsValidatedAndRestrictedToAuthorizedStaff(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	otherRamble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "Adulte")})
	other := reserve(t, h, reservationPayload{RambleID: otherRamble.ID, ContactEmail: "bob@example.com", Participants: people(1, "Adulte")})
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	rename := adminParticipantPayload{FirstName: stringPtr("Zoé")}

	for _, c := range []struct {
		name          string
		id            uint
		participantID uint
		payload       adminParticipantPayload
		token         string
		code          int
	}{
		{"anonymous visitor", created.ID, created.Participants[0].ID, rename, "", http.StatusUnauthorized},
		{"the booker", created.ID, created.Participants[0].ID, rename, h.Login("alice@example.com"), http.StatusForbidden},
		{"guide of another ramble", other.ID, other.Participants[0].ID, rename, h.Token(owningGuide), http.StatusForbidden},
		{"empty name", created.ID, created.Participants[0].ID, adminParticipantPayload{FirstName: stringPtr(" ")}, h.Token(owningGuide), http.StatusBadRequest},
		{"unknown price", created.ID, created.Participants[0].ID, adminParticipantPayload{PriceLabel: stringPtr("Senior")}, h.Token(owningGuide), http.StatusBadRequest},
		{"participant of another reservation", created.ID, other.Participants[0].ID, rename, h.Token(owningGuide), http.StatusNotFound},
		{"guide of the ramble", created.ID, created.Participants[0].ID, rename, h.Token(owningGuide), http.StatusOK},
	} {
		t.Run(c.name, func(t *testing.T) {
			if response := adminUpdateParticipant(h, c.id, c.participantID, c.payload, c.token); response.Code != c.code {
				t.Fatalf("expected %d, got %d: %s", c.code, response.Code, response.Body)
			}
		})
	}
}
