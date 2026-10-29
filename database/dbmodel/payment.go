package dbmodel

import (
	"context"
	"errors"
	"slices"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PaymentTimeouts bound the decisions on payments, which call Stripe while the
// reservation or the payment is locked. A zero duration is the default one.
type PaymentTimeouts struct {
	// Decision bounds the transaction of a decision (ForReservation, ForPayment).
	Decision time.Duration
	// Reconciliation bounds the transaction of a reconciliation in the
	// background (ForIdleReservation).
	Reconciliation time.Duration
}

// Default PaymentTimeouts. A reconciliation holds the reservation for less
// time than an operation on the places of the ramble waits for it.
const (
	defaultPaymentDecisionTimeout       = 30 * time.Second
	defaultPaymentReconciliationTimeout = 3 * time.Second
)

// decisionContext is the context of the external calls made by a decision
// bounded by timeout: it ends before the transaction, with a sixth of timeout
// left to record what Stripe answered, so that no call outlives the lock.
func decisionContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout-timeout/6)
}

type Payment struct {
	gorm.Model

	// Stripe identifiers
	StripePaymentIntentID string  `gorm:"not null;unique;index"`
	StripeChargeID        *string `gorm:"index"`

	// Payment details
	Amount        int64   `gorm:"not null"`                   // Amount in cents
	Currency      string  `gorm:"not null;default:'eur'"`     // Currency code
	Status        string  `gorm:"not null;default:'pending'"` // pending, requires_action, succeeded, failed, cancelled, partially_refunded, refunded
	PaymentMethod string  `gorm:"default:'card'"`             // card, sepa, etc.
	FailureReason *string `gorm:""`                           // Reason for payment failure

	// Reservation relationship
	ReservationID *uint `gorm:"index"`

	// Legacy registration relationship
	RegistrationID *uint `gorm:"index"`
	GroupID        *uint `gorm:"index"`

	// Payer information
	PayerEmail string `gorm:"not null"`
	PayerName  string `gorm:"not null"`

	// Guide receiving payment
	GuideID uint `gorm:"not null;index"`

	// Timestamps
	PaidAt       *time.Time
	RefundedAt   *time.Time
	RefundAmount *int64 // Amount refunded in cents

	// Refund requested from Stripe and not known made yet: its amount in cents
	// and the idempotency key that identifies it on Stripe, recorded before
	// the call so that retrying it reuses the key whatever is refunded since.
	PendingRefundAmount *int64
	PendingRefundKey    *string

	// Foreign Objects
	Guide Guide `gorm:"foreignKey:GuideID"`
}

// CollectedPaymentStatuses are the statuses of a payment paid, whether it has
// been refunded since or not.
var CollectedPaymentStatuses = []string{"succeeded", "partially_refunded", "refunded"}

// IsCollected reports whether the payment was paid, whether it has been refunded since or not.
func (p *Payment) IsCollected() bool {
	return slices.Contains(CollectedPaymentStatuses, p.Status)
}

// RefundedAmount returns the amount in cents refunded so far. A payment
// refunded without the amount refunded recorded is refunded entirely.
func (p *Payment) RefundedAmount() int64 {
	if p.RefundAmount == nil {
		if p.Status == "refunded" {
			return p.Amount
		}
		return 0
	}
	return *p.RefundAmount
}

// NetAmount returns the amount in cents collected and kept: the amount paid
// minus the refunds, 0 when the payment was not paid.
func (p *Payment) NetAmount() int64 {
	if !p.IsCollected() {
		return 0
	}
	return max(p.Amount-p.RefundedAmount(), 0)
}

// OpenPaymentStatuses are the statuses of a payment that can still be paid on
// Stripe: neither collected nor cancelled. A declined payment (failed) stays
// open, as Stripe lets the payer try again with another payment method.
var OpenPaymentStatuses = []string{"pending", "requires_action", "failed"}

// IsOpen reports whether the payment can still be paid on Stripe.
func (p *Payment) IsOpen() bool {
	return slices.Contains(OpenPaymentStatuses, p.Status)
}

