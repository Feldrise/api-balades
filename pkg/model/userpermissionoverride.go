package model

import "time"

type UserPermissionOverride struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	UserID      uint    `json:"userId"`
	Permission  string  `json:"permission"`
	Override    bool    `json:"override"`
	Description *string `json:"description,omitempty"`
}
