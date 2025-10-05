package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func SeedV4(database *gorm.DB) error {
	// Create admin registration permissions
	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 16}, Name: "manage:registration", ReadableName: strPtr("Gérer les inscriptions"), Description: strPtr("Permet de gérer toutes les inscriptions (lecture, modification, suppression)")},
		{Model: gorm.Model{ID: 17}, Name: "view:all-registrations", ReadableName: strPtr("Voir toutes les inscriptions"), Description: strPtr("Permet de voir toutes les inscriptions de toutes les balades")},
		{Model: gorm.Model{ID: 18}, Name: "update:registration-status", ReadableName: strPtr("Modifier le statut des inscriptions"), Description: strPtr("Permet de modifier le statut des inscriptions (confirmer, annuler, etc.)")},
		{Model: gorm.Model{ID: 19}, Name: "update:registration-details", ReadableName: strPtr("Modifier les détails d'inscription"), Description: strPtr("Permet de modifier les informations personnelles des inscriptions")},
		{Model: gorm.Model{ID: 20}, Name: "bulk:registration-actions", ReadableName: strPtr("Actions en lot sur les inscriptions"), Description: strPtr("Permet d'effectuer des actions en lot sur plusieurs inscriptions")},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}

	// Add permissions to admin role
	admin := dbmodel.Role{}
	database.Preload("Permissions").First(&admin, 1)

	admin.Permissions = append(
		admin.Permissions,
		permissions[0], // manage:registration
		permissions[1], // view:all-registrations
		permissions[2], // update:registration-status
		permissions[3], // update:registration-details
		permissions[4], // bulk:registration-actions
	)

	database.Save(&admin)

	return nil
}
