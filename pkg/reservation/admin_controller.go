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
