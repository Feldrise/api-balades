package seed

import (
	"log"

	"gorm.io/gorm"
)

// SeedV8 moves pending confirmation deadlines from 24h to 12h before upcoming rambles.
func SeedV8(database *gorm.DB) error {
	registrationsResult := database.Exec(`
		UPDATE ramble_registrations AS rr
		SET confirmation_deadline = r.date - INTERVAL '12 hours',
		    updated_at = NOW()
		FROM rambles AS r
		WHERE rr.ramble_id = r.id
		  AND rr.deleted_at IS NULL
		  AND r.deleted_at IS NULL
		  AND rr.status = 'pending'
		  AND rr.confirmation_deadline IS NOT NULL
		  AND r.date IS NOT NULL
		  AND r.date > NOW()
		  AND rr.confirmation_deadline > NOW()
		  AND rr.confirmation_deadline < r.date - INTERVAL '12 hours'
	`)
	if registrationsResult.Error != nil {
		return registrationsResult.Error
	}

	groupsResult := database.Exec(`
		UPDATE ramble_registration_groups AS rg
		SET confirmation_deadline = r.date - INTERVAL '12 hours',
		    updated_at = NOW()
		FROM rambles AS r
		WHERE rg.ramble_id = r.id
		  AND rg.deleted_at IS NULL
		  AND r.deleted_at IS NULL
		  AND rg.status = 'pending'
		  AND rg.confirmation_deadline IS NOT NULL
		  AND r.date IS NOT NULL
		  AND r.date > NOW()
		  AND rg.confirmation_deadline > NOW()
		  AND rg.confirmation_deadline < r.date - INTERVAL '12 hours'
	`)
	if groupsResult.Error != nil {
		return groupsResult.Error
	}

	log.Printf(
		"SeedV8: updated confirmation deadlines (registrations: %d, groups: %d)",
		registrationsResult.RowsAffected,
		groupsResult.RowsAffected,
	)

	return nil
}
