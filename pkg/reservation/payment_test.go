package reservation_test

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"feldrise.com/balade/database/dbmodel"
	reservations "feldrise.com/balade/pkg/reservation"
	"feldrise.com/balade/pkg/testharness"
	"gorm.io/gorm"
)

func TestOnlyTheBookerOrAnAuthorizedStaffCanReadAPayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	otherRamble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	paid := h.PayReservation(created.ID, 2500)

	admin := h.CreateUser("admin@example.com", "admin")
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	otherGuide := h.CreateUser("other-guide@example.com", "guide")
	h.CreateGuide(otherGuide, otherRamble)
	stranger := h.CreateUser("bob@example.com", "user")

	for _, c := range []struct {
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
	} {
		response := h.Do(http.MethodGet, fmt.Sprintf("/payments/%d", paid.ID), nil, c.token)
		if response.Code != c.code {
			t.Errorf("%s: expected %d, got %d: %s", c.name, c.code, response.Code, response.Body)
		}
	}
}

func refundPayment(h *testharness.Harness, id uint, amount *int64, token string) *testharness.Response {
	return h.Do(http.MethodPost, fmt.Sprintf("/payments/%d/refund", id), map[string]any{"amount": amount}, token)
}

func TestOnlyStaffAllowedToRefundCanRefundAPayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(familyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	paid := h.PayReservation(created.ID, 2500)
	owningGuide := h.CreateUser("guide@example.com", "guide")
	h.CreateGuide(owningGuide, ramble)
	admin := h.CreateUser("admin@example.com", "admin")

	for _, c := range []struct {
		name  string
		token string
		code  int
	}{
		{"anonymous visitor", "", http.StatusUnauthorized},
		{"booker", h.Login("alice@example.com"), http.StatusForbidden},
		{"guide of the ramble", h.Token(owningGuide), http.StatusForbidden},
	} {
		if response := refundPayment(h, paid.ID, nil, c.token); response.Code != c.code {
			t.Errorf("%s: expected %d, got %d: %s", c.name, c.code, response.Code, response.Body)
		}
	}
	if refunds := h.Stripe.Refunds(); len(refunds) != 0 {
		t.Fatalf("expected no refund by unauthorized users, got %+v", refunds)
	}

	if response := refundPayment(h, paid.ID, nil, h.Token(admin)); response.Code != http.StatusOK {
		t.Errorf("admin: expected 200, got %d: %s", response.Code, response.Body)
	}
}

// openConfirmation moves the clock to the start of the confirmation period of the ramble.
func openConfirmation(h *testharness.Harness, ramble *dbmodel.Ramble) {
	h.Clock.Set(ramble.Date.Add(-3*24*time.Hour + time.Hour))
}

func TestConfirmingAgainGivesBackThePendingPaymentOfTheSameAmount(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)

	var first, second reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &first)
	confirmReservation(h, created.ID, token).JSON(t, &second)

	if first.Payment == nil || second.Payment == nil {
		t.Fatalf("expected a payment to make both times, got %+v and %+v", first.Payment, second.Payment)
	}
	if second.Payment.Payment.ID != first.Payment.Payment.ID || second.Payment.ClientSecret != first.Payment.ClientSecret {
		t.Errorf("expected the same pending payment back, got %+v then %+v", first.Payment, second.Payment)
	}
	if intents := h.Stripe.PaymentIntents(); len(intents) != 1 {
		t.Errorf("expected a single Stripe payment for the same amount due, got %+v", intents)
	}
}

