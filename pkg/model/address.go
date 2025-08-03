package model

type Address struct {
	Number        int       `json:"number"`
	Route         string    `json:"route"`
	OptionalRoute *string   `json:"optionalRoute,omitempty"`
	City          string    `json:"city"`
	ZipCode       string    `json:"zipCode"`
	Country       string    `json:"country"`
	Coordinates   []float64 `json:"coordinates"`
}