func (p Payment) ToModel() model.Payment {
	payment := model.Payment{
		ID:                    p.ID,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
		StripePaymentIntentID: p.StripePaymentIntentID,
		StripeChargeID:        p.StripeChargeID,
		Amount:                p.Amount,
		Currency:              p.Currency,
		Status:                p.Status,
		PaymentMethod:         p.PaymentMethod,
		FailureReason:         p.FailureReason,
		ReservationID:         p.ReservationID,
		RegistrationID:        p.RegistrationID,
		GroupID:               p.GroupID,
		PayerEmail:            p.PayerEmail,
		PayerName:             p.PayerName,
		GuideID:               p.GuideID,
		PaidAt:                p.PaidAt,
		RefundedAt:            p.RefundedAt,
		RefundAmount:          p.RefundAmount,
		Guide:                 p.Guide.ToModel(),
	}

	return payment
}

type PaymentFilter struct {
	GuideID       *uint
	Status        *string
	Statuses      []string
	PayerEmail    *string
	DateFrom      *time.Time
	DateTo        *time.Time
	AmountMin     *int64
	AmountMax     *int64
	Currency      *string
	PaymentMethod *string
	Limit         *int
	Offset        *int
	SortBy        *string // "created_at", "amount", "paid_at"
	SortOrder     *string // "asc", "desc"
}

type PaymentRepository interface {
	FindByID(id uint) (*Payment, error)
	FindByStripePaymentIntentID(paymentIntentID string) (*Payment, error)
	FindAll(filter *PaymentFilter) ([]Payment, error)
	CountAll(filter *PaymentFilter) (int64, error)
	Create(payment *Payment) (*Payment, error)
	Update(payment *Payment) (*Payment, error)
	// UpdateStatus saves the status of the payment only.
	UpdateStatus(id uint, status string) error
	// UpdateByStripePaymentIntentID applies changes to the payment of the
	// Stripe payment intent in a single statement, unless its status is one of
	// kept, as it is when the statement runs. It reports whether a payment was
	// changed.
	UpdateByStripePaymentIntentID(paymentIntentID string, kept []string, changes map[string]any) (bool, error)
	// RecordRefundedAmount records, in a single statement, that refunded cents
	// of the payment of the Stripe payment intent are refunded so far, all
	// refunds included, and marks it partially_refunded, or refunded once
	// refunded entirely, unless it records as much refunded already: the amount
	// refunded recorded never decreases. The statement waits for the payment
	// locked by ForPayment. It reports whether a payment was changed.
	RecordRefundedAmount(paymentIntentID string, refunded int64, at time.Time) (bool, error)
	// SetPendingRefund records the refund of amount cents requested from
	// Stripe with the idempotency key, or clears it when amount is nil.
	SetPendingRefund(id uint, amount *int64, key *string) error
	Delete(id uint) error
	GetTotalAmountByGuide(guideID uint, status string) (int64, error)
	// ForReservation locks the reservation and runs decide in a transaction with
	// the reservation as it is once locked, with its Ramble, Participants and
	// Payments, and a repository bound to the transaction. The payments of a
	// reservation are so decided one at a time. A deleted reservation comes with
	// its payments only. decide makes its calls to Stripe with ctx, which ends
	// before the transaction.
	ForReservation(reservationID uint, decide func(ctx context.Context, payments PaymentRepository, reservation *Reservation) error) error
	// ForIdleReservation is ForReservation for a reconciliation in the
	// background, which never makes an operation of the user wait: it skips the
	// reservation, without running decide, while another operation holds it,
	// and its transaction is bounded by PaymentTimeouts.Reconciliation.
	ForIdleReservation(reservationID uint, decide func(ctx context.Context, payments PaymentRepository, reservation *Reservation) error) error
	// ForPayment locks the payment and runs decide in a transaction with the
	// payment as it is once locked, with its Guide, and a repository bound to
	// the transaction. The refunds of a payment are so decided one at a time.
	// decide gets nil when the payment does not exist. decide makes its calls
	// to Stripe with ctx, which ends before the transaction.
	ForPayment(paymentID uint, decide func(ctx context.Context, payments PaymentRepository, payment *Payment) error) error
}

type paymentRepository struct {
	db       *gorm.DB
	timeouts PaymentTimeouts
}

func NewPaymentRepository(db *gorm.DB, timeouts PaymentTimeouts) PaymentRepository {
	if timeouts.Decision == 0 {
		timeouts.Decision = defaultPaymentDecisionTimeout
	}
	if timeouts.Reconciliation == 0 {
		timeouts.Reconciliation = defaultPaymentReconciliationTimeout
	}
	return &paymentRepository{db: db, timeouts: timeouts}
}

