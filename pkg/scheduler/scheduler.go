package scheduler

import (
	"log"
	"time"

	"feldrise.com/balade/pkg/clock"
)

const (
	// tickPeriod is how often the payments left open are cancelled, so that a
	// payment can hardly be made once the ramble has started.
	tickPeriod = time.Minute
	// confirmationsPeriod is how often the confirmations are handled.
	confirmationsPeriod = time.Hour
)

// Reservations are the treatments run by the scheduler at a given time.
type Reservations interface {
	SendConfirmationRequests(now time.Time) error
	CancelUnconfirmedReservations(now time.Time) error
	CancelPaymentsOfStartedRambles(now time.Time) error
	CancelPaymentsOfCancelledReservations() error
}

// ReservationScheduler asks bookers to confirm before the ramble, cancels the
// reservations left unconfirmed at the deadline and the payments left open
// once the ramble has started or the reservation cancelled.
type ReservationScheduler struct {
	reservations Reservations
	clock        clock.Clock

	lastConfirmations time.Time
}

func NewReservationScheduler(reservations Reservations, clock clock.Clock) *ReservationScheduler {
	return &ReservationScheduler{reservations: reservations, clock: clock}
}

// StartScheduler runs the treatments now, then every minute in the background.
func (s *ReservationScheduler) StartScheduler() {
	s.Tick(s.clock.Now())

	ticker := time.NewTicker(tickPeriod)
	go func() {
		for range ticker.C {
			s.Tick(s.clock.Now())
		}
	}()

	log.Println("Reservation scheduler started")
}

// Tick runs the treatments due at now: the cancellation of the payments left
// open at every tick, the confirmations once an hour.
func (s *ReservationScheduler) Tick(now time.Time) {
	if s.lastConfirmations.IsZero() || now.Sub(s.lastConfirmations) >= confirmationsPeriod {
		s.lastConfirmations = now
		if err := s.reservations.SendConfirmationRequests(now); err != nil {
			log.Printf("Error sending confirmation requests: %v", err)
		}
		if err := s.reservations.CancelUnconfirmedReservations(now); err != nil {
			log.Printf("Error cancelling unconfirmed reservations: %v", err)
		}
	}

	if err := s.reservations.CancelPaymentsOfStartedRambles(now); err != nil {
		log.Printf("Error cancelling the payments of started rambles: %v", err)
	}
	if err := s.reservations.CancelPaymentsOfCancelledReservations(); err != nil {
		log.Printf("Error cancelling the payments of cancelled reservations: %v", err)
	}
}
