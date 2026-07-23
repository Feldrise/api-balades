package dbmodel

import (
	"context"
	"errors"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
)

type Guide struct {
	gorm.Model

	UserID *uint `gorm:"uniqueIndex"`
	User   *User `gorm:"foreignKey:UserID"`

	FirstName             string `gorm:"not null"`
	LastName              string `gorm:"not null"`
	Email                 string `gorm:"not null;unique"`
	Phone                 *string
	Bio                   *string
	Experience            *string
	Specialties           *string
	Languages             *string
	CertificationLevel    *string
	Avatar                *string
	IsActive              bool `gorm:"not null;default:true"`
	EmergencyContactName  *string
	EmergencyContactPhone *string

	// Payment configuration
	StripeAccountID     *string `gorm:"column:stripe_account_id"`
	StripePublicKey     *string `gorm:"column:stripe_public_key"`
	StripeSecretKey     *string `gorm:"column:stripe_secret_key"`     // Should be encrypted
	StripeWebhookSecret *string `gorm:"column:stripe_webhook_secret"` // Should be encrypted
	PaymentEnabled      bool    `gorm:"not null;default:false"`
}

func (g Guide) ToModel() model.Guide {
	return model.Guide{
		ID:                    g.ID,
		CreatedAt:             g.CreatedAt,
		UpdatedAt:             g.UpdatedAt,
		FirstName:             g.FirstName,
		LastName:              g.LastName,
		Email:                 g.Email,
		Phone:                 g.Phone,
		Bio:                   g.Bio,
		Experience:            g.Experience,
		Specialties:           g.Specialties,
		Languages:             g.Languages,
		CertificationLevel:    g.CertificationLevel,
		Avatar:                g.Avatar,
		IsActive:              g.IsActive,
		EmergencyContactName:  g.EmergencyContactName,
		EmergencyContactPhone: g.EmergencyContactPhone,
		StripeAccountID:       g.StripeAccountID,
		StripePublicKey:       g.StripePublicKey,
		PaymentEnabled:        g.PaymentEnabled,
	}
}

type GuideFilter struct {
	IsActive *bool
	Search   *string
}

type GuideRepository interface {
	FindByID(id uint) (*Guide, error)
	FindByUserID(userID uint) (*Guide, error)
	UserOwnsRamble(userID uint, rambleID uint) (bool, error)
	FindRambleIDsOwnedByUser(userID uint) ([]uint, error)
	FindAll(filter *GuideFilter) ([]Guide, error)
	Create(guide *Guide) (*Guide, error)
	Update(guide *Guide) (*Guide, error)
	Delete(id uint) error
}

type guideRepository struct {
	db *gorm.DB
}

func NewGuideRepository(db *gorm.DB) GuideRepository {
	return &guideRepository{db: db}
}

func (r *guideRepository) FindByUserID(userID uint) (*Guide, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var guide Guide
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&guide).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &guide, nil
}

func (r *guideRepository) UserOwnsRamble(userID uint, rambleID uint) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).
		Table("ramble_guides").
		Joins("JOIN guides ON guides.id = ramble_guides.guide_id").
		Where("guides.user_id = ? AND ramble_guides.ramble_id = ? AND guides.deleted_at IS NULL AND ramble_guides.deleted_at IS NULL", userID, rambleID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindRambleIDsOwnedByUser returns ramble IDs where the user's linked guide is assigned.
func (r *guideRepository) FindRambleIDsOwnedByUser(userID uint) ([]uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rambleIDs []uint
	err := r.db.WithContext(ctx).
		Table("ramble_guides").
		Joins("JOIN guides ON guides.id = ramble_guides.guide_id").
		Where("guides.user_id = ? AND guides.deleted_at IS NULL AND ramble_guides.deleted_at IS NULL", userID).
		Distinct().
		Pluck("ramble_guides.ramble_id", &rambleIDs).Error

	if err != nil {
		return nil, err
	}

	if rambleIDs == nil {
		rambleIDs = []uint{}
	}

	return rambleIDs, nil
}

func (r *guideRepository) FindByID(id uint) (*Guide, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var guide Guide
	tx := r.db.WithContext(ctx).Model(&guide)

	err := tx.First(&guide, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err // Other error
	}

	return &guide, nil
}

func (r *guideRepository) FindAll(filter *GuideFilter) ([]Guide, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var guides []Guide
	tx := r.db.WithContext(ctx).Model(&Guide{})

	if filter != nil {
		if filter.IsActive != nil {
			tx = tx.Where("is_active = ?", *filter.IsActive)
		}

		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			tx = tx.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", search, search, search)
		}
	}

	err := tx.Find(&guides).Error

	return guides, err
}

func (r *guideRepository) Create(guide *Guide) (*Guide, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(guide).Error

	return guide, err
}

func (r *guideRepository) Update(guide *Guide) (*Guide, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Save(guide).Error

	return guide, err
}

func (r *guideRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.WithContext(ctx)

	err := tx.Where("id = ?", id).Delete(&Guide{}).Error

	return err
}
