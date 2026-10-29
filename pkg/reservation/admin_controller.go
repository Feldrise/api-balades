package reservation

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"github.com/go-chi/render"
)

const defaultAdminPerPage = 50

// @Summary Get all reservations (admin)
// @Description Get the reservations with their participants, with filtering and pagination (admin or guide)
// @Tags admin,reservations
// @Produce json
// @Param ramble_id query int false "Ramble ID"
// @Param email query string false "Exact contact email"
// @Param search query string false "Contact email or participant name"
// @Param ramble_title query string false "Part of the ramble title"
// @Param status query string false "Status: pending, confirmed, waiting_list or cancelled"
// @Param statuses query []string false "Several statuses"
// @Param user_id query int false "Booker account ID"
// @Param date_from query string false "Created since this date (YYYY-MM-DD)"
// @Param date_to query string false "Created until this date included (YYYY-MM-DD)"
// @Param sort_by query string false "created_at (default), contact_email or status"
// @Param sort_order query string false "asc or desc (default)"
// @Param page query int false "Page, from 1"
// @Param per_page query int false "Reservations per page, 50 by default"
// @Success 200 {object} model.AdminReservationListResponse
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /reservations/admin [get]
func (config *Config) AdminList(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	filter := &model.AdminReservationFilterPayload{}
	if err := parseAdminFilter(r, filter); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	page, perPage := 1, defaultAdminPerPage
	if filter.Page != nil {
		page = *filter.Page
	}
	if filter.PerPage != nil {
		perPage = *filter.PerPage
	}
	offset := (page - 1) * perPage

	dbFilter, err := toDBFilter(filter)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	rambleIDs, ok := config.resolveGuideRambleScope(w, r, user, filter.RambleID, "view:all-registrations", "view:registrations:self")
	if !ok {
		return
	}
	dbFilter.RambleIDs = rambleIDs
	dbFilter.Limit = &perPage
	dbFilter.Offset = &offset

	total, err := config.ReservationRepository.CountAll(dbFilter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	reservations, err := config.ReservationRepository.FindAll(dbFilter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	result := make([]*model.Reservation, len(reservations))
	for i := range reservations {
		result[i] = reservations[i].ToModel()
	}

	render.JSON(w, r, &model.AdminReservationListResponse{
		Reservations: result,
		Total:        total,
		Page:         page,
		PerPage:      perPage,
		TotalPages:   int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func parseAdminFilter(r *http.Request, filter *model.AdminReservationFilterPayload) error {
	query := r.URL.Query()

	if value := query.Get("ramble_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid ramble_id")
		}
		rambleID := uint(id)
		filter.RambleID = &rambleID
	}

	if value := query.Get("email"); value != "" {
		filter.Email = &value
	}

	if value := query.Get("search"); value != "" {
		filter.Search = &value
	}

	if value := query.Get("ramble_title"); value != "" {
		filter.RambleTitle = &value
	}

	if value := query.Get("status"); value != "" {
		filter.Status = &value
	}

	filter.Statuses = query["statuses"]

	if value := query.Get("user_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid user_id")
		}
		userID := uint(id)
		filter.UserID = &userID
	}

	if value := query.Get("date_from"); value != "" {
		filter.DateFrom = &value
	}

	if value := query.Get("date_to"); value != "" {
		filter.DateTo = &value
	}

	if value := query.Get("sort_by"); value != "" {
		filter.SortBy = &value
	}

	if value := query.Get("sort_order"); value != "" {
		filter.SortOrder = &value
	}

	if value := query.Get("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid page")
		}
		filter.Page = &page
	}

	if value := query.Get("per_page"); value != "" {
		perPage, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid per_page")
		}
		filter.PerPage = &perPage
	}

	return filter.Bind(r)
}

// toDBFilter converts the filter of the request, without its ramble scope and pagination.
func toDBFilter(filter *model.AdminReservationFilterPayload) (*dbmodel.ReservationFilter, error) {
	dbFilter := &dbmodel.ReservationFilter{
		UserID:      filter.UserID,
		Email:       filter.Email,
		Search:      filter.Search,
		RambleTitle: filter.RambleTitle,
		Status:      filter.Status,
		Statuses:    filter.Statuses,
	}

	if filter.DateFrom != nil {
		from, err := time.Parse("2006-01-02", *filter.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format. Use YYYY-MM-DD")
		}
		dbFilter.CreatedFrom = &from
	}

	if filter.DateTo != nil {
		to, err := time.Parse("2006-01-02", *filter.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format. Use YYYY-MM-DD")
		}
		// Up to the end of the day.
		to = to.Add(24*time.Hour - time.Nanosecond)
		dbFilter.CreatedTo = &to
	}

	if filter.SortBy != nil {
		dbFilter.SortBy = *filter.SortBy
	}
	dbFilter.SortAsc = filter.SortOrder != nil && *filter.SortOrder == "asc"

	return dbFilter, nil
}

