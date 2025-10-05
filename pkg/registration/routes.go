package registration

import (
	"feldrise.com/balade/config"
	"feldrise.com/balade/pkg/authentication"
	"github.com/go-chi/chi"
)

func New(configuration *config.Config) *Config {
	return &Config{configuration}
}

func (config *Config) Routes() *chi.Mux {
	router := chi.NewRouter()

	// Public/user routes
	router.Post("/", config.Create)
	router.Get("/{id}", config.Get)
	router.Get("/", config.GetUserRegistrations)
	router.Put("/{id}/confirm", config.Confirm)
	router.Put("/{id}/cancel", config.Cancel)
	router.Get("/ramble/{rambleId}", config.GetRambleRegistrations)

	// Admin routes
	router.Route("/admin", func(r chi.Router) {
		r.Use(authentication.RequireAuthentication())
		r.Get("/", config.AdminGetAllRegistrations)
		r.Put("/{id}", config.AdminUpdateRegistration)
		r.Put("/{id}/status", config.AdminUpdateRegistrationStatus)
		r.Delete("/{id}", config.AdminDeleteRegistration)
		r.Post("/bulk-action", config.AdminBulkRegistrationAction)
	})

	return router
}
