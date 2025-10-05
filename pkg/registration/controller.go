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
	"feldrise.com/balade/pkg/notifications/email"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// @Summary Create a new ramble registration
// @Description Register user(s) for a ramble. Supports both single and group registrations. If ramble is full, users are added to waiting list.
// @Tags registrations
// @Accept json
// @Produce json
// @Param registration body model.RambleRegistrationCreatePayload true "Registration data"
// @Success 201 {object} model.RambleRegistration
// @Failure 400 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations [post]
func (config *Config) Create(w http.ResponseWriter, r *http.Request) {
	data := &model.RambleRegistrationCreatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Check if ramble exists
	ramble, err := config.RambleRepository.FindByID(data.RambleID)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}
	if ramble == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	// Check if ramble is cancelled
	if ramble.IsCancelled {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("cannot register for a cancelled ramble")))
		return
	}

	// If it's a group registration (multiple participants)
	if len(data.Participants) > 1 {
		registration, err := config.createGroupRegistration(data, ramble)
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, registration.ToModel())
		return
	}

	// Single registration (legacy support)
	registration, err := config.createSingleRegistration(data, ramble)
	if err != nil {
		if err.Error() == "already registered for this ramble" {
			render.Render(w, r, errors.ErrInvalidRequest(err))
			return
		}
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, registration.ToModel())
}

// @Summary Get a registration by ID
// @Description Get details of a specific registration
// @Tags registrations
// @Produce json
// @Param id path int true "Registration ID"
// @Success 200 {object} model.RambleRegistration
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/{id} [get]
func (config *Config) Get(w http.ResponseWriter, r *http.Request) {
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

	render.JSON(w, r, registration.ToModel())
}

// @Summary Get user's registrations
// @Description Get all registrations for the authenticated user
// @Tags registrations
// @Produce json
// @Success 200 {array} model.RambleRegistration
// @Failure 401 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations [get]
func (config *Config) GetUserRegistrations(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	filter := &dbmodel.RambleRegistrationFilter{
		UserID: &user.ID,
	}

	registrations, err := config.RambleRegistrationRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	result := make([]*model.RambleRegistration, len(registrations))
	for i, reg := range registrations {
		result[i] = reg.ToModel()
	}

	render.JSON(w, r, result)
}

// @Summary Confirm a registration
// @Description Confirm or deny participation in a ramble
// @Tags registrations
// @Accept json
// @Produce json
// @Param id path int true "Registration ID"
// @Param confirmation body model.RambleRegistrationConfirmPayload true "Confirmation data"
// @Success 200 {object} model.RambleRegistration
// @Failure 400 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/{id}/confirm [put]
func (config *Config) Confirm(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid ID")))
		return
	}

	data := &model.RambleRegistrationConfirmPayload{}
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

	if registration.Status != "pending" {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("registration cannot be confirmed in current status")))
		return
	}

	now := time.Now()
	if data.Confirmed {
		registration.Status = "confirmed"
		registration.ConfirmationDate = &now
	} else {
		registration.Status = "cancelled"
		registration.CancellationDate = &now
		registration.CancellationReason = func() *string { s := "User declined confirmation"; return &s }()
	}

	updatedRegistration, err := config.RambleRegistrationRepository.Update(registration)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// If cancelled, try to promote someone from waiting list
	if !data.Confirmed {
		go config.promoteFromWaitingList(registration.RambleID)
	}

	render.JSON(w, r, updatedRegistration.ToModel())
}

// @Summary Cancel a registration
// @Description Cancel a user's registration for a ramble
// @Tags registrations
// @Accept json
// @Produce json
// @Param id path int true "Registration ID"
// @Param cancellation body model.RambleRegistrationCancelPayload true "Cancellation data"
// @Success 200 {object} model.RambleRegistration
// @Failure 400 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/{id}/cancel [put]
func (config *Config) Cancel(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid ID")))
		return
	}

	data := &model.RambleRegistrationCancelPayload{}
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

	if registration.Status == "cancelled" {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("registration already cancelled")))
		return
	}

	now := time.Now()
	registration.Status = "cancelled"
	registration.CancellationDate = &now
	registration.CancellationReason = data.Reason

	updatedRegistration, err := config.RambleRegistrationRepository.Update(registration)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	// Try to promote someone from waiting list
	go config.promoteFromWaitingList(registration.RambleID)

	render.JSON(w, r, updatedRegistration.ToModel())
}

// @Summary Get ramble registrations
// @Description Get all registrations for a specific ramble (admin only)
// @Tags registrations
// @Produce json
// @Param rambleId path int true "Ramble ID"
// @Success 200 {array} model.RambleRegistration
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 500 {object} errors.ErrResponse
// @Router /registrations/ramble/{rambleId} [get]
func (config *Config) GetRambleRegistrations(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	// TODO: Add permission check for admin access

	rambleIdParam := chi.URLParam(r, "rambleId")
	rambleId, err := strconv.Atoi(rambleIdParam)
	if err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid ramble ID")))
		return
	}

	filter := &dbmodel.RambleRegistrationFilter{
		RambleID: func() *uint { id := uint(rambleId); return &id }(),
	}

	registrations, err := config.RambleRegistrationRepository.FindAll(filter)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}

	result := make([]*model.RambleRegistration, len(registrations))
	for i, reg := range registrations {
		result[i] = reg.ToModel()
	}

	render.JSON(w, r, result)
}