func TestPendingPaymentOfAnObsoleteAmountIsCancelledWhenTheAmountDueChanges(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	var confirmed reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &confirmed)

	if response := removeParticipant(h, created.ID, created.Participants[1].ID, token); response.Code != http.StatusOK {
		t.Fatalf("removing: expected 200, got %d: %s", response.Code, response.Body)
	}

	if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 1 || cancelled[0] != h.Stripe.PaymentIntents()[0].ID {
		t.Fatalf("expected the payment of 25,00 € cancelled on Stripe, got %v", cancelled)
	}
	got := getReservation(t, h, created.ID, token)
	if len(got.Payments) != 1 || got.Payments[0].Status != "cancelled" {
		t.Errorf("expected the obsolete payment cancelled, got %+v", got.Payments)
	}

	var again reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &again)
	if again.Payment == nil || again.Payment.Payment.Amount != 1250 || again.Payment.Payment.ID == confirmed.Payment.Payment.ID {
		t.Errorf("expected a new payment of 12,50 €, got %+v", again.Payment)
	}
}

func TestPendingPaymentsAreCancelledWithTheReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	confirmReservation(h, created.ID, token)

	if response := cancelReservation(h, created.ID, token); response.Code != http.StatusOK {
		t.Fatalf("cancelling: expected 200, got %d: %s", response.Code, response.Body)
	}

	if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 1 {
		t.Errorf("expected the pending payment cancelled on Stripe, got %v", cancelled)
	}
}

type refundedReservation struct {
	reservation
	HasSuccessfulPayment bool  `json:"has_successful_payment"`
	AmountPaid           int64 `json:"amount_paid"`
	AmountDue            int64 `json:"amount_due"`
	Payments             []struct {
		Status       string `json:"status"`
		Amount       int64  `json:"amount"`
		RefundAmount *int64 `json:"refund_amount"`
	} `json:"payments"`
}

func int64Ptr(value int64) *int64 {
	return &value
}

func TestPartiallyRefundedPaymentCountsForItsNetAmount(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	h.Clock.Set(ramble.Date.Add(-2 * 24 * time.Hour))
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	paid := h.PayReservation(created.ID, 2500)
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	token := h.Login("alice@example.com")

	if response := refundPayment(h, paid.ID, int64Ptr(1000), admin); response.Code != http.StatusOK {
		t.Fatalf("refunding: expected 200, got %d: %s", response.Code, response.Body)
	}

	var got refundedReservation
	h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", created.ID), nil, token).JSON(t, &got)
	if !got.HasSuccessfulPayment || got.AmountPaid != 1500 || got.AmountDue != 1000 {
		t.Errorf("expected 15,00 € paid and 10,00 € due, got paid %d, due %d, successful %v", got.AmountPaid, got.AmountDue, got.HasSuccessfulPayment)
	}
	if p := got.Payments[0]; p.Status != "partially_refunded" || p.RefundAmount == nil || *p.RefundAmount != 1000 {
		t.Errorf("expected the payment partially refunded of 10,00 €, got %+v", p)
	}
	var confirmed reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &confirmed)
	if confirmed.Payment == nil || confirmed.Payment.Payment.Amount != 1000 {
		t.Errorf("expected the 10,00 € refunded to be due again, got %+v", confirmed.Payment)
	}

	if response := refundPayment(h, paid.ID, nil, admin); response.Code != http.StatusOK {
		t.Fatalf("refunding the rest: expected 200, got %d: %s", response.Code, response.Body)
	}
	h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", created.ID), nil, token).JSON(t, &got)
	if got.HasSuccessfulPayment || got.AmountPaid != 0 || got.Payments[0].Status != "refunded" || *got.Payments[0].RefundAmount != 2500 {
		t.Errorf("expected the payment entirely refunded, got %+v", got)
	}
	if refunds := h.Stripe.Refunds(); len(refunds) != 2 || refunds[1].Amount == nil || *refunds[1].Amount != 1500 {
		t.Errorf("expected the second refund to be of the 15,00 € left, got %+v", refunds)
	}
}

type paymentView struct {
	ID                    uint   `json:"id"`
	Status                string `json:"status"`
	GuideID               uint   `json:"guide_id"`
	StripePaymentIntentID string `json:"stripe_payment_intent_id"`
}

