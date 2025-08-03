package model

import (
	"errors"
	"net/http"
	"time"
)

type RamblePrice struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type Ramble struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Title                  string        `json:"title"`
	Status                 string        `json:"status"`
	Description            *string       `json:"description,omitempty"`
	Type                   string        `json:"type"`
	Date                   *time.Time    `json:"date,omitempty"`
	Location               *string       `json:"location,omitempty"`
	MeetingPoint           *string       `json:"meeting_point,omitempty"`
	MaxParticipants        *int          `json:"max_participants,omitempty"`
	Prices                 []RamblePrice `json:"prices"`
	Difficulty             string        `json:"difficulty"`
	EstimatedDuration      *string       `json:"estimated_duration,omitempty"` // store as HH:MM:SS
	EquipmentNeeded        *string       `json:"equipment_needed,omitempty"`
	Prerequisites          *string       `json:"prerequisites,omitempty"`
	CoverImageURL          *string       `json:"cover_image_url,omitempty"`
	AdditionalDocumentsURL *string       `json:"additional_documents_url,omitempty"`
} // @name Ramble

type RambleCreatePayload struct {
	Title                  *string       `json:"title" binding:"required"`
	Description            *string       `json:"description,omitempty"`
	Type                   *string       `json:"type" binding:"required"`
	Date                   *time.Time    `json:"date,omitempty"`
	Location               *string       `json:"location,omitempty"`
	MeetingPoint           *string       `json:"meeting_point,omitempty"`
	MaxParticipants        *int          `json:"max_participants,omitempty"`
	Difficulty             *string       `json:"difficulty" binding:"required"`
	EstimatedDuration      *string       `json:"estimated_duration,omitempty"` // store as HH:MM:SS
	EquipmentNeeded        *string       `json:"equipment_needed,omitempty"`
	Prerequisites          *string       `json:"prerequisites,omitempty"`
	CoverImageURL          *string       `json:"cover_image_url,omitempty"`
	AdditionalDocumentsURL *string       `json:"additional_documents_url,omitempty"`
	Prices                 []RamblePrice `json:"prices" binding:"required"`
} // @name RambleCreatePayload

func (rb *RambleCreatePayload) Bind(r *http.Request) error {
	if rb.Title == nil {
		return errors.New("title is required")
	}

	if rb.Type == nil {
		return errors.New("type is required")
	}

	if rb.Difficulty == nil {
		return errors.New("difficulty is required")
	}

	if len(rb.Prices) == 0 {
		return errors.New("at least one price is required")
	}

	return nil
}
