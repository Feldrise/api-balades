package scheduler_test

import (
	"testing"
	"time"

	"feldrise.com/balade/pkg/scheduler"
)

// recordedReservations counts the treatments run.
type recordedReservations struct {
	confirmationRequests int
	unconfirmedCancels   int
	reconciledPayments   int
}

func (r *recordedReservations) SendConfirmationRequests(now time.Time) error {
	r.confirmationRequests++
	return nil
}

func (r *recordedReservations) CancelUnconfirmedReservations(now time.Time) error {
	r.unconfirmedCancels++
	return nil
}

func (r *recordedReservations) ReconcileOpenPayments() error {
	r.reconciledPayments++
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestSchedulerClosesPaymentsEveryMinuteAndRunsTheOtherTreatmentsEveryHour(t *testing.T) {
	start := time.Date(2026, 11, 8, 9, 0, 0, 0, time.UTC)
	reservations := &recordedReservations{}
	s := scheduler.NewReservationScheduler(reservations, fixedClock{start})

	for minute := range 121 {
		s.Tick(start.Add(time.Duration(minute) * time.Minute))
	}

	if reservations.reconciledPayments != 121 {
		t.Errorf("expected the open payments reconciled at each of the 121 minutes, got %d", reservations.reconciledPayments)
	}
	if reservations.confirmationRequests != 3 || reservations.unconfirmedCancels != 3 {
		t.Errorf("expected the confirmations handled at 9:00, 10:00 and 11:00 only, got %d and %d", reservations.confirmationRequests, reservations.unconfirmedCancels)
	}
}
