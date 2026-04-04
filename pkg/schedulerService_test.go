package pkg

import (
	"testing"
	"time"
)

func TestCalculateSchedule_DefaultTime(t *testing.T) {
	// Test with current time and default team
	result := CalculateSchedule(time.Now(), MiaroTeam)

	// Verify that the result has the correct timezone
	if result.TimeRequested.Location().String() != "Europe/Paris" {
		t.Errorf("Expected timezone Europe/Paris, got %s", result.TimeRequested.Location().String())
	}

	// Verify day in schedule is within valid range
	if result.DayInSchedule < 0 || result.DayInSchedule > 9 {
		t.Errorf("Day in schedule out of range: %d", result.DayInSchedule)
	}

	// Verify schedule type is valid
	validTypes := map[ScheduleType]bool{
		MORNING:   true,
		AFTERNOON: true,
		NIGHT:     true,
		FREE:      true,
	}
	if !validTypes[result.ScheduleType] {
		t.Errorf("Invalid schedule type: %s", result.ScheduleType)
	}
}

func TestCalculateSchedule_SpecificDates(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")

	testCases := []struct {
		name         string
		date         time.Time
		expectedDay  int
		expectedType ScheduleType
		description  string
	}{
		{
			name:         "Start date (Aug 31, 2024)",
			date:         time.Date(2024, time.August, 31, 12, 0, 0, 0, loc),
			expectedDay:  0,
			expectedType: MORNING,
			description:  "Day 0 should be MORNING",
		},
		{
			name:         "Day 1",
			date:         time.Date(2024, time.September, 1, 12, 0, 0, 0, loc),
			expectedDay:  1,
			expectedType: MORNING,
			description:  "Day 1 should be MORNING",
		},
		{
			name:         "Day 2",
			date:         time.Date(2024, time.September, 2, 12, 0, 0, 0, loc),
			expectedDay:  2,
			expectedType: AFTERNOON,
			description:  "Day 2 should be AFTERNOON",
		},
		{
			name:         "Day 3",
			date:         time.Date(2024, time.September, 3, 12, 0, 0, 0, loc),
			expectedDay:  3,
			expectedType: AFTERNOON,
			description:  "Day 3 should be AFTERNOON",
		},
		{
			name:         "Day 4",
			date:         time.Date(2024, time.September, 4, 12, 0, 0, 0, loc),
			expectedDay:  4,
			expectedType: NIGHT,
			description:  "Day 4 should be NIGHT",
		},
		{
			name:         "Day 5",
			date:         time.Date(2024, time.September, 5, 12, 0, 0, 0, loc),
			expectedDay:  5,
			expectedType: NIGHT,
			description:  "Day 5 should be NIGHT",
		},
		{
			name:         "Day 6",
			date:         time.Date(2024, time.September, 6, 12, 0, 0, 0, loc),
			expectedDay:  6,
			expectedType: FREE,
			description:  "Day 6 should be FREE",
		},
		{
			name:         "Day 7",
			date:         time.Date(2024, time.September, 7, 12, 0, 0, 0, loc),
			expectedDay:  7,
			expectedType: FREE,
			description:  "Day 7 should be FREE",
		},
		{
			name:         "Day 8",
			date:         time.Date(2024, time.September, 8, 12, 0, 0, 0, loc),
			expectedDay:  8,
			expectedType: FREE,
			description:  "Day 8 should be FREE",
		},
		{
			name:         "Day 9",
			date:         time.Date(2024, time.September, 9, 12, 0, 0, 0, loc),
			expectedDay:  9,
			expectedType: FREE,
			description:  "Day 9 should be FREE",
		},
		{
			name:         "Day 10 (wraps to day 0)",
			date:         time.Date(2024, time.September, 10, 12, 0, 0, 0, loc),
			expectedDay:  0,
			expectedType: MORNING,
			description:  "Day 10 should wrap to day 0 (MORNING)",
		},
		{
			name:         "Day 20 (wraps to day 0)",
			date:         time.Date(2024, time.September, 20, 12, 0, 0, 0, loc),
			expectedDay:  0,
			expectedType: MORNING,
			description:  "Day 20 should wrap to day 0 (MORNING)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateSchedule(tc.date, MiaroTeam) // MiaroTeam (3) has effective epoch Aug 31

			if result.DayInSchedule != tc.expectedDay {
				t.Errorf("%s: expected day %d, got %d", tc.description, tc.expectedDay, result.DayInSchedule)
			}

			if result.ScheduleType != tc.expectedType {
				t.Errorf("%s: expected type %s, got %s", tc.description, tc.expectedType, result.ScheduleType)
			}
		})
	}
}

