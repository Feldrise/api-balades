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
	Status                 string `gorm:"not null"`
	Description            *string
	Type                   string `gorm:"not null;default:'Découverte générale'"`
	Date                   *time.Time
	Location               *string
	MeetingPoint           *string
	MaxParticipants        *int
	Difficulty             string `gorm:"not null;default:'Facile'"`
	EstimatedDuration      *string
	EquipmentNeeded        *string
	Prerequisites          *string
	CoverImage             *string
	AdditionalDocumentsURL *string

	// Many-to-many relationship with guides
	Guides []Guide `gorm:"many2many:ramble_guides;"`

	// Foreign Objects
	Prices        []RamblePrice        `gorm:"foreignKey:RambleID;"`
	Registrations []RambleRegistration `gorm:"foreignKey:RambleID;"`

	// Computed fields (not stored in database)
	PlacesLeft *int `gorm:"-"`
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

	return model.Ramble{
		ID:                     rm.ID,
		CreatedAt:              rm.CreatedAt,
		UpdatedAt:              rm.UpdatedAt,
		Title:                  rm.Title,
		Status:                 rm.Status,
		Description:            rm.Description,
		Type:                   rm.Type,
		Date:                   rm.Date,
		Location:               rm.Location,
		MeetingPoint:           rm.MeetingPoint,
		MaxParticipants:        rm.MaxParticipants,
		Prices:                 prices,
		Difficulty:             rm.Difficulty,
		EstimatedDuration:      rm.EstimatedDuration,
		EquipmentNeeded:        rm.EquipmentNeeded,
		Prerequisites:          rm.Prerequisites,
		CoverImage:             rm.CoverImage,
		AdditionalDocumentsURL: rm.AdditionalDocumentsURL,
		Guides:                 guides,
		PlacesLeft:             rm.PlacesLeft,
	}
}

type RambleFilter struct {
	Status     *string
	Type       *string
	Difficulty *string
	Location   *string
	Search     *string    // Search in title, description, location
	DateFrom   *time.Time // Filter rambles from this date
	DateTo     *time.Time // Filter rambles to this date
	GuideID    *uint      // Filter rambles by specific guide
	IsActive   *bool      // Filter by active status (non-archived)
}

type RambleRepository interface {
	FindByID(id uint) (*Ramble, error)
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

func (r *rambleRepository) FindByID(id uint) (*Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ramble Ramble
	tx := r.db.WithContext(ctx).Model(&ramble)

	tx = tx.Preload("Prices").Preload("Guides").Preload("Registrations")

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

	if filter != nil {
		if filter.Status != nil {
			tx = tx.Where("status = ?", *filter.Status)
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

		if filter.IsActive != nil {
			if *filter.IsActive {
				tx = tx.Where("status != ?", "archived")
			} else {
				tx = tx.Where("status = ?", "archived")
			}
		}
	}

	tx = tx.Preload("Prices").Preload("Guides").Preload("Registrations")

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
	if updateAssociations {
		err := r.db.Unscoped().WithContext(ctx).Model(&RamblePrice{}).
			Where("ramble_id = ?", ramble.ID).
			Delete(&RamblePrice{}).Error

		err = r.db.Unscoped().WithContext(ctx).Model(&RambleGuide{}).
			Where("ramble_id = ?", ramble.ID).
			Delete(&RambleGuide{}).Error

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