func getPayment(t *testing.T, h *testharness.Harness, id uint, token string) paymentView {
	t.Helper()

	response := h.Do(http.MethodGet, fmt.Sprintf("/payments/%d", id), nil, token)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /payments/%d: expected 200, got %d: %s", id, response.Code, response.Body)
	}
	var got paymentView
	response.JSON(t, &got)
	return got
}

func TestLateWebhooksNeverDowngradeASucceededOrRefundedPayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	var confirmed reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &confirmed)
	id := confirmed.Payment.Payment.ID
	p := getPayment(t, h, id, token)

	if response := h.SendStripeWebhook(p.GuideID, "payment_intent.succeeded", p.StripePaymentIntentID); response.Code != http.StatusOK {
		t.Fatalf("succeeded webhook: expected 200, got %d: %s", response.Code, response.Body)
	}
	for _, late := range []string{"payment_intent.payment_failed", "payment_intent.requires_action", "payment_intent.canceled"} {
		h.SendStripeWebhook(p.GuideID, late, p.StripePaymentIntentID)
	}
	if got := getPayment(t, h, id, token); got.Status != "succeeded" {
		t.Fatalf("expected the payment to stay succeeded after late webhooks, got %q", got.Status)
	}

	admin := h.Token(h.CreateUser("admin@example.com", "admin"))
	if response := refundPayment(h, id, nil, admin); response.Code != http.StatusOK {
		t.Fatalf("refunding: expected 200, got %d: %s", response.Code, response.Body)
	}
	for _, late := range []string{"payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.requires_action"} {
		h.SendStripeWebhook(p.GuideID, late, p.StripePaymentIntentID)
	}
	if got := getPayment(t, h, id, token); got.Status != "refunded" {
		t.Errorf("expected the payment to stay refunded after late webhooks, got %q", got.Status)
	}
}

func TestParticipantsAddedWhenThePaymentCannotBePreparedArePayableLater(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	h.Clock.Set(ramble.Date.Add(-2 * 24 * time.Hour))
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	h.PayReservation(created.ID, 2500)
	token := h.Login("alice@example.com")
	h.Stripe.SetUnavailable(true)

	response := addParticipants(h, created.ID, []participantPayload{{FirstName: "Zoé", LastName: "Martin", PriceLabel: "Enfant"}}, token)

	if response.Code != http.StatusOK {
		t.Fatalf("expected the addition to succeed, got %d: %s", response.Code, response.Body)
	}
	var body struct {
		reservationWithPayment
		PaymentUnavailable bool `json:"payment_unavailable"`
	}
	response.JSON(t, &body)
	if len(body.Participants) != 3 || body.Payment != nil || !body.PaymentUnavailable {
		t.Errorf("expected the participant added, with the payment to make later, got %s", response.Body)
	}

	h.Stripe.SetUnavailable(false)
	var later reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &later)
	if later.Payment == nil || later.Payment.Payment.Amount != 600 {
		t.Errorf("expected the complement of 6,00 € payable later, got %+v", later.Payment)
	}
}

func TestSimultaneousConfirmationsPrepareASinglePayment(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	h.Stripe.SetCreationDelay(200 * time.Millisecond)

	responses := make([]reservationWithPayment, 2)
	var wg sync.WaitGroup
	for i := range responses {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := confirmReservation(h, created.ID, token)
			if response.Code != http.StatusOK {
				t.Errorf("confirmation %d: expected 200, got %d: %s", i, response.Code, response.Body)
				return
			}
			response.JSON(t, &responses[i])
		}()
	}
	wg.Wait()

	if intents := h.Stripe.PaymentIntents(); len(intents) != 1 || intents[0].IdempotencyKey == "" {
		t.Fatalf("expected a single Stripe payment created with an idempotency key, got %+v", intents)
	}
	first, second := responses[0].Payment, responses[1].Payment
	if first == nil || second == nil || first.Payment.ID != second.Payment.ID {
		t.Errorf("expected both confirmations to give the same payment, got %+v and %+v", first, second)
	}
	if got := getReservation(t, h, created.ID, token); len(got.Payments) != 1 {
		t.Errorf("expected a single payment recorded, got %+v", got.Payments)
	}
}