func (r *paymentRepository) FindByID(id uint) (*Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payment Payment
	err := r.db.WithContext(ctx).
		Preload("Guide").
		First(&payment, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err // Other error
	}

	return &payment, nil
}

func (r *paymentRepository) FindByStripePaymentIntentID(paymentIntentID string) (*Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payment Payment
	err := r.db.WithContext(ctx).
		Preload("Guide").
		Where("stripe_payment_intent_id = ?", paymentIntentID).
		First(&payment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err // Other error
	}

	return &payment, nil
}

func (r *paymentRepository) FindAll(filter *PaymentFilter) ([]Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payments []Payment
	tx := r.db.WithContext(ctx).Model(&Payment{})

	if filter != nil {
		if filter.GuideID != nil {
			tx = tx.Where("guide_id = ?", *filter.GuideID)
		}
		if filter.Status != nil {
			tx = tx.Where("status = ?", *filter.Status)
		}
		if len(filter.Statuses) > 0 {
			tx = tx.Where("status IN ?", filter.Statuses)
		}
		if filter.PayerEmail != nil {
			tx = tx.Where("payer_email ILIKE ?", "%"+*filter.PayerEmail+"%")
		}
		if filter.DateFrom != nil {
			tx = tx.Where("created_at >= ?", *filter.DateFrom)
		}
		if filter.DateTo != nil {
			tx = tx.Where("created_at <= ?", *filter.DateTo)
		}
		if filter.AmountMin != nil {
			tx = tx.Where("amount >= ?", *filter.AmountMin)
		}
		if filter.AmountMax != nil {
			tx = tx.Where("amount <= ?", *filter.AmountMax)
		}
		if filter.Currency != nil {
			tx = tx.Where("currency = ?", *filter.Currency)
		}
		if filter.PaymentMethod != nil {
			tx = tx.Where("payment_method = ?", *filter.PaymentMethod)
		}

		// Sorting
		sortBy := "created_at"
		sortOrder := "DESC"
		if filter.SortBy != nil && *filter.SortBy != "" {
			validSortFields := map[string]bool{
				"created_at": true,
				"amount":     true,
				"paid_at":    true,
			}
			if validSortFields[*filter.SortBy] {
				sortBy = *filter.SortBy
			}
		}
		if filter.SortOrder != nil && (*filter.SortOrder == "asc" || *filter.SortOrder == "desc") {
			if *filter.SortOrder == "asc" {
				sortOrder = "ASC"
			}
		}
		tx = tx.Order(sortBy + " " + sortOrder)

		// Pagination
		if filter.Limit != nil && *filter.Limit > 0 {
			tx = tx.Limit(*filter.Limit)
		}
		if filter.Offset != nil && *filter.Offset > 0 {
			tx = tx.Offset(*filter.Offset)
		}
	} else {
		tx = tx.Order("created_at DESC")
	}

	err := tx.Preload("Guide").
		Find(&payments).Error

	return payments, err
}

func (r *paymentRepository) CountAll(filter *PaymentFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	tx := r.db.WithContext(ctx).Model(&Payment{})

	if filter != nil {
		if filter.GuideID != nil {
			tx = tx.Where("guide_id = ?", *filter.GuideID)
		}
		if filter.Status != nil {
			tx = tx.Where("status = ?", *filter.Status)
		}
		if len(filter.Statuses) > 0 {
			tx = tx.Where("status IN ?", filter.Statuses)
		}
		if filter.PayerEmail != nil {
			tx = tx.Where("payer_email ILIKE ?", "%"+*filter.PayerEmail+"%")
		}
		if filter.DateFrom != nil {
			tx = tx.Where("created_at >= ?", *filter.DateFrom)
		}
		if filter.DateTo != nil {
			tx = tx.Where("created_at <= ?", *filter.DateTo)
		}
		if filter.AmountMin != nil {
			tx = tx.Where("amount >= ?", *filter.AmountMin)
		}
		if filter.AmountMax != nil {
			tx = tx.Where("amount <= ?", *filter.AmountMax)
		}
		if filter.Currency != nil {
			tx = tx.Where("currency = ?", *filter.Currency)
		}
		if filter.PaymentMethod != nil {
			tx = tx.Where("payment_method = ?", *filter.PaymentMethod)
		}
	}

	err := tx.Count(&count).Error
	return count, err
}

