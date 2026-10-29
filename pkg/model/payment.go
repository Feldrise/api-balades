package model

import (
	"net/http"
	"time"
)

type Payment struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Stripe identifiers
	StripePaymentIntentID string  `json:"stripe_payment_intent_id"`
	StripeChargeID        *string `json:"stripe_charge_id,omitempty"`

	// Payment details
	Amount        int64   `json:"amount"`                   // Amount in cents
	Currency      string  `json:"currency"`                 // Currency code
	Status        string  `json:"status"`                   // pending, succeeded, failed, cancelled, refunded
	PaymentMethod string  `json:"payment_method"`           // card, sepa, etc.
	FailureReason *string `json:"failure_reason,omitempty"` // Reason for payment failure

	// Reservation relationship
	ReservationID *uint `json:"reservation_id,omitempty"`

	// Legacy registration relationship
	RegistrationID *uint `json:"registration_id,omitempty"`
	GroupID        *uint `json:"group_id,omitempty"`

	// Payer information
	PayerEmail string `json:"payer_email"`
	PayerName  string `json:"payer_name"`

	// Guide receiving payment
	GuideID uint `json:"guide_id"`

	// Timestamps
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	RefundedAt   *time.Time `json:"refunded_at,omitempty"`
	RefundAmount *int64     `json:"refund_amount,omitempty"` // Amount refunded in cents

	// Foreign Objects
	Guide Guide `json:"guide"`
} // @name Payment

type PaymentRefundPayload struct {
	Amount *int64  `json:"amount,omitempty"` // Amount to refund in cents, if nil refunds full amount
	Reason *string `json:"reason,omitempty"` // Reason for refund
} // @name PaymentRefundPayload

func (p *PaymentRefundPayload) Bind(r *http.Request) error {
	// No required fields for refund
	return nil
}

type PaymentResponse struct {
	Payment        Payment `json:"payment"`
	ClientSecret   string  `json:"client_secret"`   // Stripe client secret for frontend
	PublishableKey string  `json:"publishable_key"` // Stripe publishable key for frontend
} // @name PaymentResponse
