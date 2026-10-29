package email

//go:generate mockgen -source=types.go -destination=../../mocks/mock_emailservice.go -package=mocks

// Templates

type AuthenticationData struct {
	Code string
}

type RegistrationConfirmationData struct {
	FirstName        string
	LastName         string
	Title            string
	Date             string
	Location         string
	Status           string
	IsGroup          bool
	ParticipantCount int
	ParticipantNames []string
	// Payment information
	PaymentEnabled  bool
	PaymentRequired bool
	PaymentURL      string
	Prices          []PriceInfo
}

type ReservationConfirmationData struct {
	Title    string
	Date     string
	Location string
	// Status is pending, confirmed or waiting_list.
	Status       string
	Participants []ReservationParticipantInfo
	// HasPrices tells whether the ramble has prices, so an estimated amount.
	HasPrices bool
	// EstimatedAmount is formatted, e.g. "18,50 €".
	EstimatedAmount string
	PaymentEnabled  bool
	ReservationURL  string
}

type ReservationParticipantInfo struct {
	Name       string
	PriceLabel string
	// Price is formatted, e.g. "12,50 €".
	Price string
}

type PriceInfo struct {
	Label  string
	Amount float64
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

type SpotAvailableData struct {
	FirstName string
	LastName  string
	Title     string
	Date      string
	Location  string
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
