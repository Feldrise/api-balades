package payment

import (
	"feldrise.com/balade/config"
	"feldrise.com/balade/pkg/authentication"
	"github.com/go-chi/chi"
)

func Routes(configuration *config.Config) chi.Router {
	r := chi.NewRouter()

	controller := NewController(configuration)

	// Payment endpoints
	r.Get("/{id}", controller.GetPayment)
	r.With(authentication.RequireAnyPermission("refund:payments", "manage:payments")).Post("/{id}/refund", controller.RefundPayment)
	// Public: the Stripe signature authenticates the webhook.
	r.Post("/webhook", controller.WebhookHandler)

	return r
}
