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
	ramble := rr.Ramble

	rrModel := &model.RambleRegistration{
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

	if ramble.ID != 0 {
		rrModel.Ramble = &model.RambleRegistrationSummary{
			Title:    ramble.Title,
			Date:     ramble.Date,
			Location: ramble.Location,
		}
	}

	return rrModel
}

type RambleRegistrationFilter struct {
	RambleID  *uint
	RambleIDs []uint // When set, restrict to these ramble IDs (empty = no matches)
	UserID    *uint
	Email     *string
	Status    *string
	GroupID   *uint
	// New advanced filters
	DateFrom    *time.Time
	DateTo      *time.Time
	Search      *string // Search in name, email
	RambleTitle *string
	Statuses    []string // Multiple statuses
	Limit       *int
	Offset      *int
	SortBy      *string // "created_at", "registration_date", "first_name", "last_name", "email"
	SortOrder   *string // "asc", "desc"
}

type RambleRegistrationRepository interface {
	FindByID(id uint) (*RambleRegistration, error)
	FindAll(filter *RambleRegistrationFilter) ([]RambleRegistration, error)
	CountAll(filter *RambleRegistrationFilter) (int64, error)
	Create(registration *RambleRegistration) (*RambleRegistration, error)
	Update(registration *RambleRegistration) (*RambleRegistration, error)
	Delete(id uint) error
	CountByRambleAndStatus(rambleID uint, status string) (int64, error)
	FindByRambleAndStatus(rambleID uint, status string) ([]RambleRegistration, error)
	GetNextInWaitingList(rambleID uint) (*RambleRegistration, error)
	GetRegistrationsRequiringConfirmation(beforeDate time.Time) ([]RambleRegistration, error)
	GetUnconfirmedRegistrations(deadlineBefore time.Time, rambleDateOnOrBefore time.Time) ([]RambleRegistration, error)
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
		tx = applyRambleRegistrationFilter(tx, filter)

		// Sorting
		sortBy := "created_at"
		sortOrder := "DESC"
		if filter.SortBy != nil && *filter.SortBy != "" {
			validSortFields := map[string]bool{
				"created_at":        true,
				"registration_date": true,
				"first_name":        true,
				"last_name":         true,
				"email":             true,
				"status":            true,
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

	err := tx.Preload("Ramble").Preload("User").Find(&registrations).Error
	return registrations, err
}

func (r *rambleRegistrationRepository) CountAll(filter *RambleRegistrationFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	tx := r.db.WithContext(ctx).Model(&RambleRegistration{})

	if filter != nil {
		tx = applyRambleRegistrationFilter(tx, filter)
	}

	err := tx.Count(&count).Error
	return count, err
}

func applyRambleRegistrationFilter(tx *gorm.DB, filter *RambleRegistrationFilter) *gorm.DB {
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
	if filter.DateFrom != nil {
		tx = tx.Where("registration_date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		tx = tx.Where("registration_date <= ?", *filter.DateTo)
	}
	if filter.Search != nil && *filter.Search != "" {
		searchTerm := "%" + *filter.Search + "%"
		tx = tx.Where("(first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?)", searchTerm, searchTerm, searchTerm)
	}
	if filter.RambleTitle != nil && *filter.RambleTitle != "" {
		tx = tx.Joins("JOIN rambles ON rambles.id = ramble_registrations.ramble_id").
			Where("rambles.title ILIKE ?", "%"+*filter.RambleTitle+"%")
	}
	if len(filter.Statuses) > 0 {
		tx = tx.Where("status IN ?", filter.Statuses)
	}
	return tx
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

func (r *rambleRegistrationRepository) GetUnconfirmedRegistrations(deadlineBefore time.Time, rambleDateOnOrBefore time.Time) ([]RambleRegistration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var registrations []RambleRegistration
	err := r.db.WithContext(ctx).
		Preload("Ramble").
		Joins("JOIN rambles ON rambles.id = ramble_registrations.ramble_id AND rambles.deleted_at IS NULL").
		Where(
			"ramble_registrations.status = ? AND ramble_registrations.confirmation_deadline IS NOT NULL AND (ramble_registrations.confirmation_deadline < ? OR rambles.date <= ?)",
			"pending", deadlineBefore, rambleDateOnOrBefore,
		).
		Find(&registrations).Error

	return registrations, err
}
