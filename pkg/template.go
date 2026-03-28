package pkg

import (
	"fmt"
	"time"
)

// ScheduleBeautified contains human-readable French text representations of the schedule information.
type ScheduleBeautified struct {
	Schedule               string       // e.g., "du matin", "de l'après-midi", "de nuit", "libre"
	IsWorking              string       // e.g., "est au travail", "n'est pas au travail"
	ScheduleNextWorkingDay string       // e.g., "du matin" (for the next working day)
	NextWorkingDay         string       // e.g., "demain", "dans 3 jours"
	CalendarDays           []CalendarDay // Calendar data for the month
}

// CalendarDay represents a single day in the calendar view
type CalendarDay struct {
	DayNumber     int    // Day of the month (1-31)
	ShiftType     string // "Matin", "Après-midi", "Nuit", "Libre"
	IsToday       bool   // True if this is today
	IsNextWork    bool   // True if this is the next working day
	IsEmpty       bool   // True for empty cells at start/end of month
	ShiftClass    string // CSS class: "morning", "afternoon", "night", "free"
}

// FormatScheduleBeautified converts a Schedule into human-readable French text.
// It determines the current shift type, whether the person is currently working,
// and when the next working day will be.
func FormatScheduleBeautified(schedule Schedule) ScheduleBeautified {
	currentDay := getBeautifiedSchedule(schedule.ScheduleType)

	isWorking := isWorkingString(schedule)

	scheduleNextWorkingDay, nextWorkingDay, nextWorkDays := nextWorkingDay(schedule.DayInSchedule)

	calendarDays := GenerateCalendarData(schedule, nextWorkDays)

	return ScheduleBeautified{
		Schedule:               currentDay,
		IsWorking:              isWorking,
		ScheduleNextWorkingDay: scheduleNextWorkingDay,
		NextWorkingDay:         nextWorkingDay,
		CalendarDays:           calendarDays,
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

func nextWorkingDay(day int) (string, string, int) {
	i := 1 // Start from tomorrow
	for schedule[(day+i)%10] == FREE {
		i = i + 1
	}

	var nextWorkingDayStr string
	if i == 1 {
		nextWorkingDayStr = "demain"
	} else {
		nextWorkingDayStr = fmt.Sprintf("dans %v jours", i)
	}

	return getBeautifiedSchedule(schedule[(day+i)%10]), nextWorkingDayStr, i
}

// GenerateCalendarData generates calendar data for the current month
func GenerateCalendarData(currentSchedule Schedule, nextWorkDays int) []CalendarDay {
	now := currentSchedule.TimeRequested.In(parisLoc)

	// Get first and last day of current month
	year, month, _ := now.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, parisLoc)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Calculate weekday offset (Monday = 0, Sunday = 6)
	weekdayOffset := int(firstDay.Weekday()) - 1
	if weekdayOffset == -1 {
		weekdayOffset = 6
	}

	var calendarDays []CalendarDay

	// Add empty cells for days before month starts
	for i := 0; i < weekdayOffset; i++ {
		calendarDays = append(calendarDays, CalendarDay{IsEmpty: true})
	}

	// Add actual days of the month
	for day := 1; day <= lastDay.Day(); day++ {
		currentDate := time.Date(year, month, day, 12, 0, 0, 0, parisLoc)
		daySchedule := CalculateSchedule(currentDate, currentSchedule.Team)

		isToday := day == now.Day()
		isNextWork := !isToday && daySchedule.ScheduleType != FREE &&
			(currentDate.After(now) || currentDate.Equal(now)) &&
			day == now.Day()+nextWorkDays

		var shiftType, shiftClass string
		switch daySchedule.ScheduleType {
		case MORNING:
			shiftType = "Matin"
			shiftClass = "morning"
		case AFTERNOON:
			shiftType = "Après-midi"
			shiftClass = "afternoon"
		case NIGHT:
			shiftType = "Nuit"
			shiftClass = "night"
		case FREE:
			shiftType = "Libre"
			shiftClass = "free"
		}

		calendarDays = append(calendarDays, CalendarDay{
			DayNumber:  day,
			ShiftType:  shiftType,
			IsToday:    isToday,
			IsNextWork: isNextWork,
			IsEmpty:    false,
			ShiftClass: shiftClass,
		})
	}

	return calendarDays
}
