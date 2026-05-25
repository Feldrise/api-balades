package model

import (
	"errors"
	"net/http"
	"time"
)

type Guide struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 *string `json:"phone,omitempty"`
	Bio                   *string `json:"bio,omitempty"`
	Experience            *string `json:"experience,omitempty"`
	Specialties           *string `json:"specialties,omitempty"`
	Languages             *string `json:"languages,omitempty"`
	CertificationLevel    *string `json:"certification_level,omitempty"`
	Avatar                *string `json:"avatar,omitempty"`
	IsActive              bool    `json:"is_active"`
	EmergencyContactName  *string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string `json:"emergency_contact_phone,omitempty"`

	// Payment configuration (sensitive fields excluded for security)
	StripeAccountID *string `json:"stripe_account_id,omitempty"`
	StripePublicKey *string `json:"stripe_public_key,omitempty"`
	PaymentEnabled  bool    `json:"payment_enabled"`
} // @name Guide

type GuideCreatePayload struct {
	FirstName             *string `json:"first_name" binding:"required"`
	LastName              *string `json:"last_name" binding:"required"`
	Email                 *string `json:"email" binding:"required"`
	Phone                 *string `json:"phone,omitempty"`
	Bio                   *string `json:"bio,omitempty"`
	Experience            *string `json:"experience,omitempty"`
	Specialties           *string `json:"specialties,omitempty"`
	Languages             *string `json:"languages,omitempty"`
	CertificationLevel    *string `json:"certification_level,omitempty"`
	AvatarBase64          *string `json:"avatar_base64,omitempty"` // Base64 encoded image
	EmergencyContactName  *string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string `json:"emergency_contact_phone,omitempty"`
} // @name GuideCreatePayload

type GuidePaymentConfigPayload struct {
	StripeAccountID     *string `json:"stripe_account_id,omitempty"`
	StripePublicKey     *string `json:"stripe_public_key,omitempty"`
	StripeSecretKey     *string `json:"stripe_secret_key,omitempty"`
	StripeWebhookSecret *string `json:"stripe_webhook_secret,omitempty"`
	PaymentEnabled      *bool   `json:"payment_enabled,omitempty"`
} // @name GuidePaymentConfigPayload

type GuideLinkUserPayload struct {
	UserID *uint `json:"user_id" binding:"required"`
} // @name GuideLinkUserPayload

func (g *GuideCreatePayload) Bind(r *http.Request) error {
	if g.FirstName == nil {
		return errors.New("first_name is required")
	}

	if g.LastName == nil {
		return errors.New("last_name is required")
	}

	if g.Email == nil {
		return errors.New("email is required")
	}

	return nil
}

func (g *GuideLinkUserPayload) Bind(r *http.Request) error {
	if g.UserID == nil {
		return errors.New("user_id is required")
	}

	return nil
}
