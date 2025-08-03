package model

import "time"

type Permission struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name         string  `json:"name"`
	ReadableName *string `json:"readableName,omitempty"`
	Description  *string `json:"description,omitempty"`
	Category     *string `json:"category,omitempty"`
}
