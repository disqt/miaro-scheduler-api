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

type Schedule struct {
	timeRequested time.Time
	scheduleType  ScheduleType
	dayInSchedule int
}

/*
CalculateSchedule
This defaults to time.Now() if there is no parameters.
This will exit the program if there are more than 1 parameter.
*/
func CalculateSchedule(timeReq ...time.Time) Schedule {
	var date time.Time

	// This is to support optional parameters
	if len(timeReq) == 0 {
		date = time.Now()
	} else if len(timeReq) == 1 {
		date = timeReq[0]
	} else {
		panic("Received multiple dates")
	}

	initialDate := time.Date(2024, time.August, 31, 0, 0, 0, 0, time.UTC)

	diffDays := int(date.Sub(initialDate).Hours() / 24)

	dayInSchedule := diffDays % 10

	return Schedule{
		timeRequested: date,
		scheduleType:  schedule[dayInSchedule],
		dayInSchedule: dayInSchedule,
	}
}
