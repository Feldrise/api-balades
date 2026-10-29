package model

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

type Reservation struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RambleID     uint   `json:"ramble_id"`
	UserID       uint   `json:"user_id"`
	ContactEmail string `json:"contact_email"`

	Status               string     `json:"status"` // pending, confirmed, waiting_list, cancelled
	ConfirmationDeadline *time.Time `json:"confirmation_deadline,omitempty"`
	ConfirmationDate     *time.Time `json:"confirmation_date,omitempty"`
	CancellationDate     *time.Time `json:"cancellation_date,omitempty"`
	CancellationReason   *string    `json:"cancellation_reason,omitempty"`

	Participants []Participant `json:"participants"`
	// EstimatedAmount is the amount in cents computed from the participants' known prices.
	EstimatedAmount int64 `json:"estimated_amount"`
	// HasMissingPrices tells that a participant's price is unknown, so that
	// EstimatedAmount is incomplete.
	HasMissingPrices bool                 `json:"has_missing_prices"`
	Payments         []ReservationPayment `json:"payments"`
	// HasSuccessfulPayment tells that an amount paid is kept (a payment succeeded
	// and not entirely refunded): after a removal or a cancellation, the booker
	// is invited to contact the guide for a refund.
	HasSuccessfulPayment bool `json:"has_successful_payment"`
	// AmountPaid is the amount in cents paid and kept: the payments succeeded minus their refunds.
	AmountPaid int64 `json:"amount_paid"`
	// AmountDue is the amount in cents left to pay: EstimatedAmount minus AmountPaid, never below zero.
	AmountDue int64 `json:"amount_due"`

	Ramble *ReservationRamble `json:"ramble,omitempty"`
} // @name Reservation

// ReservationWithPayment is a reservation with the payment the booker has to
// make right away, null when there is nothing to pay online.
type ReservationWithPayment struct {
	*Reservation
	Payment *PaymentResponse `json:"payment"`
} // @name ReservationWithPayment

// ReservationRamble is the ramble a reservation is for.
type ReservationRamble struct {
	Title    string     `json:"title"`
	Date     *time.Time `json:"date"`
	Location *string    `json:"location"`
} // @name ReservationRamble

type Participant struct {
	ID         uint   `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PriceLabel string `json:"price_label"`
	// PriceAmount is the amount in euros of the participant's price, 0 when the price is missing.
	PriceAmount float64 `json:"price_amount"`
	// PriceMissing tells that the ramble has prices but the participant's is
	// unknown, as for participants migrated from a ramble with several prices.
	PriceMissing bool `json:"price_missing"`
	IsBooker     bool `json:"is_booker"`
} // @name Participant

type ReservationPayment struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	// Amount is in cents.
	Amount       int64      `json:"amount"`
	Currency     string     `json:"currency"`
	Status       string     `json:"status"` // pending, requires_action, succeeded, failed, cancelled, partially_refunded, refunded
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	RefundedAt   *time.Time `json:"refunded_at,omitempty"`
	RefundAmount *int64     `json:"refund_amount,omitempty"`
} // @name ReservationPayment

type ParticipantPayload struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PriceLabel string `json:"price_label"`
	IsBooker   bool   `json:"is_booker"`
} // @name ParticipantPayload

type ReservationCreatePayload struct {
	RambleID     uint                 `json:"ramble_id"`
	ContactEmail string               `json:"contact_email"`
	ContactPhone *string              `json:"contact_phone,omitempty"`
	Participants []ParticipantPayload `json:"participants"`
} // @name ReservationCreatePayload

func (payload *ReservationCreatePayload) Bind(r *http.Request) error {
	payload.ContactEmail = strings.TrimSpace(payload.ContactEmail)

	if payload.RambleID == 0 {
		return errors.New("ramble_id is required")
	}

	if payload.ContactEmail == "" {
		return errors.New("contact_email is required")
	}

	if len(payload.Participants) == 0 {
		return errors.New("at least one participant is required")
	}

	bookers := 0
	for i := range payload.Participants {
		participant := &payload.Participants[i]
		if err := participant.normalize(); err != nil {
			return err
		}
		if participant.IsBooker {
			bookers++
		}
	}

	if bookers > 1 {
		return errors.New("only one participant can be the booker")
	}

	return nil
}

type ParticipantsAddPayload struct {
	Participants []ParticipantPayload `json:"participants"`
} // @name ParticipantsAddPayload

func (payload *ParticipantsAddPayload) Bind(r *http.Request) error {
	if len(payload.Participants) == 0 {
		return errors.New("at least one participant is required")
	}

	for i := range payload.Participants {
		if err := payload.Participants[i].normalize(); err != nil {
			return err
		}
	}

	return nil
}

type ParticipantUpdatePayload struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PriceLabel string `json:"price_label"`
} // @name ParticipantUpdatePayload

func (payload *ParticipantUpdatePayload) Bind(r *http.Request) error {
	payload.FirstName = strings.TrimSpace(payload.FirstName)
	payload.LastName = strings.TrimSpace(payload.LastName)

	if payload.FirstName == "" || payload.LastName == "" {
		return errors.New("a participant needs a first name and a last name")
	}

	return nil
}

func (participant *ParticipantPayload) normalize() error {
	participant.FirstName = strings.TrimSpace(participant.FirstName)
	participant.LastName = strings.TrimSpace(participant.LastName)

	if participant.FirstName == "" || participant.LastName == "" {
		return errors.New("each participant needs a first name and a last name")
	}

	return nil
}
