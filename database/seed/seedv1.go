package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func strPtr(s string) *string {
	return &s
}

func SeedV1(database *gorm.DB) error {
	// Create permissions
	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 1}, Name: "create:user", ReadableName: strPtr("Créer un utilisateur"), Description: strPtr("Permet de créer un nouvel utilisateur")},
		{Model: gorm.Model{ID: 2}, Name: "read:user", ReadableName: strPtr("Lire un utilisateur"), Description: strPtr("Permet de lire les informations d'un utilisateur")},
		{Model: gorm.Model{ID: 3}, Name: "read:user:self", ReadableName: strPtr("Lire son propre utilisateur"), Description: strPtr("Permet de lire ses propres informations")},
		{Model: gorm.Model{ID: 4}, Name: "update:user", ReadableName: strPtr("Mettre à jour un utilisateur"), Description: strPtr("Permet de mettre à jour les informations d'un utilisateur")},
		{Model: gorm.Model{ID: 5}, Name: "update:user:self", ReadableName: strPtr("Mettre à jour son propre utilisateur"), Description: strPtr("Permet de mettre à jour ses propres informations")},
		{Model: gorm.Model{ID: 6}, Name: "delete:user", ReadableName: strPtr("Supprimer un utilisateur"), Description: strPtr("Permet de supprimer un utilisateur")},
		{Model: gorm.Model{ID: 7}, Name: "delete:user:self", ReadableName: strPtr("Supprimer son propre utilisateur"), Description: strPtr("Permet de supprimer son propre compte")},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}

	// Create roles
	roles := []dbmodel.Role{
		{
			Model: gorm.Model{ID: 1},
			Name:  "admin",
			Permissions: []dbmodel.Permission{
				permissions[0], // create:user
				permissions[1], // read:user
				permissions[2], // read:user:self
				permissions[3], // update:user
				permissions[4], // update:user:self
				permissions[5], // delete:user
				permissions[6], // delete:user:self
			},
		},
		{
			Model: gorm.Model{ID: 2},
			Name:  "user",
			Permissions: []dbmodel.Permission{
				permissions[2], // read:user:self
				permissions[4], // update:user:self
			},
		},
	}

	for _, role := range roles {
		database.Create(&role)
	}

	return nil
}
