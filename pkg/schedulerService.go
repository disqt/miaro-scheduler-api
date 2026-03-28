package pkg

import (
	"time"
)

// parisLoc is loaded once at init to avoid repeated time.LoadLocation calls.
var parisLoc *time.Location

func init() {
	var err error
	parisLoc, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic("Failed to load Europe/Paris timezone: " + err.Error())
	}
}

type ScheduleType string

const (
	MORNING   ScheduleType = "MORNING"
	AFTERNOON ScheduleType = "AFTERNOON"
	NIGHT     ScheduleType = "NIGHT"
	FREE      ScheduleType = "FREE"
)

var schedule = map[int]ScheduleType{
	0: MORNING,
	1: MORNING,
	2: AFTERNOON,
	3: AFTERNOON,
	4: NIGHT,
	5: NIGHT,
	6: FREE,
	7: FREE,
	8: FREE,
	9: FREE,
}

// Schedule represents the work schedule information for a given time.
type Schedule struct {
	TimeRequested time.Time    `json:"time_requested"`
	ScheduleType  ScheduleType `json:"schedule_type"`
	DayInSchedule int          `json:"day_in_schedule"`
	Team          int          `json:"team"`
}

// ParisLoc returns the Europe/Paris timezone location.
func ParisLoc() *time.Location {
	return parisLoc
}

// CalculateSchedule computes the shift for the given date and team.
// Each team's epoch is offset forward by (team-1)*2 days from the base epoch (Aug 31, 2024).
// Invalid team numbers are clamped to MiaroTeam (team 1).
func CalculateSchedule(date time.Time, team int) Schedule {
	date = date.In(parisLoc)
	team = ValidateTeam(team)

	// Offset epoch forward by (team-1)*2 days
	initialDate := time.Date(2024, time.August, 31, 0, 0, 0, 0, parisLoc)
	initialDate = initialDate.AddDate(0, 0, (team-1)*2)

	diffDays := int(date.Sub(initialDate).Hours() / 24)

	// Positive modulo to handle dates before a team's epoch
	dayInSchedule := ((diffDays % 10) + 10) % 10

	return Schedule{
		TimeRequested: date,
		ScheduleType:  schedule[dayInSchedule],
		DayInSchedule: dayInSchedule,
		Team:          team,
	}
}
