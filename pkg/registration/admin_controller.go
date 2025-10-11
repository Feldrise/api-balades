package registration

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Admin-specific registration management endpoints

// @Summary Get all registrations (admin)
// @Description Get all registrations with advanced filtering and pagination (admin only)
// @Tags admin,registrations
// @Accept json
// @Produce json
// @Param filter query model.AdminRegistrationFilterPayload false "Filter parameters"
// @Success 200 {object} model.AdminRegistrationListResponse
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/admin [get]
func (config *Config) AdminGetAllRegistrations(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	if !user.HasPermission("view:all-registrations") {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	// Parse query parameters
	filter := &model.AdminRegistrationFilterPayload{}
	if err := parseAdminFilterFromQuery(r, filter); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Convert to database filter
	dbFilter, err := convertToDBFilter(filter)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Get total count
	totalCount, err := config.RambleRegistrationRepository.CountAll(dbFilter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Get registrations
	registrations, err := config.RambleRegistrationRepository.FindAll(dbFilter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Convert to models
	result := make([]*model.RambleRegistration, len(registrations))
	for i, reg := range registrations {
		result[i] = reg.ToModel()
	}

	// Calculate pagination info
	page := 1
	perPage := 50
	if filter.Page != nil {
		page = *filter.Page
	}
	if filter.PerPage != nil {
		perPage = *filter.PerPage
	}

	totalPages := int((totalCount + int64(perPage) - 1) / int64(perPage))

	response := &model.AdminRegistrationListResponse{
		Registrations: result,
		Total:         totalCount,
		Page:          page,
		PerPage:       perPage,
		TotalPages:    totalPages,
	}

	render.JSON(w, r, response)
}

// @Summary Update registration details (admin)
// @Description Update registration details like name, email, phone (admin only)
// @Tags admin,registrations
// @Accept json
// @Produce json
// @Param id path int true "Registration ID"
// @Param update body model.AdminRegistrationUpdatePayload true "Update data"
// @Success 200 {object} model.RambleRegistration
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/admin/{id} [put]
func (config *Config) AdminUpdateRegistration(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	if !user.HasPermission("update:registration-details") {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid ID")))
		return
	}

	data := &model.AdminRegistrationUpdatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	registration, err := config.RambleRegistrationRepository.FindByID(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if registration == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	// Update fields if provided
	if data.FirstName != nil {
		registration.FirstName = *data.FirstName
	}
	if data.LastName != nil {
		registration.LastName = *data.LastName
	}
	if data.Email != nil {
		registration.Email = *data.Email
	}
	if data.Phone != nil {
		registration.Phone = data.Phone
	}

	// Handle status updates specially to maintain business logic
	if data.Status != nil && *data.Status != registration.Status {
		err := config.updateRegistrationStatus(registration, *data.Status, data.Notes, true)
		if err != nil {
			render.Render(w, r, errors.ErrInvalidRequest(err))
			return
		}
	}

	updatedRegistration, err := config.RambleRegistrationRepository.Update(registration)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.JSON(w, r, updatedRegistration.ToModel())
}

// @Summary Update registration status (admin)
// @Description Update registration status with optional reason (admin only)
// @Tags admin,registrations
// @Accept json
// @Produce json
// @Param id path int true "Registration ID"
// @Param status body model.AdminRegistrationStatusUpdatePayload true "Status update data"
// @Success 200 {object} model.RambleRegistration
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/admin/{id}/status [put]
func (config *Config) AdminUpdateRegistrationStatus(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	if !user.HasPermission("update:registration-status") {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid ID")))
		return
	}

	data := &model.AdminRegistrationStatusUpdatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	registration, err := config.RambleRegistrationRepository.FindByID(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if registration == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	sendEmail := data.SendEmail != nil && *data.SendEmail
	err = config.updateRegistrationStatus(registration, data.Status, data.Reason, sendEmail)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	updatedRegistration, err := config.RambleRegistrationRepository.Update(registration)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Handle waiting list promotion if registration was cancelled
	if data.Status == "cancelled" {
		go config.promoteFromWaitingList(registration.RambleID)
	}

	render.JSON(w, r, updatedRegistration.ToModel())
}

// @Summary Delete registration (admin)
// @Description Delete a registration (admin only)
// @Tags admin,registrations
// @Produce json
// @Param id path int true "Registration ID"
// @Success 204
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/admin/{id} [delete]
func (config *Config) AdminDeleteRegistration(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	if !user.HasPermission("manage:registration") {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid ID")))
		return
	}

	registration, err := config.RambleRegistrationRepository.FindByID(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if registration == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	rambleID := registration.RambleID
	err = config.RambleRegistrationRepository.Delete(uint(id))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Try to promote someone from waiting list if this was a confirmed registration
	if registration.Status == "confirmed" {
		go config.promoteFromWaitingList(rambleID)
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Bulk action on registrations (admin)
// @Description Perform bulk actions on multiple registrations (admin only)
// @Tags admin,registrations
// @Accept json
// @Produce json
// @Param bulk body model.BulkRegistrationActionPayload true "Bulk action data"
// @Success 200 {object} model.BulkActionResult
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/admin/bulk-action [post]
func (config *Config) AdminBulkRegistrationAction(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	if !user.HasPermission("bulk:registration-actions") {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	data := &model.BulkRegistrationActionPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	result := &model.BulkActionResult{
		SuccessCount: 0,
		FailureCount: 0,
		Errors:       []model.BulkActionError{},
		Updated:      []*model.RambleRegistration{},
	}

	sendEmail := data.SendEmail != nil && *data.SendEmail

	for _, registrationID := range data.RegistrationIDs {
		registration, err := config.RambleRegistrationRepository.FindByID(registrationID)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.BulkActionError{
				RegistrationID: registrationID,
				Error:          fmt.Sprintf("Failed to find registration: %v", err),
			})
			continue
		}

		if registration == nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.BulkActionError{
				RegistrationID: registrationID,
				Error:          "Registration not found",
			})
			continue
		}

		var actionErr error
		switch data.Action {
		case "confirm":
			actionErr = config.updateRegistrationStatus(registration, "confirmed", data.Reason, sendEmail)
		case "cancel":
			actionErr = config.updateRegistrationStatus(registration, "cancelled", data.Reason, sendEmail)
		case "move_to_waiting":
			actionErr = config.updateRegistrationStatus(registration, "waiting_list", data.Reason, sendEmail)
		case "delete":
			if !user.HasPermission("manage:registration") {
				actionErr = fmt.Errorf("insufficient permissions for delete action")
			} else {
				rambleID := registration.RambleID
				actionErr = config.RambleRegistrationRepository.Delete(registrationID)
				if actionErr == nil && registration.Status == "confirmed" {
					go config.promoteFromWaitingList(rambleID)
				}
			}
		}

		if actionErr != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.BulkActionError{
				RegistrationID: registrationID,
				Error:          actionErr.Error(),
			})
			continue
		}

		if data.Action != "delete" {
			updatedRegistration, err := config.RambleRegistrationRepository.Update(registration)
			if err != nil {
				result.FailureCount++
				result.Errors = append(result.Errors, model.BulkActionError{
					RegistrationID: registrationID,
					Error:          fmt.Sprintf("Failed to update registration: %v", err),
				})
				continue
			}
			result.Updated = append(result.Updated, updatedRegistration.ToModel())

			// Handle waiting list promotion for cancellations
			if data.Action == "cancel" {
				go config.promoteFromWaitingList(registration.RambleID)
			}
		}

		result.SuccessCount++
	}

	render.JSON(w, r, result)
}

// Helper functions

func parseAdminFilterFromQuery(r *http.Request, filter *model.AdminRegistrationFilterPayload) error {
	query := r.URL.Query()

	if rambleID := query.Get("ramble_id"); rambleID != "" {
		id, err := strconv.Atoi(rambleID)
		if err != nil {
			return fmt.Errorf("invalid ramble_id")
		}
		uid := uint(id)
		filter.RambleID = &uid
	}

	if userID := query.Get("user_id"); userID != "" {
		id, err := strconv.Atoi(userID)
		if err != nil {
			return fmt.Errorf("invalid user_id")
		}
		uid := uint(id)
		filter.UserID = &uid
	}

	if email := query.Get("email"); email != "" {
		filter.Email = &email
	}

	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}

	if statuses := query["statuses"]; len(statuses) > 0 {
		filter.Statuses = statuses
	}

	if dateFrom := query.Get("date_from"); dateFrom != "" {
		filter.DateFrom = &dateFrom
	}

	if dateTo := query.Get("date_to"); dateTo != "" {
		filter.DateTo = &dateTo
	}

	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}

	if rambleTitle := query.Get("ramble_title"); rambleTitle != "" {
		filter.RambleTitle = &rambleTitle
	}

	if page := query.Get("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			return fmt.Errorf("invalid page")
		}
		filter.Page = &p
	}

	if perPage := query.Get("per_page"); perPage != "" {
		pp, err := strconv.Atoi(perPage)
		if err != nil {
			return fmt.Errorf("invalid per_page")
		}
		filter.PerPage = &pp
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		filter.SortBy = &sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		filter.SortOrder = &sortOrder
	}

	return nil
}

