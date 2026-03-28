package pkg

import (
	"github.com/go-playground/assert/v2"
	"testing"
	"time"
)

func TestGetBeautifiedSchedule(t *testing.T) {
	t.Run("NIGHT", func(t *testing.T) {
		res := getBeautifiedSchedule(NIGHT)

		assert.Equal(t, "de nuit", res)
	})

	t.Run("AFTERNOON", func(t *testing.T) {
		res := getBeautifiedSchedule(AFTERNOON)

		assert.Equal(t, "de l'après-midi", res)
	})

	t.Run("MORNING", func(t *testing.T) {
		res := getBeautifiedSchedule(MORNING)

		assert.Equal(t, "du matin", res)
	})

	t.Run("FREE", func(t *testing.T) {
		res := getBeautifiedSchedule(FREE)

		assert.Equal(t, "libre", res)
	})
}

func TestIsWorkingString(t *testing.T) {
	t.Run("7h matin", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 7, 0, 0, 0, time.UTC),
			DayInSchedule: 0,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("5h matin", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 5, 0, 0, 0, time.UTC),
			DayInSchedule: 1,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("15h aprem", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 15, 0, 0, 0, time.UTC),
			DayInSchedule: 3,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("13h aprem", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 13, 0, 0, 0, time.UTC),
			DayInSchedule: 2,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("23h premiere nuit", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 23, 0, 0, 0, time.UTC),
			DayInSchedule: 4,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("13h premiere nuit", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 13, 0, 0, 0, time.UTC),
			DayInSchedule: 4,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("5h 2eme nuit", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 5, 0, 0, 0, time.UTC),
			DayInSchedule: 5,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("8h 2eme nuit", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 8, 0, 0, 0, time.UTC),
			DayInSchedule: 5,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("23h 2eme nuit", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 23, 0, 0, 0, time.UTC),
			DayInSchedule: 5,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("5h premiere journee libre", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 5, 0, 0, 0, time.UTC),
			DayInSchedule: 6,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("8h premiere journee libre", func(t *testing.T) {
		schedule := Schedule{
			TimeRequested: time.Date(2024, time.August, 31, 8, 0, 0, 0, time.UTC),
			DayInSchedule: 6,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

}

func TestNextWorkingDay(t *testing.T) {
	t.Run("Morning 0", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(0)

		assert.Equal(t, "demain", nextDay)
		assert.Equal(t, "du matin", schedule)
	})

	t.Run("Morning 1", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(1)

		assert.Equal(t, "demain", nextDay)
		assert.Equal(t, "de l'après-midi", schedule)
	})

	t.Run("Afternoon 2", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(2)

		assert.Equal(t, "demain", nextDay)
		assert.Equal(t, "de l'après-midi", schedule)
	})

	t.Run("Afternoon 3", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(3)

		assert.Equal(t, "demain", nextDay)
		assert.Equal(t, "de nuit", schedule)
	})

	t.Run("Nuit  4", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(4)

		assert.Equal(t, "demain", nextDay)
		assert.Equal(t, "de nuit", schedule)
	})

	t.Run("Nuit  5", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(5)

		assert.Equal(t, "dans 5 jours", nextDay)
		assert.Equal(t, "du matin", schedule)
	})

	t.Run("Libre 6", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(6)

		assert.Equal(t, "dans 4 jours", nextDay)
		assert.Equal(t, "du matin", schedule)
	})

	t.Run("Libre 7", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(7)

		assert.Equal(t, "dans 3 jours", nextDay)
		assert.Equal(t, "du matin", schedule)
	})

	t.Run("Libre 8", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(8)

		assert.Equal(t, "dans 2 jours", nextDay)
		assert.Equal(t, "du matin", schedule)
	})

	t.Run("Libre 9", func(t *testing.T) {
		schedule, nextDay, _ := nextWorkingDay(9)

		assert.Equal(t, "demain", nextDay)
		assert.Equal(t, "du matin", schedule)
	})

}

func TestCalculateSchedule_UsesParisTimezone(t *testing.T) {
	utc := time.Date(2025, 6, 9, 13, 30, 0, 0, time.UTC)

	result := CalculateSchedule(utc, MiaroTeam)

	// the returned time requested should be in the Paris timezone
	// 13:30 UTC should be 15:30 in Europe/Paris (UTC + 2)
	expectedHour := 15
	expectedMinute := 30

	if result.TimeRequested.Location().String() != "Europe/Paris" {
		t.Errorf("Expected time to be in Paris timezone, got %s", result.TimeRequested.Location().String())
	}

	if result.TimeRequested.Hour() != expectedHour || result.TimeRequested.Minute() != expectedMinute {
		t.Errorf("Expected time %02d:%02d, got %02d:%02d",
			expectedHour, expectedMinute,
			result.TimeRequested.Hour(), result.TimeRequested.Minute())
	}
}
