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
			timeRequested: time.Date(2024, time.August, 31, 7, 0, 0, 0, time.UTC),
			dayInSchedule: 0,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("5h matin", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 5, 0, 0, 0, time.UTC),
			dayInSchedule: 1,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("15h aprem", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 15, 0, 0, 0, time.UTC),
			dayInSchedule: 3,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("13h aprem", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 13, 0, 0, 0, time.UTC),
			dayInSchedule: 2,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("23h premiere nuit", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 23, 0, 0, 0, time.UTC),
			dayInSchedule: 4,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("13h premiere nuit", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 13, 0, 0, 0, time.UTC),
			dayInSchedule: 4,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("5h 2eme nuit", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 5, 0, 0, 0, time.UTC),
			dayInSchedule: 5,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("8h 2eme nuit", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 8, 0, 0, 0, time.UTC),
			dayInSchedule: 5,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

	t.Run("23h 2eme nuit", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 23, 0, 0, 0, time.UTC),
			dayInSchedule: 5,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("5h premiere journee libre", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 5, 0, 0, 0, time.UTC),
			dayInSchedule: 6,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "est au travail", res)
	})

	t.Run("8h premiere journee libre", func(t *testing.T) {
		schedule := Schedule{
			timeRequested: time.Date(2024, time.August, 31, 8, 0, 0, 0, time.UTC),
			dayInSchedule: 6,
		}

		res := isWorkingString(schedule)

		assert.Equal(t, "n'est pas au travail", res)
	})

}
