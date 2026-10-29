package payments

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/clock"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/payments/stripe"
	"feldrise.com/balade/pkg/security"
	stripego "github.com/stripe/stripe-go/v83"
)

type PaymentService interface {
	// CreatePaymentForReservation creates the payment of the amount due by the
	// reservation, as it is once locked: the payments of a reservation are
	// decided one at a time. It fails with one of the errors above when the
	// reservation, once locked, cannot be paid. It gives back the pending
	// payment of the same amount to the same guide, if any, rather than
	// creating another one, and cancels every other open payment.
	CreatePaymentForReservation(reservation *dbmodel.Reservation, payerInfo PayerInfo) (*model.PaymentResponse, error)
	// CancelObsoletePayments reconciles the open payments of the reservation,
	// as it is once locked: it cancels, on Stripe too, every open payment but
	// the single one it may still pay (payable, of the amount due, to the guide
	// of the ramble), so that the others cannot be paid.
	CancelObsoletePayments(reservation *dbmodel.Reservation) error
	ProcessPaymentSuccess(paymentIntentID string) error
	ProcessPaymentFailure(paymentIntentID string, failureReason ...string) error
	RefundPayment(paymentID uint, amount *int64, reason *string) error
	GetPaymentByID(id uint) (*model.Payment, error)
	VerifyWebhookAndProcess(payload []byte, signature string, guideID uint) error
}

// Errors telling why the reservation, as it is once locked, cannot be paid.
var (
	ErrReservationCancelled    = errors.New("the reservation is cancelled")
	ErrReservationWaitingList  = errors.New("the reservation is in the waiting list")
	ErrReservationNotConfirmed = errors.New("the reservation is not confirmed")
	ErrRambleCancelled         = errors.New("the ramble is cancelled")
	ErrRambleStarted           = errors.New("payment is closed once the ramble has started")
	ErrMissingPrices           = errors.New("the price of every participant must be specified first")
	ErrNothingToPay            = errors.New("nothing to pay for this reservation")
)

type PayerInfo struct {
	Email string
	Name  string
}

type paymentService struct {
	paymentRepo       dbmodel.PaymentRepository
	rambleRepo        dbmodel.RambleRepository
	guideRepo         dbmodel.GuideRepository
	stripeService     stripe.StripeService
	encryptionService security.EncryptionService
	clock             clock.Clock
}

func NewPaymentService(
	paymentRepo dbmodel.PaymentRepository,
	rambleRepo dbmodel.RambleRepository,
	guideRepo dbmodel.GuideRepository,
	stripeService stripe.StripeService,
	encryptionService security.EncryptionService,
	clock clock.Clock,
) PaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		rambleRepo:        rambleRepo,
		guideRepo:         guideRepo,
		stripeService:     stripeService,
		encryptionService: encryptionService,
		clock:             clock,
	}
}

func (s *paymentService) CreatePaymentForReservation(reservation *dbmodel.Reservation, payerInfo PayerInfo) (*model.PaymentResponse, error) {
	var response *model.PaymentResponse
	err := s.paymentRepo.ForReservation(reservation.ID, func(ctx context.Context, payments dbmodel.PaymentRepository, locked *dbmodel.Reservation) error {
		var err error
		response, err = s.createPayment(ctx, payments, locked, payerInfo)
		return err
	})
	return response, err
}

