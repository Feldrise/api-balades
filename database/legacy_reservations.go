package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"gorm.io/gorm"
)

// legacyStatusRank orders the statuses of a split group: the reservation takes
// the most advanced status that is not cancelled.
var legacyStatusRank = map[string]int{
	dbmodel.ReservationStatusWaitingList: 1,
	dbmodel.ReservationStatusPending:     2,
	dbmodel.ReservationStatusConfirmed:   3,
}

// MigrateLegacyRegistrations turns the individual registrations and the
// registration groups into reservations, with their payments. It runs at every
// startup and only migrates what has not been migrated yet.
func MigrateLegacyRegistrations(database *gorm.DB) error {
	var migratedRegistrations, migratedGroups []uint
	if err := database.Unscoped().Model(&dbmodel.Reservation{}).Where("legacy_registration_id IS NOT NULL").Pluck("legacy_registration_id", &migratedRegistrations).Error; err != nil {
		return err
	}
	if err := database.Unscoped().Model(&dbmodel.Reservation{}).Where("legacy_group_id IS NOT NULL").Pluck("legacy_group_id", &migratedGroups).Error; err != nil {
		return err
	}

	groupsQuery := database.Preload("Registrations", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).Order("id ASC")
	if len(migratedGroups) > 0 {
		groupsQuery = groupsQuery.Where("id NOT IN ?", migratedGroups)
	}
	var groups []dbmodel.RambleRegistrationGroup
	if err := groupsQuery.Find(&groups).Error; err != nil {
		return err
	}

	registrationsQuery := database.Where("group_id IS NULL").Order("id ASC")
	if len(migratedRegistrations) > 0 {
		registrationsQuery = registrationsQuery.Where("id NOT IN ?", migratedRegistrations)
	}
	var registrations []dbmodel.RambleRegistration
	if err := registrationsQuery.Find(&registrations).Error; err != nil {
		return err
	}

	for _, group := range groups {
		if len(group.Registrations) == 0 {
			continue
		}
		if err := database.Transaction(func(tx *gorm.DB) error { return migrateLegacyGroup(tx, group) }); err != nil {
			return fmt.Errorf("migrating registration group %d: %w", group.ID, err)
		}
	}

	for _, registration := range registrations {
		if err := database.Transaction(func(tx *gorm.DB) error { return migrateLegacyRegistration(tx, registration) }); err != nil {
			return fmt.Errorf("migrating registration %d: %w", registration.ID, err)
		}
	}

	if len(groups)+len(registrations) > 0 {
		log.Printf("Migrated legacy registrations to reservations (groups: %d, registrations: %d)", len(groups), len(registrations))
	}

	return nil
}

func migrateLegacyRegistration(tx *gorm.DB, registration dbmodel.RambleRegistration) error {
	user, err := findOrCreateLegacyBooker(tx, registration.Email, registration)
	if err != nil {
		return err
	}

	priceLabel, err := legacyPriceLabel(tx, registration.RambleID)
	if err != nil {
		return err
	}

	reservation := &dbmodel.Reservation{
		Model:                gorm.Model{CreatedAt: legacyCreationDate(registration.RegistrationDate, registration.CreatedAt)},
		RambleID:             registration.RambleID,
		UserID:               user.ID,
		ContactEmail:         registration.Email,
		Status:               registration.Status,
		ConfirmationDeadline: registration.ConfirmationDeadline,
		ConfirmationDate:     registration.ConfirmationDate,
		CancellationDate:     registration.CancellationDate,
		CancellationReason:   registration.CancellationReason,
		LegacyRegistrationID: &registration.ID,
		Participants: []dbmodel.Participant{{
			FirstName:  registration.FirstName,
			LastName:   registration.LastName,
			PriceLabel: priceLabel,
			IsBooker:   true,
		}},
	}
	if err := tx.Create(reservation).Error; err != nil {
		return err
	}

	return tx.Model(&dbmodel.Payment{}).
		Where("registration_id = ? AND reservation_id IS NULL", registration.ID).
		Update("reservation_id", reservation.ID).Error
}

