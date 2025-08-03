package authentication

import (
	"context"
	"net/http"

	"feldrise.com/balade/config"
	"feldrise.com/balade/database/dbmodel"
)

var UserCtxKey = contextKey{"user"}

type contextKey struct {
	name string
}

func Middelware(c *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			// We get the user from the token
			tokenString := header
			userID, err := ParseToken(c.Constants.JWTSecret, tokenString)

			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			user, err := c.UserRepository.FindByID(userID, &dbmodel.UserFieldsToInclude{
				UserProfile:       true,
				Roles:             true,
				Roles_Permissions: true,
			})

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// We add the user to the context
			ctx := context.WithValue(r.Context(), UserCtxKey, user)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) *dbmodel.User {
	raw, _ := ctx.Value(UserCtxKey).(*dbmodel.User)
	return raw
}
