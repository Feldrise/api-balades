package database

import (
	"log"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/database/seed"
	"gorm.io/gorm"
)

func Migrate(database *gorm.DB) {
	database.AutoMigrate(
		&seed.Seed{},
		&dbmodel.Address{},
		&dbmodel.User{},
		&dbmodel.UserProfile{},
		&dbmodel.Permission{},
		&dbmodel.Role{},
		&dbmodel.UserPermissionOverride{},
		&dbmodel.Ramble{},
		&dbmodel.RamblePrice{},
		&dbmodel.Guide{},
		&dbmodel.RambleGuide{},
		&dbmodel.RambleRegistrationGroup{},
		&dbmodel.RambleRegistration{},
	)

	log.Println("Database migrated")

	ApplySeeds(database)
}

func ApplySeeds(database *gorm.DB) {
	seedsToApply := []struct {
		Name     string
		SeedFunc func(*gorm.DB) error
	}{
		{"SeedV1", seed.SeedV1},
		{"SeedV2", seed.SeedV2},
	}

	for _, seedToApply := range seedsToApply {
		if !isSeedApplied(database, seedToApply.Name) {
			log.Printf("Applying seed %s", seedToApply.Name)
			if err := seedToApply.SeedFunc(database); err != nil {
				log.Fatalf("Error applying seed %s: %s", seedToApply.Name, err)
			}
			markSeedAsApplied(database, seedToApply.Name)
		}
	}

	log.Println("Seeds applied")
}

func isSeedApplied(database *gorm.DB, name string) bool {
	var count int64
	database.Model(&seed.Seed{}).Where("name = ?", name).Count(&count)

	return count > 0
}

func markSeedAsApplied(database *gorm.DB, name string) {
	database.Create(&seed.Seed{Name: name})
}
