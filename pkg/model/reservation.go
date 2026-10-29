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
	// EstimatedAmount is the amount in cents computed from the participants' prices.
	EstimatedAmount int64                `json:"estimated_amount"`
	Payments        []ReservationPayment `json:"payments"`

	Ramble *RambleRegistrationSummary `json:"ramble,omitempty"`
} // @name Reservation

type Participant struct {
	ID         uint   `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PriceLabel string `json:"price_label"`
	// PriceAmount is the amount in euros of the participant's price.
	PriceAmount float64 `json:"price_amount"`
	IsBooker    bool    `json:"is_booker"`
} // @name Participant

type ReservationPayment struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	// Amount is in cents.
	Amount       int64      `json:"amount"`
	Currency     string     `json:"currency"`
	Status       string     `json:"status"` // pending, succeeded, failed, cancelled, refunded
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
		participant.FirstName = strings.TrimSpace(participant.FirstName)
		participant.LastName = strings.TrimSpace(participant.LastName)

		if participant.FirstName == "" || participant.LastName == "" {
			return errors.New("each participant needs a first name and a last name")
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