// @Summary Update reservation status (admin)
// @Description Set the status of a reservation, as a whole. A reservation put back in pending, confirmed or waiting list from a status that holds no place must fit entirely in the places left. The places released go to the waiting list.
// @Tags admin,reservations
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param status body model.AdminReservationStatusUpdatePayload true "Status update data"
// @Success 200 {object} model.Reservation
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "not_enough_places"
// @Router /reservations/admin/{id}/status [put]
func (config *Config) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	data := &model.AdminReservationStatusUpdatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	reservation, ok := config.reservationToAdminister(w, r, user, id, "update:registration-status", "update:registration-status:self")
	if !ok {
		return
	}

	result, err := config.setStatus(reservation, data.Status, data.Reason)
	if err != nil {
		render.Render(w, r, renderError(err))
		return
	}
	config.notifyPromoted(result.promoted)

	updated, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if result.changed && data.SendEmail != nil && *data.SendEmail {
		config.sendReservationEmail(updated, changeStatusUpdated)
	}

	render.JSON(w, r, updated.ToModel())
}

// reservationToAdminister returns the reservation when the user holds the
// blanket permission, or the self permission and guides its ramble, and renders
// the error otherwise.
func (config *Config) reservationToAdminister(w http.ResponseWriter, r *http.Request, user *dbmodel.User, id uint, blanketPerm, selfPerm string) (*dbmodel.Reservation, bool) {
	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return nil, false
	}
	if reservation == nil {
		render.Render(w, r, errors.ErrNotFound())
		return nil, false
	}

	allowed, err := config.canActOnRamble(user, reservation.RambleID, blanketPerm, selfPerm)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return nil, false
	}
	if !allowed {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return nil, false
	}

	return reservation, true
}

// @Summary Delete a reservation (admin)
// @Description Delete a reservation with its participants. The places it held go to the waiting list.
// @Tags admin,reservations
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 204
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /reservations/admin/{id} [delete]
func (config *Config) AdminDelete(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	reservation, ok := config.reservationToAdminister(w, r, user, id, "manage:registration", "manage:registration:self")
	if !ok {
		return
	}

	result, err := config.deleteReservation(reservation)
	if err != nil {
		render.Render(w, r, renderError(err))
		return
	}
	config.notifyPromoted(result.promoted)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Bulk action on reservations (admin)
// @Description Confirm, cancel, move to waiting list or delete several reservations. Each one follows the rules of the single actions, and fails separately.
// @Tags admin,reservations
// @Accept json
// @Produce json
// @Param bulk body model.BulkReservationActionPayload true "Bulk action data"
// @Success 200 {object} model.BulkReservationActionResult
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /reservations/admin/bulk-action [post]
func (config *Config) AdminBulkAction(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	if !user.HasPermission("bulk:registration-actions") && !user.HasPermission("bulk:registration-actions:self") {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	data := &model.BulkReservationActionPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	result := &model.BulkReservationActionResult{
		Errors:  []model.BulkReservationActionError{},
		Updated: []*model.Reservation{},
	}
	sendEmail := data.SendEmail != nil && *data.SendEmail

	for _, id := range data.ReservationIDs {
		updated, err := config.bulkAction(user, id, data, sendEmail)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.BulkReservationActionError{ReservationID: id, Error: err.Error()})
			continue
		}

		result.SuccessCount++
		if updated != nil {
			result.Updated = append(result.Updated, updated)
		}
	}

	render.JSON(w, r, result)
}

// bulkAction applies the action of a bulk request to one reservation, and
// returns the reservation as saved, nil when it is deleted.
func (config *Config) bulkAction(user *dbmodel.User, id uint, data *model.BulkReservationActionPayload, sendEmail bool) (*model.Reservation, error) {
	reservation, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find reservation: %w", err)
	}
	if reservation == nil {
		return nil, fmt.Errorf("reservation not found")
	}

	canAct, err := config.canActOnRamble(user, reservation.RambleID, "bulk:registration-actions", "bulk:registration-actions:self")
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !canAct {
		return nil, fmt.Errorf("insufficient permissions for this reservation")
	}

	if data.Action == "delete" {
		canDelete, err := config.canActOnRamble(user, reservation.RambleID, "manage:registration", "manage:registration:self")
		if err != nil {
			return nil, fmt.Errorf("permission check failed: %w", err)
		}
		if !canDelete {
			return nil, fmt.Errorf("insufficient permissions for delete action")
		}

		change, err := config.deleteReservation(reservation)
		if err != nil {
			return nil, err
		}
		config.notifyPromoted(change.promoted)
		return nil, nil
	}

	change, err := config.setStatus(reservation, bulkActionStatuses[data.Action], data.Reason)
	if err != nil {
		return nil, err
	}
	config.notifyPromoted(change.promoted)

	updated, err := config.ReservationRepository.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload reservation: %w", err)
	}
	if change.changed && sendEmail {
		config.sendReservationEmail(updated, changeStatusUpdated)
	}

	return updated.ToModel(), nil
}
