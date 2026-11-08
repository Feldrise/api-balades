package reservation

import (
	"net/http"
	"time"

	"feldrise.com/balade/database/dbmodel"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/errors"
	"feldrise.com/balade/pkg/model"
	"github.com/go-chi/render"
)

// @Summary Mark the presence of participants (admin or guide)
// @Description Records the presence of participants of the ramble's reservations, from the day of the ramble (Paris time). Each change says when it was made on the guide's phone: the latest change wins, so changes queued offline can be sent again safely. A change is refused per participant (outcome reservation_cancelled or unknown_participant) without failing the others.
// @Tags admin,reservations
// @Accept json
// @Produce json
// @Param rambleId path int true "Ramble ID"
// @Param body body model.PresenceUpdatePayload true "Presence changes"
// @Success 200 {object} model.PresenceUpdateResponse
// @Failure 400 {object} errors.ErrResponse
// @Failure 401 {object} errors.ErrResponse
// @Failure 403 {object} errors.ErrResponse
// @Failure 404 {object} errors.ErrResponse
// @Failure 409 {object} errors.ErrResponse "ramble_not_started"
// @Router /reservations/admin/rambles/{rambleId}/presence [put]
func (config *Config) AdminMarkPresence(w http.ResponseWriter, r *http.Request) {
	user := authentication.ForContext(r.Context())
	if user == nil {
		render.Render(w, r, errors.ErrUnauthorized("authentication required"))
		return
	}

	rambleID, ok := parseID(w, r, "rambleId")
	if !ok {
		return
	}

	data := &model.PresenceUpdatePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	allowed, err := config.canActOnRamble(user, rambleID, "update:registration-details", "update:registration-details:self")
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}
	if !allowed {
		render.Render(w, r, errors.ErrForbidden("insufficient permissions"))
		return
	}

	ramble, err := config.RambleRepository.FindByID(rambleID, true)
	if err != nil {
		render.Render(w, r, errors.ErrServerError(err))
		return
	}
	if ramble == nil {
		render.Render(w, r, errors.ErrNotFound())
		return
	}

	now := config.Clock.Now()
	if !rambleDayStarted(ramble.Date, now) {
		render.Render(w, r, renderError(errRambleNotStarted()))
		return
	}

	response := &model.PresenceUpdateResponse{Participants: make([]model.PresenceOutcome, 0, len(data.Changes))}
	for _, change := range data.Changes {
		markedAt := now
		if change.MarkedAt != nil && change.MarkedAt.Before(now) {
			markedAt = *change.MarkedAt
		}

		result, err := config.ReservationRepository.MarkPresence(rambleID, dbmodel.PresenceChange{
			ParticipantID: change.ParticipantID,
			Present:       change.Present,
			MarkedAt:      markedAt,
			MarkedByID:    user.ID,
		})
		if err != nil {
			render.Render(w, r, errors.ErrServerError(err))
			return
		}

		response.Participants = append(response.Participants, presenceOutcome(change, result))
	}

	render.JSON(w, r, response)
}

func presenceOutcome(change model.PresenceChangePayload, result dbmodel.PresenceResult) model.PresenceOutcome {
	outcome := model.PresenceOutcome{ParticipantID: change.ParticipantID}
	if result.Participant == nil {
		outcome.Outcome = "unknown_participant"
		return outcome
	}

	outcome.Present = result.Participant.Present
	outcome.PresenceMarkedAt = result.Participant.PresenceMarkedAt
	outcome.PresenceMarkedBy = result.Participant.PresenceMarkedByID

	switch {
	case result.ReservationStatus == dbmodel.ReservationStatusCancelled:
		outcome.Outcome = "reservation_cancelled"
	case result.Applied:
		outcome.Outcome = "applied"
	case samePresence(result.Participant.Present, change.Present):
		outcome.Outcome = "unchanged"
	default:
		outcome.Outcome = "outdated"
	}
	return outcome
}

func samePresence(a, b *bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// rambleDayStarted reports whether now is on or after the day of the ramble, in Paris time.
func rambleDayStarted(rambleDate *time.Time, now time.Time) bool {
	if rambleDate == nil {
		return false
	}

	location := time.Local
	if paris, err := time.LoadLocation("Europe/Paris"); err == nil {
		location = paris
	}

	day := rambleDate.In(location)
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, location)
	return !now.Before(startOfDay)
}
