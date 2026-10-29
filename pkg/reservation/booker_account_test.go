package reservation_test

import (
	"net/http"
	"sync"
	"testing"
	"time"

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

// reserveAtOnce sends the payloads simultaneously and returns the responses in order.
func reserveAtOnce(h *testharness.Harness, payloads []reservationPayload) []*testharness.Response {
	responses := make([]*testharness.Response, len(payloads))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, payload := range payloads {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			responses[i] = h.Do(http.MethodPost, "/reservations", payload, "")
		}()
	}
	close(start)
	wg.Wait()
	return responses
}

func TestSimultaneousReservationsOfANewEmailShareASingleAccount(t *testing.T) {
	h := testharness.New(t)

	var payloads []reservationPayload
	for range 30 {
		ramble := h.CreateRamble(familyRamble())
		payloads = append(payloads, reservationPayload{RambleID: ramble.ID, ContactEmail: "new@example.com", Participants: people(1, "Adulte")})
	}

	for i, response := range reserveAtOnce(h, payloads) {
		if response.Code != http.StatusCreated {
			t.Errorf("reservation %d: expected 201, got %d: %s", i, response.Code, response.Body)
		}
	}

	var mine []reservation
	h.Do(http.MethodGet, "/reservations", nil, h.Login("new@example.com")).JSON(t, &mine)
	if len(mine) != len(payloads) {
		t.Errorf("expected the %d reservations on a single account, got %d", len(payloads), len(mine))
	}
}

func TestSimultaneousReservationsOfANewEmailOnTheSameRambleAreRefusedButOne(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())

	var payloads []reservationPayload
	for range 30 {
		payloads = append(payloads, reservationPayload{RambleID: ramble.ID, ContactEmail: "new@example.com", Participants: people(1, "Adulte")})
	}

	created := 0
	for i, response := range reserveAtOnce(h, payloads) {
		switch response.Code {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			var body errorBody
			response.JSON(t, &body)
			if body.Reason != "reservation_already_exists" {
				t.Errorf("reservation %d: expected reason reservation_already_exists, got %s", i, response.Body)
			}
		default:
			t.Errorf("reservation %d: expected 201 or 409, got %d: %s", i, response.Code, response.Body)
		}
	}
	if created != 1 {
		t.Errorf("expected a single reservation created, got %d", created)
	}
}

func TestReservationDecidesItsPlacesOnTheRambleAsItIsOnceLocked(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(testharness.RambleOptions{MaxParticipants: testharness.Int(10)})
	reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(1, "")})

	// An update of the ramble holds its lock while the reservation arrives.
	update := h.DB.Begin()
	if err := update.Exec("SELECT id FROM rambles WHERE id = ? FOR UPDATE", ramble.ID).Error; err != nil {
		t.Fatalf("locking the ramble: %v", err)
	}
	if err := update.Exec("UPDATE rambles SET max_participants = 1 WHERE id = ?", ramble.ID).Error; err != nil {
		t.Fatalf("lowering the capacity: %v", err)
	}
	responses := make(chan *testharness.Response, 1)
	go func() {
		responses <- h.Do(http.MethodPost, "/reservations", reservationPayload{RambleID: ramble.ID, ContactEmail: "bob@example.com", Participants: people(1, "")}, "")
	}()
	time.Sleep(300 * time.Millisecond)
	if err := update.Commit().Error; err != nil {
		t.Fatalf("committing the update: %v", err)
	}

	response := <-responses
	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body)
	}
	var created reservation
	response.JSON(t, &created)
	if created.Status != "waiting_list" {
		t.Errorf("expected the reservation to wait on the full ramble, got %q", created.Status)
	}
}
