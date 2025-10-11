package guide

import (
	"feldrise.com/balade/config"
	"github.com/go-chi/chi"
)

func New(configuration *config.Config) *Config {
	return &Config{configuration}
}

func (config *Config) Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/{id}", config.Get)
	router.Get("/", config.GetAll)
	router.Post("/", config.Create)
	router.Put("/{id}", config.Update)
	router.Delete("/{id}", config.Delete)
	router.Put("/{id}/payment-config", config.UpdatePaymentConfig)

	return router
}
