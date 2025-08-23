package registration

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
	router.Get("/{id}", config.Get)
	router.Get("/", config.GetUserRegistrations)
	router.Put("/{id}/confirm", config.Confirm)
	router.Put("/{id}/cancel", config.Cancel)
	router.Get("/ramble/{rambleId}", config.GetRambleRegistrations)

	return router
}
