package reservation_test

import (
	"net/http"
	"testing"

	"feldrise.com/balade/pkg/testharness"
)

func TestBookerAccountIsFoundWhateverTheCaseOfTheEmail(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	h.CreateUser("alice@example.com", "user")

	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "Alice@Example.com", Participants: people(1, "Adulte")})

	response := h.Do(http.MethodGet, "/reservations", nil, h.Login("alice@example.com"))
	var mine []reservation
	response.JSON(t, &mine)
	if len(mine) != 1 || mine[0].ID != created.ID {
		t.Fatalf("expected the existing account to hold reservation %d, got %s", created.ID, response.Body)
	}
}
