package model

import "time"

type User struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`

	Profile *UserProfile `json:"profile,omitempty"`
}