// createPayment gives the payment of the amount due by the reservation, locked.
func (s *paymentService) createPayment(ctx context.Context, payments dbmodel.PaymentRepository, reservation *dbmodel.Reservation, payerInfo PayerInfo) (*model.PaymentResponse, error) {
	if err := s.checkPayable(reservation); err != nil {
		return nil, err
	}
	ramble := &reservation.Ramble
	if !ramble.PaymentEnabled {
		return nil, errors.New("payment is not enabled for this ramble")
	}
	amountCents := reservation.AmountDue()

	guide, err := s.getPaymentGuide(ramble)
	if err != nil {
		return nil, err
	}

	credentials, err := s.getGuideCredentials(guide)
	if err != nil {
		return nil, err
	}

	if err := s.cancelObsoletePayments(ctx, payments, reservation, guide); err != nil {
		return nil, err
	}

	if kept := keptPayment(reservation, guide); kept != nil {
		paymentIntent, err := s.stripeService.RetrievePaymentIntent(ctx, credentials, kept.StripePaymentIntentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
		}
		switch paymentIntent.Status {
		case stripego.PaymentIntentStatusCanceled:
			// Cancelled on Stripe without the webhook being received.
			kept.Status = "cancelled"
			if err := payments.UpdateStatus(kept.ID, kept.Status); err != nil {
				return nil, fmt.Errorf("failed to update payment: %w", err)
			}
		case stripego.PaymentIntentStatusSucceeded, stripego.PaymentIntentStatusProcessing:
			return nil, errors.New("the payment is being collected, awaiting its confirmation by Stripe")
		default:
			kept.Guide = *guide
			return &model.PaymentResponse{
				Payment:        kept.ToModel(),
				ClientSecret:   paymentIntent.ClientSecret,
				PublishableKey: *guide.StripePublicKey,
			}, nil
		}
	}

	reservationID := reservation.ID
	metadata := map[string]string{
		"reservation_id": strconv.FormatUint(uint64(reservationID), 10),
		"ramble_id":      strconv.FormatUint(uint64(ramble.ID), 10),
		"participants":   strconv.Itoa(len(reservation.Participants)),
		"payer_email":    payerInfo.Email,
		"payer_name":     payerInfo.Name,
	}

	idempotencyKey := paymentIdempotencyKey(reservationID, len(reservation.Payments)+1, amountCents, "eur", metadata)
	paymentIntent, err := s.stripeService.CreatePaymentIntent(ctx, credentials, amountCents, "eur", metadata, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	payment, err := payments.Create(&dbmodel.Payment{
		StripePaymentIntentID: paymentIntent.ID,
		Amount:                amountCents,
		Currency:              "eur",
		Status:                "pending",
		PaymentMethod:         "card",
		ReservationID:         &reservationID,
		PayerEmail:            payerInfo.Email,
		PayerName:             payerInfo.Name,
		GuideID:               guide.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}
	payment.Guide = *guide

	return &model.PaymentResponse{
		Payment:        payment.ToModel(),
		ClientSecret:   paymentIntent.ClientSecret,
		PublishableKey: *guide.StripePublicKey,
	}, nil
}

// paymentIdempotencyKey is the idempotency key of the creation of a payment
// intent. It changes with every payment recorded and with the parameters sent,
// so that a retry after a failure to record the payment gives back the same
// payment intent, and other parameters (amount due, payer...) a new one:
// Stripe refuses a key reused with other parameters.
func paymentIdempotencyKey(reservationID uint, paymentNumber int, amount int64, currency string, metadata map[string]string) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "amount=%d\ncurrency=%s\n", amount, currency)
	for _, key := range slices.Sorted(maps.Keys(metadata)) {
		fmt.Fprintf(hash, "metadata[%q]=%q\n", key, metadata[key])
	}
	return fmt.Sprintf("reservation-%d-payment-%d-%x", reservationID, paymentNumber, hash.Sum(nil))
}

func (s *paymentService) CancelObsoletePayments(reservation *dbmodel.Reservation) error {
	return s.paymentRepo.ForReservation(reservation.ID, func(ctx context.Context, payments dbmodel.PaymentRepository, locked *dbmodel.Reservation) error {
		guide, err := s.payableTo(locked)
		if err != nil {
			return err
		}
		return s.cancelObsoletePayments(ctx, payments, locked, guide)
	})
}

// payableTo returns the guide the reservation, locked, pays what it owes to,
// nil when it cannot be paid.
func (s *paymentService) payableTo(reservation *dbmodel.Reservation) (*dbmodel.Guide, error) {
	ramble := &reservation.Ramble
	if s.checkPayable(reservation) != nil || !ramble.PaymentEnabled || ramble.PaymentGuideID == nil {
		return nil, nil
	}
	guide, err := s.guideRepo.FindByID(*ramble.PaymentGuideID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment guide: %w", err)
	}
	if guide == nil || !guide.PaymentEnabled {
		return nil, nil
	}
	return guide, nil
}