func (r *paymentRepository) Create(payment *Payment) (*Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(payment).Error
	return payment, err
}

func (r *paymentRepository) Update(payment *Payment) (*Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Save(payment).Error
	return payment, err
}

func (r *paymentRepository) UpdateStatus(id uint, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *paymentRepository) UpdateByStripePaymentIntentID(paymentIntentID string, kept []string, changes map[string]any) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Model(&Payment{}).
		Where("stripe_payment_intent_id = ? AND status NOT IN ?", paymentIntentID, kept).
		Updates(changes)
	return result.RowsAffected > 0, result.Error
}

func (r *paymentRepository) RecordRefundedAmount(paymentIntentID string, refunded int64, at time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// The amount refunded recorded is RefundedAmount, in SQL.
	result := r.db.WithContext(ctx).Model(&Payment{}).
		Where("stripe_payment_intent_id = ?", paymentIntentID).
		Where("COALESCE(refund_amount, CASE WHEN status = 'refunded' THEN amount ELSE 0 END) < ?", refunded).
		Updates(map[string]any{
			"refund_amount": refunded,
			"refunded_at":   at,
			"paid_at":       gorm.Expr("COALESCE(paid_at, ?)", at),
			"status":        gorm.Expr("CASE WHEN ? >= amount THEN 'refunded' ELSE 'partially_refunded' END", refunded),
		})
	return result.RowsAffected > 0, result.Error
}

func (r *paymentRepository) SetPendingRefund(id uint, amount *int64, key *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Updates(map[string]any{
		"pending_refund_amount": amount,
		"pending_refund_key":    key,
	}).Error
}

func (r *paymentRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&Payment{}, id).Error
}

func (r *paymentRepository) GetTotalAmountByGuide(guideID uint, status string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int64
	err := r.db.WithContext(ctx).Model(&Payment{}).
		Where("guide_id = ? AND status = ?", guideID, status).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error

	return total, err
}

func (r *paymentRepository) ForReservation(reservationID uint, decide func(ctx context.Context, payments PaymentRepository, reservation *Reservation) error) error {
	return r.forReservation(reservationID, r.timeouts.Decision, false, decide)
}

func (r *paymentRepository) ForIdleReservation(reservationID uint, decide func(ctx context.Context, payments PaymentRepository, reservation *Reservation) error) error {
	return r.forReservation(reservationID, r.timeouts.Reconciliation, true, decide)
}

// forReservation locks the reservation, unless skipBusy and another
// transaction holds it, and runs decide in a transaction bounded by timeout.
func (r *paymentRepository) forReservation(reservationID uint, timeout time.Duration, skipBusy bool, decide func(ctx context.Context, payments PaymentRepository, reservation *Reservation) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	callCtx, cancelCalls := decisionContext(ctx, timeout)
	defer cancelCalls()

	locking := clause.Locking{Strength: "UPDATE"}
	if skipBusy {
		locking.Options = "SKIP LOCKED"
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reservation Reservation
		err := tx.Unscoped().Clauses(locking).First(&reservation, reservationID).Error
		if skipBusy && errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		if reservation.DeletedAt.Valid {
			if err := tx.Where("reservation_id = ?", reservationID).Order("created_at ASC").Find(&reservation.Payments).Error; err != nil {
				return err
			}
		} else if err := preloadReservation(tx).First(&reservation, reservationID).Error; err != nil {
			return err
		}

		return decide(callCtx, &paymentRepository{db: tx, timeouts: r.timeouts}, &reservation)
	})
}

func (r *paymentRepository) ForPayment(paymentID uint, decide func(ctx context.Context, payments PaymentRepository, payment *Payment) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeouts.Decision)
	defer cancel()
	callCtx, cancelCalls := decisionContext(ctx, r.timeouts.Decision)
	defer cancelCalls()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payment Payment
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, paymentID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return decide(callCtx, &paymentRepository{db: tx, timeouts: r.timeouts}, nil)
		}
		if err != nil {
			return err
		}
		if err := tx.First(&payment.Guide, payment.GuideID).Error; err != nil {
			return err
		}
		return decide(callCtx, &paymentRepository{db: tx, timeouts: r.timeouts}, &payment)
	})
}
