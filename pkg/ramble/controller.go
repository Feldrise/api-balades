package ramble

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
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

// Get godoc
// @Summary Get Ramble
// @Description Get a ramble by ID
// @ID get-ramble
// @Param id path int true "Ramble ID"
// @Success 200 {object} Ramble
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /rambles/{id} [get]
func (config *Config) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Convert id to uint
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	dbRamble, err := config.RambleRepository.FindByID(uint(idUint))

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbRamble == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	render.JSON(w, r, dbRamble.ToModel())
}

// GetAll godoc
// @Summary Get all Rambles
// @Description Get all rambles with optional status filter
// @ID get-all-rambles
// @Param status query string false "Filter by status (e.g., 'active', 'archived')"
// @Success 200 {array} Ramble
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /rambles [get]
func (config *Config) GetAll(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	var filter *dbmodel.RambleFilter

	if status != "" {
		filter = &dbmodel.RambleFilter{
			Status: &status,
		}
	}

	dbRambles, err := config.RambleRepository.FindAll(filter)

	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	rambles := make([]model.Ramble, len(dbRambles))
	for i, dbRamble := range dbRambles {
		rambles[i] = dbRamble.ToModel()
	}

	render.JSON(w, r, rambles)

}

// Create godoc
// @Summary Create Ramble
// @Description Create a new ramble
// @ID create-ramble
// @Accept json
// @Produce json
// @Param ramble body RambleCreatePayload true "Ramble data"
// @Success 201 {object} Ramble
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /rambles [post]
func (config *Config) Create(w http.ResponseWriter, r *http.Request) {
	loggedUser := authentication.ForContext(r.Context())

	if loggedUser == nil {
		render.Render(w, r, errors.ErrUnauthorized("You must be logged in to access this resource"))
		return
	}

	var payload model.RambleCreatePayload

	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	dbRamble := &dbmodel.Ramble{
		Title:             *payload.Title,
		Description:       payload.Description,
		Type:              *payload.Type,
		Date:              payload.Date,
		Location:          payload.Location,
		MeetingPoint:      payload.MeetingPoint,
		MaxParticipants:   payload.MaxParticipants,
		Difficulty:        *payload.Difficulty,
		EstimatedDuration: payload.EstimatedDuration,
		EquipmentNeeded:   payload.EquipmentNeeded,
		Prerequisites:     payload.Prerequisites,
		Prices:            make([]dbmodel.RamblePrice, len(payload.Prices)),
	}

	for i, price := range payload.Prices {
		dbRamble.Prices[i] = dbmodel.RamblePrice{
			Label:  price.Label,
			Amount: price.Amount,
		}
	}

	// Handle guide associations if provided
	if len(payload.GuideIDs) > 0 {
		guides := make([]dbmodel.Guide, len(payload.GuideIDs))
		for i, guideID := range payload.GuideIDs {
			guides[i] = dbmodel.Guide{Model: gorm.Model{ID: guideID}}
		}
		dbRamble.Guides = guides
	}

	// Create the ramble first to get the ID
	dbRamble, err := config.RambleRepository.Create(dbRamble)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Handle cover image upload if provided
	if payload.CoverImageBase64 != nil && *payload.CoverImageBase64 != "" {
		rambleID := fmt.Sprintf("%d", dbRamble.ID)
		filename, err := helper.SaveBase64Image(*payload.CoverImageBase64, config.Constants.DataPath, "ramble", rambleID)
		if err != nil {
			// Log the error but don't fail the creation
			fmt.Printf("Failed to save ramble cover image: %v\n", err)
		} else {
			// Update the ramble with the cover image filename
			dbRamble.CoverImageURL = &filename
		}
	}

	// Handle additional document upload if provided
	if payload.AdditionalDocumentBase64 != nil && *payload.AdditionalDocumentBase64 != "" {
		rambleID := fmt.Sprintf("%d", dbRamble.ID)
		filename, err := helper.SaveBase64Document(*payload.AdditionalDocumentBase64, config.Constants.DataPath, "ramble", rambleID, "document_")
		if err != nil {
			// Log the error but don't fail the creation
			fmt.Printf("Failed to save ramble additional document: %v\n", err)
		} else {
			// Update the ramble with the additional document filename
			dbRamble.AdditionalDocumentsURL = &filename
		}
	}

	// Update the ramble if any files were uploaded
	if (payload.CoverImageBase64 != nil && *payload.CoverImageBase64 != "") ||
		(payload.AdditionalDocumentBase64 != nil && *payload.AdditionalDocumentBase64 != "") {
		dbRamble, err = config.RambleRepository.Update(dbRamble)
		if err != nil {
			// Log the error but don't fail the creation
			fmt.Printf("Failed to update ramble with file URLs: %v\n", err)
		}
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dbRamble.ToModel())
}

// Update godoc
// @Summary Update an existing ramble
// @Description Update an existing ramble with the provided data
// @ID update-ramble
// @Accept json
// @Produce json
// @Param id path int true "Ramble ID"
// @Param ramble body map[string]interface{} true "Ramble data"
// @Success 200 {object} model.Ramble
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /rambles/{id} [put]
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

	dbRamble, err := config.RambleRepository.FindByID(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbRamble == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	helper.ApplyChanges(data, dbRamble)

	// Handle guide IDs if provided in the update
	if guideIDsInterface, exists := data["guide_ids"]; exists {
		if guideIDsSlice, ok := guideIDsInterface.([]interface{}); ok {
			guides := make([]dbmodel.Guide, len(guideIDsSlice))
			for i, guideIDInterface := range guideIDsSlice {
				if guideIDFloat, ok := guideIDInterface.(float64); ok {
					guideID := uint(guideIDFloat)
					guides[i] = dbmodel.Guide{Model: gorm.Model{ID: guideID}}
				}
			}
			dbRamble.Guides = guides
		}
	}

	dbRamble, err = config.RambleRepository.Update(dbRamble)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, dbRamble.ToModel())
}

// Delete godoc
// @Summary Delete a ramble
// @Description Delete a ramble by its ID
// @ID delete-ramble
// @Param id path int true "Ramble ID"
// @Success 204
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /rambles/{id} [delete]
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

	dbRamble, err := config.RambleRepository.FindByID(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if dbRamble == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	err = config.RambleRepository.Delete(uint(idUint))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.NoContent(w, r)
}
