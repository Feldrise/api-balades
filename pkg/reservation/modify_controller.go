package reservation

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// @Summary Add participants to a reservation
// @Description Add participants to a pending reservation if they all fit in the places left, or to a reservation in waiting list. The booker receives a summary email.
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param participants body model.ParticipantsAddPayload true "Participants to add"
// @Success 200 {object} model.Reservation
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "not_enough_places, reservation_cancelled or reservation_confirmed"
// @Router /reservations/{id}/participants [post]
func (config *Config) AddParticipants(w http.ResponseWriter, r *http.Request) {
	reservation, ok := config.reservationToModify(w, r)
	if !ok {
		return
	}

	data := &model.ParticipantsAddPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	if err := config.addParticipants(reservation, data); err != nil {
		render.Render(w, r, renderError(err))
		return
	}

	config.respondModified(w, r, reservation.ID, changeParticipantsAdded, nil)
}

// @Summary Remove a participant from a reservation
// @Description Remove a participant. Removing the last one cancels the reservation. The places released go to the waiting list. has_successful_payment tells that the booker should contact the guide for a refund.
// @Tags reservations
// @Produce json
// @Param id path int true "Reservation ID"
// @Param participantId path int true "Participant ID"
// @Success 200 {object} model.Reservation
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "reservation_cancelled"
// @Router /reservations/{id}/participants/{participantId} [delete]
func (config *Config) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	reservation, ok := config.reservationToModify(w, r)
	if !ok {
		return
	}
	participantID, ok := parseID(w, r, "participantId")
	if !ok {
		return
	}

	result, err := config.removeParticipant(reservation, participantID)
	if err != nil {
		render.Render(w, r, renderError(err))
		return
	}

	config.respondModified(w, r, reservation.ID, changeParticipantsRemoved, result.promoted)
}

// @Summary Update a participant of a reservation
// @Description Change the name or the price of a participant, as long as the reservation is not paid.
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param participantId path int true "Participant ID"
// @Param participant body model.ParticipantUpdatePayload true "Participant data"
// @Success 200 {object} model.Reservation
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "reservation_paid or reservation_cancelled"
// @Router /reservations/{id}/participants/{participantId} [put]
func (config *Config) UpdateParticipant(w http.ResponseWriter, r *http.Request) {
	reservation, ok := config.reservationToModify(w, r)
	if !ok {
		return
	}
	participantID, ok := parseID(w, r, "participantId")
	if !ok {
		return
	}

	data := &model.ParticipantUpdatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	if err := config.updateParticipant(reservation, participantID, data); err != nil {
		render.Render(w, r, renderError(err))
		return
	}

	config.respondModified(w, r, reservation.ID, changeNone, nil)
}

// @Summary Cancel a reservation
// @Description Cancel the whole reservation. The places released go to the waiting list. No automatic refund: has_successful_payment tells that the booker should contact the guide.
// @Tags reservations
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 200 {object} model.Reservation
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "reservation_cancelled"
// @Router /reservations/{id}/cancel [put]
func (config *Config) Cancel(w http.ResponseWriter, r *http.Request) {
	reservation, ok := config.reservationToModify(w, r)
	if !ok {
		return
	}

	result, err := config.cancelReservation(reservation)
	if err != nil {
		render.Render(w, r, renderError(err))
		return
	}

	config.respondModified(w, r, reservation.ID, changeNone, result.promoted)
}

// reservationToModify returns the reservation of the request when the user is
// its booker or an admin or guide allowed on its ramble, and renders the error otherwise.
func (config *Config) reservationToModify(w http.ResponseWriter, r *http.Request) (*dbmodel.Reservation, bool) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return nil, false
	}

	id, ok := parseID(w, r, "id")
	if !ok {
		return nil, false
	}

	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return nil, false
	}
	if reservation == nil {
		render.Render(w, r, errors.ErrNotFound())
		return nil, false
	}

	allowed, err := config.canView(user, reservation)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return nil, false
	}
	if !allowed {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return nil, false
	}

	return reservation, true
}

// respondModified renders the reservation as saved, sends the booker the
// summary of the change and tells the promoted reservations they left the waiting list.
func (config *Config) respondModified(w http.ResponseWriter, r *http.Request, id uint, change string, promoted []uint) {
	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if change != changeNone {
		config.sendReservationEmail(reservation, change)
	}
	config.notifyPromoted(promoted)

	render.JSON(w, r, reservation.ToModel())
}

func parseID(w http.ResponseWriter, r *http.Request, param string) (uint, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, param), 10, 64)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid %s", param)))
		return 0, false
	}
	return uint(id), true
}

// notifyPromoted tells the bookers of the promoted reservations they left the waiting list.
func (config *Config) notifyPromoted(promoted []uint) {
	for _, promotedID := range promoted {
		promotedReservation, err := config.ReservationRepository.FindByID(promotedID)
		if err != nil || promotedReservation == nil {
			log.Printf("Failed to load promoted reservation %d: %v", promotedID, err)
			continue
		}
		config.sendReservationEmail(promotedReservation, changePromoted)
	}
}
