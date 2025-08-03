package model

import (
	"errors"
	"net/http"
)

type AuthenticatePostPayload struct {
	Email     *string `json:"email" validate:"required" example:"admin@feldrise.com"` // the user's email
	Phone     *string `json:"phone" example:"+33612345678"`                           // the user's phone
	FirstName *string `json:"first_name" example:"Victor"`                            // the user's first name
	LastName  *string `json:"last_name" example:"DENIS"`                              // the user's last name
} // @name AuthenticatePostPayload

func (ap *AuthenticatePostPayload) Bind(r *http.Request) error {
	if ap.Email == nil {
		return errors.New("missing email property")
	}

	return nil
}

type VerifyAuthenticatePostPayload struct {
	Email *string `json:"email" validate:"required" example:"admin@feldrise.com"` // the user's email
	Code  *string `json:"code" validate:"required" example:"123456"`              // the verification code
} // @name VerifyAuthenticatePostPayload

func (vp *VerifyAuthenticatePostPayload) Bind(r *http.Request) error {
	if vp.Email == nil {
		return errors.New("missing email property")
	}

	if vp.Code == nil {
		return errors.New("missing code property")
	}

	return nil
}
