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
