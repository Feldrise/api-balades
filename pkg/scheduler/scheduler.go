package scheduler

import (
	"log"
	"time"

	"feldrise.com/balade/pkg/clock"
)

// Reservations are the treatments run by the scheduler at a given time.
type Reservations interface {
	SendConfirmationRequests(now time.Time) error
	CancelUnconfirmedReservations(now time.Time) error
	CancelPaymentsOfStartedRambles(now time.Time) error
}

// ReservationScheduler asks bookers to confirm before the ramble, cancels the
// reservations left unconfirmed at the deadline and the payments left open
// once the ramble has started.
type ReservationScheduler struct {
	reservations Reservations
	clock        clock.Clock
}

func NewReservationScheduler(reservations Reservations, clock clock.Clock) *ReservationScheduler {
	return &ReservationScheduler{reservations: reservations, clock: clock}
}

// StartScheduler runs the treatments now, then every hour in the background.
func (s *ReservationScheduler) StartScheduler() {
	s.run()

	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			s.run()
		}
	}()

	log.Println("Reservation scheduler started")
}

func (s *ReservationScheduler) run() {
	now := s.clock.Now()
	if err := s.reservations.SendConfirmationRequests(now); err != nil {
		log.Printf("Error sending confirmation requests: %v", err)
	}
	if err := s.reservations.CancelUnconfirmedReservations(now); err != nil {
		log.Printf("Error cancelling unconfirmed reservations: %v", err)
	}
	if err := s.reservations.CancelPaymentsOfStartedRambles(now); err != nil {
		log.Printf("Error cancelling the payments of started rambles: %v", err)
	}
}
