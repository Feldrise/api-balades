package payment

import (
	"feldrise.com/balade/pkg/payments"
	"github.com/go-chi/chi"
)

func Routes(paymentService payments.PaymentService) chi.Router {
	r := chi.NewRouter()

	controller := NewController(paymentService)

	// Payment endpoints
	r.Get("/{id}", controller.GetPayment)
	r.Post("/{id}/refund", controller.RefundPayment)
	r.Post("/webhook", controller.WebhookHandler)

	return r
}
