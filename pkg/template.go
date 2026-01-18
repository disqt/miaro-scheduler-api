package pkg

import "fmt"

// ScheduleBeautified contains human-readable French text representations of the schedule information.
type ScheduleBeautified struct {
	Schedule               string // e.g., "du matin", "de l'après-midi", "de nuit", "libre"
	IsWorking              string // e.g., "est au travail", "n'est pas au travail"
	ScheduleNextWorkingDay string // e.g., "du matin" (for the next working day)
	NextWorkingDay         string // e.g., "demain", "dans 3 jours"
}

// FormatScheduleBeautified converts a Schedule into human-readable French text.
// It determines the current shift type, whether the person is currently working,
// and when the next working day will be.
func FormatScheduleBeautified(schedule Schedule) ScheduleBeautified {
	currentDay := getBeautifiedSchedule(schedule.ScheduleType)

	isWorking := isWorkingString(schedule)

	scheduleNextWorkingDay, nextWorkingDay := nextWorkingDay(schedule.DayInSchedule)

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

	hour := schedule.TimeRequested.Hour()
	dayInSchedule := schedule.DayInSchedule

	if (dayInSchedule == 0 || dayInSchedule == 1) && (hour >= 6 && hour < 14) {
		isWorking = true
	} else if (dayInSchedule == 2 || dayInSchedule == 3) && (hour >= 14 && hour < 22) {
		isWorking = true
	} else if dayInSchedule == 4 && hour >= 22 {
		isWorking = true
	} else if dayInSchedule == 5 && (hour < 6 || hour >= 22) {
		isWorking = true
	} else if dayInSchedule == 6 && hour < 6 {
		isWorking = true
	}

	if isWorking {
		return "est au travail"
	} else {
		return "n'est pas au travail"
	}
}

func nextWorkingDay(day int) (string, string) {
	i := 1 // We initialise at 1 to start from tomorrow
	for schedule[(day+i)%10] == FREE {
		i = i + 1
	}

	var nextWorkingDay string

	if i == 1 {
		nextWorkingDay = "demain"
	} else {
		nextWorkingDay = fmt.Sprintf("dans %v jours", i)
	}

	return getBeautifiedSchedule(schedule[(day+i)%10]), nextWorkingDay
}
