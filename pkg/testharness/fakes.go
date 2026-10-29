package testharness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"feldrise.com/balade/pkg/notifications/email"
	"feldrise.com/balade/pkg/payments/stripe"
	stripego "github.com/stripe/stripe-go/v83"
)

// SentEmail is an email recorded by FakeEmailService.
type SentEmail struct {
	To       string
	Subject  string
	Template string
	Data     any
	// Body is the email rendered with the real template.
	Body string
}

// FakeEmailService records emails instead of sending them.
type FakeEmailService struct {
	templatesPath string

	mu      sync.Mutex
	sent    []SentEmail
	failing bool
}

// SetFailing makes the sending of emails fail, as when the SMTP server is down.
func (s *FakeEmailService) SetFailing(failing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failing = failing
}

func NewFakeEmailService(templatesPath string) *FakeEmailService {
	return &FakeEmailService{templatesPath: templatesPath}
}

func (s *FakeEmailService) Send(to string, subject string, templateName string, data interface{}, attachments ...string) error {
	body, err := email.Render(s.templatesPath, templateName, data)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.failing {
		return errors.New("the SMTP server is down")
	}

	s.sent = append(s.sent, SentEmail{
		To:       to,
		Subject:  subject,
		Template: templateName,
		Data:     data,
		Body:     body,
	})

	return nil
}

// Sent returns the emails sent so far.
func (s *FakeEmailService) Sent() []SentEmail {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]SentEmail(nil), s.sent...)
}

// SentWithTemplate returns the emails sent so far with the given template.
func (s *FakeEmailService) SentWithTemplate(templateName string) []SentEmail {
	var result []SentEmail
	for _, sent := range s.Sent() {
		if sent.Template == templateName {
			result = append(result, sent)
		}
	}
	return result
}

// CreatedPaymentIntent is a payment intent recorded by FakeStripeService.
type CreatedPaymentIntent struct {
	ID          string
	Credentials stripe.GuideStripeCredentials
	Amount      int64
	Currency    string
	Metadata    map[string]string
	// IdempotencyKey is the key the payment intent was created with: a creation
	// with the same key gives back this payment intent, as Stripe does.
	IdempotencyKey string
}

// Refund is a refund recorded by FakeStripeService.
type Refund struct {
	PaymentIntentID string
	Amount          int64
	// IdempotencyKey is the key the refund was made with: a refund with the
	// same key refunds nothing more, as on Stripe.
	IdempotencyKey string
}

// FakeStripeService records payment intents and refunds instead of calling Stripe.
type FakeStripeService struct {
	mu             sync.Mutex
	paymentIntents []CreatedPaymentIntent
	cancelled      []string
	refunds        []Refund
	unavailable    bool
	cancelFailing  bool
	creationDelay  time.Duration
	refundDelay    time.Duration
}

// SetCancellationFailing makes the cancellation of payment intents fail, as
// when Stripe is unreachable.
func (s *FakeStripeService) SetCancellationFailing(failing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancelFailing = failing
}

// SetCreationDelay makes the creation of payment intents take delay, as a
// slow network does, so that simultaneous requests overlap.
func (s *FakeStripeService) SetCreationDelay(delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.creationDelay = delay
}

// SetRefundDelay makes refunds take delay, as a slow network does, so that
// simultaneous requests overlap. A refund whose context ends first fails, but
// is made all the same, as Stripe does once it received the request.
func (s *FakeStripeService) SetRefundDelay(delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refundDelay = delay
}

// SetUnavailable makes the creation of payment intents fail, as when Stripe is unreachable.
func (s *FakeStripeService) SetUnavailable(unavailable bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.unavailable = unavailable
}

func NewFakeStripeService() *FakeStripeService {
	return &FakeStripeService{}
}

// respond waits delay, as a slow network does, and fails if ctx ends first.
func respond(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("stripe did not answer in time: %w", ctx.Err())
	}
}

func (s *FakeStripeService) CreatePaymentIntent(ctx context.Context, credentials stripe.GuideStripeCredentials, amount int64, currency string, metadata map[string]string, idempotencyKey string) (*stripego.PaymentIntent, error) {
	s.mu.Lock()
	delay := s.creationDelay
	s.mu.Unlock()
	if err := respond(ctx, delay); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.unavailable {
		return nil, errors.New("stripe is unavailable")
	}

	for _, intent := range s.paymentIntents {
		if idempotencyKey != "" && intent.IdempotencyKey == idempotencyKey {
			// Stripe refuses to reuse a key with other parameters.
			if intent.Amount != amount || intent.Currency != currency || !maps.Equal(intent.Metadata, metadata) {
				return nil, fmt.Errorf("idempotency key %q already used with other parameters", idempotencyKey)
			}
			return &stripego.PaymentIntent{
				ID:           intent.ID,
				Amount:       intent.Amount,
				Currency:     stripego.Currency(intent.Currency),
				ClientSecret: intent.ID + "_secret",
				Status:       stripego.PaymentIntentStatusRequiresPaymentMethod,
				Metadata:     intent.Metadata,
			}, nil
		}
	}

	id := fmt.Sprintf("pi_test_%d", len(s.paymentIntents)+1)
	s.paymentIntents = append(s.paymentIntents, CreatedPaymentIntent{
		ID:             id,
		Credentials:    credentials,
		Amount:         amount,
		Currency:       currency,
		Metadata:       metadata,
		IdempotencyKey: idempotencyKey,
	})

	return &stripego.PaymentIntent{
		ID:           id,
		Amount:       amount,
		Currency:     stripego.Currency(currency),
		ClientSecret: id + "_secret",
		Status:       stripego.PaymentIntentStatusRequiresPaymentMethod,
		Metadata:     metadata,
	}, nil
}

