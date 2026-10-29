package reservation

import (
	"fmt"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/model"
)

func (config *Config) createReservation(data *model.ReservationCreatePayload, ramble *dbmodel.Ramble) (*dbmodel.Reservation, error) {
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
		return &dbmodel.Reservation{
			RambleID:     ramble.ID,
			UserID:       user.ID,
			ContactEmail: data.ContactEmail,
			Status:       dbmodel.ReservationStatusPending,
			Participants: participants,
		}, nil
	})
}

// findOrCreateBooker returns the single account of the booker, found or created
// (without password) from the contact email.
func (config *Config) findOrCreateBooker(data *model.ReservationCreatePayload) (*dbmodel.User, error) {
	user, err := config.UserRepository.FindByEmail(data.ContactEmail, nil)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	profile := dbmodel.UserProfile{Phone: data.ContactPhone}
	for _, participant := range data.Participants {
		if participant.IsBooker {
			lastName := participant.LastName
			profile.FirstName = participant.FirstName
			profile.LastName = &lastName
		}
	}

	return config.UserRepository.Create(&dbmodel.User{
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
