package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func SeedV6(database *gorm.DB) error {
	configureGuidePaymentsSelfDesc := "Configure own guide payment settings (Stripe credentials)"
	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 26}, Name: "configure:guide-payments:self", ReadableName: strPtr("Configurer ses paiements guide"), Description: &configureGuidePaymentsSelfDesc},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}

	// Load existing permissions needed for the guide role
	var ramblePermissions []dbmodel.Permission
	database.Where("name IN ?", []string{
		"read:ramble",
		"read:ramble:self",
		"create:ramble",
		"update:ramble:self",
		"delete:ramble:self",
		"read:user:self",
		"update:user:self",
	}).Find(&ramblePermissions)

	permissionByName := make(map[string]dbmodel.Permission, len(ramblePermissions))
	for _, permission := range ramblePermissions {
		permissionByName[permission.Name] = permission
	}

	guideRole := dbmodel.Role{
		Model: gorm.Model{ID: 3},
		Name:  "guide",
		Permissions: []dbmodel.Permission{
			permissionByName["read:ramble"],
			permissionByName["read:ramble:self"],
			permissionByName["create:ramble"],
			permissionByName["update:ramble:self"],
			permissionByName["delete:ramble:self"],
			permissionByName["read:user:self"],
			permissionByName["update:user:self"],
			permissions[0], // configure:guide-payments:self
		},
	}

	database.Create(&guideRole)

	return nil
}
