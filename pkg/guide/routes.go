package guide

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

	router.With(authentication.RequireAuthentication()).Get("/me", config.GetMe)
	router.With(authentication.RequireAuthentication()).Put("/me", config.UpdateMe)
	router.With(authentication.RequirePermission("configure:guide-payments:self")).Put("/me/payment-config", config.UpdateMyPaymentConfig)

	router.Get("/{id}", config.Get)
	router.Get("/", config.GetAll)
	router.Post("/", config.Create)
	router.Put("/{id}", config.Update)
	router.Delete("/{id}", config.Delete)
	router.Put("/{id}/payment-config", config.UpdatePaymentConfig)
	router.With(authentication.RequirePermission("create:user")).Post("/{id}/link-user", config.LinkUser)

	return router
}
