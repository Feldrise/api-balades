package reservation

import (
	"fmt"
	"net/http"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"github.com/go-chi/render"
)

// @Summary Create a reservation
// @Description Reserve a ramble for one or more participants. Public: the booker's account is found or created from the contact email.
// @Tags reservations
// @Accept json
// @Produce json
// @Param reservation body model.ReservationCreatePayload true "Reservation data"
// @Success 201 {object} model.Reservation
// @Failure 400 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /reservations [post]
func (config *Config) Create(w http.ResponseWriter, r *http.Request) {
	data := &model.ReservationCreatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	ramble, err := config.RambleRepository.FindByID(data.RambleID, false)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}
	if ramble == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}
	if ramble.IsCancelled {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("cannot reserve a cancelled ramble")))
		return
	}

	reservation, err := config.createReservation(data, ramble)
	if err != nil {
		render.Render(w, r, renderError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, reservation.ToModel())
}

// @Summary Get the user's reservations
// @Description Get the reservations of the authenticated booker, with participants, prices, status, estimated amount and payments
// @Tags reservations
// @Produce json
// @Success 200 {array} model.Reservation
// @Failure 401 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /reservations [get]
func (config *Config) GetUserReservations(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	reservations, err := config.ReservationRepository.FindAll(&dbmodel.ReservationFilter{UserID: &user.ID})
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	result := make([]*model.Reservation, len(reservations))
	for i := range reservations {
		result[i] = reservations[i].ToModel()
	}

	render.JSON(w, r, result)
}
