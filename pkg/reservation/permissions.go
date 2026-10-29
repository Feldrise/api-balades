package reservation

import (
	"net/http"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/errors"
	"github.com/go-chi/render"
)

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

// resolveGuideRambleScope returns the ramble IDs a user with the blanket or the
// self permission may act on. With a blanket permission, the slice holds only
// rambleID when it is set and is nil otherwise (every ramble). With the self
// permission, it holds the rambles the user guides, restricted to rambleID when
// it is set. ok is false when the error is already rendered.
func (config *Config) resolveGuideRambleScope(w http.ResponseWriter, r *http.Request, user *dbmodel.User, rambleID *uint, blanketPerm, selfPerm string) (rambleIDs []uint, ok bool) {
	if user.HasPermission(blanketPerm) {
		if rambleID != nil {
			return []uint{*rambleID}, true
		}
		return nil, true
	}

	if !user.HasPermission(selfPerm) {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return nil, false
	}

	ownedIDs, err := config.GuideRepository.FindRambleIDsOwnedByUser(user.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return nil, false
	}

	if rambleID != nil {
		for _, id := range ownedIDs {
			if id == *rambleID {
				return []uint{*rambleID}, true
			}
		}
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return nil, false
	}

	if ownedIDs == nil {
		ownedIDs = []uint{}
	}
	return ownedIDs, true
}

// canActOnRamble reports whether user holds the blanket permission, or the self
// permission and guides the ramble.
func (config *Config) canActOnRamble(user *dbmodel.User, rambleID uint, blanketPerm, selfPerm string) (bool, error) {
	if user.HasPermission(blanketPerm) {
		return true, nil
	}

	if user.HasPermission(selfPerm) {
		return config.GuideRepository.UserOwnsRamble(user.ID, rambleID)
	}

	return false, nil
}
