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
// ramble takes place within the confirmation period to confirm it, once per reservation.
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

		claimed, err := config.ReservationRepository.ClaimConfirmationRequest(reservation.ID, deadline)
		if err != nil {
			return err
		}
		if claimed {
			config.sendConfirmationRequest(reservation, deadline)
		}
	}
	return nil
}

// CancelUnconfirmedReservations cancels entirely the reservations still pending
// at their confirmation deadline, then admits the waiting list into the places released.
func (config *Config) CancelUnconfirmedReservations(now time.Time) error {
	reservations, err := config.ReservationRepository.FindPendingRambleOnOrBefore(now.Add(confirmationDeadlineBeforeRamble))
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
		config.notifyPromoted(promoted)
	}
	return nil
}

func (config *Config) sendConfirmationRequest(reservation *dbmodel.Reservation, deadline time.Time) {
	firstName, lastName := bookerName(reservation)
	data := email.EventConfirmationRequestData{
		FirstName:            firstName,
		LastName:             lastName,
		Title:                reservation.Ramble.Title,
		Date:                 formatDate(reservation.Ramble.Date),
		Location:             formatLocation(reservation.Ramble.Location),
		ConfirmationDeadline: formatDate(&deadline),
		ReservationURL:       fmt.Sprintf("%s/mes-reservations/%d", config.ApplicationURL, reservation.ID),
	}

	err := config.EmailService.Send(
		reservation.ContactEmail,
		"Confirmation requise - Balade Écologique",
		"event_confirmation_request",
		data,
	)
	if err != nil {
		log.Printf("Failed to send confirmation request for reservation %d: %v", reservation.ID, err)
	}
}

// bookerName is the name of the booker, from their participant line or else their account.
func bookerName(reservation *dbmodel.Reservation) (string, string) {
	for _, participant := range reservation.Participants {
		if participant.IsBooker {
			return participant.FirstName, participant.LastName
		}
	}

	lastName := ""
	if reservation.User.UserProfile.LastName != nil {
		lastName = *reservation.User.UserProfile.LastName
	}
	return reservation.User.UserProfile.FirstName, lastName
}
