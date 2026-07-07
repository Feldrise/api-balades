package guide

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/helper"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"feldrise.com/balade/pkg/security"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Get godoc
// @Summary Get Guide
// @Description Get a guide by ID
// @ID get-guide
// @Param id path int true "Guide ID"
// @Success 200 {object} Guide
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/{id} [get]
func (config *Config) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Convert id to uint
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	dbGuide, err := config.GuideRepository.FindByID(uint(idUint))

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	render.JSON(w, r, dbGuide.ToModel())
}

// GetAll godoc
// @Summary Get all Guides
// @Description Get all guides with optional active filter
// @ID get-all-guides
// @Param is_active query bool false "Filter by active status (true/false)"
// @Success 200 {array} Guide
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /guides [get]
func (config *Config) GetAll(w http.ResponseWriter, r *http.Request) {
	isActiveStr := r.URL.Query().Get("is_active")
	search := r.URL.Query().Get("search")
	filter := &dbmodel.GuideFilter{}

	if isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			render.Render(w, r, errors.ErrInvalidRequest(err))
			return
		}
		filter.IsActive = &isActive
	}

	if search != "" {
		filter.Search = &search
	}

	dbGuides, err := config.GuideRepository.FindAll(filter)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	guides := make([]model.Guide, len(dbGuides))
	for i, dbGuide := range dbGuides {
		guides[i] = dbGuide.ToModel()
	}

	render.JSON(w, r, guides)
}

// Create godoc
// @Summary Create Guide
// @Description Create a new guide
// @ID create-guide
// @Accept json
// @Produce json
// @Param guide body GuideCreatePayload true "Guide data"
// @Success 201 {object} Guide
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /guides [post]
func (config *Config) Create(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())

	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	var payload model.GuideCreatePayload

	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	dbGuide := &dbmodel.Guide{
		FirstName:             *payload.FirstName,
		LastName:              *payload.LastName,
		Email:                 *payload.Email,
		Phone:                 payload.Phone,
		Bio:                   payload.Bio,
		Experience:            payload.Experience,
		Specialties:           payload.Specialties,
		Languages:             payload.Languages,
		CertificationLevel:    payload.CertificationLevel,
		IsActive:              true, // Default to active
		EmergencyContactName:  payload.EmergencyContactName,
		EmergencyContactPhone: payload.EmergencyContactPhone,
	}

	// Create the guide first to get the ID
	dbGuide, err := config.GuideRepository.Create(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Handle avatar upload if provided
	if payload.AvatarBase64 != nil && *payload.AvatarBase64 != "" {
		guideID := fmt.Sprintf("%d", dbGuide.ID)
		filename, err := helper.SaveBase64Image(*payload.AvatarBase64, config.Constants.DataPath, "guide", guideID)
		if err != nil {
			render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("failed to save avatar: %w", err)))
			return
		}
		dbGuide.Avatar = &filename
		dbGuide, err = config.GuideRepository.Update(dbGuide)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	if err := config.linkGuideToUserByEmail(dbGuide); err != nil {
		fmt.Printf("Failed to auto-link guide to user: %v\n", err)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dbGuide.ToModel())
}

// Update godoc
// @Summary Update an existing guide
// @Description Update an existing guide with the provided data
// @ID update-guide
// @Accept json
// @Produce json
// @Param id path int true "Guide ID"
// @Param guide body map[string]interface{} true "Guide data"
// @Success 200 {object} model.Guide
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/{id} [put]
func (config *Config) Update(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())

	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
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

	dbGuide, err := config.GuideRepository.FindByID(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	var avatarBase64 string
	if avatarInterface, exists := data["avatar_base64"]; exists {
		if avatarStr, ok := avatarInterface.(string); ok {
			avatarBase64 = avatarStr
		}
	}
	delete(data, "avatar_base64")

	helper.ApplyChanges(data, dbGuide)

	if avatarBase64 != "" {
		guideID := fmt.Sprintf("%d", dbGuide.ID)
		filename, err := helper.SaveBase64Image(avatarBase64, config.Constants.DataPath, "guide", guideID)
		if err != nil {
			render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("failed to save avatar: %w", err)))
			return
		}
		dbGuide.Avatar = &filename
	}

	dbGuide, err = config.GuideRepository.Update(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, dbGuide.ToModel())
}

