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
	CoverImage             *string       `json:"cover_image,omitempty"`
	AdditionalDocumentsURL *string       `json:"additional_documents_url,omitempty"`
	IsCancelled            bool          `json:"is_cancelled"`
	CancellationDate       *time.Time    `json:"cancellation_date,omitempty"`
	CancellationReason     *string       `json:"cancellation_reason,omitempty"`
	Guides                 []Guide       `json:"guides,omitempty"`
	PlacesLeft             *int          `json:"places_left,omitempty"`
} // @name Ramble

type RambleCreatePayload struct {
	Title                    *string       `json:"title" binding:"required"`
	Description              *string       `json:"description,omitempty"`
	Type                     *string       `json:"type" binding:"required"`
	Date                     *time.Time    `json:"date,omitempty"`
	Location                 *string       `json:"location,omitempty"`
	MeetingPoint             *string       `json:"meeting_point,omitempty"`
	MaxParticipants          *int          `json:"max_participants,omitempty"`
	Difficulty               *string       `json:"difficulty" binding:"required"`
	EstimatedDuration        *string       `json:"estimated_duration,omitempty"` // store as HH:MM:SS
	EquipmentNeeded          *string       `json:"equipment_needed,omitempty"`
	Prerequisites            *string       `json:"prerequisites,omitempty"`
	CoverImageBase64         *string       `json:"cover_image_base64,omitempty"`         // Base64 encoded cover image
	AdditionalDocumentBase64 *string       `json:"additional_document_base64,omitempty"` // Base64 encoded document
	GuideIDs                 []uint        `json:"guide_ids,omitempty"`                  // Array of guide IDs
	Prices                   []RamblePrice `json:"prices" binding:"required"`
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

type RambleCancelPayload struct {
	Reason *string `json:"reason" binding:"required"`
} // @name RambleCancelPayload

func (rc *RambleCancelPayload) Bind(r *http.Request) error {
	if rc.Reason == nil || *rc.Reason == "" {
		return errors.New("cancellation reason is required")
	}

	return nil
}
