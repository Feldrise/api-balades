package dbmodel

import (
	"time"

	"gorm.io/gorm"
)

// RambleRegistrationGroup is the legacy group of registrations, replaced by reservations.
// It only remains to be migrated by database.MigrateLegacyReservations.
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
