package pkg

import (
	"fmt"
	"time"
)

// ScheduleBeautified contains human-readable French text representations of the schedule information.
type ScheduleBeautified struct {
	Schedule               string        // e.g., "du matin", "de l'après-midi", "de nuit", "libre"
	IsWorking              string        // e.g., "est au travail", "n'est pas au travail"
	ScheduleNextWorkingDay string        // e.g., "du matin" (for the next working day)
	NextWorkingDay         string        // e.g., "demain", "dans 3 jours"
	CalendarDays           []CalendarDay // Calendar data for the month
	CalendarMonthLabel     string        // e.g., "Mars 2026"
	PrevMonth              string        // e.g., "2026-02" for prev link
	NextMonth              string        // e.g., "2026-04" for next link
	IsCurrentMonth         bool          // True when viewing current month
}

// CalendarDay represents a single day in the calendar view
type CalendarDay struct {
	DayNumber    int    // Day of the month (1-31)
	ShiftType    string // "Matin", "Après-midi", "Nuit", "Libre"
	IsToday      bool   // True if this is today
	IsNextWork   bool   // True if this is the next working day
	IsOtherMonth bool   // True for days from adjacent months
	ShiftClass   string // CSS class: "morning", "afternoon", "night", "free"
}

var frenchMonths = map[time.Month]string{
	time.January:   "Janvier",
	time.February:  "Février",
	time.March:     "Mars",
	time.April:     "Avril",
	time.May:       "Mai",
	time.June:      "Juin",
	time.July:      "Juillet",
	time.August:    "Août",
	time.September: "Septembre",
	time.October:   "Octobre",
	time.November:  "Novembre",
	time.December:  "Décembre",
}

// FormatScheduleBeautified converts a Schedule into human-readable French text.
// It determines the current shift type, whether the person is currently working,
// and when the next working day will be.
func FormatScheduleBeautified(schedule Schedule, targetMonth time.Time) ScheduleBeautified {
	currentDay := getBeautifiedSchedule(schedule.ScheduleType)
	isWorking := isWorkingString(schedule)
	scheduleNextWorkingDay, nextWorkingDay, nextWorkDays := nextWorkingDay(schedule.DayInSchedule)

	calendarDays := GenerateCalendarData(schedule, targetMonth, nextWorkDays)

	now := schedule.TimeRequested.In(parisLoc)
	year, month, _ := targetMonth.In(parisLoc).Date()

	prevMonth := targetMonth.AddDate(0, -1, 0)
	nextMonthDate := targetMonth.AddDate(0, 1, 0)

	return ScheduleBeautified{
		Schedule:               currentDay,
		IsWorking:              isWorking,
		ScheduleNextWorkingDay: scheduleNextWorkingDay,
		NextWorkingDay:         nextWorkingDay,
		CalendarDays:           calendarDays,
		CalendarMonthLabel:     fmt.Sprintf("%s %d", frenchMonths[month], year),
		PrevMonth:              prevMonth.Format("2006-01"),
		NextMonth:              nextMonthDate.Format("2006-01"),
		IsCurrentMonth:         now.Year() == year && now.Month() == month,
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

// GenerateCalendarData generates calendar data for the target month with full-week grids.
// Leading and trailing days from adjacent months are included to complete each row.
func GenerateCalendarData(currentSchedule Schedule, targetMonth time.Time, nextWorkDays int) []CalendarDay {
	now := currentSchedule.TimeRequested.In(parisLoc)
	targetMonth = targetMonth.In(parisLoc)

	year, month, _ := targetMonth.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, parisLoc)
	lastDay := firstDay.AddDate(0, 1, -1)

	isCurrentMonth := now.Year() == year && now.Month() == month

	// Monday = 0, Sunday = 6
	weekdayOffset := int(firstDay.Weekday()) - 1
	if weekdayOffset == -1 {
		weekdayOffset = 6
	}

	var calendarDays []CalendarDay

	// Leading days from previous month
	for i := weekdayOffset; i > 0; i-- {
		d := firstDay.AddDate(0, 0, -i)
		calendarDays = append(calendarDays, buildCalendarDay(d, currentSchedule.Team, now, isCurrentMonth, nextWorkDays, true))
	}

	// Days of the target month
	for day := 1; day <= lastDay.Day(); day++ {
		d := time.Date(year, month, day, 12, 0, 0, 0, parisLoc)
		calendarDays = append(calendarDays, buildCalendarDay(d, currentSchedule.Team, now, isCurrentMonth, nextWorkDays, false))
	}

	// Trailing days from next month
	trailingDays := 7 - (len(calendarDays) % 7)
	if trailingDays < 7 {
		for i := 1; i <= trailingDays; i++ {
			d := lastDay.AddDate(0, 0, i)
			calendarDays = append(calendarDays, buildCalendarDay(d, currentSchedule.Team, now, isCurrentMonth, nextWorkDays, true))
		}
	}

	return calendarDays
}

func buildCalendarDay(date time.Time, team int, now time.Time, isCurrentMonth bool, nextWorkDays int, isOtherMonth bool) CalendarDay {
	daySchedule := CalculateSchedule(date, team)
	_, _, day := date.Date()

	isToday := false
	isNextWork := false
	if isCurrentMonth && !isOtherMonth {
		isToday = day == now.Day()
		isNextWork = !isToday && daySchedule.ScheduleType != FREE &&
			(date.After(now) || date.Equal(now)) &&
			day == now.Day()+nextWorkDays
	}

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

	return CalendarDay{
		DayNumber:    day,
		ShiftType:    shiftType,
		IsToday:      isToday,
		IsNextWork:   isNextWork,
		IsOtherMonth: isOtherMonth,
		ShiftClass:   shiftClass,
	}
}