func (s *FakeStripeService) ConfirmPayment(credentials stripe.GuideStripeCredentials, paymentIntentID string) (*stripego.PaymentIntent, error) {
	return &stripego.PaymentIntent{ID: paymentIntentID, Status: stripego.PaymentIntentStatusSucceeded}, nil
}

func (s *FakeStripeService) RetrievePaymentIntent(ctx context.Context, credentials stripe.GuideStripeCredentials, paymentIntentID string) (*stripego.PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, intent := range s.paymentIntents {
		if intent.ID == paymentIntentID {
			status := stripego.PaymentIntentStatusRequiresPaymentMethod
			if slices.Contains(s.cancelled, paymentIntentID) {
				status = stripego.PaymentIntentStatusCanceled
			}
			return &stripego.PaymentIntent{
				ID:           intent.ID,
				Amount:       intent.Amount,
				Currency:     stripego.Currency(intent.Currency),
				ClientSecret: intent.ID + "_secret",
				Status:       status,
				Metadata:     intent.Metadata,
			}, nil
		}
	}
	return nil, fmt.Errorf("no such payment intent %q", paymentIntentID)
}

// CancelPaymentIntent cancels the payment intent, and fails as Stripe does
// when it is cancelled already.
func (s *FakeStripeService) CancelPaymentIntent(ctx context.Context, credentials stripe.GuideStripeCredentials, paymentIntentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelFailing {
		return errors.New("stripe is unavailable")
	}
	if slices.Contains(s.cancelled, paymentIntentID) {
		return fmt.Errorf("payment intent %q is cancelled already", paymentIntentID)
	}
	s.cancelled = append(s.cancelled, paymentIntentID)
	return nil
}

func (s *FakeStripeService) RefundPayment(ctx context.Context, credentials stripe.GuideStripeCredentials, paymentIntentID string, amount int64, idempotencyKey string) (int64, error) {
	s.mu.Lock()
	delay := s.refundDelay
	if !slices.ContainsFunc(s.refunds, func(r Refund) bool { return r.IdempotencyKey == idempotencyKey }) {
		s.refunds = append(s.refunds, Refund{PaymentIntentID: paymentIntentID, Amount: amount, IdempotencyKey: idempotencyKey})
	}
	var refunded int64
	for _, r := range s.refunds {
		if r.PaymentIntentID == paymentIntentID {
			refunded += r.Amount
		}
	}
	s.mu.Unlock()

	if err := respond(ctx, delay); err != nil {
		return 0, err
	}
	return refunded, nil
}

// FakeWebhookSignature is the signature the fake Stripe service accepts for
// webhooks signed with secret.
func FakeWebhookSignature(secret string) string {
	return "fake_signature_" + secret
}

// VerifyWebhookSignature accepts the payload of an event signed with FakeWebhookSignature.
func (s *FakeStripeService) VerifyWebhookSignature(payload []byte, signature string, secret string) (*stripego.Event, error) {
	if signature != FakeWebhookSignature(secret) {
		return nil, errors.New("invalid webhook signature")
	}

	var event stripego.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("invalid webhook payload: %w", err)
	}
	return &event, nil
}

// ProcessWebhookEvent reads the event as the real Stripe service does.
func (s *FakeStripeService) ProcessWebhookEvent(event *stripego.Event) (*stripe.WebhookEventData, error) {
	return stripe.NewStripeService().ProcessWebhookEvent(event)
}

// PaymentIntents returns the payment intents created so far.
func (s *FakeStripeService) PaymentIntents() []CreatedPaymentIntent {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]CreatedPaymentIntent(nil), s.paymentIntents...)
}

// CancelledPaymentIntents returns the ids of the payment intents cancelled so far.
func (s *FakeStripeService) CancelledPaymentIntents() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]string(nil), s.cancelled...)
}

// Refunds returns the refunds made so far.
func (s *FakeStripeService) Refunds() []Refund {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]Refund(nil), s.refunds...)
}

// FakeClock is a clock set by the test. It starts at the current time.
type FakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewFakeClock() *FakeClock {
	return &FakeClock{now: time.Now().Truncate(time.Second)}
}

func (c *FakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *FakeClock) Set(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = now
}

func (c *FakeClock) Advance(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(duration)
}
