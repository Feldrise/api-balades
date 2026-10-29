package reservation

import (
	"slices"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/model"
)

// modification is the outcome of a change on a reservation: the reservations
// that left the waiting list are told by email once the change is saved.
type modification struct {
	promoted []uint
	// changed is false when the change left the reservation as it was.
	changed bool
}

// addParticipants adds participants to the reservation, all of them or none:
// to a pending or confirmed reservation only if they fit in the places left
// (they are confirmed with a confirmed reservation), to a reservation in
// waiting list which keeps waiting as a whole.
func (config *Config) addParticipants(reservation *dbmodel.Reservation, payload *model.ParticipantsAddPayload) error {
	return config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		active := locked.Reservation(reservation.ID)
		if active == nil {
			return errReservationCancelled()
		}

		hasBooker := false
		for _, participant := range active.Participants {
			hasBooker = hasBooker || participant.IsBooker
		}

		participants := make([]dbmodel.Participant, len(payload.Participants))
		for i, participant := range payload.Participants {
			priceLabel, err := resolvePriceLabel(locked.Ramble, participant.PriceLabel)
			if err != nil {
				return err
			}
			if participant.IsBooker {
				if hasBooker {
					return errSecondBooker()
				}
				hasBooker = true
			}

			participants[i] = dbmodel.Participant{
				FirstName:  participant.FirstName,
				LastName:   participant.LastName,
				PriceLabel: priceLabel,
				IsBooker:   participant.IsBooker,
			}
		}

		if max := locked.Ramble.MaxParticipantsPerReservation(); len(active.Participants)+len(participants) > max {
			return errTooManyParticipants(max)
		}
		if active.TakesPlaces() {
			if placesLeft := locked.Ramble.PlacesLeftFor(locked.Active); placesLeft != nil && len(participants) > *placesLeft {
				return errNotEnoughPlaces(*placesLeft)
			}
		}

		return locked.AddParticipants(active, participants)
	})
}

// removeParticipant removes a participant from the reservation. Removing the
// last one cancels the reservation. The places released go to the waiting list.
func (config *Config) removeParticipant(reservation *dbmodel.Reservation, participantID uint) (*modification, error) {
	result := &modification{}
	err := config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		active := locked.Reservation(reservation.ID)
		if active == nil {
			return errReservationCancelled()
		}
		if findParticipant(active, participantID) == nil {
			return errParticipantNotFound()
		}

		if len(active.Participants) == 1 {
			if err := config.cancel(locked, active); err != nil {
				return err
			}
		} else if err := locked.RemoveParticipant(active, participantID); err != nil {
			return err
		}

		promoted, err := config.promoteWaitingList(locked)
		result.promoted = promoted
		return err
	})
	return result, err
}

// updateParticipant changes the name and price of a participant, as long as
// the reservation is not paid.
func (config *Config) updateParticipant(reservation *dbmodel.Reservation, participantID uint, payload *model.ParticipantUpdatePayload) error {
	return config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		active := locked.Reservation(reservation.ID)
		if active == nil {
			return errReservationCancelled()
		}
		participant := findParticipant(active, participantID)
		if participant == nil {
			return errParticipantNotFound()
		}
		if active.HasSuccessfulPayment() {
			return errReservationPaid()
		}

		priceLabel, err := resolvePriceLabel(locked.Ramble, payload.PriceLabel)
		if err != nil {
			return err
		}

		participant.FirstName = payload.FirstName
		participant.LastName = payload.LastName
		participant.PriceLabel = priceLabel
		return locked.UpdateParticipant(participant)
	})
}

// confirmReservation confirms a pending reservation, once every participant's
// price is known so that the amount to pay is complete. Confirming a
// reservation already confirmed changes nothing.
func (config *Config) confirmReservation(reservation *dbmodel.Reservation) error {
	return config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		active := locked.Reservation(reservation.ID)
		if active == nil {
			return errReservationCancelled()
		}
		if active.Status == dbmodel.ReservationStatusWaitingList {
			return errReservationWaitingList()
		}
		for _, participant := range active.Participants {
			if locked.Ramble.PriceMissing(participant.PriceLabel) {
				return errMissingPrices()
			}
		}
		if active.Status == dbmodel.ReservationStatusConfirmed {
			return nil
		}

		now := config.Clock.Now()
		active.Status = dbmodel.ReservationStatusConfirmed
		active.ConfirmationDate = &now
		return locked.SaveStatus(active)
	})
}

// cancelReservation cancels the whole reservation. The places released go to the waiting list.
func (config *Config) cancelReservation(reservation *dbmodel.Reservation) (*modification, error) {
	result := &modification{}
	err := config.ReservationRepository.ModifyForRamble(reservation.RambleID, func(locked *dbmodel.LockedRamble) error {
		active := locked.Reservation(reservation.ID)
		if active == nil {
			return errReservationCancelled()
		}
		if err := config.cancel(locked, active); err != nil {
			return err
		}

		promoted, err := config.promoteWaitingList(locked)
		result.promoted = promoted
		return err
	})
	return result, err
}

func (config *Config) cancel(locked *dbmodel.LockedRamble, reservation *dbmodel.Reservation) error {
	now := config.Clock.Now()
	reservation.Status = dbmodel.ReservationStatusCancelled
	reservation.CancellationDate = &now
	return locked.SaveStatus(reservation)
}

// promoteWaitingList goes through the reservations in waiting list by creation
// date and admits each one that fits entirely in the places left, never splitting
// a reservation. It returns the admitted reservations. The reservations of the
// skipped ids are left in the waiting list, as when an admin just moved them there.
func (config *Config) promoteWaitingList(locked *dbmodel.LockedRamble, skipped ...uint) ([]uint, error) {
	var promoted []uint
	for i := range locked.Active {
		reservation := &locked.Active[i]
		if reservation.Status != dbmodel.ReservationStatusWaitingList || slices.Contains(skipped, reservation.ID) {
			continue
		}

		placesLeft := locked.Ramble.PlacesLeftFor(locked.Active)
		if placesLeft != nil && *placesLeft <= 0 {
			break
		}
		if placesLeft != nil && len(reservation.Participants) > *placesLeft {
			continue
		}

		config.admit(reservation, locked.Ramble)
		if err := locked.SaveStatus(reservation); err != nil {
			return nil, err
		}
		promoted = append(promoted, reservation.ID)
	}
	return promoted, nil
}

func findParticipant(reservation *dbmodel.Reservation, participantID uint) *dbmodel.Participant {
	for i := range reservation.Participants {
		if reservation.Participants[i].ID == participantID {
			return &reservation.Participants[i]
		}
	}
	return nil
}
