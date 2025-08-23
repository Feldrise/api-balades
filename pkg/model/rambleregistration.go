package model

import (
	"errors"
	"net/http"
	"time"
)

type RambleRegistration struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	Email                string     `json:"email"`
	Phone                *string    `json:"phone,omitempty"`
	Status               string     `json:"status"`
	RegistrationDate     time.Time  `json:"registration_date"`
	ConfirmationDate     *time.Time `json:"confirmation_date,omitempty"`
	ConfirmationDeadline *time.Time `json:"confirmation_deadline,omitempty"`
	CancellationDate     *time.Time `json:"cancellation_date,omitempty"`
	CancellationReason   *string    `json:"cancellation_reason,omitempty"`

	RambleID uint  `json:"ramble_id"`
	UserID   *uint `json:"user_id,omitempty"`
	GroupID  *uint `json:"group_id,omitempty"`
} // @name RambleRegistration

type RambleRegistrationGroup struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PrimaryEmail         string                `json:"primary_email"`
	Status               string                `json:"status"`
	RegistrationDate     time.Time             `json:"registration_date"`
	ConfirmationDate     *time.Time            `json:"confirmation_date,omitempty"`
	ConfirmationDeadline *time.Time            `json:"confirmation_deadline,omitempty"`
	CancellationDate     *time.Time            `json:"cancellation_date,omitempty"`
	CancellationReason   *string               `json:"cancellation_reason,omitempty"`
	RambleID             uint                  `json:"ramble_id"`
	Registrations        []*RambleRegistration `json:"registrations"`
	ParticipantCount     int                   `json:"participant_count"`
} // @name RambleRegistrationGroup

type ParticipantInfo struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
} // @name ParticipantInfo

type RambleRegistrationCreatePayload struct {
	RambleID     uint              `json:"ramble_id"`
	Participants []ParticipantInfo `json:"participants"`
	PrimaryEmail *string           `json:"primary_email,omitempty"` // Optional for single registration
} // @name RambleRegistrationCreatePayload

func (payload *RambleRegistrationCreatePayload) Bind(r *http.Request) error {
	if payload.RambleID == 0 {
		return errors.New("ramble ID is required")
	}

	if len(payload.Participants) == 0 {
		return errors.New("at least one participant is required")
	}

	if len(payload.Participants) > 10 {
		return errors.New("maximum 10 participants allowed per registration")
	}

	// Validate each participant
	for i, participant := range payload.Participants {
		if participant.FirstName == "" {
			return errors.New("participant " + string(rune(i+1)) + ": first name is required")
		}
		if participant.LastName == "" {
			return errors.New("participant " + string(rune(i+1)) + ": last name is required")
		}
		if participant.Email == "" {
			return errors.New("participant " + string(rune(i+1)) + ": email is required")
		}
	}

	// For single registration, use participant's email as primary
	if len(payload.Participants) == 1 && payload.PrimaryEmail == nil {
		payload.PrimaryEmail = &payload.Participants[0].Email
	}

	// For group registration, primary email is required
	if len(payload.Participants) > 1 && (payload.PrimaryEmail == nil || *payload.PrimaryEmail == "") {
		return errors.New("primary email is required for group registrations")
	}

	return nil
}

type RambleRegistrationConfirmPayload struct {
	Confirmed bool `json:"confirmed"`
} // @name RambleRegistrationConfirmPayload

func (payload *RambleRegistrationConfirmPayload) Bind(r *http.Request) error {
	return nil
}

type RambleRegistrationCancelPayload struct {
	Reason *string `json:"reason,omitempty"`
} // @name RambleRegistrationCancelPayload

func (payload *RambleRegistrationCancelPayload) Bind(r *http.Request) error {
	return nil
}