func TestPaymentDeclinedOnceStaysPayableThenIsCancelledWithTheReservation(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	var confirmed reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &confirmed)
	p := getPayment(t, h, confirmed.Payment.Payment.ID, token)
	if response := h.SendStripeWebhook(p.GuideID, "payment_intent.payment_failed", p.StripePaymentIntentID); response.Code != http.StatusOK {
		t.Fatalf("failed webhook: expected 200, got %d: %s", response.Code, response.Body)
	}

	var again reservationWithPayment
	confirmReservation(h, created.ID, token).JSON(t, &again)
	if again.Payment == nil || again.Payment.Payment.ID != confirmed.Payment.Payment.ID {
		t.Errorf("expected the declined payment, still payable on Stripe, given back, got %+v", again.Payment)
	}
	if intents := h.Stripe.PaymentIntents(); len(intents) != 1 {
		t.Errorf("expected no other Stripe payment, got %+v", intents)
	}

	if response := cancelReservation(h, created.ID, token); response.Code != http.StatusOK {
		t.Fatalf("cancelling: expected 200, got %d: %s", response.Code, response.Body)
	}
	if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 1 || cancelled[0] != p.StripePaymentIntentID {
		t.Errorf("expected the declined payment cancelled on Stripe, got %v", cancelled)
	}
	if got := getPayment(t, h, p.ID, token); got.Status != "cancelled" {
		t.Errorf("expected the declined payment cancelled, got %q", got.Status)
	}
}

func TestOpenPaymentsAreCancelledWhenAnAdminDeletesTheReservation(t *testing.T) {
	for _, bulk := range []bool{false, true} {
		t.Run(fmt.Sprintf("bulk %v", bulk), func(t *testing.T) {
			h := testharness.New(t)
			ramble := h.CreateRamble(payingFamilyRamble())
			created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
			openConfirmation(h, ramble)
			confirmReservation(h, created.ID, h.Login("alice@example.com"))
			admin := h.Token(h.CreateUser("admin@example.com", "admin"))

			if bulk {
				if result := adminBulk(t, h, bulkPayload{ReservationIDs: []uint{created.ID}, Action: "delete"}, admin); result.SuccessCount != 1 {
					t.Fatalf("expected the deletion to succeed, got %+v", result)
				}
			} else if response := adminDelete(h, created.ID, admin); response.Code != http.StatusNoContent {
				t.Fatalf("deleting: expected 204, got %d: %s", response.Code, response.Body)
			}

			if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 1 || cancelled[0] != h.Stripe.PaymentIntents()[0].ID {
				t.Errorf("expected the open payment cancelled on Stripe, got %v", cancelled)
			}
		})
	}
}

