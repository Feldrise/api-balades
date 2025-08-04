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
	CoverImageURL          *string
	AdditionalDocumentsURL *string

	// Many-to-many relationship with guides
	Guides []Guide `gorm:"many2many:ramble_guides;"`

	// Foreign Objects
	Prices []RamblePrice `gorm:"foreignKey:RambleID;"`
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
		CoverImageURL:          rm.CoverImageURL,
		AdditionalDocumentsURL: rm.AdditionalDocumentsURL,
		Guides:                 guides,
	}
}

type RambleFilter struct {
	Status *string
}

type RambleRepository interface {
	FindByID(id uint) (*Ramble, error)
	FindAll(filter *RambleFilter) ([]Ramble, error)
	Create(ramble *Ramble) (*Ramble, error)
	Update(ramble *Ramble) (*Ramble, error)
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

	tx = tx.Preload("Prices").Preload("Guides")

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

	if filter != nil && filter.Status != nil {
		tx = tx.Where("status = ?", *filter.Status)
	}

	tx = tx.Preload("Prices").Preload("Guides")

	err := tx.Find(&rambles).Error

	return rambles, err
}

func (r *rambleRepository) Create(ramble *Ramble) (*Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(ramble).Error

	return ramble, err
}

func (r *rambleRepository) Update(ramble *Ramble) (*Ramble, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Model(&RamblePrice{}).
		Where("ramble_id = ?", ramble.ID).
		Delete(&RamblePrice{}).Error

	tx := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true})

	err = tx.Save(ramble).Error

	return ramble, err
}

func (r *rambleRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.WithContext(ctx)

	err := tx.Where("id = ?", id).Delete(&Ramble{}).Error

	return err
}
