package reservation

import (
	"feldrise.com/balade/config"
	"github.com/go-chi/chi"
)

func New(configuration *config.Config) *Config {
	return &Config{configuration}
}

func (config *Config) Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", config.Create)
	router.Get("/", config.GetUserReservations)
	router.Get("/{id}", config.Get)
	router.Put("/{id}/cancel", config.Cancel)
	router.Post("/{id}/participants", config.AddParticipants)
	router.Put("/{id}/participants/{participantId}", config.UpdateParticipant)
	router.Delete("/{id}/participants/{participantId}", config.RemoveParticipant)

	return router
}
