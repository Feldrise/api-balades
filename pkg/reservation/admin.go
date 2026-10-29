package reservation

import (
	"strings"

	"feldrise.com/balade/database/dbmodel"
)

// setStatus sets the status of the reservation on behalf of an admin or a guide.
// The reservation is never split: it takes places as a whole or none, so one that
// does not take places yet must fit entirely in the places left.
func (config *Config) setStatus(reservation *dbmodel.Reservation, status string, reason *string) (*modification, error) {
	result := &modification{}
	err := config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		active := locked.Reservation(reservation.ID)
		if active == nil {
			if status == dbmodel.ReservationStatusCancelled {
				return nil
			}
			var err error
			if active, err = config.reactivable(locked, reservation.ID); err != nil || active == nil {
				return err
			}
		}
		if active.Status == status {
			return nil
		}
		result.changed = true

		if !active.TakesPlaces() && (status == dbmodel.ReservationStatusPending || status == dbmodel.ReservationStatusConfirmed) {
			if placesLeft := locked.Ramble.PlacesLeftFor(locked.Active); placesLeft != nil && len(active.Participants) > *placesLeft {
				return errNotEnoughPlaces(*placesLeft)
			}
		}

		releasing := active.TakesPlaces()

		active.CancellationDate = nil
		active.CancellationReason = nil
		switch status {
		case dbmodel.ReservationStatusCancelled:
			active.CancellationReason = reason
			if err := config.cancel(locked, active); err != nil {
				return err
			}
		case dbmodel.ReservationStatusWaitingList:
			active.Status = status
			active.ConfirmationDate = nil
		case dbmodel.ReservationStatusConfirmed:
			now := config.Clock.Now()
			active.Status = status
			active.ConfirmationDate = &now
		case dbmodel.ReservationStatusPending:
			active.Status = status
			active.ConfirmationDate = nil
		}
		if err := locked.SaveStatus(active); err != nil {
			return err
		}

		if releasing && !active.TakesPlaces() {
			promoted, err := config.promoteWaitingList(locked, active.ID)
			result.promoted = promoted
			return err
		}
		return nil
	})
	return result, err
}

// reactivable returns the cancelled reservation with the given id, which an
// admin may bring back unless its booker has another reservation on the ramble
// since. It returns nil when the reservation is not cancelled.
func (config *Config) reactivable(locked *dbmodel.LockedRamble, id uint) (*dbmodel.Reservation, error) {
	cancelled, err := locked.CancelledReservation(id)
	if err != nil || cancelled == nil {
		return nil, err
	}

	for _, existing := range locked.Active {
		if strings.EqualFold(existing.ContactEmail, cancelled.ContactEmail) {
			return nil, errReservationAlreadyExists(existing.ID)
		}
	}

	return cancelled, nil
}

// deleteReservation deletes the reservation on behalf of an admin or a guide.
// The places it held go to the waiting list.
func (config *Config) deleteReservation(reservation *dbmodel.Reservation) (*modification, error) {
	result := &modification{}
	err := config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		releasing := reservation.Status != dbmodel.ReservationStatusCancelled
		if active := locked.Reservation(reservation.ID); active != nil {
			releasing = active.TakesPlaces()
		}

		if err := locked.Delete(reservation); err != nil {
			return err
		}

		if releasing {
			promoted, err := config.promoteWaitingList(locked)
			result.promoted = promoted
			return err
		}
		return nil
	})
	return result, err
}

// bulkActionStatuses are the statuses set by the bulk actions other than delete.
var bulkActionStatuses = map[string]string{
	"confirm":         dbmodel.ReservationStatusConfirmed,
	"cancel":          dbmodel.ReservationStatusCancelled,
	"move_to_waiting": dbmodel.ReservationStatusWaitingList,
}
