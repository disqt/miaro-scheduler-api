package pkg

import "fmt"

type ScheduleBeautified struct {
	Schedule               string
	IsWorking              string
	ScheduleNextWorkingDay string
	NextWorkingDay         string
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

	hour := schedule.timeRequested.Hour()
	dayInSchedule := schedule.dayInSchedule

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
	for schedule[day+i] == FREE {
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
