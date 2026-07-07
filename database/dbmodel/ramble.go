package dbmodel

import (
	"context"
	"errors"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
)

type RamblePrice struct {
	gorm.Model

	Label  string  `gorm:"not null"`
	Amount float64 `gorm:"not null"`

	RambleID string `gorm:"not null"` // Foreign key to Ramble

	// Foreign Objects
	Ramble Ramble `gorm:"foreignKey:RambleID;"`
}

type Ramble struct {
	gorm.Model

	Title                  string `gorm:"not null"`
	Description            *string
	Type                   string `gorm:"not null;default:'Découverte générale'"`
	Date                   *time.Time
	Location               *string
	MeetingPoint           *string
	MeetingLatitude        *float64
	MeetingLongitude       *float64
	MaxParticipants        *int
	Difficulty             string `gorm:"not null;default:'Facile'"`
	EstimatedDuration      *string
	EquipmentNeeded        *string
	Prerequisites          *string
	CoverImage             *string
	AdditionalDocumentsURL *string

	// Publication
	PublishedAt *time.Time

	// Cancellation fields
	IsCancelled        bool `gorm:"not null;default:false"`
	CancellationDate   *time.Time
	CancellationReason *string

	// Payment fields
	PaymentGuideID  *uint `gorm:"column:payment_guide_id;index"`
	PaymentEnabled  bool  `gorm:"not null;default:false"`
	PaymentRequired bool  `gorm:"not null;default:false"`

	// Many-to-many relationship with guides
	Guides []Guide `gorm:"many2many:ramble_guides;"`

	// Foreign Objects
	Prices        []RamblePrice        `gorm:"foreignKey:RambleID;"`
	Registrations []RambleRegistration `gorm:"foreignKey:RambleID;"`
	PaymentGuide  *Guide               `gorm:"foreignKey:PaymentGuideID"`

	// Computed fields (not stored in database)
	PlacesLeft *int `gorm:"-"`
}

// IsPublished reports whether the ramble is visible to the public.
func (rm *Ramble) IsPublished() bool {
	if rm.PublishedAt == nil {
		return false
	}
	return !rm.PublishedAt.After(time.Now())
}

// CalculatePlacesLeft computes the number of available places for the ramble
func (rm *Ramble) CalculatePlacesLeft() {
	if rm.MaxParticipants == nil {
		rm.PlacesLeft = nil // Unlimited places
		return
	}

	reservationCount := 0
	for _, registration := range rm.Registrations {
		if registration.Status != "cancelled" {
			reservationCount++
		}
	}

	placesLeft := *rm.MaxParticipants - reservationCount

	// We keep the negative value to indicate waitlist status
	rm.PlacesLeft = &placesLeft
}

func (rm Ramble) ToModel() model.Ramble {
	prices := []model.RamblePrice{}
	for _, price := range rm.Prices {
		prices = append(prices, model.RamblePrice{
			Label:  price.Label,
			Amount: price.Amount,
		})
	}

	guides := []model.Guide{}
	for _, guide := range rm.Guides {
		guides = append(guides, guide.ToModel())
	}

	// Calculate places left
	rm.CalculatePlacesLeft()

	rambleModel := model.Ramble{
		ID:                     rm.ID,
		CreatedAt:              rm.CreatedAt,
		UpdatedAt:              rm.UpdatedAt,
		Title:                  rm.Title,
		Description:            rm.Description,
		Type:                   rm.Type,
		Date:                   rm.Date,
		Location:               rm.Location,
		MeetingPoint:           rm.MeetingPoint,
		MeetingLatitude:        rm.MeetingLatitude,
		MeetingLongitude:       rm.MeetingLongitude,
		MaxParticipants:        rm.MaxParticipants,
		Prices:                 prices,
		Difficulty:             rm.Difficulty,
		EstimatedDuration:      rm.EstimatedDuration,
		EquipmentNeeded:        rm.EquipmentNeeded,
		Prerequisites:          rm.Prerequisites,
		CoverImage:             rm.CoverImage,
		AdditionalDocumentsURL: rm.AdditionalDocumentsURL,
		PublishedAt:            rm.PublishedAt,
		IsCancelled:            rm.IsCancelled,
		CancellationDate:       rm.CancellationDate,
		CancellationReason:     rm.CancellationReason,
		Guides:                 guides,
		PlacesLeft:             rm.PlacesLeft,
		PaymentGuideID:         rm.PaymentGuideID,
		PaymentEnabled:         rm.PaymentEnabled,
		PaymentRequired:        rm.PaymentRequired,
	}

	if rm.PaymentGuide != nil {
		paymentGuide := rm.PaymentGuide.ToModel()
		rambleModel.PaymentGuide = &paymentGuide
	}

	return rambleModel
}

