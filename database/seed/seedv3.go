package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func SeedV3(database *gorm.DB) error {
	// Create cancellation permission
	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 15}, Name: "cancel:ramble", ReadableName: strPtr("Annuler une balade"), Description: strPtr("Permet d'annuler une balade avec un motif")},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}

	// Add permission to admin role
	admin := dbmodel.Role{}
	database.Preload("Permissions").First(&admin, 1)

	admin.Permissions = append(
		admin.Permissions,
		permissions[0], // cancel:ramble
	)

	database.Save(&admin)

	return nil
}
