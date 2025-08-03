package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/helper"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/notifications/email"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

// Authenticate godoc
// @Summary Authenticate a user
// @Description Authenticate a user by sending a unique code
// @ID authenticate
// @Tags autentication
// @Param request body AuthenticatePostPayload true "user's info"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /authentication/authenticate [post]
func (config *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	data := &model.AuthenticatePostPayload{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	dbUser, err := config.UserRepository.FindByEmail(*data.Email, &dbmodel.UserFieldsToInclude{
		UserProfile: true,
	})

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbUser == nil {
		if data.FirstName == nil || *data.FirstName == "" {
			render.Render(w, r, errors.ErrInvalidRequest(nil))
			return
		}

		// We create a new user
		dbUser = &dbmodel.User{
			Email: *data.Email,
			UserProfile: dbmodel.UserProfile{
				FirstName: *data.FirstName,
				LastName:  data.LastName,
				Phone:     data.Phone,
			},
		}

		dbUser, err = config.UserRepository.Create(dbUser)

		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	// Generate a new code
	code, err := GenerateOTP(6)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	nowPlusFive := time.Now().Add(5 * time.Minute)

	dbUser.AuthenticationCode = &code
	dbUser.AuthenticationExpireAt = &nowPlusFive

	dbUser, err = config.UserRepository.Update(dbUser)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	err = config.EmailService.Send(
		dbUser.Email,
		"Code de vérification",
		"authentication",
		email.AuthenticationData{
			Code: code,
		},
	)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, "success")
}

// VerifyAuthenticate godoc
// @Summary Verify the authentication code
// @Description Verify the authentication code
// @ID verify-authenticate
// @Tags autentication
// @Param request body VerifyAuthenticatePostPayload true "user's info"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /authentication/verify-authentication [post]
func (config *Config) VerifyAuthenticate(w http.ResponseWriter, r *http.Request) {
	data := &model.VerifyAuthenticatePostPayload{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	dbUser, err := config.UserRepository.FindByEmail(*data.Email, &dbmodel.UserFieldsToInclude{
		UserProfile: true,
	})

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("invalid email"))
		return
	}

	if dbUser.AuthenticationCode == nil {
		render.Render(w, r, errors.ErrUnauthorized("invalid authentication code"))
		return
	}

	if *dbUser.AuthenticationCode != *data.Code {
		render.Render(w, r, errors.ErrUnauthorized("invalid authentication code"))
		return
	}

	if dbUser.AuthenticationExpireAt == nil || dbUser.AuthenticationExpireAt.Before(time.Now()) {
		render.Render(w, r, errors.ErrUnauthorized("authentication code expired"))
		return
	}

	tokenString, err := GenerateToken(config.Constants.JWTSecret, dbUser.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	dbUser, err = config.UserRepository.Update(dbUser)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	type UserWithToken struct {
		User  model.User `json:"user"`
		Token string     `json:"token"`
	}

	render.JSON(w, r, &UserWithToken{
		User:  *dbUser.ToModel(),
		Token: tokenString,
	})
}

// Update godoc
// @Summary Update the current user
// @Description Update the current user
// @ID update
// @Tags autentication
// @Param request body map[string]interface{} true "user's info"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /authentication/{id} [put]
func (config *Config) Update(w http.ResponseWriter, r *http.Request) {
	loggedUser := ForContext(r.Context())

	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("unauthorized"))
		return
	}

	var data map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	id := chi.URLParam(r, "id")

	// Convert id to uint
	idUint, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if loggedUser.ID != uint(idUint) {
		render.Render(w, r, errors.ErrUnauthorized("unauthorized"))
		return
	}

	helper.ApplyChanges(data, loggedUser)

	user, err := config.UserRepository.Update(loggedUser)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, user.ToModel())
}

// Me godoc
// @Summary Get the current user
// @Description Get the current user
// @ID me
// @Tags autentication
// @Success 200 {string} string "ok"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /authentication/me [get]
func (config *Config) Me(w http.ResponseWriter, r *http.Request) {
	ctxUser := ForContext(r.Context())

	if ctxUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("invalid user"))
		return
	}

	render.JSON(w, r, ctxUser.ToModel())
}

// CheckIfEmailExists godoc
// @Summary Check if the email exists
// @Description Check if the email exists
// @ID check-if-email-exists
// @Tags autentication
// @Param email query string true "email"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /authentication/check-email [get]
func (config *Config) CheckIfEmailExists(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email == "" {
		render.Render(w, r, errors.ErrInvalidRequest(nil))
		return
	}

	user, err := config.UserRepository.FindByEmail(email, nil)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if user != nil {
		// Here if it exists we return a 401 to be able to display an error message
		render.Render(w, r, errors.ErrUnauthorized("email already exists"))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, "not exists")
}

// Private

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}

	return string(bytes), err
}

func checkPassword(initialPassword string, providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(initialPassword), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func GenerateOTP(length int) (string, error) {
	const otpChars = "1234567890"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}
