package clock

import "time"

// Clock gives the current time. It is injected so that time-dependent rules
// (auto-confirmation, scheduler deadlines) can be driven from tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

// Real returns a Clock backed by the system time.
func Real() Clock {
	return realClock{}
}

func (realClock) Now() time.Time {
	return time.Now()
}
