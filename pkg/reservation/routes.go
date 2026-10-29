package reservation

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

	router.Route("/admin", func(r chi.Router) {
		r.Use(authentication.RequireAuthentication())
		r.Get("/", config.AdminList)
		r.Put("/{id}/status", config.AdminUpdateStatus)
		r.Delete("/{id}", config.AdminDelete)
		r.Post("/bulk-action", config.AdminBulkAction)
	})

	router.Post("/", config.Create)
	router.Get("/", config.GetUserReservations)
	router.Get("/{id}", config.Get)
	router.Put("/{id}/confirm", config.Confirm)
	router.Put("/{id}/cancel", config.Cancel)
	router.Post("/{id}/participants", config.AddParticipants)
	router.Put("/{id}/participants/{participantId}", config.UpdateParticipant)
	router.Delete("/{id}/participants/{participantId}", config.RemoveParticipant)

	return router
}