// Delete godoc
// @Summary Delete a guide
// @Description Delete a guide by its ID
// @ID delete-guide
// @Param id path int true "Guide ID"
// @Success 204
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/{id} [delete]
func (config *Config) Delete(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())

	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	id := chi.URLParam(r, "id")

	// Convert id to uint
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	dbGuide, err := config.GuideRepository.FindByID(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	err = config.GuideRepository.Delete(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.NoContent(w, r)
}

// UpdatePaymentConfig godoc
// @Summary Update payment configuration for a guide
// @Description Update payment configuration for a guide including Stripe credentials
// @ID update-guide-payment-config
// @Accept json
// @Produce json
// @Param id path int true "Guide ID"
// @Param config body model.GuidePaymentConfigPayload true "Payment configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/{id}/payment-config [put]
func (config *Config) UpdatePaymentConfig(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())

	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	id := chi.URLParam(r, "id")

	// Convert id to uint
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	var payload model.GuidePaymentConfigPayload
	if err := render.Decode(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	dbGuide, err := config.GuideRepository.FindByID(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	if err := config.applyPaymentConfig(dbGuide, payload); err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	dbGuide, err = config.GuideRepository.Update(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, map[string]string{"message": "Payment configuration updated successfully"})
}

// GetMe godoc
// @Summary Get current user's guide profile
// @Description Get the guide profile linked to the authenticated user
// @ID get-guide-me
// @Success 200 {object} Guide
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/me [get]
func (config *Config) GetMe(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())
	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	dbGuide, err := config.GuideRepository.FindByUserID(loggedUser.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	render.JSON(w, r, dbGuide.ToModel())
}

// UpdateMe godoc
// @Summary Update current user's guide profile
// @Description Update the guide profile linked to the authenticated user
// @ID update-guide-me
// @Accept json
// @Produce json
// @Param guide body map[string]interface{} true "Guide data"
// @Success 200 {object} model.Guide
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/me [put]
func (config *Config) UpdateMe(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())
	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	dbGuide, err := config.GuideRepository.FindByUserID(loggedUser.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	config.updateGuideFromRequest(w, r, dbGuide)
}

// UpdateMyPaymentConfig godoc
// @Summary Update payment configuration for the current guide
// @Description Update payment configuration for the guide linked to the authenticated user
// @ID update-guide-me-payment-config
// @Accept json
// @Produce json
// @Param config body model.GuidePaymentConfigPayload true "Payment configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /guides/me/payment-config [put]
func (config *Config) UpdateMyPaymentConfig(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())
	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	dbGuide, err := config.GuideRepository.FindByUserID(loggedUser.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	var payload model.GuidePaymentConfigPayload
	if err := render.Decode(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	if err := config.applyPaymentConfig(dbGuide, payload); err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	dbGuide, err = config.GuideRepository.Update(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, map[string]string{"message": "Payment configuration updated successfully"})
}

// LinkUser godoc
// @Summary Link a guide to a user account
// @Description Link an existing guide profile to a user and grant the guide role
// @ID link-guide-user
// @Accept json
// @Produce json
// @Param id path int true "Guide ID"
// @Param payload body model.GuideLinkUserPayload true "User link data"
// @Success 200 {object} model.Guide
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 409 {string} string "conflict"
// @Failure 500 {string} string "internal server error"
// @Router /guides/{id}/link-user [post]
func (config *Config) LinkUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	var payload model.GuideLinkUserPayload
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	dbGuide, err := config.GuideRepository.FindByID(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbGuide == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	dbUser, err := config.UserRepository.FindByID(*payload.UserID, &dbmodel.UserFieldsToInclude{
		Roles:             true,
		Roles_Permissions: true,
	})
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbUser == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	if dbGuide.UserID != nil && *dbGuide.UserID != dbUser.ID {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("guide is already linked to another user")))
		return
	}

	existingGuide, err := config.GuideRepository.FindByUserID(dbUser.ID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if existingGuide != nil && existingGuide.ID != dbGuide.ID {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("user is already linked to another guide profile")))
		return
	}

	userID := dbUser.ID
	dbGuide.UserID = &userID

	dbGuide, err = config.GuideRepository.Update(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if err := authentication.GrantGuideRole(config.UserRepository, config.RoleRepository, dbUser); err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, dbGuide.ToModel())
}

