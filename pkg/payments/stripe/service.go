package stripe

import (
	"context"
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

// StripeService calls Stripe. The calls taking a context give up once it ends:
// Stripe may still have done what was asked.
type StripeService interface {
	// CreatePaymentIntent creates a payment intent. A creation with the
	// idempotency key of a previous one gives back the same payment intent.
	CreatePaymentIntent(ctx context.Context, credentials GuideStripeCredentials, amount int64, currency string, metadata map[string]string, idempotencyKey string) (*stripe.PaymentIntent, error)
	ConfirmPayment(credentials GuideStripeCredentials, paymentIntentID string) (*stripe.PaymentIntent, error)
	// RetrievePaymentIntent returns the payment intent with its client secret.
	RetrievePaymentIntent(ctx context.Context, credentials GuideStripeCredentials, paymentIntentID string) (*stripe.PaymentIntent, error)
	// CancelPaymentIntent cancels a payment intent not paid yet, so that it can no longer be paid.
	CancelPaymentIntent(ctx context.Context, credentials GuideStripeCredentials, paymentIntentID string) error
	// RefundPayment refunds amount cents of the payment intent, and returns the
	// amount refunded on it so far, all refunds included, as Stripe knows it.
	// A refund with the idempotency key of a previous one refunds nothing more.
	RefundPayment(ctx context.Context, credentials GuideStripeCredentials, paymentIntentID string, amount int64, idempotencyKey string) (int64, error)
	VerifyWebhookSignature(payload []byte, signature string, secret string) (*stripe.Event, error)
	ProcessWebhookEvent(event *stripe.Event) (*WebhookEventData, error)
}

type WebhookEventData struct {
	EventType         string
	PaymentIntentID   string
	PaymentIntentData *stripe.PaymentIntent
	// Metadata is the metadata of the payment intent, or of the charge.
	Metadata map[string]string
	// AmountRefunded is, for a charge.refunded event, the amount in cents
	// refunded on the charge so far, all refunds included.
	AmountRefunded int64
}

type stripeService struct{}

func NewStripeService() StripeService {
	return &stripeService{}
}

func (s *stripeService) CreatePaymentIntent(ctx context.Context, credentials GuideStripeCredentials, amount int64, currency string, metadata map[string]string, idempotencyKey string) (*stripe.PaymentIntent, error) {
	// Set the API key for this specific guide
	stripe.Key = credentials.SecretKey

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Params: stripe.Params{
			Context:  ctx,
			Metadata: metadata,
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	params.SetIdempotencyKey(idempotencyKey)

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

func (s *stripeService) RetrievePaymentIntent(ctx context.Context, credentials GuideStripeCredentials, paymentIntentID string) (*stripe.PaymentIntent, error) {
	stripe.Key = credentials.SecretKey

	params := &stripe.PaymentIntentParams{Params: stripe.Params{Context: ctx}}

	if credentials.AccountID != "" {
		params.SetStripeAccount(credentials.AccountID)
	}

	pi, err := paymentintent.Get(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	return pi, nil
}

func (s *stripeService) CancelPaymentIntent(ctx context.Context, credentials GuideStripeCredentials, paymentIntentID string) error {
	stripe.Key = credentials.SecretKey

	params := &stripe.PaymentIntentCancelParams{Params: stripe.Params{Context: ctx}}

	if credentials.AccountID != "" {
		params.SetStripeAccount(credentials.AccountID)
	}

	if _, err := paymentintent.Cancel(paymentIntentID, params); err != nil {
		return fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	return nil
}

func (s *stripeService) RefundPayment(ctx context.Context, credentials GuideStripeCredentials, paymentIntentID string, amount int64, idempotencyKey string) (int64, error) {
	stripe.Key = credentials.SecretKey

	params := &stripe.RefundParams{
		Params:        stripe.Params{Context: ctx},
		PaymentIntent: stripe.String(paymentIntentID),
		Amount:        stripe.Int64(amount),
	}
	params.SetIdempotencyKey(idempotencyKey)

	if credentials.AccountID != "" {
		params.SetStripeAccount(credentials.AccountID)
	}

	if _, err := refund.New(params); err != nil {
		return 0, fmt.Errorf("failed to create refund: %w", err)
	}

	// The refund replayed for a reused idempotency key is as it was then: the
	// amount refunded is read afresh on the charge.
	retrieveParams := &stripe.PaymentIntentParams{Params: stripe.Params{Context: ctx}}
	retrieveParams.AddExpand("latest_charge")
	if credentials.AccountID != "" {
		retrieveParams.SetStripeAccount(credentials.AccountID)
	}
	pi, err := paymentintent.Get(paymentIntentID, retrieveParams)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve the refunded amount: %w", err)
	}
	if pi.LatestCharge == nil {
		return 0, fmt.Errorf("payment intent %s has no charge", paymentIntentID)
	}

	return pi.LatestCharge.AmountRefunded, nil
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
			Metadata:          paymentIntent.Metadata,
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
			Metadata:          paymentIntent.Metadata,
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
			Metadata:          paymentIntent.Metadata,
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
			Metadata:          paymentIntent.Metadata,
		}, nil

	case "charge.refunded":
		// Sent for every refund, made from the API or from the dashboard.
		var charge stripe.Charge
		err := json.Unmarshal(event.Data.Raw, &charge)
		if err != nil {
			return nil, fmt.Errorf("error parsing charge.refunded event: %v", err)
		}

		data := &WebhookEventData{
			EventType:      string(event.Type),
			Metadata:       charge.Metadata,
			AmountRefunded: charge.AmountRefunded,
		}
		if charge.PaymentIntent != nil {
			data.PaymentIntentID = charge.PaymentIntent.ID
		}
		return data, nil

	default:
		// Ignore unsupported event types
		return nil, fmt.Errorf("unsupported event type: %s", event.Type)
	}
}
