package seed

import (
	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

func SeedV5(database *gorm.DB) error {
	// Create payment-related permissions
	managePaymentsDesc := "Full payment management - create, view, refund payments"
	viewPaymentsDesc := "View payment information"
	configureGuidePaymentsDesc := "Configure guide payment settings (Stripe credentials)"
	refundPaymentsDesc := "Process payment refunds"
	webhookPaymentsDesc := "Handle payment webhook events"

	permissions := []dbmodel.Permission{
		{Model: gorm.Model{ID: 21}, Name: "manage:payments", Description: &managePaymentsDesc},
		{Model: gorm.Model{ID: 22}, Name: "view:payments", Description: &viewPaymentsDesc},
		{Model: gorm.Model{ID: 23}, Name: "configure:guide-payments", Description: &configureGuidePaymentsDesc},
		{Model: gorm.Model{ID: 24}, Name: "refund:payments", Description: &refundPaymentsDesc},
		{Model: gorm.Model{ID: 25}, Name: "webhook:payments", Description: &webhookPaymentsDesc},
	}

	for _, permission := range permissions {
		database.Create(&permission)
	}

	// Add permissions to admin role
	admin := dbmodel.Role{}
	database.Preload("Permissions").First(&admin, 1)

	admin.Permissions = append(
		admin.Permissions,
		permissions[0], // manage:payments
		permissions[1], // view:payments
		permissions[2], // configure:guide-payments
		permissions[3], // refund:payments
		permissions[4], // webhook:payments
	)

	database.Save(&admin)

	return nil
}
