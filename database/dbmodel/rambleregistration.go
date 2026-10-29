package dbmodel

import (
	"time"

	"gorm.io/gorm"
)

// RambleRegistration is the legacy individual registration, replaced by reservations.
// It only remains to be migrated by database.MigrateLegacyReservations.
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
