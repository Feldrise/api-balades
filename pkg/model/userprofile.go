package model

import "time"

type UserProfile struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	FirstName  string  `json:"firstName"`
	LastName   *string `json:"lastName,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	AddressID  *uint   `json:"addressId,omitempty"`
	AvatarName *string `json:"avatarName,omitempty"`
}
