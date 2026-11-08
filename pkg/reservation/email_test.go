package reservation

import (
	"strings"
	"testing"
	"time"

	"feldrise.com/balade/config"
	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/notifications/email"
)

// renderingEmailService keeps the bodies of the emails rendered with the real templates.
type renderingEmailService struct {
	bodies []string
}

func (s *renderingEmailService) Send(to string, subject string, templateName string, data interface{}, attachments ...string) error {
	body, err := email.Render("../../emails", templateName, data)
	if err != nil {
		return err
	}
	s.bodies = append(s.bodies, body)
	return nil
}

// No route sends the summary of a reservation with missing prices yet (only
// migrated reservations have some), so the email is sent directly.
func TestReservationEmailDoesNotPresentAMissingPriceAsFree(t *testing.T) {
	emails := &renderingEmailService{}
	reservationConfig := &Config{&config.Config{EmailService: emails}}
	date := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)

	reservationConfig.sendReservationConfirmationEmail(&dbmodel.Reservation{
		ContactEmail: "parent@example.com",
		Status:       dbmodel.ReservationStatusPending,
		Ramble: dbmodel.Ramble{
			Title:  "Balade des oiseaux",
			Date:   &date,
			Prices: []dbmodel.RamblePrice{{Label: "Adulte", Amount: 12.5}, {Label: "Enfant", Amount: 6}},
		},
		Participants: []dbmodel.Participant{
			{FirstName: "Paul", LastName: "Martin", PriceLabel: "Adulte"},
			{FirstName: "Léa", LastName: "Martin"},
		},
	})

	if len(emails.bodies) != 1 {
		t.Fatalf("expected a single rendered email, got %d", len(emails.bodies))
	}
	body := emails.bodies[0]
	if !strings.Contains(body, "Léa Martin – tarif à préciser") {
		t.Errorf("expected Léa's price to be announced as to be specified, got %s", body)
	}
	if !strings.Contains(body, "12,50 € hors tarifs à préciser") {
		t.Errorf("expected the estimated amount to leave out the missing price, got %s", body)
	}
	if strings.Contains(body, "0,00 €") {
		t.Errorf("expected no price presented as 0,00 €, got %s", body)
	}
}
