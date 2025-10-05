package model

import (
	"errors"
	"net/http"
)

// Admin-specific payloads for registration management

type AdminRegistrationUpdatePayload struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Status    *string `json:"status,omitempty"`
	Notes     *string `json:"notes,omitempty"`
} // @name AdminRegistrationUpdatePayload

func (payload *AdminRegistrationUpdatePayload) Bind(r *http.Request) error {
	if payload.Status != nil {
		validStatuses := map[string]bool{
			"pending":      true,
			"confirmed":    true,
			"waiting_list": true,
			"cancelled":    true,
		}
		if !validStatuses[*payload.Status] {
			return errors.New("invalid status. Valid statuses are: pending, confirmed, waiting_list, cancelled")
		}
	}

	if payload.Email != nil && *payload.Email == "" {
		return errors.New("email cannot be empty")
	}

	if payload.FirstName != nil && *payload.FirstName == "" {
		return errors.New("first name cannot be empty")
	}

	if payload.LastName != nil && *payload.LastName == "" {
		return errors.New("last name cannot be empty")
	}

	return nil
}

type AdminRegistrationStatusUpdatePayload struct {
	Status    string  `json:"status"`
	Reason    *string `json:"reason,omitempty"`
	SendEmail *bool   `json:"send_email,omitempty"` // Whether to send notification email
} // @name AdminRegistrationStatusUpdatePayload

func (payload *AdminRegistrationStatusUpdatePayload) Bind(r *http.Request) error {
	if payload.Status == "" {
		return errors.New("status is required")
	}

	validStatuses := map[string]bool{
		"pending":      true,
		"confirmed":    true,
		"waiting_list": true,
		"cancelled":    true,
	}

	if !validStatuses[payload.Status] {
		return errors.New("invalid status. Valid statuses are: pending, confirmed, waiting_list, cancelled")
	}

	return nil
}

type BulkRegistrationActionPayload struct {
	RegistrationIDs []uint  `json:"registration_ids"`
	Action          string  `json:"action"` // "confirm", "cancel", "move_to_waiting", "delete"
	Reason          *string `json:"reason,omitempty"`
	SendEmail       *bool   `json:"send_email,omitempty"` // Whether to send notification emails
} // @name BulkRegistrationActionPayload

func (payload *BulkRegistrationActionPayload) Bind(r *http.Request) error {
	if len(payload.RegistrationIDs) == 0 {
		return errors.New("at least one registration ID is required")
	}

	if len(payload.RegistrationIDs) > 100 {
		return errors.New("maximum 100 registrations can be processed at once")
	}

	if payload.Action == "" {
		return errors.New("action is required")
	}

	validActions := map[string]bool{
		"confirm":         true,
		"cancel":          true,
		"move_to_waiting": true,
		"delete":          true,
	}

	if !validActions[payload.Action] {
		return errors.New("invalid action. Valid actions are: confirm, cancel, move_to_waiting, delete")
	}

	return nil
}

type AdminRegistrationFilterPayload struct {
	RambleID    *uint    `json:"ramble_id,omitempty"`
	UserID      *uint    `json:"user_id,omitempty"`
	Email       *string  `json:"email,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Statuses    []string `json:"statuses,omitempty"`
	DateFrom    *string  `json:"date_from,omitempty"` // YYYY-MM-DD format
	DateTo      *string  `json:"date_to,omitempty"`   // YYYY-MM-DD format
	Search      *string  `json:"search,omitempty"`
	RambleTitle *string  `json:"ramble_title,omitempty"`
	Page        *int     `json:"page,omitempty"`
	PerPage     *int     `json:"per_page,omitempty"`
	SortBy      *string  `json:"sort_by,omitempty"`
	SortOrder   *string  `json:"sort_order,omitempty"`
} // @name AdminRegistrationFilterPayload

func (payload *AdminRegistrationFilterPayload) Bind(r *http.Request) error {
	if payload.Page != nil && *payload.Page < 1 {
		return errors.New("page must be greater than 0")
	}

	if payload.PerPage != nil && (*payload.PerPage < 1 || *payload.PerPage > 500) {
		return errors.New("per_page must be between 1 and 500")
	}

	if payload.SortOrder != nil && (*payload.SortOrder != "asc" && *payload.SortOrder != "desc") {
		return errors.New("sort_order must be 'asc' or 'desc'")
	}

	return nil
}

type AdminRegistrationListResponse struct {
	Registrations []*RambleRegistration `json:"registrations"`
	Total         int64                 `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
	TotalPages    int                   `json:"total_pages"`
} // @name AdminRegistrationListResponse

type BulkActionResult struct {
	SuccessCount int                   `json:"success_count"`
	FailureCount int                   `json:"failure_count"`
	Errors       []BulkActionError     `json:"errors,omitempty"`
	Updated      []*RambleRegistration `json:"updated,omitempty"`
} // @name BulkActionResult

type BulkActionError struct {
	RegistrationID uint   `json:"registration_id"`
	Error          string `json:"error"`
} // @name BulkActionError