// Helper methods

func (config *Config) findOrCreateUser(email, firstName, lastName string, phone *string) (*dbmodel.User, error) {
	// Try to find existing user
	user, err := config.UserRepository.FindByEmail(email, nil)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	// Create new user with profile
	newUser := &dbmodel.User{
		Email: email,
		UserProfile: dbmodel.UserProfile{
			FirstName: firstName,
			LastName:  &lastName,
			Phone:     phone,
		},
	}

	createdUser, err := config.UserRepository.Create(newUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (config *Config) sendRegistrationConfirmationEmail(registration *dbmodel.RambleRegistration, ramble *dbmodel.Ramble) {
	data := email.RegistrationConfirmationData{
		FirstName:        registration.FirstName,
		LastName:         registration.LastName,
		Title:            ramble.Title,
		Date:             formatDate(ramble.Date),
		Location:         formatLocation(ramble.Location),
		Status:           registration.Status,
		IsGroup:          false,
		ParticipantCount: 1,
		ParticipantNames: []string{registration.FirstName + " " + registration.LastName},
	}

	err := config.EmailService.Send(
		registration.Email,
		"Confirmation d'inscription - Balade Écologique",
		"registration_confirmation",
		data,
	)

	if err != nil {
		// Log error but don't fail the registration
		fmt.Printf("Failed to send registration confirmation email: %v\n", err)
	}
}

func (config *Config) promoteFromWaitingList(rambleID uint) {
	// Get next person from waiting list
	nextRegistration, err := config.RambleRegistrationRepository.GetNextInWaitingList(rambleID)
	if err != nil || nextRegistration == nil {
		return
	}

	// Get ramble details
	ramble, err := config.RambleRepository.FindByID(rambleID)
	if err != nil || ramble == nil {
		return
	}

	// Check if there's still space
	confirmedCount, err := config.RambleRegistrationRepository.CountByRambleAndStatus(rambleID, "confirmed")
	if err != nil {
		return
	}

	pendingCount, err := config.RambleRegistrationRepository.CountByRambleAndStatus(rambleID, "pending")
	if err != nil {
		return
	}

	if ramble.MaxParticipants != nil {
		totalRegistered := confirmedCount + pendingCount
		if totalRegistered >= int64(*ramble.MaxParticipants) {
			return // Still full
		}
	}

	// Promote from waiting list
	nextRegistration.Status = "pending"
	_, err = config.RambleRegistrationRepository.Update(nextRegistration)
	if err != nil {
		return
	}

	// Send notification email
	data := email.SpotAvailableData{
		FirstName: nextRegistration.FirstName,
		LastName:  nextRegistration.LastName,
		Title:     ramble.Title,
		Date:      formatDate(ramble.Date),
		Location:  formatLocation(ramble.Location),
	}

	config.EmailService.Send(
		nextRegistration.Email,
		"Place disponible - Balade Écologique",
		"spot_available",
		data,
	)
}

func formatDate(date *time.Time) string {
	if date == nil {
		return "Date à confirmer"
	}
	return date.Format("2 January 2006 à 15:04")
}

func formatLocation(location *string) string {
	if location == nil {
		return "Lieu à confirmer"
	}
	return *location
}

// createSingleRegistration handles the creation of a single participant registration
func (config *Config) createSingleRegistration(data *model.RambleRegistrationCreatePayload, ramble *dbmodel.Ramble) (*dbmodel.RambleRegistration, error) {
	// Use first participant data for backward compatibility
	participant := data.Participants[0]

	// Check if user already registered for this ramble
	existingFilter := &dbmodel.RambleRegistrationFilter{
		RambleID: &data.RambleID,
		Email:    &participant.Email,
	}
	existing, err := config.RambleRegistrationRepository.FindAll(existingFilter)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return nil, fmt.Errorf("already registered for this ramble")
	}

	// Create or find user
	user, err := config.findOrCreateUser(participant.Email, participant.FirstName, participant.LastName, participant.Phone)
	if err != nil {
		return nil, err
	}

	// Check availability and determine status
	confirmedCount, err := config.RambleRegistrationRepository.CountByRambleAndStatus(data.RambleID, "confirmed")
	if err != nil {
		return nil, err
	}

	pendingCount, err := config.RambleRegistrationRepository.CountByRambleAndStatus(data.RambleID, "pending")
	if err != nil {
		return nil, err
	}

	status := "pending"
	if ramble.MaxParticipants != nil {
		totalRegistered := confirmedCount + pendingCount
		if totalRegistered >= int64(*ramble.MaxParticipants) {
			status = "waiting_list"
		}
	}

	// Create registration
	registration := &dbmodel.RambleRegistration{
		FirstName:        participant.FirstName,
		LastName:         participant.LastName,
		Email:            participant.Email,
		Phone:            participant.Phone,
		Status:           status,
		RegistrationDate: time.Now(),
		RambleID:         data.RambleID,
		UserID:           &user.ID,
	}

	createdRegistration, err := config.RambleRegistrationRepository.Create(registration)
	if err != nil {
		return nil, err
	}

	// Send confirmation email
	go config.sendRegistrationConfirmationEmail(createdRegistration, ramble)

	return createdRegistration, nil
}

// createGroupRegistration handles the creation of a group registration with multiple participants
func (config *Config) createGroupRegistration(data *model.RambleRegistrationCreatePayload, ramble *dbmodel.Ramble) (*dbmodel.RambleRegistration, error) {
	// Check if any participant is already registered
	for _, participant := range data.Participants {
		existingFilter := &dbmodel.RambleRegistrationFilter{
			RambleID: &data.RambleID,
			Email:    &participant.Email,
		}
		existing, err := config.RambleRegistrationRepository.FindAll(existingFilter)
		if err != nil {
			return nil, err
		}
		if len(existing) > 0 {
			return nil, fmt.Errorf("participant %s %s is already registered for this ramble", participant.FirstName, participant.LastName)
		}
	}

	// Check availability for the entire group
	confirmedCount, err := config.RambleRegistrationRepository.CountByRambleAndStatus(data.RambleID, "confirmed")
	if err != nil {
		return nil, err
	}

	pendingCount, err := config.RambleRegistrationRepository.CountByRambleAndStatus(data.RambleID, "pending")
	if err != nil {
		return nil, err
	}

	groupSize := len(data.Participants)
	status := "pending"
	if ramble.MaxParticipants != nil {
		totalRegistered := confirmedCount + pendingCount
		if totalRegistered+int64(groupSize) > int64(*ramble.MaxParticipants) {
			status = "waiting_list"
		}
	}

	// Create group record
	group := &dbmodel.RambleRegistrationGroup{
		RambleID:         data.RambleID,
		PrimaryEmail:     data.Participants[0].Email, // First participant is primary contact
		Status:           status,
		RegistrationDate: time.Now(),
	}

	createdGroup, err := config.RambleRegistrationGroupRepository.Create(group)
	if err != nil {
		return nil, err
	}

	// Create individual registrations for each participant
	var primaryRegistration *dbmodel.RambleRegistration
	for i, participant := range data.Participants {
		// Create or find user
		user, err := config.findOrCreateUser(participant.Email, participant.FirstName, participant.LastName, participant.Phone)
		if err != nil {
			return nil, err
		}

		// Create registration
		registration := &dbmodel.RambleRegistration{
			FirstName:        participant.FirstName,
			LastName:         participant.LastName,
			Email:            participant.Email,
			Phone:            participant.Phone,
			Status:           status,
			RegistrationDate: time.Now(),
			RambleID:         data.RambleID,
			UserID:           &user.ID,
			GroupID:          &createdGroup.ID,
		}

		createdRegistration, err := config.RambleRegistrationRepository.Create(registration)
		if err != nil {
			return nil, err
		}

		// Keep track of the primary contact's registration
		if i == 0 {
			primaryRegistration = createdRegistration
		}
	}

	// Send confirmation email to primary contact
	go config.sendGroupRegistrationConfirmationEmail(primaryRegistration, ramble, createdGroup)

	return primaryRegistration, nil
}

// sendGroupRegistrationConfirmationEmail sends confirmation email for group registrations
func (config *Config) sendGroupRegistrationConfirmationEmail(registration *dbmodel.RambleRegistration, ramble *dbmodel.Ramble, group *dbmodel.RambleRegistrationGroup) {
	// Get all participants in the group
	groupFilter := &dbmodel.RambleRegistrationFilter{
		GroupID: &group.ID,
	}
	participants, err := config.RambleRegistrationRepository.FindAll(groupFilter)
	if err != nil {
		return // Silently fail email sending
	}

	// Build participant list for email
	participantNames := make([]string, len(participants))
	for i, p := range participants {
		participantNames[i] = p.FirstName + " " + p.LastName
	}

	data := email.RegistrationConfirmationData{
		FirstName:        registration.FirstName,
		LastName:         registration.LastName,
		Title:            ramble.Title,
		Date:             formatDate(ramble.Date),
		Location:         formatLocation(ramble.Location),
		Status:           registration.Status,
		IsGroup:          true,
		ParticipantCount: len(participants),
		ParticipantNames: participantNames,
	}

	config.EmailService.Send(
		registration.Email,
		"Confirmation d'inscription - Balade Écologique",
		"registration_confirmation",
		data,
	)
}
