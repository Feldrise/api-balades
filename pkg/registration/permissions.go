package registration

import (
	"net/http"

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
