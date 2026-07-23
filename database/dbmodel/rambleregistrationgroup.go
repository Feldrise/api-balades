package dbmodel

import (
	"context"
	"errors"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
)

type RambleRegistrationGroup struct {
	gorm.Model

	// Group information
	PrimaryEmail string `gorm:"not null"`
	Status       string `gorm:"not null;default:'pending'"` // pending, confirmed, waiting_list, cancelled

	// Registration details
	RegistrationDate     time.Time `gorm:"not null"`
	ConfirmationDate     *time.Time
	ConfirmationDeadline *time.Time
	CancellationDate     *time.Time
	CancellationReason   *string

	// Foreign keys
	RambleID uint `gorm:"not null;index"`

	// Foreign Objects
	Ramble        Ramble               `gorm:"foreignKey:RambleID"`
	Registrations []RambleRegistration `gorm:"foreignKey:GroupID"`
}

func (grg *RambleRegistrationGroup) ToModel() *model.RambleRegistrationGroup {
	registrations := make([]*model.RambleRegistration, len(grg.Registrations))
	for i, reg := range grg.Registrations {
		registrations[i] = reg.ToModel()
	}

	return &model.RambleRegistrationGroup{
		ID:                   grg.ID,
		CreatedAt:            grg.CreatedAt,
		UpdatedAt:            grg.UpdatedAt,
		PrimaryEmail:         grg.PrimaryEmail,
		Status:               grg.Status,
		RegistrationDate:     grg.RegistrationDate,
		ConfirmationDate:     grg.ConfirmationDate,
		ConfirmationDeadline: grg.ConfirmationDeadline,
		CancellationDate:     grg.CancellationDate,
		CancellationReason:   grg.CancellationReason,
		RambleID:             grg.RambleID,
		Registrations:        registrations,
		ParticipantCount:     len(grg.Registrations),
	}
}

type RambleRegistrationGroupFilter struct {
	RambleID     *uint
	RambleIDs    []uint // When set, restrict to these ramble IDs (empty = no matches)
	PrimaryEmail *string
	Status       *string
}

type RambleRegistrationGroupRepository interface {
	FindByID(id uint) (*RambleRegistrationGroup, error)
	FindAll(filter *RambleRegistrationGroupFilter) ([]RambleRegistrationGroup, error)
	Create(group *RambleRegistrationGroup) (*RambleRegistrationGroup, error)
	Update(group *RambleRegistrationGroup) (*RambleRegistrationGroup, error)
	Delete(id uint) error
	CountByRambleAndStatus(rambleID uint, status string) (int64, error)
	FindByRambleAndStatus(rambleID uint, status string) ([]RambleRegistrationGroup, error)
	GetNextInWaitingList(rambleID uint) (*RambleRegistrationGroup, error)
	GetGroupsRequiringConfirmation(beforeDate time.Time) ([]RambleRegistrationGroup, error)
	GetUnconfirmedGroups(beforeDate time.Time) ([]RambleRegistrationGroup, error)
}

type rambleRegistrationGroupRepository struct {
	db *gorm.DB
}

func NewRambleRegistrationGroupRepository(db *gorm.DB) RambleRegistrationGroupRepository {
	return &rambleRegistrationGroupRepository{db: db}
}

func (r *rambleRegistrationGroupRepository) FindByID(id uint) (*RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var group RambleRegistrationGroup
	err := r.db.WithContext(ctx).Preload("Ramble").Preload("Registrations").First(&group, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &group, nil
}

func (r *rambleRegistrationGroupRepository) FindAll(filter *RambleRegistrationGroupFilter) ([]RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var groups []RambleRegistrationGroup
	tx := r.db.WithContext(ctx).Model(&RambleRegistrationGroup{})

	if filter != nil {
		if filter.RambleID != nil {
			tx = tx.Where("ramble_id = ?", *filter.RambleID)
		}
		if filter.RambleIDs != nil {
			if len(filter.RambleIDs) == 0 {
				tx = tx.Where("1 = 0")
			} else {
				tx = tx.Where("ramble_id IN ?", filter.RambleIDs)
			}
		}
		if filter.PrimaryEmail != nil {
			tx = tx.Where("primary_email = ?", *filter.PrimaryEmail)
		}
		if filter.Status != nil {
			tx = tx.Where("status = ?", *filter.Status)
		}
	}

	err := tx.Preload("Ramble").Preload("Registrations").Find(&groups).Error
	return groups, err
}

func (r *rambleRegistrationGroupRepository) Create(group *RambleRegistrationGroup) (*RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(group).Error
	return group, err
}

func (r *rambleRegistrationGroupRepository) Update(group *RambleRegistrationGroup) (*RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Save(group).Error
	return group, err
}

func (r *rambleRegistrationGroupRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&RambleRegistrationGroup{}, id).Error
}

func (r *rambleRegistrationGroupRepository) CountByRambleAndStatus(rambleID uint, status string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).Model(&RambleRegistrationGroup{}).
		Where("ramble_id = ? AND status = ?", rambleID, status).
		Count(&count).Error

	return count, err
}

func (r *rambleRegistrationGroupRepository) FindByRambleAndStatus(rambleID uint, status string) ([]RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var groups []RambleRegistrationGroup
	err := r.db.WithContext(ctx).
		Where("ramble_id = ? AND status = ?", rambleID, status).
		Order("created_at ASC").
		Preload("Registrations").
		Find(&groups).Error

	return groups, err
}

func (r *rambleRegistrationGroupRepository) GetNextInWaitingList(rambleID uint) (*RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var group RambleRegistrationGroup
	err := r.db.WithContext(ctx).
		Where("ramble_id = ? AND status = ?", rambleID, "waiting_list").
		Order("created_at ASC").
		Preload("Registrations").
		First(&group).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &group, nil
}

func (r *rambleRegistrationGroupRepository) GetGroupsRequiringConfirmation(beforeDate time.Time) ([]RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var groups []RambleRegistrationGroup
	err := r.db.WithContext(ctx).
		Preload("Ramble").
		Preload("Registrations").
		Joins("JOIN rambles ON rambles.id = ramble_registration_groups.ramble_id").
		Where("ramble_registration_groups.status IN (?, ?) AND rambles.date IS NOT NULL AND rambles.date::date = ?::date AND ramble_registration_groups.confirmation_deadline IS NULL",
			"pending", "confirmed", beforeDate.Format("2006-01-02")).
		Find(&groups).Error

	return groups, err
}

func (r *rambleRegistrationGroupRepository) GetUnconfirmedGroups(beforeDate time.Time) ([]RambleRegistrationGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var groups []RambleRegistrationGroup
	err := r.db.WithContext(ctx).
		Preload("Ramble").
		Preload("Registrations").
		Where("status = ? AND confirmation_deadline IS NOT NULL AND confirmation_deadline < ?", "pending", beforeDate).
		Find(&groups).Error

	return groups, err
}