func migrateLegacyGroup(tx *gorm.DB, group dbmodel.RambleRegistrationGroup) error {
	status := dbmodel.ReservationStatusCancelled
	for _, registration := range group.Registrations {
		if legacyStatusRank[registration.Status] > legacyStatusRank[status] {
			status = registration.Status
		}
	}

	// Cancelled participants of an active group are not migrated; a cancelled
	// group keeps all its participants.
	kept := group.Registrations
	if status != dbmodel.ReservationStatusCancelled {
		kept = nil
		for _, registration := range group.Registrations {
			if registration.Status != dbmodel.ReservationStatusCancelled {
				kept = append(kept, registration)
			}
		}
	}

	var contact *dbmodel.RambleRegistration
	for i := range group.Registrations {
		if strings.EqualFold(group.Registrations[i].Email, group.PrimaryEmail) {
			contact = &group.Registrations[i]
			break
		}
	}
	if contact == nil {
		contact = &group.Registrations[0]
	}

	user, err := findOrCreateLegacyBooker(tx, group.PrimaryEmail, *contact)
	if err != nil {
		return err
	}

	priceLabel, err := legacyPriceLabel(tx, group.RambleID)
	if err != nil {
		return err
	}

	reservation := &dbmodel.Reservation{
		Model:                gorm.Model{CreatedAt: legacyCreationDate(group.RegistrationDate, group.CreatedAt)},
		RambleID:             group.RambleID,
		UserID:               user.ID,
		ContactEmail:         group.PrimaryEmail,
		Status:               status,
		ConfirmationDeadline: group.ConfirmationDeadline,
		ConfirmationDate:     group.ConfirmationDate,
		LegacyGroupID:        &group.ID,
	}
	if status == dbmodel.ReservationStatusCancelled {
		reservation.CancellationDate = group.CancellationDate
		reservation.CancellationReason = group.CancellationReason
	}

	registrationIDs := make([]uint, len(group.Registrations))
	for i, registration := range group.Registrations {
		registrationIDs[i] = registration.ID
	}

	for _, registration := range kept {
		reservation.Participants = append(reservation.Participants, dbmodel.Participant{
			FirstName:  registration.FirstName,
			LastName:   registration.LastName,
			PriceLabel: priceLabel,
			IsBooker:   strings.EqualFold(registration.Email, group.PrimaryEmail),
		})

		// Fill the dates the group lacks from its participants.
		if reservation.ConfirmationDeadline == nil {
			reservation.ConfirmationDeadline = registration.ConfirmationDeadline
		}
		if reservation.ConfirmationDate == nil && status == dbmodel.ReservationStatusConfirmed {
			reservation.ConfirmationDate = registration.ConfirmationDate
		}
		if reservation.CancellationDate == nil && status == dbmodel.ReservationStatusCancelled {
			reservation.CancellationDate = registration.CancellationDate
			reservation.CancellationReason = registration.CancellationReason
		}
	}

	if err := tx.Create(reservation).Error; err != nil {
		return err
	}

	return tx.Model(&dbmodel.Payment{}).
		Where("(group_id = ? OR registration_id IN ?) AND reservation_id IS NULL", group.ID, registrationIDs).
		Update("reservation_id", reservation.ID).Error
}

// findOrCreateLegacyBooker returns the account of email, preferring the account
// already linked to the registration.
func findOrCreateLegacyBooker(tx *gorm.DB, email string, registration dbmodel.RambleRegistration) (*dbmodel.User, error) {
	var user dbmodel.User

	if registration.UserID != nil && strings.EqualFold(registration.Email, email) {
		err := tx.First(&user, *registration.UserID).Error
		if err == nil {
			return &user, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	err := tx.Where("email = ?", email).First(&user).Error
	if err == nil {
		return &user, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	lastName := registration.LastName
	user = dbmodel.User{
		Email: email,
		UserProfile: dbmodel.UserProfile{
			FirstName: registration.FirstName,
			LastName:  &lastName,
			Phone:     registration.Phone,
		},
	}
	if err := tx.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// legacyPriceLabel is the price applied to migrated participants: registrations
// had no price, so only a ramble with a single price gives one.
func legacyPriceLabel(tx *gorm.DB, rambleID uint) (string, error) {
	var labels []string
	if err := tx.Model(&dbmodel.RamblePrice{}).Where("ramble_id = ?", rambleID).Pluck("label", &labels).Error; err != nil {
		return "", err
	}
	if len(labels) == 1 {
		return labels[0], nil
	}
	return "", nil
}

func legacyCreationDate(registrationDate time.Time, createdAt time.Time) time.Time {
	if registrationDate.IsZero() {
		return createdAt
	}
	return registrationDate
}
