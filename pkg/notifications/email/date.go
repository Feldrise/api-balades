package email

import (
	"fmt"
	"time"
)

var frenchWeekdays = [...]string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"}

var frenchMonths = [...]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}

// FormatDate formats a date for the emails, in French and Paris time, such as
// « samedi 15 juin 2030 à 11h00 ».
func FormatDate(date *time.Time) string {
	if date == nil {
		return "Date à confirmer"
	}

	local := date.Local()
	if paris, err := time.LoadLocation("Europe/Paris"); err == nil {
		local = date.In(paris)
	}

	return fmt.Sprintf("%s %d %s %d à %dh%02d",
		frenchWeekdays[local.Weekday()], local.Day(), frenchMonths[local.Month()-1], local.Year(), local.Hour(), local.Minute())
}
