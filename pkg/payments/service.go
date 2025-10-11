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
)

type PaymentService interface {
	CreatePaymentForRegistration(registrationID uint, priceLabel string, payerInfo PayerInfo, returnURL string) (*model.PaymentResponse, error)
	CreatePaymentForGroup(groupID uint, priceLabel string, payerInfo PayerInfo, returnURL string, priceSelections []model.PriceSelection) (*model.PaymentResponse, error)
	ProcessPaymentSuccess(paymentIntentID string) error
	ProcessPaymentFailure(paymentIntentID string, failureReason ...string) error
	RefundPayment(paymentID uint, amount *int64, reason *string) error
	GetPaymentByID(id uint) (*model.Payment, error)
	GetPaymentsByRegistration(registrationID uint) ([]model.Payment, error)
	GetPaymentsByGroup(groupID uint) ([]model.Payment, error)
	VerifyWebhookAndProcess(payload []byte, signature string, guideID uint) error
}

type PayerInfo struct {
	Email string
	Name  string
}

type paymentService struct {
	paymentRepo           dbmodel.PaymentRepository
	registrationRepo      dbmodel.RambleRegistrationRepository
	registrationGroupRepo dbmodel.RambleRegistrationGroupRepository
	rambleRepo            dbmodel.RambleRepository
	guideRepo             dbmodel.GuideRepository
	stripeService         stripe.StripeService
	encryptionService     security.EncryptionService
}

func NewPaymentService(
	paymentRepo dbmodel.PaymentRepository,
	registrationRepo dbmodel.RambleRegistrationRepository,
	registrationGroupRepo dbmodel.RambleRegistrationGroupRepository,
	rambleRepo dbmodel.RambleRepository,
	guideRepo dbmodel.GuideRepository,
	stripeService stripe.StripeService,
	encryptionService security.EncryptionService,
) PaymentService {
	return &paymentService{
		paymentRepo:           paymentRepo,
		registrationRepo:      registrationRepo,
		registrationGroupRepo: registrationGroupRepo,
		rambleRepo:            rambleRepo,
		guideRepo:             guideRepo,
		stripeService:         stripeService,
		encryptionService:     encryptionService,
	}
}

