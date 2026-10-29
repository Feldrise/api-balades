package dbmodel

import (
	"context"
	"errors"
	"math"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Reservation statuses. The status applies to the whole reservation.
const (
	ReservationStatusPending     = "pending"
	ReservationStatusConfirmed   = "confirmed"
	ReservationStatusWaitingList = "waiting_list"
	ReservationStatusCancelled   = "cancelled"
)

// Reservation is the commitment of a booker for one or more participants to a ramble.
// It is never split: its status applies to all its participants.
type Reservation struct {
	gorm.Model

	RambleID uint `gorm:"not null;index"`
	// UserID is the account of the booker, found or created from the contact email.
	UserID       uint   `gorm:"not null;index"`
	ContactEmail string `gorm:"not null;index"`

	// Status is pending, confirmed, waiting_list or cancelled.
	Status               string `gorm:"not null;default:'pending';index"`
	ConfirmationDeadline *time.Time
	ConfirmationDate     *time.Time
	CancellationDate     *time.Time
	CancellationReason   *string

	// Legacy registration or group this reservation was migrated from.
	LegacyRegistrationID *uint `gorm:"uniqueIndex"`
	LegacyGroupID        *uint `gorm:"uniqueIndex"`

	// Foreign Objects
	Ramble       Ramble        `gorm:"foreignKey:RambleID"`
	User         User          `gorm:"foreignKey:UserID"`
	Participants []Participant `gorm:"foreignKey:ReservationID"`
	Payments     []Payment     `gorm:"foreignKey:ReservationID"`
}

// Participant is a person attending a ramble under a reservation.
type Participant struct {
	gorm.Model

	ReservationID uint   `gorm:"not null;index"`
	FirstName     string `gorm:"not null"`
	LastName      string `gorm:"not null"`
	// PriceLabel is one of the ramble's price labels, empty when the ramble has no price.
	PriceLabel string `gorm:"not null;default:''"`
	// IsBooker marks the participant who is the booker themself.
	IsBooker bool `gorm:"not null;default:false"`
}

// TakesPlaces reports whether the reservation's participants occupy places on the ramble.
func (r *Reservation) TakesPlaces() bool {
	return r.Status == ReservationStatusPending || r.Status == ReservationStatusConfirmed
}

// OccupiedPlaces counts the participants of the reservations that take places
// (pending or confirmed). It is the single computation of occupied places.
func OccupiedPlaces(reservations []Reservation) int {
	occupied := 0
	for _, reservation := range reservations {
		if reservation.TakesPlaces() {
			occupied += len(reservation.Participants)
		}
	}
	return occupied
}

// PriceAmount returns the amount in euros of the price labelled label, 0 when unknown.
func (rm *Ramble) PriceAmount(label string) float64 {
	for _, price := range rm.Prices {
		if price.Label == label {
			return price.Amount
		}
	}
	return 0
}

// PriceMissing reports whether the participant's price is unknown: the ramble
// has prices but none is labelled label, as for participants migrated from a
// ramble with several prices.
func (rm *Ramble) PriceMissing(label string) bool {
	if len(rm.Prices) == 0 {
		return false
	}
	for _, price := range rm.Prices {
		if price.Label == label {
			return false
		}
	}
	return true
}

// HasMissingPrices reports whether a participant of the reservation has an unknown price.
func (r *Reservation) HasMissingPrices() bool {
	for _, participant := range r.Participants {
		if r.Ramble.PriceMissing(participant.PriceLabel) {
			return true
		}
	}
	return false
}

// HasSuccessfulPayment reports whether a payment of the reservation succeeded.
func (r *Reservation) HasSuccessfulPayment() bool {
	for _, payment := range r.Payments {
		if payment.Status == "succeeded" {
			return true
		}
	}
	return false
}

// EstimatedAmount returns the amount in cents of the reservation, computed from
// the known prices of its participants.
func (r *Reservation) EstimatedAmount() int64 {
	var total int64
	for _, participant := range r.Participants {
		total += toCents(r.Ramble.PriceAmount(participant.PriceLabel))
	}
	return total
}

func toCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

// ToModel requires the Ramble (with its prices), Participants and Payments to be loaded.
func (r *Reservation) ToModel() *model.Reservation {
	participants := make([]model.Participant, len(r.Participants))
	for i, participant := range r.Participants {
		participants[i] = model.Participant{
			ID:           participant.ID,
			FirstName:    participant.FirstName,
			LastName:     participant.LastName,
			PriceLabel:   participant.PriceLabel,
			PriceAmount:  r.Ramble.PriceAmount(participant.PriceLabel),
			PriceMissing: r.Ramble.PriceMissing(participant.PriceLabel),
			IsBooker:     participant.IsBooker,
		}
	}

	payments := make([]model.ReservationPayment, len(r.Payments))
	for i, payment := range r.Payments {
		payments[i] = model.ReservationPayment{
			ID:           payment.ID,
			CreatedAt:    payment.CreatedAt,
			Amount:       payment.Amount,
			Currency:     payment.Currency,
			Status:       payment.Status,
			PaidAt:       payment.PaidAt,
			RefundedAt:   payment.RefundedAt,
			RefundAmount: payment.RefundAmount,
		}
	}

	reservation := &model.Reservation{
		ID:                   r.ID,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
		RambleID:             r.RambleID,
		UserID:               r.UserID,
		ContactEmail:         r.ContactEmail,
		Status:               r.Status,
		ConfirmationDeadline: r.ConfirmationDeadline,
		ConfirmationDate:     r.ConfirmationDate,
		CancellationDate:     r.CancellationDate,
		CancellationReason:   r.CancellationReason,
		Participants:         participants,
		EstimatedAmount:      r.EstimatedAmount(),
		HasMissingPrices:     r.HasMissingPrices(),
		Payments:             payments,
		HasSuccessfulPayment: r.HasSuccessfulPayment(),
	}

	if r.Ramble.ID != 0 {
		reservation.Ramble = &model.RambleRegistrationSummary{
			Title:    r.Ramble.Title,
			Date:     r.Ramble.Date,
			Location: r.Ramble.Location,
		}
	}

	return reservation
}

type ReservationFilter struct {
	UserID *uint
}

type ReservationRepository interface {
	FindByID(id uint) (*Reservation, error)
	FindAll(filter *ReservationFilter) ([]Reservation, error)
	// CreateForRamble locks the ramble, hands the ramble's reservations that are
	// not cancelled to build, and creates the reservation build returns. The lock
	// makes the capacity and uniqueness decisions of build safe from concurrent creations.
	CreateForRamble(rambleID uint, build func(active []Reservation) (*Reservation, error)) (*Reservation, error)
	// ModifyForRamble locks the ramble and runs modify in a transaction with the
	// ramble (and its prices) and its reservations that are not cancelled. The
	// lock makes the capacity decisions of modify safe from concurrent changes.
	ModifyForRamble(rambleID uint, modify func(locked *LockedRamble) error) error
}

// LockedRamble gives the changes allowed on the reservations of a ramble while it is locked.
type LockedRamble struct {
	tx *gorm.DB

	Ramble *Ramble
	// Active are the ramble's reservations that are not cancelled, by creation date.
	Active []Reservation
}

// Reservation returns the active reservation with the given id, nil when there is none.
func (l *LockedRamble) Reservation(id uint) *Reservation {
	for i := range l.Active {
		if l.Active[i].ID == id {
			return &l.Active[i]
		}
	}
	return nil
}

// AddParticipants saves participants as new participants of reservation.
func (l *LockedRamble) AddParticipants(reservation *Reservation, participants []Participant) error {
	for i := range participants {
		participants[i].ReservationID = reservation.ID
	}
	if err := l.tx.Create(&participants).Error; err != nil {
		return err
	}
	reservation.Participants = append(reservation.Participants, participants...)
	return nil
}

// UpdateParticipant saves the name and price of participant.
func (l *LockedRamble) UpdateParticipant(participant *Participant) error {
	return l.tx.Model(participant).Select("FirstName", "LastName", "PriceLabel").Updates(participant).Error
}

// RemoveParticipant removes the participant from reservation.
func (l *LockedRamble) RemoveParticipant(reservation *Reservation, participantID uint) error {
	if err := l.tx.Where("reservation_id = ?", reservation.ID).Delete(&Participant{}, participantID).Error; err != nil {
		return err
	}
	kept := reservation.Participants[:0]
	for _, participant := range reservation.Participants {
		if participant.ID != participantID {
			kept = append(kept, participant)
		}
	}
	reservation.Participants = kept
	return nil
}

// SaveStatus saves the status of reservation with its confirmation and cancellation details.
func (l *LockedRamble) SaveStatus(reservation *Reservation) error {
	return l.tx.Model(reservation).
		Select("Status", "ConfirmationDate", "CancellationDate", "CancellationReason").
		Updates(reservation).Error
}

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func preloadReservation(tx *gorm.DB) *gorm.DB {
	return tx.
		Preload("Ramble").
		Preload("Ramble.Prices").
		Preload("Participants", func(db *gorm.DB) *gorm.DB { return db.Order("participants.id ASC") }).
		Preload("Payments", func(db *gorm.DB) *gorm.DB { return db.Order("payments.created_at ASC") })
}

func (r *reservationRepository) FindByID(id uint) (*Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var reservation Reservation
	err := preloadReservation(r.db.WithContext(ctx)).First(&reservation, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &reservation, nil
}

func (r *reservationRepository) FindAll(filter *ReservationFilter) ([]Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.WithContext(ctx).Model(&Reservation{})
	if filter != nil {
		if filter.UserID != nil {
			tx = tx.Where("user_id = ?", *filter.UserID)
		}
	}

	var reservations []Reservation
	err := preloadReservation(tx).Order("created_at DESC").Find(&reservations).Error
	return reservations, err
}

func (r *reservationRepository) CreateForRamble(rambleID uint, build func(active []Reservation) (*Reservation, error)) (*Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var created *Reservation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ramble Ramble
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&ramble, rambleID).Error; err != nil {
			return err
		}

		var active []Reservation
		err := tx.Preload("Participants").
			Where("ramble_id = ? AND status <> ?", rambleID, ReservationStatusCancelled).
			Order("created_at ASC").
			Find(&active).Error
		if err != nil {
			return err
		}

		reservation, err := build(active)
		if err != nil {
			return err
		}

		if err := tx.Create(reservation).Error; err != nil {
			return err
		}

		created = reservation
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.FindByID(created.ID)
}

func (r *reservationRepository) ModifyForRamble(rambleID uint, modify func(locked *LockedRamble) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ramble Ramble
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ramble, rambleID).Error; err != nil {
			return err
		}
		if err := tx.Where("ramble_id = ?", rambleID).Find(&ramble.Prices).Error; err != nil {
			return err
		}

		var active []Reservation
		err := tx.Preload("Participants", func(db *gorm.DB) *gorm.DB { return db.Order("participants.id ASC") }).
			Preload("Payments").
			Where("ramble_id = ? AND status <> ?", rambleID, ReservationStatusCancelled).
			Order("created_at ASC").
			Find(&active).Error
		if err != nil {
			return err
		}

		return modify(&LockedRamble{tx: tx, Ramble: &ramble, Active: active})
	})
}
