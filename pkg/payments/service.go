package payments

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/payments/stripe"
	"feldrise.com/balade/pkg/security"
	stripego "github.com/stripe/stripe-go/v83"
)

type PaymentService interface {
	// CreatePaymentForReservation creates the payment of the amount due by the
	// reservation, as it is once locked: the payments of a reservation are
	// decided one at a time. It gives back the pending payment of the same
	// amount, if any, rather than creating another one, and cancels the pending
	// payments of other amounts.
	CreatePaymentForReservation(reservation *dbmodel.Reservation, payerInfo PayerInfo) (*model.PaymentResponse, error)
	// CancelObsoletePayments cancels, on Stripe too, the pending payments of the
	// reservation, as it is once locked, that no longer match what it owes, so
	// that they cannot be paid.
	CancelObsoletePayments(reservation *dbmodel.Reservation) error
	ProcessPaymentSuccess(paymentIntentID string) error
	ProcessPaymentFailure(paymentIntentID string, failureReason ...string) error
	RefundPayment(paymentID uint, amount *int64, reason *string) error
	GetPaymentByID(id uint) (*model.Payment, error)
	VerifyWebhookAndProcess(payload []byte, signature string, guideID uint) error
}

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
}

func NewPaymentService(
	paymentRepo dbmodel.PaymentRepository,
	rambleRepo dbmodel.RambleRepository,
	guideRepo dbmodel.GuideRepository,
	stripeService stripe.StripeService,
	encryptionService security.EncryptionService,
) PaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		rambleRepo:        rambleRepo,
		guideRepo:         guideRepo,
		stripeService:     stripeService,
		encryptionService: encryptionService,
	}
}

func (s *paymentService) CreatePaymentForReservation(reservation *dbmodel.Reservation, payerInfo PayerInfo) (*model.PaymentResponse, error) {
	var response *model.PaymentResponse
	err := s.paymentRepo.ForReservation(reservation.ID, func(payments dbmodel.PaymentRepository, locked *dbmodel.Reservation) error {
		var err error
		response, err = s.createPayment(payments, locked, payerInfo)
		return err
	})
	return response, err
}

// createPayment gives the payment of the amount due by the reservation, locked.
func (s *paymentService) createPayment(payments dbmodel.PaymentRepository, reservation *dbmodel.Reservation, payerInfo PayerInfo) (*model.PaymentResponse, error) {
	ramble := &reservation.Ramble
	if !ramble.PaymentEnabled {
		return nil, errors.New("payment is not enabled for this ramble")
	}

	amountCents := reservation.AmountDue()
	if amountCents <= 0 {
		return nil, errors.New("nothing to pay for this reservation")
	}

	guide, err := s.getPaymentGuide(ramble)
	if err != nil {
		return nil, err
	}

	credentials, err := s.getGuideCredentials(guide)
	if err != nil {
		return nil, err
	}

	for i := range reservation.Payments {
		open := &reservation.Payments[i]
		if !open.IsOpen() || open.Amount != amountCents || open.GuideID != guide.ID {
			continue
		}

		paymentIntent, err := s.stripeService.RetrievePaymentIntent(credentials, open.StripePaymentIntentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
		}
		switch paymentIntent.Status {
		case stripego.PaymentIntentStatusCanceled:
			// Cancelled on Stripe without the webhook being received.
			open.Status = "cancelled"
			if err := payments.UpdateStatus(open.ID, open.Status); err != nil {
				return nil, fmt.Errorf("failed to update payment: %w", err)
			}
			continue
		case stripego.PaymentIntentStatusSucceeded, stripego.PaymentIntentStatusProcessing:
			return nil, errors.New("the payment is being collected, awaiting its confirmation by Stripe")
		}

		return &model.PaymentResponse{
			Payment:        open.ToModel(),
			ClientSecret:   paymentIntent.ClientSecret,
			PublishableKey: *guide.StripePublicKey,
		}, nil
	}

	if err := s.cancelObsoletePayments(payments, reservation); err != nil {
		return nil, err
	}

	reservationID := reservation.ID
	metadata := map[string]string{
		"reservation_id": strconv.FormatUint(uint64(reservationID), 10),
		"ramble_id":      strconv.FormatUint(uint64(ramble.ID), 10),
		"participants":   strconv.Itoa(len(reservation.Participants)),
		"payer_email":    payerInfo.Email,
		"payer_name":     payerInfo.Name,
	}

	// The key changes with every payment recorded and with the amount due, so
	// that a retry after a failure to record the payment gives back the same
	// payment intent, and a new amount due a new one.
	idempotencyKey := fmt.Sprintf("reservation-%d-payment-%d-amount-%d", reservationID, len(reservation.Payments)+1, amountCents)
	paymentIntent, err := s.stripeService.CreatePaymentIntent(credentials, amountCents, "eur", metadata, idempotencyKey)
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

	return &model.PaymentResponse{
		Payment:        payment.ToModel(),
		ClientSecret:   paymentIntent.ClientSecret,
		PublishableKey: *guide.StripePublicKey,
	}, nil
}

