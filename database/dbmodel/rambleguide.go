package dbmodel

import (
	"gorm.io/gorm"
)

type RambleGuide struct {
	gorm.Model

	RambleID uint `gorm:"not null;index"`
	GuideID  uint `gorm:"not null;index"`

	// Foreign Objects
	Ramble Ramble `gorm:"foreignKey:RambleID"`
	Guide  Guide  `gorm:"foreignKey:GuideID"`
}