// keptPayment returns the single open payment the reservation, locked, may
// still pay to guide: the first of the amount it owes to guide. It is nil when
// guide is nil, the reservation being then not payable.
func keptPayment(reservation *dbmodel.Reservation, guide *dbmodel.Guide) *dbmodel.Payment {
	if guide == nil {
		return nil
	}
	amountDue := reservation.AmountDue()
	for i := range reservation.Payments {
		open := &reservation.Payments[i]
		if open.IsOpen() && open.Amount == amountDue && open.GuideID == guide.ID {
			return open
		}
	}
	return nil
}

// cancelObsoletePayments cancels every open payment of the reservation, locked,
// but the one it may still pay to guide (see keptPayment), so that a single
// payment can be paid at a time.
func (s *paymentService) cancelObsoletePayments(ctx context.Context, payments dbmodel.PaymentRepository, reservation *dbmodel.Reservation, guide *dbmodel.Guide) error {
	kept := keptPayment(reservation, guide)

	for i := range reservation.Payments {
		open := &reservation.Payments[i]
		if !open.IsOpen() || open == kept {
			continue
		}

		guide, err := s.guideRepo.FindByID(open.GuideID)
		if err != nil {
			return fmt.Errorf("failed to find payment guide: %w", err)
		}
		if guide == nil {
			return errors.New("payment guide not found")
		}
		credentials, err := s.getGuideCredentials(guide)
		if err != nil {
			return err
		}
		if err := s.stripeService.CancelPaymentIntent(ctx, credentials, open.StripePaymentIntentID); err != nil {
			// Stripe refuses to cancel a payment intent cancelled already.
			paymentIntent, retrieveErr := s.stripeService.RetrievePaymentIntent(ctx, credentials, open.StripePaymentIntentID)
			if retrieveErr != nil || paymentIntent.Status != stripego.PaymentIntentStatusCanceled {
				return fmt.Errorf("failed to cancel obsolete payment %d: %w", open.ID, err)
			}
		}

		open.Status = "cancelled"
		if err := payments.UpdateStatus(open.ID, open.Status); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}
	return nil
}

// checkPayable tells why the reservation, locked, cannot be paid, if it cannot.
func (s *paymentService) checkPayable(reservation *dbmodel.Reservation) error {
	switch {
	case reservation.DeletedAt.Valid || reservation.Status == dbmodel.ReservationStatusCancelled:
		return ErrReservationCancelled
	case reservation.Status == dbmodel.ReservationStatusWaitingList:
		return ErrReservationWaitingList
	case reservation.Status != dbmodel.ReservationStatusConfirmed:
		return ErrReservationNotConfirmed
	case reservation.Ramble.IsCancelled:
		return ErrRambleCancelled
	case reservation.Ramble.HasStarted(s.clock.Now()):
		return ErrRambleStarted
	case reservation.HasMissingPrices():
		return ErrMissingPrices
	case reservation.AmountDue() <= 0:
		return ErrNothingToPay
	}
	return nil
}

func (s *paymentService) ProcessPaymentSuccess(paymentIntentID string) error {
	// A late event never brings back a refunded payment.
	return s.recordStatus(paymentIntentID, dbmodel.CollectedPaymentStatuses, map[string]any{
		"status":  "succeeded",
		"paid_at": time.Now(),
	})
}

func (s *paymentService) ProcessPaymentFailure(paymentIntentID string, failureReason ...string) error {
	return s.closePayment(paymentIntentID, "failed", failureReason...)
}

// closePayment records that the payment failed, or was cancelled on Stripe.
func (s *paymentService) closePayment(paymentIntentID string, status string, failureReason ...string) error {
	changes := map[string]any{"status": status}
	if len(failureReason) > 0 && failureReason[0] != "" {
		changes["failure_reason"] = failureReason[0]
	}
	return s.recordStatus(paymentIntentID, uncollectedOnly, changes)
}

// uncollectedOnly are the statuses a failure, a cancellation or a required
// action never replaces: a late event never downgrades a payment collected,
// nor reopens a payment cancelled.
var uncollectedOnly = append([]string{"cancelled"}, dbmodel.CollectedPaymentStatuses...)