func (s *paymentService) CancelObsoletePayments(reservation *dbmodel.Reservation) error {
	return s.paymentRepo.ForReservation(reservation.ID, func(payments dbmodel.PaymentRepository, locked *dbmodel.Reservation) error {
		return s.cancelObsoletePayments(payments, locked)
	})
}

// cancelObsoletePayments cancels the open payments of the reservation, locked,
// that no longer match what it owes.
func (s *paymentService) cancelObsoletePayments(payments dbmodel.PaymentRepository, reservation *dbmodel.Reservation) error {
	amountDue := reservation.AmountDue()
	payable := !reservation.DeletedAt.Valid && reservation.Status == dbmodel.ReservationStatusConfirmed &&
		!reservation.HasMissingPrices() && !reservation.Ramble.IsCancelled

	for i := range reservation.Payments {
		open := &reservation.Payments[i]
		if !open.IsOpen() || (payable && open.Amount == amountDue) {
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
		if err := s.stripeService.CancelPaymentIntent(credentials, open.StripePaymentIntentID); err != nil {
			// Stripe refuses to cancel a payment intent cancelled already.
			paymentIntent, retrieveErr := s.stripeService.RetrievePaymentIntent(credentials, open.StripePaymentIntentID)
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
		return errors.New("payment not found")
	}
	return nil
}

func (s *paymentService) RefundPayment(paymentID uint, amount *int64, reason *string) error {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
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

	// Process refund with Stripe
	_, err = s.stripeService.RefundPayment(credentials, payment.StripePaymentIntentID, amount)
	if err != nil {
		return fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment record: the refunds add up, and the payment is refunded once entirely
	refunded := payment.RefundedAmount() + *amount
	payment.RefundAmount = &refunded
	payment.Status = "partially_refunded"
	if refunded >= payment.Amount {
		payment.Status = "refunded"
	}
	now := time.Now()
	payment.RefundedAt = &now

	_, err = s.paymentRepo.Update(payment)
	if err != nil {
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
		err = s.ProcessPaymentSuccess(eventData.PaymentIntentID)
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

		err = s.ProcessPaymentFailure(eventData.PaymentIntentID, failureReason)
		if err != nil {
			log.Printf("Failed to process payment failure for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Successfully processed payment failure for %s", eventData.PaymentIntentID)

	case "payment_intent.canceled":
		err = s.closePayment(eventData.PaymentIntentID, "cancelled", "Payment canceled")
		if err != nil {
			log.Printf("Failed to process payment cancellation for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Successfully processed payment cancellation for %s", eventData.PaymentIntentID)

	case "payment_intent.requires_action":
		err = s.recordStatus(eventData.PaymentIntentID, uncollectedOnly, map[string]any{"status": "requires_action"})
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