func convertToDBFilter(filter *model.AdminRegistrationFilterPayload) (*dbmodel.RambleRegistrationFilter, error) {
	dbFilter := &dbmodel.RambleRegistrationFilter{
		RambleID:    filter.RambleID,
		UserID:      filter.UserID,
		Email:       filter.Email,
		Status:      filter.Status,
		Statuses:    filter.Statuses,
		Search:      filter.Search,
		RambleTitle: filter.RambleTitle,
		SortBy:      filter.SortBy,
		SortOrder:   filter.SortOrder,
	}

	// Parse dates
	if filter.DateFrom != nil && *filter.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", *filter.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format. Use YYYY-MM-DD")
		}
		dbFilter.DateFrom = &dateFrom
	}

	if filter.DateTo != nil && *filter.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", *filter.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format. Use YYYY-MM-DD")
		}
		// Set to end of day
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		dbFilter.DateTo = &dateTo
	}

	// Handle pagination
	if filter.Page != nil && filter.PerPage != nil {
		offset := (*filter.Page - 1) * *filter.PerPage
		dbFilter.Offset = &offset
		dbFilter.Limit = filter.PerPage
	} else if filter.PerPage != nil {
		dbFilter.Limit = filter.PerPage
	}

	return dbFilter, nil
}

func (config *Config) updateRegistrationStatus(registration *dbmodel.RambleRegistration, newStatus string, reason *string, sendEmail bool) error {
	if registration.Status == newStatus {
		return nil // No change needed
	}

	now := time.Now()

	switch newStatus {
	case "confirmed":
		registration.Status = "confirmed"
		registration.ConfirmationDate = &now
		registration.CancellationDate = nil
		registration.CancellationReason = nil
	case "cancelled":
		registration.Status = "cancelled"
		registration.CancellationDate = &now
		registration.CancellationReason = reason
		registration.ConfirmationDate = nil
	case "pending":
		registration.Status = "pending"
		registration.ConfirmationDate = nil
		registration.CancellationDate = nil
		registration.CancellationReason = nil
	case "waiting_list":
		registration.Status = "waiting_list"
		registration.ConfirmationDate = nil
		registration.CancellationDate = nil
		registration.CancellationReason = nil
	default:
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	// Send email notification if sendEmail is true and status is confirmed
	if sendEmail && newStatus == "confirmed" {
		go config.sendAdminConfirmationEmail(registration)
	}

	return nil
}

// sendAdminConfirmationEmail sends a confirmation email when admin confirms a registration
func (config *Config) sendAdminConfirmationEmail(registration *dbmodel.RambleRegistration) {
	// Load the ramble with all necessary data
	ramble, err := config.RambleRepository.FindByID(registration.RambleID)
	if err != nil || ramble == nil {
		fmt.Printf("Failed to load ramble for admin confirmation email: %v\n", err)
		return
	}

	// Check if this is a group registration
	if registration.GroupID != nil {
		// Load the group
		group, err := config.RambleRegistrationGroupRepository.FindByID(*registration.GroupID)
		if err != nil || group == nil {
			fmt.Printf("Failed to load group for admin confirmation email: %v\n", err)
			return
		}
		config.sendGroupRegistrationConfirmationEmail(registration, ramble, group)
	} else {
		config.sendRegistrationConfirmationEmail(registration, ramble)
	}
}