// recordStatus applies the changes to the payment of the payment intent, in a
// single statement, unless its status is one of kept when it runs, so that a
// concurrent event is never overwritten.
func (s *paymentService) recordStatus(paymentIntentID string, kept []string, changes map[string]any) error {
	updated, err := s.paymentRepo.UpdateByStripePaymentIntentID(paymentIntentID, kept, changes)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	if updated {
		return nil
	}

	payment, err := s.paymentRepo.FindByStripePaymentIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return errPaymentNotFound
	}
	return nil
}

var errPaymentNotFound = errors.New("payment not found")

// ignoreForeignPayment ignores the failure to find the payment of the event when
// its payment intent was not created by the application (it has no reservation
// in its metadata), as another application on the same Stripe account may
// create some. A reservation payment not recorded yet fails, so that Stripe
// delivers the event again.
func ignoreForeignPayment(eventData *stripe.WebhookEventData, err error) error {
	if !errors.Is(err, errPaymentNotFound) || eventData.PaymentIntentData == nil {
		return err
	}
	if _, ok := eventData.PaymentIntentData.Metadata["reservation_id"]; ok {
		return err
	}
	log.Printf("Ignoring webhook event %s on payment intent %s, not created by the application", eventData.EventType, eventData.PaymentIntentID)
	return nil
}

// RefundPayment refunds the payment, locked for the whole operation, so that
// simultaneous refunds add up and never exceed the amount paid. A refund is
// identified on Stripe by the payment, its amount and the amount refunded
// before it: retried after its answer was lost, it is not made twice.
func (s *paymentService) RefundPayment(paymentID uint, amount *int64, reason *string) error {
	return s.paymentRepo.ForPayment(paymentID, func(ctx context.Context, payments dbmodel.PaymentRepository, payment *dbmodel.Payment) error {
		return s.refund(ctx, payments, payment, amount)
	})
}

func (s *paymentService) refund(ctx context.Context, payments dbmodel.PaymentRepository, payment *dbmodel.Payment, amount *int64) error {
	if payment == nil {
		return errors.New("payment not found")
	}

	if payment.Status != "succeeded" && payment.Status != "partially_refunded" {
		return errors.New("can only refund succeeded payments")
	}

	refundable := payment.Amount - payment.RefundedAmount()
	if amount == nil {
		amount = &refundable
	}
	if *amount <= 0 || *amount > refundable {
		return fmt.Errorf("the refund must be between 1 and %d cents", refundable)
	}

	// Get guide credentials
	credentials, err := s.getGuideCredentials(&payment.Guide)
	if err != nil {
		return err
	}

	// Process refund with Stripe, keyed so that retrying a refund whose answer
	// was lost refunds nothing more
	refundedBefore := payment.RefundedAmount()
	idempotencyKey := fmt.Sprintf("payment-%d-refund-%d-after-%d", payment.ID, *amount, refundedBefore)
	refunded, err := s.stripeService.RefundPayment(ctx, credentials, payment.StripePaymentIntentID, *amount, idempotencyKey)
	if err != nil {
		return fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment record with the amount refunded as Stripe knows it, which
	// catches up on a refund not recorded, and the payment is refunded once entirely
	payment.RefundAmount = &refunded
	payment.Status = "partially_refunded"
	if refunded >= payment.Amount {
		payment.Status = "refunded"
	}
	now := time.Now()
	payment.RefundedAt = &now

	if _, err := payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	return nil
}

func (s *paymentService) GetPaymentByID(id uint) (*model.Payment, error) {
	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return nil, nil
	}

	paymentModel := payment.ToModel()
	return &paymentModel, nil
}

