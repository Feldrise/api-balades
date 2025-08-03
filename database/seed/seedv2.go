package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func SeedV2(database *gorm.DB) error {
	// Create permissions
	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 8}, Name: "create:ramble", ReadableName: strPtr("Créer une balade"), Description: strPtr("Permet de créer une nouvelle balade")},
		{Model: gorm.Model{ID: 9}, Name: "read:ramble", ReadableName: strPtr("Lire une balade"), Description: strPtr("Permet de lire les informations d'une balade")},
		{Model: gorm.Model{ID: 10}, Name: "read:ramble:self", ReadableName: strPtr("Lire sa propre balade"), Description: strPtr("Permet de lire ses propres informations")},
		{Model: gorm.Model{ID: 11}, Name: "update:ramble", ReadableName: strPtr("Mettre à jour une balade"), Description: strPtr("Permet de mettre à jour les informations d'une balade")},
		{Model: gorm.Model{ID: 12}, Name: "update:ramble:self", ReadableName: strPtr("Mettre à jour sa propre balade"), Description: strPtr("Permet de mettre à jour ses propres informations")},
		{Model: gorm.Model{ID: 13}, Name: "delete:ramble", ReadableName: strPtr("Supprimer une balade"), Description: strPtr("Permet de supprimer une balade")},
		{Model: gorm.Model{ID: 14}, Name: "delete:ramble:self", ReadableName: strPtr("Supprimer sa propre balade"), Description: strPtr("Permet de supprimer ses propres informations")},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}
	// Add roles to admin
	admin := dbmodel.Role{}
	database.Preload("Permissions").First(&admin, 1)

	admin.Permissions = append(
		admin.Permissions,
		permissions[0], // create:ramble
		permissions[1], // read:ramble
		permissions[2], // read:ramble:self
		permissions[3], // update:ramble
		permissions[4], // update:ramble:self
		permissions[5], // delete:ramble
		permissions[6], // delete:ramble:self
	)

	database.Save(&admin)

	// Add roles to user
	user := dbmodel.Role{}
	database.Preload("Permissions").First(&user, 2)

	user.Permissions = append(
		user.Permissions,
		permissions[1], // read:ramble
		permissions[2], // read:ramble:self
		permissions[4], // update:ramble:self
	)

	database.Save(&user)

	return nil
}
