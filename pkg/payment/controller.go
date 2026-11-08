package payment

import (
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"feldrise.com/balade/config"
	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/payments"
)

type Controller struct {
	paymentService        payments.PaymentService
	reservationRepository dbmodel.ReservationRepository
	guideRepository       dbmodel.GuideRepository
}

func NewController(configuration *config.Config) *Controller {
	return &Controller{
		paymentService:        configuration.PaymentService,
		reservationRepository: configuration.ReservationRepository,
		guideRepository:       configuration.GuideRepository,
	}
}

// canView reports whether user may read the payment: the booker of its
// reservation, an admin or guide allowed to view the registrations of the
// reservation's ramble, or staff allowed to view every payment.
func (pc *Controller) canView(user *dbmodel.User, payment *model.Payment) (bool, error) {
	if user.HasPermission("view:payments") || user.HasPermission("manage:payments") {
		return true, nil
	}
	if payment.ReservationID == nil {
		return false, nil
	}

	reservation, err := pc.reservationRepository.FindByID(*payment.ReservationID)
	if err != nil || reservation == nil {
		return false, err
	}

	if reservation.UserID == user.ID || user.HasPermission("view:all-registrations") {
		return true, nil
	}
	if user.HasPermission("view:registrations:self") {
		return pc.guideRepository.UserOwnsRamble(user.ID, reservation.RambleID)
	}
	return false, nil
}

// GetPayment retrieves a payment by ID
// @Summary Get payment by ID
// @Description Get payment details by ID, for the booker of its reservation or authorized staff
// @Tags payments
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} model.Payment
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /payments/{id} [get]
func (pc *Controller) GetPayment(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

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

	allowed, err := pc.canView(user, payment)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}
	if !allowed {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	render.JSON(w, r, payment)
}

// RefundPayment processes a refund for a payment
// @Summary Refund payment
// @Description Process a refund for a payment (refund:payments or manage:payments)
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param payload body model.PaymentRefundPayload true "Refund payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.ErrResponse "reason: invalid_refund_amount (an amount below 1 cent)"
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "reason: refund_exceeds_refundable (more than is left to refund, nothing refunded), payment_not_refundable (the payment is not succeeded)"
// @Failure 422 {object} errors.ErrResponse "reason: refund_refused (Stripe refused the refund, nothing refunded)"
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
		render.Render(w, r, refundError(err))
		return
	}

	render.JSON(w, r, map[string]string{"message": "Refund processed successfully"})
}

// refundError reports a refund refused with the reason why.
func refundError(err error) render.Renderer {
	switch {
	case stderrors.Is(err, payments.ErrInvalidRefundAmount):
		return errors.ErrBusiness(http.StatusBadRequest, "invalid_refund_amount", err.Error())
	case stderrors.Is(err, payments.ErrRefundExceedsRefundable):
		return errors.ErrBusiness(http.StatusConflict, "refund_exceeds_refundable", err.Error())
	case stderrors.Is(err, payments.ErrPaymentNotRefundable):
		return errors.ErrBusiness(http.StatusConflict, "payment_not_refundable", err.Error())
	case stderrors.Is(err, payments.ErrRefundRefused):
		return errors.ErrBusiness(http.StatusUnprocessableEntity, "refund_refused", err.Error())
	}
	return errors.ErrServerError(err)
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
