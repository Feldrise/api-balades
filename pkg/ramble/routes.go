package ramble

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

	router.Get("/{id}", config.Get)
	router.Get("/", config.GetAll)
	router.With(authentication.RequirePermission("create:ramble")).Post("/", config.Create)
	router.With(authentication.RequireRamblePermission(config.GuideRepository, "update:ramble", "update:ramble:self")).Put("/{id}", config.Update)
	router.With(authentication.RequireRamblePermission(config.GuideRepository, "update:ramble", "update:ramble:self")).Put("/{id}/cancel", config.Cancel)
	router.With(authentication.RequireRamblePermission(config.GuideRepository, "delete:ramble", "delete:ramble:self")).Delete("/{id}", config.Delete)

	return router
}
