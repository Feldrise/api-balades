package authentication

import (
	"net/http"

	"feldrise.com/balade/pkg/errors"
	"github.com/go-chi/render"
)

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
