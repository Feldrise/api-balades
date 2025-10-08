package scheduler

import (
	"fmt"
	"log"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/notifications/email"
)

type RegistrationScheduler struct {
	rambleRepo       dbmodel.RambleRepository
	registrationRepo dbmodel.RambleRegistrationRepository
	groupRepo        dbmodel.RambleRegistrationGroupRepository
	emailService     email.EmailService
	applicationURL   string
}

func NewRegistrationScheduler(
	rambleRepo dbmodel.RambleRepository,
	registrationRepo dbmodel.RambleRegistrationRepository,
	groupRepo dbmodel.RambleRegistrationGroupRepository,
	emailService email.EmailService,
	applicationURL string,
) *RegistrationScheduler {
	return &RegistrationScheduler{
		rambleRepo:       rambleRepo,
		registrationRepo: registrationRepo,
		groupRepo:        groupRepo,
		emailService:     emailService,
		applicationURL:   applicationURL,
	}
}

// StartScheduler starts the background tasks for registration management
func (s *RegistrationScheduler) StartScheduler() {
	// Run every hour
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.processConfirmationRequests()
				s.processUnconfirmedRegistrations()
			}
		}
	}()

	s.processConfirmationRequests()
	s.processUnconfirmedRegistrations()

	log.Println("Registration scheduler started")
}

// processConfirmationRequests sends confirmation requests 3 days before the ramble
func (s *RegistrationScheduler) processConfirmationRequests() {
	threeDaysFromNow := time.Now().AddDate(0, 0, 3)

	registrations, err := s.registrationRepo.GetRegistrationsRequiringConfirmation(threeDaysFromNow)
	if err != nil {
		log.Printf("Error getting registrations requiring confirmation: %v", err)
		return
	}

	for _, registration := range registrations {
		s.sendConfirmationRequest(&registration)
	}
}

// processUnconfirmedRegistrations releases spots for unconfirmed registrations 24h before the ramble
func (s *RegistrationScheduler) processUnconfirmedRegistrations() {
	twentyFourHoursFromNow := time.Now().Add(24 * time.Hour)

	registrations, err := s.registrationRepo.GetUnconfirmedRegistrations(twentyFourHoursFromNow)
	if err != nil {
		log.Printf("Error getting unconfirmed registrations: %v", err)
		return
	}

	for _, registration := range registrations {
		s.cancelUnconfirmedRegistration(&registration)
	}
}

func (s *RegistrationScheduler) sendConfirmationRequest(registration *dbmodel.RambleRegistration) {
	if registration.Ramble.Date == nil {
		return
	}

	// Skip if ramble is cancelled
	if registration.Ramble.IsCancelled {
		return
	}

	// Set confirmation deadline to 24 hours before the ramble
	confirmationDeadline := registration.Ramble.Date.Add(-24 * time.Hour)
	registration.ConfirmationDeadline = &confirmationDeadline

	// Update registration with deadline
	_, err := s.registrationRepo.Update(registration)
	if err != nil {
		log.Printf("Error updating registration with confirmation deadline: %v", err)
		return
	}

	// Build reservation URL
	reservationURL := fmt.Sprintf("%s/mes-reservations", s.applicationURL)

	// Send confirmation email
	data := email.EventConfirmationRequestData{
		FirstName:            registration.FirstName,
		LastName:             registration.LastName,
		Title:                registration.Ramble.Title,
		Date:                 formatDate(registration.Ramble.Date),
		Location:             formatLocation(registration.Ramble.Location),
		ConfirmationDeadline: confirmationDeadline.Format("2 January 2006 à 15:04"),
		ReservationURL:       reservationURL,
	}

	err = s.emailService.Send(
		registration.Email,
		"Confirmation requise - Balade Écologique",
		"event_confirmation_request",
		data,
	)

	if err != nil {
		log.Printf("Failed to send confirmation request email to %s: %v", registration.Email, err)
	} else {
		log.Printf("Sent confirmation request to %s for ramble: %s", registration.Email, registration.Ramble.Title)
	}
}

func (s *RegistrationScheduler) cancelUnconfirmedRegistration(registration *dbmodel.RambleRegistration) {
	// Cancel the registration
	now := time.Now()
	registration.Status = "cancelled"
	registration.CancellationDate = &now
	registration.CancellationReason = func() *string { s := "No confirmation before deadline"; return &s }()

	_, err := s.registrationRepo.Update(registration)
	if err != nil {
		log.Printf("Error cancelling unconfirmed registration: %v", err)
		return
	}

	log.Printf("Cancelled unconfirmed registration for %s %s (ramble: %s)",
		registration.FirstName, registration.LastName, registration.Ramble.Title)

	// Try to promote someone from waiting list
	s.promoteFromWaitingList(registration.RambleID)
}

func (s *RegistrationScheduler) promoteFromWaitingList(rambleID uint) {
	// Get next person from waiting list
	nextRegistration, err := s.registrationRepo.GetNextInWaitingList(rambleID)
	if err != nil || nextRegistration == nil {
		return
	}

	// Get ramble details
	ramble, err := s.rambleRepo.FindByID(rambleID)
	if err != nil || ramble == nil {
		return
	}

	// Skip if ramble is cancelled
	if ramble.IsCancelled {
		return
	}

	// Check if there's still space
	confirmedCount, err := s.registrationRepo.CountByRambleAndStatus(rambleID, "confirmed")
	if err != nil {
		return
	}

	pendingCount, err := s.registrationRepo.CountByRambleAndStatus(rambleID, "pending")
	if err != nil {
		return
	}

	if ramble.MaxParticipants != nil {
		totalRegistered := confirmedCount + pendingCount
		if totalRegistered >= int64(*ramble.MaxParticipants) {
			return // Still full
		}
	}

	// Promote from waiting list
	nextRegistration.Status = "pending"
	_, err = s.registrationRepo.Update(nextRegistration)
	if err != nil {
		return
	}

	log.Printf("Promoted %s %s from waiting list for ramble: %s",
		nextRegistration.FirstName, nextRegistration.LastName, ramble.Title)

	// Send notification email
	data := email.SpotAvailableData{
		FirstName: nextRegistration.FirstName,
		LastName:  nextRegistration.LastName,
		Title:     ramble.Title,
		Date:      formatDate(ramble.Date),
		Location:  formatLocation(ramble.Location),
	}

	err = s.emailService.Send(
		nextRegistration.Email,
		"Place disponible - Balade Écologique",
		"spot_available",
		data,
	)

	if err != nil {
		log.Printf("Failed to send spot available email to %s: %v", nextRegistration.Email, err)
	}
}

func formatDate(date *time.Time) string {
	if date == nil {
		return "Date à confirmer"
	}
	return date.Format("2 January 2006 à 15:04")
}

func formatLocation(location *string) string {
	if location == nil {
		return "Lieu à confirmer"
	}
	return *location
}
