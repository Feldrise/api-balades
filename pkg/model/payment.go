package model

import (
	"errors"
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

	// Registration relationship
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
	Registration *RambleRegistration      `json:"registration,omitempty"`
	Group        *RambleRegistrationGroup `json:"group,omitempty"`
	Guide        Guide                    `json:"guide"`
} // @name Payment

type PriceSelection struct {
	PriceLabel string `json:"price_label" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
} // @name PriceSelection

type PaymentCreatePayload struct {
	RegistrationID  *uint            `json:"registration_id,omitempty"`
	GroupID         *uint            `json:"group_id,omitempty"`
	PriceLabel      string           `json:"price_label,omitempty"`      // Deprecated: use PriceSelections for groups
	PriceSelections []PriceSelection `json:"price_selections,omitempty"` // For group payments with multiple prices
	PayerEmail      string           `json:"payer_email" binding:"required"`
	PayerName       string           `json:"payer_name" binding:"required"`
	ReturnURL       string           `json:"return_url" binding:"required"`
} // @name PaymentCreatePayload

func (p *PaymentCreatePayload) Bind(r *http.Request) error {
	if p.PayerEmail == "" {
		return errors.New("payer_email is required")
	}

	if p.PayerName == "" {
		return errors.New("payer_name is required")
	}

	if p.ReturnURL == "" {
		return errors.New("return_url is required")
	}

	if p.RegistrationID == nil && p.GroupID == nil {
		return errors.New("either registration_id or group_id is required")
	}

	if p.RegistrationID != nil && p.GroupID != nil {
		return errors.New("cannot specify both registration_id and group_id")
	}

	// For individual registrations, require single price_label
	if p.RegistrationID != nil && p.PriceLabel == "" {
		return errors.New("price_label is required for individual registration")
	}

	// For group registrations, require either price_label (legacy) or price_selections
	if p.GroupID != nil && p.PriceLabel == "" && len(p.PriceSelections) == 0 {
		return errors.New("either price_label or price_selections is required for group payment")
	}

	// Validate price_selections if provided
	if len(p.PriceSelections) > 0 {
		for i, selection := range p.PriceSelections {
			if selection.PriceLabel == "" {
				return errors.New("price_label is required for all price_selections")
			}
			if selection.Quantity < 1 {
				return errors.New("quantity must be at least 1 for all price_selections")
			}
			// Check for duplicate price labels
			for j := i + 1; j < len(p.PriceSelections); j++ {
				if p.PriceSelections[j].PriceLabel == selection.PriceLabel {
					return errors.New("duplicate price_label in price_selections")
				}
			}
		}
	}

	return nil
}

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
