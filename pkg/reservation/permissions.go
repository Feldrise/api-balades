package reservation

import "feldrise.com/balade/database/dbmodel"

// canView reports whether user may read the reservation: the booker, or an admin
// or guide allowed to view the registrations of the reservation's ramble.
func (config *Config) canView(user *dbmodel.User, reservation *dbmodel.Reservation) (bool, error) {
	if reservation.UserID == user.ID {
		return true, nil
	}

	if user.HasPermission("view:all-registrations") {
		return true, nil
	}

	if user.HasPermission("view:registrations:self") {
		return config.GuideRepository.UserOwnsRamble(user.ID, reservation.RambleID)
	}

	return false, nil
}
