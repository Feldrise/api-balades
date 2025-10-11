package registration

import (
	"feldrise.com/balade/config"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/payment"
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

	// Group routes
	router.Route("/groups", func(r chi.Router) {
		r.Get("/{id}", config.GetGroup)
		r.Get("/ramble/{rambleId}", config.GetGroupsByRamble)
		r.Put("/{id}/confirm", config.ConfirmGroup)
		r.Put("/{id}/cancel", config.CancelGroup)
	})

	// Admin routes
	router.Route("/admin", func(r chi.Router) {
		r.Use(authentication.RequireAuthentication())
		r.Get("/", config.AdminGetAllRegistrations)
		r.Put("/{id}", config.AdminUpdateRegistration)
		r.Put("/{id}/status", config.AdminUpdateRegistrationStatus)
		r.Delete("/{id}", config.AdminDeleteRegistration)
		r.Post("/bulk-action", config.AdminBulkRegistrationAction)

		// Admin group routes
		r.Route("/groups", func(gr chi.Router) {
			gr.Get("/", config.AdminGetAllGroups)
			gr.Put("/{id}/status", config.AdminUpdateGroupStatus)
			gr.Delete("/{id}", config.AdminDeleteGroup)
		})
	})

	// Payment routes
	paymentController := payment.NewController(config.PaymentService)

	router.Get("/{registration_id}/payments", paymentController.GetRegistrationPayments)

	return router
}
