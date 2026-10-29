package reservation_test

import (
	"fmt"
	"net/http"
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
