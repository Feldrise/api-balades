package reservation

import (
	"fmt"
	"log"
	"strings"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/notifications/email"
)

// sendReservationConfirmationEmail sends the reservation summary to the booker
// only. A failure is logged without failing the reservation.
func (config *Config) sendReservationConfirmationEmail(reservation *dbmodel.Reservation) {
	ramble := reservation.Ramble

	participants := make([]email.ReservationParticipantInfo, len(reservation.Participants))
	for i, participant := range reservation.Participants {
		participants[i] = email.ReservationParticipantInfo{
			Name:         participant.FirstName + " " + participant.LastName,
			PriceLabel:   participant.PriceLabel,
			Price:        formatEuros(ramble.PriceAmount(participant.PriceLabel)),
			PriceMissing: ramble.PriceMissing(participant.PriceLabel),
		}
	}

	data := email.ReservationConfirmationData{
		Title:            ramble.Title,
		Date:             formatDate(ramble.Date),
		Location:         formatLocation(ramble.Location),
		Status:           reservation.Status,
		Participants:     participants,
		HasPrices:        len(ramble.Prices) > 0,
		EstimatedAmount:  formatCents(reservation.EstimatedAmount()),
		HasMissingPrices: reservation.HasMissingPrices(),
		PaymentEnabled:   ramble.PaymentEnabled,
		ReservationURL:   fmt.Sprintf("%s/mes-reservations/%d", config.ApplicationURL, reservation.ID),
	}

	err := config.EmailService.Send(
		reservation.ContactEmail,
		"Confirmation de réservation - Balade Écologique",
		"reservation_confirmation",
		data,
	)
	if err != nil {
		log.Printf("Failed to send reservation confirmation email for reservation %d: %v", reservation.ID, err)
	}
}

func formatEuros(amount float64) string {
	return strings.Replace(fmt.Sprintf("%.2f €", amount), ".", ",", 1)
}

func formatCents(amount int64) string {
	return formatEuros(float64(amount) / 100)
}

func formatDate(date *time.Time) string {
	if date == nil {
		return "Date à confirmer"
	}
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return date.Local().Format("2 January 2006 à 15:04")
	}
	return date.In(loc).Format("2 January 2006 à 15:04")
}

func formatLocation(location *string) string {
	if location == nil {
		return "Lieu à confirmer"
	}
	return *location
}
