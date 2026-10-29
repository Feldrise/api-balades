package reservation

import (
	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/model"
)

func (config *Config) createReservation(data *model.ReservationCreatePayload, ramble *dbmodel.Ramble) (*dbmodel.Reservation, error) {
	participants := make([]dbmodel.Participant, len(data.Participants))
	for i, participant := range data.Participants {
		participants[i] = dbmodel.Participant{
			FirstName:  participant.FirstName,
			LastName:   participant.LastName,
			PriceLabel: participant.PriceLabel,
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