func TestCalculateSchedule_YearBoundary(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")

	// Test around year boundaries (only dates after start date Aug 31, 2024)
	testCases := []struct {
		name string
		date time.Time
	}{
		{
			name: "New Year 2025",
			date: time.Date(2025, time.January, 1, 0, 0, 0, 0, loc),
		},
		{
			name: "End of 2024",
			date: time.Date(2024, time.December, 31, 23, 59, 59, 0, loc),
		},
		{
			name: "One year after start",
			date: time.Date(2025, time.August, 31, 0, 0, 0, 0, loc),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateSchedule(tc.date, MiaroTeam)

			// Just verify we get valid results
			if result.DayInSchedule < 0 || result.DayInSchedule > 9 {
				t.Errorf("Day out of range: %d", result.DayInSchedule)
			}

			if result.TimeRequested.Location().String() != "Europe/Paris" {
				t.Errorf("Wrong timezone: %s", result.TimeRequested.Location().String())
			}
		})
	}
}

func TestCalculateSchedule_TimezoneConversion(t *testing.T) {
	// Test that UTC time is properly converted to Paris time
	utc := time.Date(2024, time.September, 1, 22, 0, 0, 0, time.UTC) // 10 PM UTC

	result := CalculateSchedule(utc, MiaroTeam) // MiaroTeam (3) has effective epoch Aug 31

	// Verify timezone is Paris
	if result.TimeRequested.Location().String() != "Europe/Paris" {
		t.Errorf("Expected Paris timezone, got %s", result.TimeRequested.Location().String())
	}

	// In September, Paris is UTC+2, so 22:00 UTC = 00:00 Paris (next day)
	// This means it should be September 2nd in Paris time
	expectedHour := 0
	if result.TimeRequested.Hour() != expectedHour {
		t.Errorf("Expected hour %d in Paris time, got %d", expectedHour, result.TimeRequested.Hour())
	}

	// Day 2 should be AFTERNOON
	if result.DayInSchedule != 2 {
		t.Errorf("Expected day 2, got %d", result.DayInSchedule)
	}
}

func TestCalculateSchedule_TeamOffset(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")
	// On Sep 2, 2024 (base epoch Aug 27):
	// Team 1 (epoch Aug 27): diffDays=6, 6%10=6 -> FREE
	// Team 2 (epoch Aug 29): diffDays=4, 4%10=4 -> NIGHT
	// Team 3 (epoch Aug 31): diffDays=2, 2%10=2 -> AFTERNOON
	date := time.Date(2024, time.September, 2, 12, 0, 0, 0, loc)

	tests := []struct {
		name         string
		team         int
		expectedDay  int
		expectedType ScheduleType
	}{
		{"team 1 on Sep 2", 1, 6, FREE},
		{"team 2 on Sep 2", 2, 4, NIGHT},
		{"team 3 on Sep 2", 3, 2, AFTERNOON},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateSchedule(date, tc.team)
			if result.DayInSchedule != tc.expectedDay {
				t.Errorf("expected day %d, got %d", tc.expectedDay, result.DayInSchedule)
			}
			if result.ScheduleType != tc.expectedType {
				t.Errorf("expected type %s, got %s", tc.expectedType, result.ScheduleType)
			}
			if result.Team != tc.team {
				t.Errorf("expected team %d, got %d", tc.team, result.Team)
			}
		})
	}
}

func TestCalculateSchedule_InvalidTeamDefaultsToMiaroTeam(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")
	date := time.Date(2024, time.September, 2, 12, 0, 0, 0, loc)

	result := CalculateSchedule(date, 0)
	expected := CalculateSchedule(date, MiaroTeam)

	if result.DayInSchedule != expected.DayInSchedule {
		t.Errorf("invalid team should default to MiaroTeam (%d)", MiaroTeam)
	}
}
