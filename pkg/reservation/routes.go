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

	return router
}
