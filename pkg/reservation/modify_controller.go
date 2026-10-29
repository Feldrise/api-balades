package reservation

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/payments"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// @Summary Add participants to a reservation
// @Description Add participants to a pending or confirmed reservation if they all fit in the places left, or to a reservation in waiting list. Participants added to a confirmed reservation are confirmed, and payment holds the complementary payment of their prices when online payment is enabled. The booker receives a summary email.
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param participants body model.ParticipantsAddPayload true "Participants to add"
// @Success 200 {object} model.ReservationWithPayment
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "not_enough_places, reservation_cancelled or ramble_cancelled"
// @Router /reservations/{id}/participants [post]
func (config *Config) AddParticipants(w http.ResponseWriter, r *http.Request) {
	reservation, ok := config.reservationToModify(w, r, permissionsChangeParticipants)
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

	added, err := config.ReservationRepository.FindByID(reservation.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}
	config.sendReservationEmail(added, changeParticipantsAdded)

	config.respondWithPayment(w, r, reservation.ID)
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
	reservation, ok := config.reservationToModify(w, r, permissionsChangeParticipants)
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
// @Description Change the name or the price of a participant, as long as the reservation is not paid. Once paid, only a missing price (price_missing) can still be specified.
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
	reservation, ok := config.reservationToModify(w, r, permissionsChangeParticipants)
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

// @Summary Confirm a reservation
// @Description Confirm a pending reservation, for its booker only, from 3 days before the ramble until the confirmation deadline 12 hours before. When online payment is enabled, payment holds the payment of the amount due, computed from the participants' prices. On a confirmed reservation, it gives the payment of the balance until the ramble starts.
// @Tags reservations
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 200 {object} model.ReservationWithPayment
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "missing_prices, reservation_waiting_list, reservation_cancelled, ramble_cancelled, confirmation_not_open, confirmation_deadline_passed or ramble_started"
// @Router /reservations/{id}/confirm [put]
func (config *Config) Confirm(w http.ResponseWriter, r *http.Request) {
	reservation, ok := config.bookerReservation(w, r)
	if !ok {
		return
	}

	if err := config.confirmReservation(reservation); err != nil {
		render.Render(w, r, renderError(err))
		return
	}

	config.respondWithPayment(w, r, reservation.ID)
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
	reservation, ok := config.reservationToModify(w, r, permissionsChangeStatus)
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

// modifyPermissions are the blanket and self permissions that allow staff to
// make a change on the reservations of a ramble, as on the admin routes.
type modifyPermissions struct {
	blanket string
	self    string
}

var (
	permissionsChangeParticipants = modifyPermissions{"update:registration-details", "update:registration-details:self"}
	permissionsChangeStatus       = modifyPermissions{"update:registration-status", "update:registration-status:self"}
)

// reservationToModify returns the reservation of the request when the user is
// its booker, or holds the permissions to make the change on its ramble, and
// renders the error otherwise.
func (config *Config) reservationToModify(w http.ResponseWriter, r *http.Request, permissions modifyPermissions) (*dbmodel.Reservation, bool) {
	user, reservation, ok := config.requestReservation(w, r)
	if !ok {
		return nil, false
	}
	if reservation.UserID == user.ID {
		return reservation, true
	}

	allowed, err := config.canActOnRamble(user, reservation.RambleID, permissions.blanket, permissions.self)
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

// bookerReservation returns the reservation of the request when the user is
// its booker, and renders the error otherwise.
func (config *Config) bookerReservation(w http.ResponseWriter, r *http.Request) (*dbmodel.Reservation, bool) {
	user, reservation, ok := config.requestReservation(w, r)
	if !ok {
		return nil, false
	}

	if reservation.UserID != user.ID {
		render.Render(w, r, errors.ErrForbidden("only the booker can do this"))
		return nil, false
	}

	return reservation, true
}

// requestReservation returns the authenticated user and the reservation of the
// request, and renders the error when there is none.
func (config *Config) requestReservation(w http.ResponseWriter, r *http.Request) (*dbmodel.User, *dbmodel.Reservation, bool) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return nil, nil, false
	}

	id, ok := parseID(w, r, "id")
	if !ok {
		return nil, nil, false
	}

	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return nil, nil, false
	}
	if reservation == nil {
		render.Render(w, r, errors.ErrNotFound())
		return nil, nil, false
	}

	return user, reservation, true
}

// respondWithPayment renders the reservation as saved with, when it is
// confirmed and online payment is enabled, the payment of the amount it still owes.
func (config *Config) respondWithPayment(w http.ResponseWriter, r *http.Request, id uint) {
	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	config.cancelObsoletePayments(reservation)

	var payment *model.PaymentResponse
	if reservation.Status == dbmodel.ReservationStatusConfirmed && reservation.Ramble.PaymentEnabled && !reservation.Ramble.IsCancelled &&
		!reservation.HasMissingPrices() && reservation.AmountDue() > 0 {
		firstName, lastName := reservation.BookerName()
		payment, err = config.PaymentService.CreatePaymentForReservation(reservation, payments.PayerInfo{
			Email: reservation.ContactEmail,
			Name:  strings.TrimSpace(firstName + " " + lastName),
		})
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}

		if reservation, err = config.ReservationRepository.FindByID(id); err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	render.JSON(w, r, &model.ReservationWithPayment{Reservation: reservation.ToModel(), Payment: payment})
}

// respondModified renders the reservation as saved, sends the booker the
// summary of the change and tells the promoted reservations they left the waiting list.
func (config *Config) respondModified(w http.ResponseWriter, r *http.Request, id uint, change string, promoted []uint) {
	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	config.cancelObsoletePayments(reservation)
	if change != changeNone {
		config.sendReservationEmail(reservation, change)
	}
	config.notifyPromoted(promoted)

	render.JSON(w, r, reservation.ToModel())
}

// cancelObsoletePayments cancels the pending payments of the reservation that
// no longer match what it owes. A failure is logged without failing the request.
func (config *Config) cancelObsoletePayments(reservation *dbmodel.Reservation) {
	if err := config.PaymentService.CancelObsoletePayments(reservation); err != nil {
		log.Printf("Failed to cancel obsolete payments of reservation %d: %v", reservation.ID, err)
	}
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