func (s *paymentService) CreatePaymentForRegistration(registrationID uint, priceLabel string, payerInfo PayerInfo, returnURL string) (*model.PaymentResponse, error) {
	// Get registration
	registration, err := s.registrationRepo.FindByID(registrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find registration: %w", err)
	}
	if registration == nil {
		return nil, errors.New("registration not found")
	}

	// Get ramble
	ramble, err := s.rambleRepo.FindByID(registration.RambleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find ramble: %w", err)
	}
	if ramble == nil {
		return nil, errors.New("ramble not found")
	}

	// Check if payment is enabled for this ramble
	if !ramble.PaymentEnabled {
		return nil, errors.New("payment is not enabled for this ramble")
	}

	// Find the price
	var selectedPrice *dbmodel.RamblePrice
	for _, price := range ramble.Prices {
		if price.Label == priceLabel {
			selectedPrice = &price
			break
		}
	}
	if selectedPrice == nil {
		return nil, errors.New("price not found")
	}

	// Get payment guide
	guide, err := s.getPaymentGuide(ramble)
	if err != nil {
		return nil, err
	}

	// Check if payment already exists for this registration
	existingPayments, err := s.paymentRepo.FindByRegistrationID(registrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payments: %w", err)
	}
	for _, payment := range existingPayments {
		if payment.Status == "succeeded" {
			return nil, errors.New("payment already exists for this registration")
		}
	}

	// Convert amount to cents
	amountCents := int64(selectedPrice.Amount * 100)

	// Get guide credentials
	credentials, err := s.getGuideCredentials(guide)
	if err != nil {
		return nil, err
	}

	// Create payment intent
	metadata := map[string]string{
		"registration_id": strconv.FormatUint(uint64(registrationID), 10),
		"ramble_id":       strconv.FormatUint(uint64(ramble.ID), 10),
		"price_label":     priceLabel,
		"payer_email":     payerInfo.Email,
		"payer_name":      payerInfo.Name,
	}

	paymentIntent, err := s.stripeService.CreatePaymentIntent(credentials, amountCents, "eur", metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Create payment record
	payment := &dbmodel.Payment{
		StripePaymentIntentID: paymentIntent.ID,
		Amount:                amountCents,
		Currency:              "eur",
		Status:                "pending",
		PaymentMethod:         "card",
		RegistrationID:        &registrationID,
		PayerEmail:            payerInfo.Email,
		PayerName:             payerInfo.Name,
		GuideID:               guide.ID,
	}

	payment, err = s.paymentRepo.Create(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return &model.PaymentResponse{
		Payment:        payment.ToModel(),
		ClientSecret:   paymentIntent.ClientSecret,
		PublishableKey: *guide.StripePublicKey,
	}, nil
}

func (s *paymentService) CreatePaymentForGroup(groupID uint, priceLabel string, payerInfo PayerInfo, returnURL string, priceSelections []model.PriceSelection) (*model.PaymentResponse, error) {
	// Get group
	group, err := s.registrationGroupRepo.FindByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to find group: %w", err)
	}
	if group == nil {
		return nil, errors.New("group not found")
	}

	// Get ramble
	ramble, err := s.rambleRepo.FindByID(group.RambleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find ramble: %w", err)
	}
	if ramble == nil {
		return nil, errors.New("ramble not found")
	}

	// Check if payment is enabled for this ramble
	if !ramble.PaymentEnabled {
		return nil, errors.New("payment is not enabled for this ramble")
	}

	// Get payment guide
	guide, err := s.getPaymentGuide(ramble)
	if err != nil {
		return nil, err
	}

	// Check if payment already exists for this group
	existingPayments, err := s.paymentRepo.FindByGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payments: %w", err)
	}
	for _, payment := range existingPayments {
		if payment.Status == "succeeded" {
			return nil, errors.New("payment already exists for this group")
		}
	}

	var amountCents int64
	var metadata map[string]string

	// Handle new price selections format or legacy single price
	if len(priceSelections) > 0 {
		// New format: multiple prices with quantities
		priceBreakdown := make(map[string]int) // price_label -> quantity
		totalAmount := 0.0

		// Create a map of available prices for quick lookup
		priceMap := make(map[string]*dbmodel.RamblePrice)
		for i := range ramble.Prices {
			priceMap[ramble.Prices[i].Label] = &ramble.Prices[i]
		}

		// Calculate total amount from all price selections
		for _, selection := range priceSelections {
			price, exists := priceMap[selection.PriceLabel]
			if !exists {
				return nil, fmt.Errorf("price '%s' not found for this ramble", selection.PriceLabel)
			}
			priceBreakdown[selection.PriceLabel] = selection.Quantity
			totalAmount += price.Amount * float64(selection.Quantity)
		}

		amountCents = int64(totalAmount * 100)

		// Build metadata with price breakdown
		metadata = map[string]string{
			"group_id":    strconv.FormatUint(uint64(groupID), 10),
			"ramble_id":   strconv.FormatUint(uint64(ramble.ID), 10),
			"payer_email": payerInfo.Email,
			"payer_name":  payerInfo.Name,
		}

		// Add price breakdown to metadata
		for label, qty := range priceBreakdown {
			metadata["price_"+label] = strconv.Itoa(qty)
		}
	} else {
		// Legacy format: single price label multiplied by participant count
		var selectedPrice *dbmodel.RamblePrice
		for i := range ramble.Prices {
			if ramble.Prices[i].Label == priceLabel {
				selectedPrice = &ramble.Prices[i]
				break
			}
		}
		if selectedPrice == nil {
			return nil, errors.New("price not found")
		}

		// Calculate total amount (price * number of participants)
		participantCount := len(group.Registrations)
		if participantCount == 0 {
			return nil, errors.New("no participants in group")
		}

		amountCents = int64(selectedPrice.Amount * float64(participantCount) * 100)

		// Build metadata
		metadata = map[string]string{
			"group_id":          strconv.FormatUint(uint64(groupID), 10),
			"ramble_id":         strconv.FormatUint(uint64(ramble.ID), 10),
			"price_label":       priceLabel,
			"participant_count": strconv.Itoa(participantCount),
			"payer_email":       payerInfo.Email,
			"payer_name":        payerInfo.Name,
		}
	}

	// Get guide credentials
	credentials, err := s.getGuideCredentials(guide)
	if err != nil {
		return nil, err
	}

	// Create payment intent
	paymentIntent, err := s.stripeService.CreatePaymentIntent(credentials, amountCents, "eur", metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Create payment record
	payment := &dbmodel.Payment{
		StripePaymentIntentID: paymentIntent.ID,
		Amount:                amountCents,
		Currency:              "eur",
		Status:                "pending",
		PaymentMethod:         "card",
		GroupID:               &groupID,
		PayerEmail:            payerInfo.Email,
		PayerName:             payerInfo.Name,
		GuideID:               guide.ID,
	}

	payment, err = s.paymentRepo.Create(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return &model.PaymentResponse{
		Payment:        payment.ToModel(),
		ClientSecret:   paymentIntent.ClientSecret,
		PublishableKey: *guide.StripePublicKey,
	}, nil
}

func (s *paymentService) ProcessPaymentSuccess(paymentIntentID string) error {
	payment, err := s.paymentRepo.FindByStripePaymentIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return errors.New("payment not found")
	}

	// Update payment status
	payment.Status = "succeeded"
	now := time.Now()
	payment.PaidAt = &now

	_, err = s.paymentRepo.Update(payment)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	return nil
}

func (s *paymentService) ProcessPaymentFailure(paymentIntentID string, failureReason ...string) error {
	payment, err := s.paymentRepo.FindByStripePaymentIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return errors.New("payment not found")
	}

	// Update payment status
	payment.Status = "failed"

	// Set failure reason if provided
	if len(failureReason) > 0 && failureReason[0] != "" {
		payment.FailureReason = &failureReason[0]
	}

	_, err = s.paymentRepo.Update(payment)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
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

	if payment.Status != "succeeded" {
		return errors.New("can only refund succeeded payments")
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

	// Update payment record
	payment.Status = "refunded"
	now := time.Now()
	payment.RefundedAt = &now
	if amount != nil {
		payment.RefundAmount = amount
	} else {
		payment.RefundAmount = &payment.Amount
	}

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

func (s *paymentService) GetPaymentsByRegistration(registrationID uint) ([]model.Payment, error) {
	payments, err := s.paymentRepo.FindByRegistrationID(registrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}

	paymentModels := make([]model.Payment, len(payments))
	for i, payment := range payments {
		paymentModels[i] = payment.ToModel()
	}

	return paymentModels, nil
}

func (s *paymentService) GetPaymentsByGroup(groupID uint) ([]model.Payment, error) {
	payments, err := s.paymentRepo.FindByGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}

	paymentModels := make([]model.Payment, len(payments))
	for i, payment := range payments {
		paymentModels[i] = payment.ToModel()
	}

	return paymentModels, nil
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
		err = s.ProcessPaymentFailure(eventData.PaymentIntentID, "Payment canceled")
		if err != nil {
			log.Printf("Failed to process payment cancellation for %s: %v", eventData.PaymentIntentID, err)
			return err
		}
		log.Printf("Successfully processed payment cancellation for %s", eventData.PaymentIntentID)

	case "payment_intent.requires_action":
		// Update status to requires_action if needed
		payment, err := s.paymentRepo.FindByStripePaymentIntentID(eventData.PaymentIntentID)
		if err != nil {
			log.Printf("Failed to find payment for intent %s: %v", eventData.PaymentIntentID, err)
			return err
		}

		if payment == nil {
			log.Printf("No payment found for intent %s", eventData.PaymentIntentID)
			return nil
		}

		payment.Status = "requires_action"
		_, err = s.paymentRepo.Update(payment)
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
