package payment

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/payments"
)

type Controller struct {
	paymentService payments.PaymentService
}

func NewController(paymentService payments.PaymentService) *Controller {
	return &Controller{
		paymentService: paymentService,
	}
}

// GetPayment retrieves a payment by ID
// @Summary Get payment by ID
// @Description Get payment details by ID
// @Tags payments
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} model.Payment
// @Failure 400 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /payments/{id} [get]
func (pc *Controller) GetPayment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	payment, err := pc.paymentService.GetPaymentByID(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if payment == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	render.JSON(w, r, payment)
}

// RefundPayment processes a refund for a payment
// @Summary Refund payment
// @Description Process a refund for a payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param payload body model.PaymentRefundPayload true "Refund payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /payments/{id}/refund [post]
func (pc *Controller) RefundPayment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	var payload model.PaymentRefundPayload
	if err := render.Decode(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	err = pc.paymentService.RefundPayment(uint(id), payload.Amount, payload.Reason)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, map[string]string{"message": "Refund processed successfully"})
}

// WebhookHandler handles Stripe webhook events
// @Summary Handle Stripe webhook
// @Description Handle Stripe webhook events for payment processing
// @Tags payments
// @Accept json
// @Produce json
// @Param guide_id query int true "Guide ID"
// @Param Stripe-Signature header string true "Stripe signature"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /payments/webhook [post]
func (pc *Controller) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Get guide ID from query parameter
	guideIDStr := r.URL.Query().Get("guide_id")
	if guideIDStr == "" {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("guide_id is required")))
		return
	}

	guideID, err := strconv.ParseUint(guideIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Get the Stripe signature
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("missing Stripe-Signature header")))
		return
	}

	// Verify and process webhook
	err = pc.paymentService.VerifyWebhookAndProcess(payload, signature, uint(guideID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, map[string]string{"message": "Webhook processed successfully"})
}
