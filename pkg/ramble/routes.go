package ramble

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
	router.Put("/{id}/cancel", config.Cancel)
	router.Delete("/{id}", config.Delete)

	return router
}
