package pkg

import (
	"fmt"
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

func FormatScheduleBeautified(schedule Schedule) ScheduleBeautified {
	currentDay := getBeautifiedSchedule(schedule.scheduleType)

	isWorking := isWorkingString(schedule)

	scheduleNextWorkingDay, nextWorkingDay := nextWorkingDay(schedule.dayInSchedule)

	return ScheduleBeautified{
		Schedule:               currentDay,
		IsWorking:              isWorking,
		ScheduleNextWorkingDay: scheduleNextWorkingDay,
		NextWorkingDay:         nextWorkingDay,
	}
}

func getBeautifiedSchedule(scheduleType ScheduleType) string {
	var beautifiedSchedule string
	if scheduleType == MORNING {
		beautifiedSchedule = "du matin"
	} else if scheduleType == AFTERNOON {
		beautifiedSchedule = "de l'après-midi"
	} else if scheduleType == NIGHT {
		beautifiedSchedule = "de nuit"
	} else if scheduleType == FREE {
		beautifiedSchedule = "libre"
	}

	return beautifiedSchedule
}

func isWorkingString(schedule Schedule) string {
	isWorking := false

	if schedule.scheduleType == FREE {
		isWorking = false
	}

	hourRequested := schedule.timeRequested.Hour()
	if schedule.scheduleType == MORNING && (hourRequested >= 6 && hourRequested <= 14) {
		isWorking = true
	} else if schedule.scheduleType == AFTERNOON && (hourRequested >= 14 && hourRequested <= 22) {
		isWorking = true
	} else if schedule.scheduleType == NIGHT && (hourRequested >= 22 || hourRequested <= 6) {
		isWorking = true
	}

	if isWorking {
		return "est au travail"
	} else {
		return "n'est pas au travail"
	}
}

func nextWorkingDay(day int) (string, string) {
	day++ // This is to not return today's date
	var i = 1
	for schedule[day+i] == FREE {
		day = (day + 1) % 10
		i++
	}

	var nextWorkingDay string

	if i == 1 {
		nextWorkingDay = "demain"
	} else {
		nextWorkingDay = fmt.Sprintf("dans %v jours", i)
	}

	return getBeautifiedSchedule(schedule[(day+i)%10]), nextWorkingDay
}
