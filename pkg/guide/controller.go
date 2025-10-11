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
			// Log the error but don't fail the creation
			// You might want to add proper logging here
			fmt.Printf("Failed to save guide avatar: %v\n", err)
		} else {
			// Update the guide with the avatar filename
			dbGuide.Avatar = &filename
			dbGuide, err = config.GuideRepository.Update(dbGuide)
			if err != nil {
				// Log the error but don't fail the creation
				fmt.Printf("Failed to update guide with avatar: %v\n", err)
			}
		}
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

	helper.ApplyChanges(data, dbGuide)

	// Handle avatar upload if provided
	if avatarInterface, exists := data["avatar_base64"]; exists {
		if avatarStr, ok := avatarInterface.(string); ok && avatarStr != "" {
			guideID := fmt.Sprintf("%d", dbGuide.ID)
			filename, err := helper.SaveBase64Image(avatarStr, config.Constants.DataPath, "guide", guideID)
			if err != nil {
				// Log the error but don't fail the update
				fmt.Printf("Failed to save guide avatar: %v\n", err)
			} else {
				// Update the guide with the avatar filename
				dbGuide.Avatar = &filename
			}
		}
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

	// Initialize encryption service
	encryptionService := security.NewEncryptionService(config.Constants.JWTSecret)

	// Update payment configuration
	if payload.StripeAccountID != nil {
		dbGuide.StripeAccountID = payload.StripeAccountID
	}

	if payload.StripePublicKey != nil {
		dbGuide.StripePublicKey = payload.StripePublicKey
	}

	if payload.StripeSecretKey != nil && *payload.StripeSecretKey != "" {
		// Encrypt the secret key before storing
		encryptedSecretKey, err := encryptionService.Encrypt(*payload.StripeSecretKey)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(fmt.Errorf("failed to encrypt secret key: %w", err)))
			return
		}
		dbGuide.StripeSecretKey = &encryptedSecretKey
	}

	if payload.StripeWebhookSecret != nil && *payload.StripeWebhookSecret != "" {
		// Encrypt the webhook secret before storing
		encryptedWebhookSecret, err := encryptionService.Encrypt(*payload.StripeWebhookSecret)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(fmt.Errorf("failed to encrypt webhook secret: %w", err)))
			return
		}
		dbGuide.StripeWebhookSecret = &encryptedWebhookSecret
	}

	if payload.PaymentEnabled != nil {
		dbGuide.PaymentEnabled = *payload.PaymentEnabled
	}

	dbGuide, err = config.GuideRepository.Update(dbGuide)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, map[string]string{"message": "Payment configuration updated successfully"})
}
