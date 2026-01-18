package pkg

import (
	"time"
)

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
}

/*
CalculateSchedule
This defaults to time.Now() if there is no parameters.
This will exit the program if there are more than 1 parameter.
*/
func CalculateSchedule(timeReq ...time.Time) Schedule {
	var date time.Time
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		panic("Failed to load Europe/Paris timezone: " + err.Error())
	}

	// This is to support optional parameters
	if len(timeReq) == 0 {
		date = time.Now().In(loc)
	} else if len(timeReq) == 1 {
		date = timeReq[0].In(loc)
	} else {
		panic("Received multiple dates")
	}

	// Use the same timezone for the initial date to ensure consistent day calculations
	initialDate := time.Date(2024, time.August, 31, 0, 0, 0, 0, loc)

	diffDays := int(date.Sub(initialDate).Hours() / 24)

	dayInSchedule := diffDays % 10

	return Schedule{
		TimeRequested: date,
		ScheduleType:  schedule[dayInSchedule],
		DayInSchedule: dayInSchedule,
	}
}
