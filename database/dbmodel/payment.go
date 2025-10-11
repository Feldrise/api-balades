package dbmodel

import (
	"context"
	"errors"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model

	// Stripe identifiers
	StripePaymentIntentID string  `gorm:"not null;unique;index"`
	StripeChargeID        *string `gorm:"index"`

	// Payment details
	Amount        int64   `gorm:"not null"`                   // Amount in cents
	Currency      string  `gorm:"not null;default:'eur'"`     // Currency code
	Status        string  `gorm:"not null;default:'pending'"` // pending, succeeded, failed, cancelled, refunded
	PaymentMethod string  `gorm:"default:'card'"`             // card, sepa, etc.
	FailureReason *string `gorm:""`                           // Reason for payment failure

	// Registration relationship
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

	// Foreign Objects
	Registration *RambleRegistration      `gorm:"foreignKey:RegistrationID"`
	Group        *RambleRegistrationGroup `gorm:"foreignKey:GroupID"`
	Guide        Guide                    `gorm:"foreignKey:GuideID"`
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

	if p.Registration != nil {
		regModel := p.Registration.ToModel()
		payment.Registration = regModel
	}

	if p.Group != nil {
		groupModel := p.Group.ToModel()
		payment.Group = groupModel
	}

	return payment
}

type PaymentFilter struct {
	RegistrationID *uint
	GroupID        *uint
	GuideID        *uint
	Status         *string
	Statuses       []string
	PayerEmail     *string
	DateFrom       *time.Time
	DateTo         *time.Time
	AmountMin      *int64
	AmountMax      *int64
	Currency       *string
	PaymentMethod  *string
	Limit          *int
	Offset         *int
	SortBy         *string // "created_at", "amount", "paid_at"
	SortOrder      *string // "asc", "desc"
}

type PaymentRepository interface {
	FindByID(id uint) (*Payment, error)
	FindByStripePaymentIntentID(paymentIntentID string) (*Payment, error)
	FindAll(filter *PaymentFilter) ([]Payment, error)
	CountAll(filter *PaymentFilter) (int64, error)
	Create(payment *Payment) (*Payment, error)
	Update(payment *Payment) (*Payment, error)
	Delete(id uint) error
	FindByRegistrationID(registrationID uint) ([]Payment, error)
	FindByGroupID(groupID uint) ([]Payment, error)
	GetTotalAmountByGuide(guideID uint, status string) (int64, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) FindByID(id uint) (*Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payment Payment
	err := r.db.WithContext(ctx).
		Preload("Registration").
		Preload("Group").
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
		Preload("Registration").
		Preload("Group").
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
		if filter.RegistrationID != nil {
			tx = tx.Where("registration_id = ?", *filter.RegistrationID)
		}
		if filter.GroupID != nil {
			tx = tx.Where("group_id = ?", *filter.GroupID)
		}
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

	err := tx.Preload("Registration").
		Preload("Group").
		Preload("Guide").
		Find(&payments).Error

	return payments, err
}

func (r *paymentRepository) CountAll(filter *PaymentFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	tx := r.db.WithContext(ctx).Model(&Payment{})

	if filter != nil {
		if filter.RegistrationID != nil {
			tx = tx.Where("registration_id = ?", *filter.RegistrationID)
		}
		if filter.GroupID != nil {
			tx = tx.Where("group_id = ?", *filter.GroupID)
		}
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

func (r *paymentRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&Payment{}, id).Error
}

func (r *paymentRepository) FindByRegistrationID(registrationID uint) ([]Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payments []Payment
	err := r.db.WithContext(ctx).
		Preload("Guide").
		Where("registration_id = ?", registrationID).
		Order("created_at DESC").
		Find(&payments).Error

	return payments, err
}

func (r *paymentRepository) FindByGroupID(groupID uint) ([]Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payments []Payment
	err := r.db.WithContext(ctx).
		Preload("Guide").
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Find(&payments).Error

	return payments, err
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
