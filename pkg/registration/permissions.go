package registration

import (
	"net/http"
	"strings"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"github.com/go-chi/render"
)

func hasBlanketOrOwnsRamble(user *dbmodel.User, guideRepo dbmodel.GuideRepository, rambleID uint, blanketPerm, selfPerm string) (bool, error) {
	if user == nil {
		return false, nil
	}

	if user.HasPermission(blanketPerm) {
		return true, nil
	}

	if user.HasPermission(selfPerm) {
		return guideRepo.UserOwnsRamble(user.ID, rambleID)
	}

	return false, nil
}

func (config *Config) canViewRambleRegistrations(user *dbmodel.User, rambleID uint) (bool, error) {
	return hasBlanketOrOwnsRamble(user, config.GuideRepository, rambleID, "view:all-registrations", "view:registrations:self")
}

func (config *Config) canUpdateRegistrationDetails(user *dbmodel.User, rambleID uint) (bool, error) {
	return hasBlanketOrOwnsRamble(user, config.GuideRepository, rambleID, "update:registration-details", "update:registration-details:self")
}

func (config *Config) canUpdateRegistrationStatus(user *dbmodel.User, rambleID uint) (bool, error) {
	return hasBlanketOrOwnsRamble(user, config.GuideRepository, rambleID, "update:registration-status", "update:registration-status:self")
}

func (config *Config) canManageRegistration(user *dbmodel.User, rambleID uint) (bool, error) {
	return hasBlanketOrOwnsRamble(user, config.GuideRepository, rambleID, "manage:registration", "manage:registration:self")
}

func (config *Config) canBulkRegistrationActions(user *dbmodel.User, rambleID uint) (bool, error) {
	return hasBlanketOrOwnsRamble(user, config.GuideRepository, rambleID, "bulk:registration-actions", "bulk:registration-actions:self")
}

func (config *Config) hasAnyBulkPermission(user *dbmodel.User) bool {
	return user.HasPermission("bulk:registration-actions") || user.HasPermission("bulk:registration-actions:self")
}

func isGroupMember(user *dbmodel.User, group *dbmodel.RambleRegistrationGroup) bool {
	if user == nil || group == nil {
		return false
	}

	if strings.EqualFold(group.PrimaryEmail, user.Email) {
		return true
	}

	for _, registration := range group.Registrations {
		if registration.UserID != nil && *registration.UserID == user.ID {
			return true
		}
	}

	return false
}

func (config *Config) requireAuthenticatedUser(w http.ResponseWriter, r *http.Request) *dbmodel.User {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return nil
	}
	return user
}

func (config *Config) requireRambleRegistrationAccess(w http.ResponseWriter, r *http.Request, user *dbmodel.User, rambleID uint, blanketPerm, selfPerm string) bool {
	allowed, err := hasBlanketOrOwnsRamble(user, config.GuideRepository, rambleID, blanketPerm, selfPerm)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return false
	}

	if !allowed {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return false
	}

	return true
}

// resolveGuideRambleScope returns ramble IDs a self-scoped guide may access.
// When rambleID is set, the slice contains only that ID if owned (or forbidden).
// When rambleID is nil, all assigned ramble IDs are returned (may be empty).
// ok is false when the caller should stop (403 already rendered, or missing self perm).
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

	return ownedIDs, true
}