func (config *Config) linkGuideToUserByEmail(dbGuide *dbmodel.Guide) error {
	if dbGuide.UserID != nil {
		return nil
	}

	dbUser, err := config.UserRepository.FindByEmail(dbGuide.Email, &dbmodel.UserFieldsToInclude{
		Roles:             true,
		Roles_Permissions: true,
	})
	if err != nil {
		return err
	}

	if dbUser == nil {
		return nil
	}

	existingGuide, err := config.GuideRepository.FindByUserID(dbUser.ID)
	if err != nil {
		return err
	}

	if existingGuide != nil && existingGuide.ID != dbGuide.ID {
		return fmt.Errorf("user is already linked to another guide profile")
	}

	userID := dbUser.ID
	dbGuide.UserID = &userID

	if _, err := config.GuideRepository.Update(dbGuide); err != nil {
		return err
	}

	return authentication.GrantGuideRole(config.UserRepository, config.RoleRepository, dbUser)
}

func (config *Config) updateGuideFromRequest(w http.ResponseWriter, r *http.Request, dbGuide *dbmodel.Guide) {
	var data map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	var avatarBase64 string
	if avatarInterface, exists := data["avatar_base64"]; exists {
		if avatarStr, ok := avatarInterface.(string); ok {
			avatarBase64 = avatarStr
		}
	}
	delete(data, "avatar_base64")

	helper.ApplyChanges(data, dbGuide)

	if avatarBase64 != "" {
		guideID := fmt.Sprintf("%d", dbGuide.ID)
		filename, err := helper.SaveBase64Image(avatarBase64, config.Constants.DataPath, "guide", guideID)
		if err != nil {
			render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("failed to save avatar: %w", err)))
			return
		}
		dbGuide.Avatar = &filename
	}

	dbGuide, err = config.GuideRepository.Update(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, dbGuide.ToModel())
}

func (config *Config) applyPaymentConfig(dbGuide *dbmodel.Guide, payload model.GuidePaymentConfigPayload) error {
	encryptionService := security.NewEncryptionService(config.Constants.JWTSecret)

	if payload.StripeAccountID != nil {
		dbGuide.StripeAccountID = payload.StripeAccountID
	}

	if payload.StripePublicKey != nil {
		dbGuide.StripePublicKey = payload.StripePublicKey
	}

	if payload.StripeSecretKey != nil && *payload.StripeSecretKey != "" {
		encryptedSecretKey, err := encryptionService.Encrypt(*payload.StripeSecretKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret key: %w", err)
		}
		dbGuide.StripeSecretKey = &encryptedSecretKey
	}

	if payload.StripeWebhookSecret != nil && *payload.StripeWebhookSecret != "" {
		encryptedWebhookSecret, err := encryptionService.Encrypt(*payload.StripeWebhookSecret)
		if err != nil {
			return fmt.Errorf("failed to encrypt webhook secret: %w", err)
		}
		dbGuide.StripeWebhookSecret = &encryptedWebhookSecret
	}

	if payload.PaymentEnabled != nil {
		dbGuide.PaymentEnabled = *payload.PaymentEnabled
	}

	return nil
}