func (s *paymentService) VerifyWebhookAndProcess(payload []byte, signature string, guideID uint) error {
	// Get guide's Stripe credentials
	guide, err := s.guideRepo.FindByID(guideID)
	if err != nil {
		return fmt.Errorf("failed to get guide: %w", err)
	}

	if guide == nil {
		return fmt.Errorf("guide %d not found", guideID)
	}

	// Decrypt the webhook secret
	if guide.StripeWebhookSecret == nil || *guide.StripeWebhookSecret == "" {
		return fmt.Errorf("guide %d does not have webhook secret configured", guideID)
	}

	webhookSecret, err := s.encryptionService.Decrypt(*guide.StripeWebhookSecret)
	if err != nil {
		return fmt.Errorf("failed to decrypt webhook secret: %w", err)
	}

	// Verify webhook signature
	event, err := s.stripeService.VerifyWebhookSignature(payload, signature, webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Process the webhook event
	eventData, err := s.stripeService.ProcessWebhookEvent(event)
	if err != nil {
		// Log but don't fail for unsupported events
		if strings.Contains(err.Error(), "unsupported event type") {
			log.Printf("Ignoring unsupported webhook event: %s", event.Type)
			return nil
		}
		return fmt.Errorf("failed to process webhook event: %w", err)
	}

	// Update payment status based on event type
	switch eventData.EventType {
	case "payment_intent.succeeded":
		err = ignoreForeignPayment(eventData, s.ProcessPaymentSuccess(eventData.PaymentIntentID))
		if err != nil {
			log.Printf("Failed to process payment success for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Successfully processed payment success for %s", eventData.PaymentIntentID)

	case "payment_intent.payment_failed":
		// Extract failure reason if available
		var failureReason string
		if eventData.PaymentIntentData != nil && eventData.PaymentIntentData.LastPaymentError != nil {
			failureReason = string(eventData.PaymentIntentData.LastPaymentError.Code)
		}
		if failureReason == "" {
			failureReason = "Payment failed"
		}

		err = ignoreForeignPayment(eventData, s.ProcessPaymentFailure(eventData.PaymentIntentID, failureReason))
		if err != nil {
			log.Printf("Failed to process payment failure for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Successfully processed payment failure for %s", eventData.PaymentIntentID)

	case "payment_intent.canceled":
		err = ignoreForeignPayment(eventData, s.closePayment(eventData.PaymentIntentID, "cancelled", "Payment canceled"))
		if err != nil {
			log.Printf("Failed to process payment cancellation for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Successfully processed payment cancellation for %s", eventData.PaymentIntentID)

	case "payment_intent.requires_action":
		err = ignoreForeignPayment(eventData, s.recordStatus(eventData.PaymentIntentID, uncollectedOnly, map[string]any{"status": "requires_action"}))
		if err != nil {
			log.Printf("Failed to update payment status to requires_action for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Updated payment status to requires_action for %s", eventData.PaymentIntentID)

	default:
		log.Printf("Unhandled webhook event type: %s", eventData.EventType)
	}

	return nil
}

func (s *paymentService) getPaymentGuide(ramble *dbmodel.Ramble) (*dbmodel.Guide, error) {
	if ramble.PaymentGuideID == nil {
		return nil, errors.New("no payment guide configured for this ramble")
	}

	guide, err := s.guideRepo.FindByID(*ramble.PaymentGuideID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment guide: %w", err)
	}
	if guide == nil {
		return nil, errors.New("payment guide not found")
	}

	if !guide.PaymentEnabled {
		return nil, errors.New("payment is not enabled for the guide")
	}

	return guide, nil
}

func (s *paymentService) getGuideCredentials(guide *dbmodel.Guide) (stripe.GuideStripeCredentials, error) {
	if guide.StripeSecretKey == nil || guide.StripePublicKey == nil {
		return stripe.GuideStripeCredentials{}, errors.New("guide stripe credentials not configured")
	}

	// Decrypt secret key
	secretKey, err := s.encryptionService.Decrypt(*guide.StripeSecretKey)
	if err != nil {
		return stripe.GuideStripeCredentials{}, fmt.Errorf("failed to decrypt secret key: %w", err)
	}

	credentials := stripe.GuideStripeCredentials{
		SecretKey: secretKey,
		PublicKey: *guide.StripePublicKey,
	}

	if guide.StripeAccountID != nil {
		credentials.AccountID = *guide.StripeAccountID
	}

	if guide.StripeWebhookSecret != nil {
		webhookSecret, err := s.encryptionService.Decrypt(*guide.StripeWebhookSecret)
		if err != nil {
			return stripe.GuideStripeCredentials{}, fmt.Errorf("failed to decrypt webhook secret: %w", err)
		}
		credentials.WebhookSecret = webhookSecret
	}

	return credentials, nil
}
