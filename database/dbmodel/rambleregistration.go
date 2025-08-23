package dbmodel

import (
	"context"
	"errors"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
)

type RambleRegistration struct {
	gorm.Model

	// User information
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	Email     string `gorm:"not null"`
	Phone     *string

	// Registration details
	Status               string    `gorm:"not null;default:'pending'"` // pending, confirmed, waiting_list, cancelled
	RegistrationDate     time.Time `gorm:"not null"`
	ConfirmationDate     *time.Time
	ConfirmationDeadline *time.Time
	CancellationDate     *time.Time
	CancellationReason   *string

	// Foreign keys
	RambleID uint  `gorm:"not null;index"`
	UserID   *uint `gorm:"index"` // Nullable for guest registrations
	GroupID  *uint `gorm:"index"` // Nullable for individual registrations

	// Foreign Objects
	Ramble Ramble                   `gorm:"foreignKey:RambleID"`
	User   *User                    `gorm:"foreignKey:UserID"`
	Group  *RambleRegistrationGroup `gorm:"foreignKey:GroupID"`
}

func (rr *RambleRegistration) ToModel() *model.RambleRegistration {
	return &model.RambleRegistration{
		ID:                   rr.ID,
		CreatedAt:            rr.CreatedAt,
		UpdatedAt:            rr.UpdatedAt,
		FirstName:            rr.FirstName,
		LastName:             rr.LastName,
		Email:                rr.Email,
		Phone:                rr.Phone,
		Status:               rr.Status,
		RegistrationDate:     rr.RegistrationDate,
		ConfirmationDate:     rr.ConfirmationDate,
		ConfirmationDeadline: rr.ConfirmationDeadline,
		CancellationDate:     rr.CancellationDate,
		CancellationReason:   rr.CancellationReason,
		RambleID:             rr.RambleID,
		UserID:               rr.UserID,
		GroupID:              rr.GroupID,
	}
}

type RambleRegistrationFilter struct {
	RambleID *uint
	UserID   *uint
	Email    *string
	Status   *string
	GroupID  *uint
}

type RambleRegistrationRepository interface {
	FindByID(id uint) (*RambleRegistration, error)
	FindAll(filter *RambleRegistrationFilter) ([]RambleRegistration, error)
	Create(registration *RambleRegistration) (*RambleRegistration, error)
	Update(registration *RambleRegistration) (*RambleRegistration, error)
	Delete(id uint) error
	CountByRambleAndStatus(rambleID uint, status string) (int64, error)
	FindByRambleAndStatus(rambleID uint, status string) ([]RambleRegistration, error)
	GetNextInWaitingList(rambleID uint) (*RambleRegistration, error)
	GetRegistrationsRequiringConfirmation(beforeDate time.Time) ([]RambleRegistration, error)
	GetUnconfirmedRegistrations(beforeDate time.Time) ([]RambleRegistration, error)
}

type rambleRegistrationRepository struct {
	db *gorm.DB
}

func NewRambleRegistrationRepository(db *gorm.DB) RambleRegistrationRepository {
	return &rambleRegistrationRepository{db: db}
}

func (r *rambleRegistrationRepository) FindByID(id uint) (*RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registration RambleRegistration
	err := r.db.WithContext(ctx).Preload("Ramble").Preload("User").First(&registration, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err // Other error
	}

	return &registration, nil
}

func (r *rambleRegistrationRepository) FindAll(filter *RambleRegistrationFilter) ([]RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registrations []RambleRegistration
	tx := r.db.WithContext(ctx).Model(&RambleRegistration{})

	if filter != nil {
		if filter.RambleID != nil {
			tx = tx.Where("ramble_id = ?", *filter.RambleID)
		}
		if filter.UserID != nil {
			tx = tx.Where("user_id = ?", *filter.UserID)
		}
		if filter.Email != nil {
			tx = tx.Where("email = ?", *filter.Email)
		}
		if filter.Status != nil {
			tx = tx.Where("status = ?", *filter.Status)
		}
		if filter.GroupID != nil {
			tx = tx.Where("group_id = ?", *filter.GroupID)
		}
	}

	err := tx.Preload("Ramble").Preload("User").Find(&registrations).Error
	return registrations, err
}

func (r *rambleRegistrationRepository) Create(registration *RambleRegistration) (*RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(registration).Error
	return registration, err
}

func (r *rambleRegistrationRepository) Update(registration *RambleRegistration) (*RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Save(registration).Error
	return registration, err
}

func (r *rambleRegistrationRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&RambleRegistration{}, id).Error
}

func (r *rambleRegistrationRepository) CountByRambleAndStatus(rambleID uint, status string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).Model(&RambleRegistration{}).
		Where("ramble_id = ? AND status = ?", rambleID, status).
		Count(&count).Error

	return count, err
}

func (r *rambleRegistrationRepository) FindByRambleAndStatus(rambleID uint, status string) ([]RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registrations []RambleRegistration
	err := r.db.WithContext(ctx).
		Where("ramble_id = ? AND status = ?", rambleID, status).
		Order("created_at ASC").
		Find(&registrations).Error

	return registrations, err
}

func (r *rambleRegistrationRepository) GetNextInWaitingList(rambleID uint) (*RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registration RambleRegistration
	err := r.db.WithContext(ctx).
		Where("ramble_id = ? AND status = ?", rambleID, "waiting_list").
		Order("created_at ASC").
		First(&registration).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No one in waiting list
		}
		return nil, err
	}

	return &registration, nil
}

func (r *rambleRegistrationRepository) GetRegistrationsRequiringConfirmation(beforeDate time.Time) ([]RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registrations []RambleRegistration
	err := r.db.WithContext(ctx).
		Preload("Ramble").
		Joins("JOIN rambles ON rambles.id = ramble_registrations.ramble_id").
		Where("ramble_registrations.status IN (?, ?) AND rambles.date IS NOT NULL AND rambles.date::date = ?::date AND ramble_registrations.confirmation_deadline IS NULL",
			"pending", "confirmed", beforeDate.Format("2006-01-02")).
		Find(&registrations).Error

	return registrations, err
}

func (r *rambleRegistrationRepository) GetUnconfirmedRegistrations(beforeDate time.Time) ([]RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registrations []RambleRegistration
	err := r.db.WithContext(ctx).
		Preload("Ramble").
		Where("status = ? AND confirmation_deadline IS NOT NULL AND confirmation_deadline < ?", "pending", beforeDate).
		Find(&registrations).Error

	return registrations, err
}
