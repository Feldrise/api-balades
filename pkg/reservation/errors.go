package reservation

import (
	"errors"
	"net/http"

	apierrors "feldrise.com/balade/pkg/errors"
	"github.com/go-chi/render"
)

// businessError is a reservation rule violation reported to the client with a reason.
type businessError struct {
	status  int
	reason  string
	message string
}

func (e *businessError) Error() string {
	return e.message
}

func errInvalidPrice(message string) error {
	return &businessError{status: http.StatusBadRequest, reason: "invalid_price", message: message}
}

func renderError(err error) render.Renderer {
	var business *businessError
	if errors.As(err, &business) {
		return apierrors.ErrBusiness(business.status, business.reason, business.message)
	}

	return apierrors.ErrServerError(err)
}
