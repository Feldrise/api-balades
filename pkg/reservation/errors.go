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
