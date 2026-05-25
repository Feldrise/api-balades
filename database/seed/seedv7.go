package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func SeedV7(database *gorm.DB) error {
	viewRegistrationsSelfDesc := "View registrations for own guided rambles"
	manageRegistrationSelfDesc := "Delete registrations for own guided rambles"
	updateRegistrationStatusSelfDesc := "Update registration status for own guided rambles"
	updateRegistrationDetailsSelfDesc := "Update registration details for own guided rambles"
	bulkRegistrationActionsSelfDesc := "Bulk actions on registrations for own guided rambles"

	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 27}, Name: "view:registrations:self", ReadableName: strPtr("Voir les inscriptions de ses balades"), Description: &viewRegistrationsSelfDesc},
		{Model: gorm.Model{ID: 28}, Name: "manage:registration:self", ReadableName: strPtr("Gérer les inscriptions de ses balades"), Description: &manageRegistrationSelfDesc},
		{Model: gorm.Model{ID: 29}, Name: "update:registration-status:self", ReadableName: strPtr("Modifier le statut des inscriptions de ses balades"), Description: &updateRegistrationStatusSelfDesc},
		{Model: gorm.Model{ID: 30}, Name: "update:registration-details:self", ReadableName: strPtr("Modifier les détails des inscriptions de ses balades"), Description: &updateRegistrationDetailsSelfDesc},
		{Model: gorm.Model{ID: 31}, Name: "bulk:registration-actions:self", ReadableName: strPtr("Actions en lot sur les inscriptions de ses balades"), Description: &bulkRegistrationActionsSelfDesc},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}

	guide := dbmodel.Role{}
	database.Preload("Permissions").First(&guide, 3)

	guide.Permissions = append(
		guide.Permissions,
		permissions[0],
		permissions[1],
		permissions[2],
		permissions[3],
		permissions[4],
	)

	database.Save(&guide)

	return nil
}