type RambleFilter struct {
	IsCancelled *bool // Filter by cancellation status
	Type        *string
	Difficulty  *string
	Location    *string
	Search      *string    // Search in title, description, location
	DateFrom    *time.Time // Filter rambles from this date
	DateTo      *time.Time // Filter rambles to this date
	GuideID             *uint // Filter rambles by specific guide
	IncludeUnpublished  bool  // When false, only return publicly visible rambles
}

type RambleRepository interface {
	FindByID(id uint, includeUnpublished bool) (*Ramble, error)
	FindAll(filter *RambleFilter) ([]Ramble, error)
	Create(ramble *Ramble) (*Ramble, error)
	Update(ramble *Ramble, updateAssociations bool) (*Ramble, error)
	Delete(id uint) error
}

type rambleRepository struct {
	db *gorm.DB
}

func NewRambleRepository(db *gorm.DB) RambleRepository {
	return &rambleRepository{db: db}
}

func applyPublishedFilter(tx *gorm.DB, includeUnpublished bool) *gorm.DB {
	if !includeUnpublished {
		tx = tx.Where("published_at IS NOT NULL AND published_at <= ?", time.Now())
	}
	return tx
}

func (r *rambleRepository) FindByID(id uint, includeUnpublished bool) (*Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ramble Ramble
	tx := r.db.WithContext(ctx).Model(&ramble)
	tx = applyPublishedFilter(tx, includeUnpublished)

	tx = tx.Preload("Prices").Preload("Guides").Preload("Registrations").Preload("PaymentGuide")

	err := tx.First(&ramble, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err // Other error
	}

	return &ramble, nil
}

func (r *rambleRepository) FindAll(filter *RambleFilter) ([]Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rambles []Ramble
	tx := r.db.WithContext(ctx).Model(&Ramble{})

	includeUnpublished := false
	if filter != nil {
		includeUnpublished = filter.IncludeUnpublished
	}
	tx = applyPublishedFilter(tx, includeUnpublished)

	if filter != nil {
		if filter.IsCancelled != nil {
			tx = tx.Where("is_cancelled = ?", *filter.IsCancelled)
		}

		if filter.Type != nil {
			tx = tx.Where("type = ?", *filter.Type)
		}

		if filter.Difficulty != nil {
			tx = tx.Where("difficulty = ?", *filter.Difficulty)
		}

		if filter.Location != nil {
			tx = tx.Where("location ILIKE ?", "%"+*filter.Location+"%")
		}

		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			tx = tx.Where("title ILIKE ? OR description ILIKE ? OR location ILIKE ?", search, search, search)
		}

		if filter.DateFrom != nil {
			tx = tx.Where("date >= ?", *filter.DateFrom)
		}

		if filter.DateTo != nil {
			tx = tx.Where("date <= ?", *filter.DateTo)
		}

		if filter.GuideID != nil {
			tx = tx.Joins("JOIN ramble_guides ON rambles.id = ramble_guides.ramble_id").
				Where("ramble_guides.guide_id = ?", *filter.GuideID)
		}
	}

	tx = tx.Preload("Prices").Preload("Guides").Preload("Registrations").Preload("PaymentGuide")

	err := tx.Find(&rambles).Error

	return rambles, err
}

func (r *rambleRepository) Create(ramble *Ramble) (*Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(ramble).Error

	return ramble, err
}

func (r *rambleRepository) Update(ramble *Ramble, updateAssociations bool) (*Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete existing prices and guides if updating associations
	if updateAssociations || len(ramble.Prices) > 0 || len(ramble.Guides) > 0 {
		var err error

		if updateAssociations || len(ramble.Prices) > 0 {
			err = r.db.Unscoped().WithContext(ctx).Model(&RamblePrice{}).
				Where("ramble_id = ?", ramble.ID).
				Delete(&RamblePrice{}).Error
		}

		if updateAssociations || len(ramble.Guides) > 0 {
			err = r.db.Unscoped().WithContext(ctx).Model(&RambleGuide{}).
				Where("ramble_id = ?", ramble.ID).
				Delete(&RambleGuide{}).Error
		}

		if err != nil {
			return nil, err
		}
	}

	tx := r.db.WithContext(ctx)

	err := tx.Save(ramble).Error

	return ramble, err
}

func (r *rambleRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.WithContext(ctx)

	err := tx.Where("id = ?", id).Delete(&Ramble{}).Error

	return err
}
