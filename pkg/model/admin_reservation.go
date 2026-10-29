package model

import (
	"errors"
	"net/http"
)

// AdminReservationFilterPayload filters the reservations listed to admins and guides.
type AdminReservationFilterPayload struct {
	RambleID    *uint    `json:"ramble_id,omitempty"`
	Email       *string  `json:"email,omitempty"`
	Search      *string  `json:"search,omitempty"`
	RambleTitle *string  `json:"ramble_title,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Statuses    []string `json:"statuses,omitempty"`
	UserID      *uint    `json:"user_id,omitempty"`
	DateFrom    *string  `json:"date_from,omitempty"` // YYYY-MM-DD format
	DateTo      *string  `json:"date_to,omitempty"`   // YYYY-MM-DD format
	Page        *int     `json:"page,omitempty"`
	PerPage     *int     `json:"per_page,omitempty"`
	SortBy      *string  `json:"sort_by,omitempty"`    // created_at, contact_email or status
	SortOrder   *string  `json:"sort_order,omitempty"` // asc or desc
} // @name AdminReservationFilterPayload

func (payload *AdminReservationFilterPayload) Bind(r *http.Request) error {
	if payload.Page != nil && *payload.Page < 1 {
		return errors.New("page must be greater than 0")
	}

	if payload.PerPage != nil && (*payload.PerPage < 1 || *payload.PerPage > 500) {
		return errors.New("per_page must be between 1 and 500")
	}

	if payload.SortOrder != nil && *payload.SortOrder != "asc" && *payload.SortOrder != "desc" {
		return errors.New("sort_order must be 'asc' or 'desc'")
	}

	return nil
}

type AdminReservationListResponse struct {
	Reservations []*Reservation `json:"reservations"`
	Total        int64          `json:"total"`
	Page         int            `json:"page"`
	PerPage      int            `json:"per_page"`
	TotalPages   int            `json:"total_pages"`
} // @name AdminReservationListResponse

// ReservationStatuses are the statuses an admin can set on a reservation.
var reservationStatuses = map[string]bool{
	"pending":      true,
	"confirmed":    true,
	"waiting_list": true,
	"cancelled":    true,
}

type AdminReservationStatusUpdatePayload struct {
	Status    string  `json:"status"`
	Reason    *string `json:"reason,omitempty"`
	SendEmail *bool   `json:"send_email,omitempty"` // Whether to send the reservation summary to the booker
} // @name AdminReservationStatusUpdatePayload

func (payload *AdminReservationStatusUpdatePayload) Bind(r *http.Request) error {
	if payload.Status == "" {
		return errors.New("status is required")
	}

	if !reservationStatuses[payload.Status] {
		return errors.New("invalid status. Valid statuses are: pending, confirmed, waiting_list, cancelled")
	}

	return nil
}
