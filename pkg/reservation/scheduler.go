package reservation

import (
	"fmt"
	"log"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/notifications/email"
)

// confirmationRequestPeriod is how long before the ramble the booker is asked to confirm.
const confirmationRequestPeriod = 3 * 24 * time.Hour

// confirmationDeadlineBeforeRamble is how long before the ramble unconfirmed reservations are cancelled.
const confirmationDeadlineBeforeRamble = 12 * time.Hour

// SendConfirmationRequests asks the booker of each pending reservation whose
// ramble takes place within the confirmation period to confirm it, once per
// reservation. The request is marked as sent, by the confirmation deadline,
// only once the email is sent: a failed request is sent again at the next pass.
func (config *Config) SendConfirmationRequests(now time.Time) error {
	reservations, err := config.ReservationRepository.FindAwaitingConfirmationRequest(
		now.Add(confirmationDeadlineBeforeRamble),
		now.Add(confirmationRequestPeriod),
	)
	if err != nil {
		return err
	}

	for i := range reservations {
		reservation := &reservations[i]
		deadline := reservation.Ramble.Date.Add(-confirmationDeadlineBeforeRamble)

		if err := config.sendConfirmationRequest(reservation, deadline); err != nil {
			log.Printf("Failed to send confirmation request for reservation %d, retrying at the next pass: %v", reservation.ID, err)
			continue
		}
		if err := config.ReservationRepository.MarkConfirmationRequested(reservation.ID, deadline); err != nil {
			return err
		}
	}
	return nil
}

// CancelUnconfirmedReservations cancels entirely the reservations still pending
// at their confirmation deadline, then admits the waiting list into the places
// released. A reservation whose booker was never sent the confirmation request
// (the email failed until the deadline) is not cancelled: it keeps its places.
func (config *Config) CancelUnconfirmedReservations(now time.Time) error {
	reservations, err := config.ReservationRepository.FindRequestedPendingRambleOnOrBefore(now.Add(confirmationDeadlineBeforeRamble))
	if err != nil {
		return err
	}

	for _, reservation := range reservations {
		var promoted []uint
		err := config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
			active := locked.Reservation(reservation.ID)
			if active == nil || active.Status != dbmodel.ReservationStatusPending {
				return nil
			}

			reason := "No confirmation before deadline"
			active.CancellationReason = &reason
			if err := config.cancel(locked, active); err != nil {
				return err
			}

			promoted, err = config.promoteWaitingList(locked)
			return err
		})
		if err != nil {
			return err
		}

		log.Printf("Cancelled unconfirmed reservation %d", reservation.ID)
		if cancelled, err := config.ReservationRepository.FindByID(reservation.ID); err == nil && cancelled != nil {
			config.cancelObsoletePayments(cancelled)
		}
		config.notifyPromoted(promoted)
	}
	return nil
}

// ReconcileOpenPayments cancels the open payments that the reservation can no
// longer pay (see payments.PaymentService.ReconcileObsoletePayments): of a
// reservation cancelled or deleted, of a ramble cancelled or started, of a
// former amount due or guide. It catches up on the cancellations on Stripe
// that failed at the time. A reservation held by an operation of the user, or
// whose cancellation Stripe is slow to answer, is reconciled at the next pass.
func (config *Config) ReconcileOpenPayments() error {
	reservations, err := config.ReservationRepository.FindWithOpenPayments()
	if err != nil {
		return err
	}

	for i := range reservations {
		if err := config.PaymentService.ReconcileObsoletePayments(&reservations[i]); err != nil {
			log.Printf("Failed to reconcile the open payments of reservation %d, retrying at the next pass: %v", reservations[i].ID, err)
		}
	}
	return nil
}

func (config *Config) sendConfirmationRequest(reservation *dbmodel.Reservation, deadline time.Time) error {
	firstName, lastName := reservation.BookerName()
	data := email.EventConfirmationRequestData{
		FirstName:            firstName,
		LastName:             lastName,
		Title:                reservation.Ramble.Title,
		Date:                 formatDate(reservation.Ramble.Date),
		Location:             formatLocation(reservation.Ramble.Location),
		ConfirmationDeadline: formatDate(&deadline),
		ReservationURL:       fmt.Sprintf("%s/mes-reservations/%d", config.ApplicationURL, reservation.ID),
	}

	return config.EmailService.Send(
		reservation.ContactEmail,
		"Confirmation requise - Balade Écologique",
		"event_confirmation_request",
		data,
	)
}