// waitForLockWait waits until a query of the database waits for a lock.
func waitForLockWait(t *testing.T, h *testharness.Harness) {
	t.Helper()

	for range 100 {
		var waiting int64
		h.DB.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND datname = current_database()").Scan(&waiting)
		if waiting > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected a query to wait for the payment being collected")
}

func TestWebhookReceivedWhileThePaymentSucceedsNeverDowngradesIt(t *testing.T) {
	for _, late := range []string{"payment_intent.payment_failed", "payment_intent.requires_action", "payment_intent.canceled"} {
		t.Run(late, func(t *testing.T) {
			h := testharness.New(t)
			ramble := h.CreateRamble(payingFamilyRamble())
			created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
			token := h.Login("alice@example.com")
			openConfirmation(h, ramble)
			var confirmed reservationWithPayment
			confirmReservation(h, created.ID, token).JSON(t, &confirmed)
			p := getPayment(t, h, confirmed.Payment.Payment.ID, token)

			// The success is being recorded while the other webhook arrives.
			succeeding := h.DB.Begin()
			if err := succeeding.Exec("UPDATE payments SET status = 'succeeded' WHERE id = ?", p.ID).Error; err != nil {
				t.Fatalf("recording the success: %v", err)
			}
			done := make(chan *testharness.Response)
			go func() { done <- h.SendStripeWebhook(p.GuideID, late, p.StripePaymentIntentID) }()
			waitForLockWait(t, h)
			if err := succeeding.Commit().Error; err != nil {
				t.Fatalf("committing the success: %v", err)
			}
			if response := <-done; response.Code != http.StatusOK {
				t.Fatalf("webhook: expected 200, got %d: %s", response.Code, response.Body)
			}

			if got := getPayment(t, h, p.ID, token); got.Status != "succeeded" {
				t.Errorf("expected the payment to stay succeeded, got %q", got.Status)
			}
		})
	}
}

func TestParticipantsCannotBeAddedOnceTheRambleHasStarted(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	confirmReservation(h, created.ID, token)
	h.Clock.Set(ramble.Date.Add(time.Minute))

	response := addParticipants(h, created.ID, []participantPayload{{FirstName: "Zoé", LastName: "Martin", PriceLabel: "Enfant"}}, token)

	if reason := confirmationRefusal(t, response); reason != "ramble_started" {
		t.Errorf("expected reason ramble_started, got %q", reason)
	}
	if intents := h.Stripe.PaymentIntents(); len(intents) != 1 {
		t.Errorf("expected no payment prepared once the ramble started, got %+v", intents)
	}
}

func TestSchedulerCancelsTheOpenPaymentsOfStartedRambles(t *testing.T) {
	h := testharness.New(t)
	started := h.CreateRamble(payingFamilyRamble())
	later := h.CreateRamble(payingFamilyRamble())
	startedReservation := reserve(t, h, reservationPayload{RambleID: started.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	openConfirmation(h, started)
	confirmReservation(h, startedReservation.ID, h.Login("alice@example.com"))
	startedIntent := h.Stripe.PaymentIntents()[0].ID
	h.DB.Model(&dbmodel.Ramble{}).Where("id = ?", later.ID).Update("date", started.Date.Add(48*time.Hour))
	laterReservation := reserve(t, h, reservationPayload{RambleID: later.ID, ContactEmail: "bob@example.com", Participants: people(1, "Adulte")})
	h.DB.Model(&dbmodel.Reservation{}).Where("id = ?", laterReservation.ID).Update("status", "confirmed")
	var laterPayment reservationWithPayment
	confirmReservation(h, laterReservation.ID, h.Login("bob@example.com")).JSON(t, &laterPayment)
	if laterPayment.Payment == nil {
		t.Fatalf("expected a payment for the later ramble, got %+v", laterPayment)
	}

	h.Clock.Set(started.Date.Add(time.Minute))
	if err := reservations.New(h.Config).CancelPaymentsOfStartedRambles(h.Clock.Now()); err != nil {
		t.Fatalf("cancelling the payments of started rambles: %v", err)
	}

	if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 1 || cancelled[0] != startedIntent {
		t.Errorf("expected only the payment of the started ramble cancelled on Stripe, got %v", cancelled)
	}
	got := getReservation(t, h, startedReservation.ID, h.Login("alice@example.com"))
	if len(got.Payments) != 1 || got.Payments[0].Status != "cancelled" {
		t.Errorf("expected the open payment of the started ramble cancelled, got %+v", got.Payments)
	}
}

func TestPaymentToMakeNamesItsGuide(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)

	for _, attempt := range []string{"created", "given back"} {
		var body struct {
			Payment struct {
				Payment struct {
					Guide struct {
						ID uint `json:"id"`
					} `json:"guide"`
				} `json:"payment"`
			} `json:"payment"`
		}
		confirmReservation(h, created.ID, token).JSON(t, &body)
		if got := body.Payment.Payment.Guide.ID; got != *ramble.PaymentGuideID {
			t.Errorf("payment %s: expected its guide %d, got %d", attempt, *ramble.PaymentGuideID, got)
		}
	}
}

func TestRefundedPaymentWithoutTheAmountRefundedCountsAsEntirelyRefunded(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	h.Clock.Set(ramble.Date.Add(-2 * 24 * time.Hour))
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
	paid := h.PayReservation(created.ID, 2500)
	// As a legacy refund recorded without the amount refunded.
	h.DB.Model(&dbmodel.Payment{}).Where("id = ?", paid.ID).Update("status", "refunded")

	var got refundedReservation
	h.Do(http.MethodGet, fmt.Sprintf("/reservations/%d", created.ID), nil, h.Login("alice@example.com")).JSON(t, &got)
	if got.HasSuccessfulPayment || got.AmountPaid != 0 || got.AmountDue != 2500 {
		t.Errorf("expected nothing kept and 25,00 € due, got paid %d, due %d, successful %v", got.AmountPaid, got.AmountDue, got.HasSuccessfulPayment)
	}
}

func TestPaymentIsNotPreparedWhenTheReservationStopsBeingPayableDuringTheConfirmation(t *testing.T) {
	cases := []struct {
		name   string
		change func(tx *gorm.DB, created reservation, ramble *dbmodel.Ramble)
		reason string
	}{
		{"reservation cancelled", func(tx *gorm.DB, created reservation, ramble *dbmodel.Ramble) {
			tx.Exec("UPDATE reservations SET status = 'cancelled' WHERE id = ?", created.ID)
		}, "reservation_cancelled"},
		{"ramble cancelled", func(tx *gorm.DB, created reservation, ramble *dbmodel.Ramble) {
			tx.Exec("UPDATE rambles SET is_cancelled = true WHERE id = ?", ramble.ID)
		}, "ramble_cancelled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := testharness.New(t)
			ramble := h.CreateRamble(payingFamilyRamble())
			created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
			token := h.Login("alice@example.com")
			openConfirmation(h, ramble)
			h.DB.Model(&dbmodel.Reservation{}).Where("id = ?", created.ID).Update("status", "confirmed")

			// The change is being made while the confirmation decides the payment.
			changing := h.DB.Begin()
			if err := changing.Exec("SELECT id FROM reservations WHERE id = ? FOR UPDATE", created.ID).Error; err != nil {
				t.Fatalf("locking the reservation: %v", err)
			}
			done := make(chan *testharness.Response)
			go func() { done <- confirmReservation(h, created.ID, token) }()
			waitForLockWait(t, h)
			c.change(changing, created, ramble)
			if err := changing.Commit().Error; err != nil {
				t.Fatalf("committing the change: %v", err)
			}

			if reason := confirmationRefusal(t, <-done); reason != c.reason {
				t.Errorf("expected reason %s, got %q", c.reason, reason)
			}
			if intents := h.Stripe.PaymentIntents(); len(intents) != 0 {
				t.Errorf("expected no payment prepared, got %+v", intents)
			}
		})
	}
}

func TestSchedulerCancelsTheOpenPaymentsLeftByACancellationOrADeletion(t *testing.T) {
	for _, closing := range []string{"cancelled by the booker", "deleted by an admin"} {
		t.Run(closing, func(t *testing.T) {
			h := testharness.New(t)
			ramble := h.CreateRamble(payingFamilyRamble())
			created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: people(2, "Adulte")})
			token := h.Login("alice@example.com")
			openConfirmation(h, ramble)
			confirmReservation(h, created.ID, token)
			intent := h.Stripe.PaymentIntents()[0].ID

			h.Stripe.SetCancellationFailing(true)
			if closing == "deleted by an admin" {
				if response := adminDelete(h, created.ID, h.Token(h.CreateUser("admin@example.com", "admin"))); response.Code != http.StatusNoContent {
					t.Fatalf("deleting: expected 204, got %d: %s", response.Code, response.Body)
				}
			} else if response := cancelReservation(h, created.ID, token); response.Code != http.StatusOK {
				t.Fatalf("cancelling: expected 200, got %d: %s", response.Code, response.Body)
			}
			if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 0 {
				t.Fatalf("expected the cancellation on Stripe to fail, got %v", cancelled)
			}

			h.Stripe.SetCancellationFailing(false)
			if err := reservations.New(h.Config).CancelPaymentsOfCancelledReservations(); err != nil {
				t.Fatalf("cancelling the payments of cancelled reservations: %v", err)
			}

			if cancelled := h.Stripe.CancelledPaymentIntents(); len(cancelled) != 1 || cancelled[0] != intent {
				t.Errorf("expected the open payment cancelled on Stripe at the next pass, got %v", cancelled)
			}
			var status string
			h.DB.Raw("SELECT status FROM payments WHERE stripe_payment_intent_id = ?", intent).Scan(&status)
			if status != "cancelled" {
				t.Errorf("expected the payment recorded cancelled, got %q", status)
			}
		})
	}
}

