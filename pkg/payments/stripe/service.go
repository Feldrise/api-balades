package stripe

import (
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/paymentintent"
	"github.com/stripe/stripe-go/v83/refund"
	"github.com/stripe/stripe-go/v83/webhook"
)

type GuideStripeCredentials struct {
	AccountID     string
	SecretKey     string
	PublicKey     string
	WebhookSecret string
}

type StripeService interface {
	CreatePaymentIntent(credentials GuideStripeCredentials, amount int64, currency string, metadata map[string]string) (*stripe.PaymentIntent, error)
	ConfirmPayment(credentials GuideStripeCredentials, paymentIntentID string) (*stripe.PaymentIntent, error)
	RefundPayment(credentials GuideStripeCredentials, paymentIntentID string, amount *int64) (*stripe.Refund, error)
	VerifyWebhookSignature(payload []byte, signature string, secret string) (*stripe.Event, error)
	ProcessWebhookEvent(event *stripe.Event) (*WebhookEventData, error)
}

type WebhookEventData struct {
	EventType         string
	PaymentIntentID   string
	PaymentIntentData *stripe.PaymentIntent
}

type stripeService struct{}

func NewStripeService() StripeService {
	return &stripeService{}
}

func (s *stripeService) CreatePaymentIntent(credentials GuideStripeCredentials, amount int64, currency string, metadata map[string]string) (*stripe.PaymentIntent, error) {
	// Set the API key for this specific guide
	stripe.Key = credentials.SecretKey

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Params: stripe.Params{
			Metadata: metadata,
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	// If the guide has a connected account, use it
	if credentials.AccountID != "" {
		params.SetStripeAccount(credentials.AccountID)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return pi, nil
}

func (s *stripeService) ConfirmPayment(credentials GuideStripeCredentials, paymentIntentID string) (*stripe.PaymentIntent, error) {
	stripe.Key = credentials.SecretKey

	params := &stripe.PaymentIntentConfirmParams{}

	if credentials.AccountID != "" {
		params.SetStripeAccount(credentials.AccountID)
	}

	pi, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}

	return pi, nil
}

func (s *stripeService) RefundPayment(credentials GuideStripeCredentials, paymentIntentID string, amount *int64) (*stripe.Refund, error) {
	stripe.Key = credentials.SecretKey

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	if amount != nil {
		params.Amount = stripe.Int64(*amount)
	}

	if credentials.AccountID != "" {
		params.SetStripeAccount(credentials.AccountID)
	}

	refund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return refund, nil
}

func (s *stripeService) VerifyWebhookSignature(payload []byte, signature string, secret string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	return &event, nil
}

func (s *stripeService) ProcessWebhookEvent(event *stripe.Event) (*WebhookEventData, error) {
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return nil, fmt.Errorf("error parsing payment_intent.succeeded event: %v", err)
		}

		return &WebhookEventData{
			EventType:         string(event.Type),
			PaymentIntentID:   paymentIntent.ID,
			PaymentIntentData: &paymentIntent,
		}, nil

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return nil, fmt.Errorf("error parsing payment_intent.payment_failed event: %v", err)
		}

		return &WebhookEventData{
			EventType:         string(event.Type),
			PaymentIntentID:   paymentIntent.ID,
			PaymentIntentData: &paymentIntent,
		}, nil

	case "payment_intent.canceled":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return nil, fmt.Errorf("error parsing payment_intent.canceled event: %v", err)
		}

		return &WebhookEventData{
			EventType:         string(event.Type),
			PaymentIntentID:   paymentIntent.ID,
			PaymentIntentData: &paymentIntent,
		}, nil

	case "payment_intent.requires_action":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return nil, fmt.Errorf("error parsing payment_intent.requires_action event: %v", err)
		}

		return &WebhookEventData{
			EventType:         string(event.Type),
			PaymentIntentID:   paymentIntent.ID,
			PaymentIntentData: &paymentIntent,
		}, nil

	default:
		// Ignore unsupported event types
		return nil, fmt.Errorf("unsupported event type: %s", event.Type)
	}
}
