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

type PayerInfo struct {
	Email string
	Name  string
}

type Controller struct {
	paymentService payments.PaymentService
}

func NewController(paymentService payments.PaymentService) *Controller {
	return &Controller{
		paymentService: paymentService,
	}
}

// CreatePayment creates a new payment intent for a registration or group
// @Summary Create payment intent
// @Description Create a new payment intent for a registration or group
// @Tags payments
// @Accept json
// @Produce json
// @Param payload body model.PaymentCreatePayload true "Payment creation payload"
// @Success 200 {object} model.PaymentResponse
// @Failure 400 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /payments [post]
func (pc *Controller) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var payload model.PaymentCreatePayload
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	payerInfo := payments.PayerInfo{
		Email: payload.PayerEmail,
		Name:  payload.PayerName,
	}

	var response *model.PaymentResponse
	var err error

	if payload.RegistrationID != nil {
		response, err = pc.paymentService.CreatePaymentForRegistration(*payload.RegistrationID, payload.PriceLabel, payerInfo, payload.ReturnURL)
	} else if payload.GroupID != nil {
		response, err = pc.paymentService.CreatePaymentForGroup(*payload.GroupID, payload.PriceLabel, payerInfo, payload.ReturnURL, payload.PriceSelections)
	} else {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("either registration_id or group_id is required")))
		return
	}

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, response)
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

// GetRegistrationPayments gets all payments for a registration
// @Summary Get payments for registration
// @Description Get all payments associated with a registration
// @Tags payments
// @Produce json
// @Param registration_id path int true "Registration ID"
// @Success 200 {array} model.Payment
// @Failure 400 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/{registration_id}/payments [get]
func (pc *Controller) GetRegistrationPayments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "registration_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	payments, err := pc.paymentService.GetPaymentsByRegistration(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, payments)
}

// GetGroupPayments gets all payments for a group
// @Summary Get payments for group
// @Description Get all payments associated with a group
// @Tags payments
// @Produce json
// @Param group_id path int true "Group ID"
// @Success 200 {array} model.Payment
// @Failure 400 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /groups/{group_id}/payments [get]
func (pc *Controller) GetGroupPayments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "group_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	payments, err := pc.paymentService.GetPaymentsByGroup(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, payments)
}
