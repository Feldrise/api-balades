package seed

import (
	"log"

	"gorm.io/gorm"
)

// SeedV9 backfills published_at for existing rambles so they remain publicly visible.
func SeedV9(database *gorm.DB) error {
	result := database.Exec(`
		UPDATE rambles
		SET published_at = COALESCE(date, created_at, NOW()),
		    updated_at = NOW()
		WHERE published_at IS NULL
		  AND deleted_at IS NULL
	`)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("SeedV9: backfilled published_at for %d rambles", result.RowsAffected)
	return nil
}
