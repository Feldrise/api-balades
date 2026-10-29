package email

//go:generate mockgen -source=types.go -destination=../../mocks/mock_emailservice.go -package=mocks

// Templates

type AuthenticationData struct {
	Code string
}

type ReservationConfirmationData struct {
	Title    string
	Date     string
	Location string
	// Status is pending, confirmed, waiting_list or cancelled.
	Status string
	// Change is empty for a new reservation, or participants_added,
	// participants_removed, promoted (left the waiting list) or status_updated
	// (set by an admin).
	Change       string
	Participants []ReservationParticipantInfo
	// HasPrices tells whether the ramble has prices, so an estimated amount.
	HasPrices bool
	// EstimatedAmount is formatted, e.g. "18,50 €", and sums the known prices.
	EstimatedAmount string
	// HasMissingPrices tells that a participant's price is unknown.
	HasMissingPrices bool
	PaymentEnabled   bool
	ReservationURL   string
}

type ReservationParticipantInfo struct {
	Name       string
	PriceLabel string
	// Price is formatted, e.g. "12,50 €".
	Price string
	// PriceMissing tells that the ramble has prices but the participant's is unknown.
	PriceMissing bool
}

type EventConfirmationRequestData struct {
	FirstName            string
	LastName             string
	Title                string
	Date                 string
	Location             string
	ConfirmationDeadline string
	ReservationURL       string
}

type RambleCancellationData struct {
	FirstName          string
	LastName           string
	Title              string
	Date               string
	Location           string
	CancellationReason string
}

// Service

type Credentials struct {
	Host     string
	Port     int
	Email    string
	Password string
}

type EmailService interface {
	Send(to string, subject string, templateName string, data interface{}, attachments ...string) error
}
