package reservation

import (
	"errors"
	"fmt"
	"net/http"

	apierrors "feldrise.com/balade/pkg/errors"
	"github.com/go-chi/render"
)

// businessError is a reservation rule violation reported to the client with a reason.
type businessError struct {
	status  int
	reason  string
	message string
	// reservationID is the existing reservation a duplicate attempt refers to.
	reservationID uint
}

type businessErrorResponse struct {
	*apierrors.ErrResponse
	ReservationID uint `json:"reservation_id,omitempty"`
}

func (e *businessError) Error() string {
	return e.message
}

func errInvalidPrice(message string) error {
	return &businessError{status: http.StatusBadRequest, reason: "invalid_price", message: message}
}

func errTooManyParticipants(max int) error {
	return &businessError{
		status:  http.StatusBadRequest,
		reason:  "too_many_participants",
		message: fmt.Sprintf("a reservation cannot have more than %d participants for this ramble", max),
	}
}

func errReservationAlreadyExists(reservationID uint) error {
	return &businessError{
		status:        http.StatusConflict,
		reason:        "reservation_already_exists",
		message:       "a reservation already exists for this ramble and email",
		reservationID: reservationID,
	}
}

func errNotEnoughPlaces(placesLeft int) error {
	return &businessError{
		status:  http.StatusConflict,
		reason:  "not_enough_places",
		message: fmt.Sprintf("the participants do not fit in the %d places left", max(placesLeft, 0)),
	}
}

func errReservationCancelled() error {
	return &businessError{status: http.StatusConflict, reason: "reservation_cancelled", message: "the reservation is cancelled"}
}

func errReservationPaid() error {
	return &businessError{
		status:  http.StatusConflict,
		reason:  "reservation_paid",
		message: "participants cannot be modified once the reservation is paid",
	}
}

func errReservationWaitingList() error {
	return &businessError{
		status:  http.StatusConflict,
		reason:  "reservation_waiting_list",
		message: "a reservation in waiting list cannot be confirmed",
	}
}

func errMissingPrices() error {
	return &businessError{
		status:  http.StatusConflict,
		reason:  "missing_prices",
		message: "the price of every participant must be specified first",
	}
}

func errConfirmationNotOpen() error {
	return &businessError{
		status:  http.StatusConflict,
		reason:  "confirmation_not_open",
		message: "the reservation can be confirmed from 3 days before the ramble",
	}
}

func errConfirmationDeadlinePassed() error {
	return &businessError{
		status:  http.StatusConflict,
		reason:  "confirmation_deadline_passed",
		message: "the confirmation deadline, 12 hours before the ramble, has passed",
	}
}

func errRambleCancelled() error {
	return &businessError{status: http.StatusConflict, reason: "ramble_cancelled", message: "the ramble is cancelled"}
}

func errRambleStarted() error {
	return &businessError{status: http.StatusConflict, reason: "ramble_started", message: "the ramble has started"}
}

func errParticipantNotFound() error {
	return &businessError{status: http.StatusNotFound, reason: "participant_not_found", message: "no such participant in this reservation"}
}

func errSecondBooker() error {
	return &businessError{status: http.StatusBadRequest, reason: "second_booker", message: "only one participant can be the booker"}
}

func renderError(err error) render.Renderer {
	var business *businessError
	if errors.As(err, &business) {
		return &businessErrorResponse{
			ErrResponse:   apierrors.ErrBusiness(business.status, business.reason, business.message),
			ReservationID: business.reservationID,
		}
	}

	return apierrors.ErrServerError(err)
}
