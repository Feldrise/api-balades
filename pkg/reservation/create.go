package reservation

import (
	"fmt"
	"strings"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/model"
)

// autoConfirmationPeriod is how close to the ramble a reservation is confirmed on creation.
const autoConfirmationPeriod = 3 * 24 * time.Hour

func (config *Config) createReservation(data *model.ReservationCreatePayload, ramble *dbmodel.Ramble) (*dbmodel.Reservation, error) {
	if max := ramble.MaxParticipantsPerReservation(); len(data.Participants) > max {
		return nil, errTooManyParticipants(max)
	}

	participants := make([]dbmodel.Participant, len(data.Participants))
	for i, participant := range data.Participants {
		priceLabel, err := resolvePriceLabel(ramble, participant.PriceLabel)
		if err != nil {
			return nil, err
		}

		participants[i] = dbmodel.Participant{
			FirstName:  participant.FirstName,
			LastName:   participant.LastName,
			PriceLabel: priceLabel,
			IsBooker:   participant.IsBooker,
		}
	}

	user, err := config.findOrCreateBooker(data)
	if err != nil {
		return nil, err
	}

	return config.ReservationRepository.CreateForRamble(ramble.ID, func(active []dbmodel.Reservation) (*dbmodel.Reservation, error) {
		for _, existing := range active {
			if strings.EqualFold(existing.ContactEmail, data.ContactEmail) {
				return nil, errReservationAlreadyExists(existing.ID)
			}
		}

		reservation := &dbmodel.Reservation{
			RambleID:     ramble.ID,
			UserID:       user.ID,
			ContactEmail: data.ContactEmail,
			Participants: participants,
		}
		config.decideInitialStatus(reservation, ramble, active)

		return reservation, nil
	})
}

// decideInitialStatus puts the whole reservation in waiting list when its
// participants do not all fit in the places left, and confirms it right away
// when the ramble is within the confirmation period.
func (config *Config) decideInitialStatus(reservation *dbmodel.Reservation, ramble *dbmodel.Ramble, active []dbmodel.Reservation) {
	placesLeft := ramble.PlacesLeftFor(active)
	if placesLeft != nil && len(reservation.Participants) > *placesLeft {
		reservation.Status = dbmodel.ReservationStatusWaitingList
		return
	}

	now := config.Clock.Now()
	if ramble.Date != nil {
		untilRamble := ramble.Date.Sub(now)
		if untilRamble >= 0 && untilRamble <= autoConfirmationPeriod {
			reservation.Status = dbmodel.ReservationStatusConfirmed
			reservation.ConfirmationDate = &now
			return
		}
	}

	reservation.Status = dbmodel.ReservationStatusPending
}

// findOrCreateBooker returns the single account of the booker, found or created
// (without password) from the contact email.
func (config *Config) findOrCreateBooker(data *model.ReservationCreatePayload) (*dbmodel.User, error) {
	profile := dbmodel.UserProfile{Phone: data.ContactPhone}
	for _, participant := range data.Participants {
		if participant.IsBooker {
			lastName := participant.LastName
			profile.FirstName = participant.FirstName
			profile.LastName = &lastName
		}
	}

	return config.UserRepository.FindOrCreateByEmail(&dbmodel.User{
		Email:       data.ContactEmail,
		UserProfile: profile,
	})
}

// resolvePriceLabel checks that label is one of the ramble's prices. A ramble
// without price takes no label, and a ramble with a single price applies it by default.
func resolvePriceLabel(ramble *dbmodel.Ramble, label string) (string, error) {
	if len(ramble.Prices) == 0 {
		return "", nil
	}

	if label == "" {
		if len(ramble.Prices) == 1 {
			return ramble.Prices[0].Label, nil
		}
		return "", errInvalidPrice("each participant needs a price of the ramble")
	}

	for _, price := range ramble.Prices {
		if price.Label == label {
			return label, nil
		}
	}

	return "", errInvalidPrice(fmt.Sprintf("unknown price %q for this ramble", label))
}
