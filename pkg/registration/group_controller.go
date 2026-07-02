package registration

import (
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

// GetGroup retrieves a specific group by ID
func (config *Config) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	group, err := config.RambleRegistrationGroupRepository.FindByID(uint(groupID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if group == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	user := authentication.ForContext(r.Context())
	if user != nil {
		canView, err := config.canViewRambleRegistrations(user, group.RambleID)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}

		if !canView && !isGroupMember(user, group) {
			render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
			return
		}
	}

	render.JSON(w, r, group.ToModel())
}

// GetGroupsByRamble retrieves all groups for a specific ramble
func (config *Config) GetGroupsByRamble(w http.ResponseWriter, r *http.Request) {
	user := config.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	rambleIDStr := chi.URLParam(r, "rambleId")
	rambleID, err := strconv.ParseUint(rambleIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	if !config.requireRambleRegistrationAccess(w, r, user, uint(rambleID), "view:all-registrations", "view:registrations:self") {
		return
	}

	filter := &dbmodel.RambleRegistrationGroupFilter{
		RambleID: func() *uint { id := uint(rambleID); return &id }(),
	}

	groups, err := config.RambleRegistrationGroupRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	response := make([]*model.RambleRegistrationGroup, len(groups))
	for i, group := range groups {
		response[i] = group.ToModel()
	}

	render.JSON(w, r, response)
}

// ConfirmGroup confirms all registrations in a group
func (config *Config) ConfirmGroup(w http.ResponseWriter, r *http.Request) {
	user := config.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	var payload model.RambleRegistrationConfirmPayload
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	group, err := config.RambleRegistrationGroupRepository.FindByID(uint(groupID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if group == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	if !config.requireRambleRegistrationAccess(w, r, user, group.RambleID, "update:registration-status", "update:registration-status:self") {
		return
	}

	// Update group status
	if payload.Confirmed {
		if group.Status == "pending" {
			group.Status = "confirmed"
			if group.ConfirmationDate == nil {
				now := time.Now()
				group.ConfirmationDate = &now
			}
		}
	} else {
		group.Status = "cancelled"
		now := time.Now()
		group.CancellationDate = &now
		reason := "Declined by group"
		group.CancellationReason = &reason
	}

	_, err = config.RambleRegistrationGroupRepository.Update(group)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Update all individual registrations in the group
	filter := &dbmodel.RambleRegistrationFilter{
		GroupID: &group.ID,
	}
	registrations, err := config.RambleRegistrationRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	for _, registration := range registrations {
		registration.Status = group.Status
		registration.ConfirmationDate = group.ConfirmationDate
		registration.CancellationDate = group.CancellationDate
		registration.CancellationReason = group.CancellationReason

		_, err = config.RambleRegistrationRepository.Update(&registration)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	// If we're declining a group that was pending, try to move someone from waiting list
	if !payload.Confirmed && group.Status == "cancelled" {
		go config.promoteFromWaitingList(group.RambleID)
	}

	render.JSON(w, r, group.ToModel())
}

// CancelGroup cancels all registrations in a group
func (config *Config) CancelGroup(w http.ResponseWriter, r *http.Request) {
	user := config.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	var payload model.RambleRegistrationCancelPayload
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	group, err := config.RambleRegistrationGroupRepository.FindByID(uint(groupID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if group == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	if !config.requireRambleRegistrationAccess(w, r, user, group.RambleID, "update:registration-status", "update:registration-status:self") {
		return
	}

	// Update group status to cancelled
	group.Status = "cancelled"
	now := time.Now()
	group.CancellationDate = &now
	if payload.Reason != nil {
		group.CancellationReason = payload.Reason
	} else {
		reason := "Cancelled by group"
		group.CancellationReason = &reason
	}

	_, err = config.RambleRegistrationGroupRepository.Update(group)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Update all individual registrations in the group
	filter := &dbmodel.RambleRegistrationFilter{
		GroupID: &group.ID,
	}
	registrations, err := config.RambleRegistrationRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	for _, registration := range registrations {
		registration.Status = "cancelled"
		registration.CancellationDate = group.CancellationDate
		registration.CancellationReason = group.CancellationReason

		_, err = config.RambleRegistrationRepository.Update(&registration)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	// Process waiting list to fill open spots
	go config.promoteFromWaitingList(group.RambleID)

	render.JSON(w, r, group.ToModel())
}

// AdminGetAllGroups retrieves all groups with optional filtering
func (config *Config) AdminGetAllGroups(w http.ResponseWriter, r *http.Request) {
	user := config.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	var filter dbmodel.RambleRegistrationGroupFilter

	// Parse query parameters
	if rambleIDStr := r.URL.Query().Get("ramble_id"); rambleIDStr != "" {
		if rambleID, err := strconv.ParseUint(rambleIDStr, 10, 32); err == nil {
			id := uint(rambleID)
			filter.RambleID = &id
		}
	}

	if !user.HasPermission("view:all-registrations") {
		if filter.RambleID == nil {
			render.Render(w, r, errors.ErrForbidden("ramble_id is required"))
			return
		}

		if !config.requireRambleRegistrationAccess(w, r, user, *filter.RambleID, "view:all-registrations", "view:registrations:self") {
			return
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	if primaryEmail := r.URL.Query().Get("primary_email"); primaryEmail != "" {
		filter.PrimaryEmail = &primaryEmail
	}

	groups, err := config.RambleRegistrationGroupRepository.FindAll(&filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	response := make([]*model.RambleRegistrationGroup, len(groups))
	for i, group := range groups {
		response[i] = group.ToModel()
	}

	render.JSON(w, r, response)
}

// AdminUpdateGroupStatus updates the status of a group (admin only)
func (config *Config) AdminUpdateGroupStatus(w http.ResponseWriter, r *http.Request) {
	user := config.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	var payload struct {
		Status string  `json:"status"`
		Reason *string `json:"reason,omitempty"`
	}

	if err := render.DecodeJSON(r.Body, &payload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	group, err := config.RambleRegistrationGroupRepository.FindByID(uint(groupID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if group == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	if !config.requireRambleRegistrationAccess(w, r, user, group.RambleID, "update:registration-status", "update:registration-status:self") {
		return
	}

	// Update group status
	oldStatus := group.Status
	group.Status = payload.Status

	now := time.Now()
	switch payload.Status {
	case "confirmed":
		if group.ConfirmationDate == nil {
			group.ConfirmationDate = &now
		}
	case "cancelled":
		if group.CancellationDate == nil {
			group.CancellationDate = &now
		}
		if payload.Reason != nil {
			group.CancellationReason = payload.Reason
		}
	}

	_, err = config.RambleRegistrationGroupRepository.Update(group)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Update all individual registrations in the group
	filter := &dbmodel.RambleRegistrationFilter{
		GroupID: &group.ID,
	}
	registrations, err := config.RambleRegistrationRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	for _, registration := range registrations {
		registration.Status = group.Status
		registration.ConfirmationDate = group.ConfirmationDate
		registration.CancellationDate = group.CancellationDate
		registration.CancellationReason = group.CancellationReason

		_, err = config.RambleRegistrationRepository.Update(&registration)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	// If status changed from confirmed to something else, or cancelled, process waiting list
	if (oldStatus == "confirmed" && payload.Status != "confirmed") || payload.Status == "cancelled" {
		go config.promoteFromWaitingList(group.RambleID)
	}

	render.JSON(w, r, group.ToModel())
}

// AdminDeleteGroup deletes a group and all its registrations (admin only)
func (config *Config) AdminDeleteGroup(w http.ResponseWriter, r *http.Request) {
	user := config.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	group, err := config.RambleRegistrationGroupRepository.FindByID(uint(groupID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	if group == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	if !config.requireRambleRegistrationAccess(w, r, user, group.RambleID, "manage:registration", "manage:registration:self") {
		return
	}

	// Delete all individual registrations first
	filter := &dbmodel.RambleRegistrationFilter{
		GroupID: &group.ID,
	}
	registrations, err := config.RambleRegistrationRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	for _, registration := range registrations {
		err = config.RambleRegistrationRepository.Delete(registration.ID)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
	}

	// Delete the group
	err = config.RambleRegistrationGroupRepository.Delete(uint(groupID))
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Process waiting list to fill open spots
	go config.promoteFromWaitingList(group.RambleID)

	w.WriteHeader(http.StatusNoContent)
}
