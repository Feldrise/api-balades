package authentication

import (
	"net/http"
	"strconv"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

const GuideRoleName = "guide"

// IsGuide returns true if the user has the guide role.
func IsGuide(user *dbmodel.User) bool {
	if user == nil {
		return false
	}

	for _, role := range user.Roles {
		if role.Name == GuideRoleName {
			return true
		}
	}

	return false
}

// GrantGuideRole assigns the guide role to a user if they do not already have it.
func GrantGuideRole(userRepo dbmodel.UserRepository, roleRepo dbmodel.RoleRepository, user *dbmodel.User) error {
	if IsGuide(user) {
		return nil
	}

	role, err := roleRepo.FindByName(GuideRoleName)
	if err != nil {
		return err
	}

	if role == nil {
		return gorm.ErrRecordNotFound
	}

	user.Roles = append(user.Roles, *role)

	_, err = userRepo.Update(user)
	return err
}

// RequirePermission is a middleware that checks if the user has the required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := ForContext(r.Context())
			if user == nil {
				render.Render(w, r, errors.ErrUnauthorized("authentication required"))
				return
			}

			if !user.HasPermission(permission) {
				render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission is a middleware that checks if the user has any of the required permissions
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := ForContext(r.Context())
			if user == nil {
				render.Render(w, r, errors.ErrUnauthorized("authentication required"))
				return
			}

			hasPermission := false
			for _, permission := range permissions {
				if user.HasPermission(permission) {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRamblePermission allows access when the user has a blanket permission,
// or when they have a self permission and own the ramble as a linked guide.
func RequireRamblePermission(guideRepo dbmodel.GuideRepository, blanketPermission, selfPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := ForContext(r.Context())
			if user == nil {
				render.Render(w, r, errors.ErrUnauthorized("authentication required"))
				return
			}

			if user.HasPermission(blanketPermission) {
				next.ServeHTTP(w, r)
				return
			}

			if user.HasPermission(selfPermission) {
				idStr := chi.URLParam(r, "id")
				idUint, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					render.Render(w, r, errors.ErrInvalidRequest(err))
					return
				}

				owns, err := guideRepo.UserOwnsRamble(user.ID, uint(idUint))
				if err != nil {
					render.Render(w, r, errors.ErrServerError(err))
					return
				}

				if owns {
					next.ServeHTTP(w, r)
					return
				}
			}

			render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		})
	}
}

// RequireAuthentication is a middleware that checks if the user is authenticated
func RequireAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := ForContext(r.Context())
			if user == nil {
				render.Render(w, r, errors.ErrUnauthorized("authentication required"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