func TestPaymentNotRecordedIsPreparedAgainAfterTheBookerNameChanges(t *testing.T) {
	h := testharness.New(t)
	ramble := h.CreateRamble(payingFamilyRamble())
	participants := people(2, "Adulte")
	participants[0].IsBooker = true
	created := reserve(t, h, reservationPayload{RambleID: ramble.ID, ContactEmail: "alice@example.com", Participants: participants})
	token := h.Login("alice@example.com")
	openConfirmation(h, ramble)
	confirmReservation(h, created.ID, token)
	// As if the payment created on Stripe had failed to be recorded.
	h.DB.Unscoped().Where("reservation_id = ?", created.ID).Delete(&dbmodel.Payment{})

	booker := created.Participants[0]
	if response := updateParticipant(h, created.ID, booker.ID, participantUpdatePayload{FirstName: "Alicia", LastName: booker.LastName, PriceLabel: "Adulte"}, token); response.Code != http.StatusOK {
		t.Fatalf("renaming the booker: expected 200, got %d: %s", response.Code, response.Body)
	}

	var again struct {
		reservationWithPayment
		PaymentUnavailable bool `json:"payment_unavailable"`
	}
	confirmReservation(h, created.ID, token).JSON(t, &again)
	if again.PaymentUnavailable || again.Payment == nil || again.Payment.Payment.Amount != 2500 {
		t.Errorf("expected the payment of 25,00 € prepared, got %+v, unavailable %v", again.Payment, again.PaymentUnavailable)
	}
}

func TestWebhookOnAPaymentUnknownIsRetriedOnlyWhenItIsAReservationPayment(t *testing.T) {
	for _, event := range []string{"payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled", "payment_intent.requires_action"} {
		t.Run(event, func(t *testing.T) {
			h := testharness.New(t)
			ramble := h.CreateRamble(payingFamilyRamble())

			if response := h.SendStripeWebhookWithMetadata(*ramble.PaymentGuideID, event, "pi_other_app", map[string]string{"order": "42"}); response.Code != http.StatusOK {
				t.Errorf("payment of another application: expected 200, got %d: %s", response.Code, response.Body)
			}
			if response := h.SendStripeWebhookWithMetadata(*ramble.PaymentGuideID, event, "pi_not_recorded_yet", map[string]string{"reservation_id": "1"}); response.Code == http.StatusOK {
				t.Errorf("reservation payment not recorded yet: expected an error for Stripe to deliver it again, got 200")
			}
		})
	}
}
