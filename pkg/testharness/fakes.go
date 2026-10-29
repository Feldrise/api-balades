package testharness

import (
	"errors"
	"fmt"
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

	mu   sync.Mutex
	sent []SentEmail
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
}

// Refund is a refund recorded by FakeStripeService.
type Refund struct {
	PaymentIntentID string
	Amount          *int64
}

// FakeStripeService records payment intents and refunds instead of calling Stripe.
type FakeStripeService struct {
	mu             sync.Mutex
	paymentIntents []CreatedPaymentIntent
	cancelled      []string
	refunds        []Refund
}

func NewFakeStripeService() *FakeStripeService {
	return &FakeStripeService{}
}

func (s *FakeStripeService) CreatePaymentIntent(credentials stripe.GuideStripeCredentials, amount int64, currency string, metadata map[string]string) (*stripego.PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("pi_test_%d", len(s.paymentIntents)+1)
	s.paymentIntents = append(s.paymentIntents, CreatedPaymentIntent{
		ID:          id,
		Credentials: credentials,
		Amount:      amount,
		Currency:    currency,
		Metadata:    metadata,
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

func (s *FakeStripeService) RetrievePaymentIntent(credentials stripe.GuideStripeCredentials, paymentIntentID string) (*stripego.PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, intent := range s.paymentIntents {
		if intent.ID == paymentIntentID {
			return &stripego.PaymentIntent{
				ID:           intent.ID,
				Amount:       intent.Amount,
				Currency:     stripego.Currency(intent.Currency),
				ClientSecret: intent.ID + "_secret",
				Metadata:     intent.Metadata,
			}, nil
		}
	}
	return nil, fmt.Errorf("no such payment intent %q", paymentIntentID)
}

func (s *FakeStripeService) CancelPaymentIntent(credentials stripe.GuideStripeCredentials, paymentIntentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancelled = append(s.cancelled, paymentIntentID)
	return nil
}

func (s *FakeStripeService) RefundPayment(credentials stripe.GuideStripeCredentials, paymentIntentID string, amount *int64) (*stripego.Refund, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refunds = append(s.refunds, Refund{PaymentIntentID: paymentIntentID, Amount: amount})

	return &stripego.Refund{ID: fmt.Sprintf("re_test_%d", len(s.refunds))}, nil
}

func (s *FakeStripeService) VerifyWebhookSignature(payload []byte, signature string, secret string) (*stripego.Event, error) {
	return nil, errors.New("webhooks are not supported by the fake Stripe service")
}

func (s *FakeStripeService) ProcessWebhookEvent(event *stripego.Event) (*stripe.WebhookEventData, error) {
	return nil, errors.New("webhooks are not supported by the fake Stripe service")
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
