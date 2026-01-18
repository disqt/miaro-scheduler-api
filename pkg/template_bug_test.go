package pkg

import (
	"testing"
)

// TestNextWorkingDayAccessesInvalidMapKeys demonstrates that the nextWorkingDay
// function accesses map keys that don't exist (keys 10, 11, etc.) when checking
// if days are FREE. This works by accident because the zero value for ScheduleType
// is "" which doesn't equal "FREE", but it's semantically incorrect for a circular
// schedule.
func TestNextWorkingDayAccessesInvalidMapKeys(t *testing.T) {
	// This test demonstrates the issue by showing that when we're on day 5
	// and looking for the next working day, the loop checks schedule[6],
	// schedule[7], schedule[8], schedule[9], and then schedule[10].
	//
	// schedule[10] doesn't exist in the map, so it returns the zero value "".
	// Since "" != "FREE", the loop exits, but this is semantically wrong.
	// The loop should wrap around using modulo and check schedule[0] instead.

	// Starting from day 5 (second night shift)
	// Days 6,7,8,9 are all FREE
	// Next working day should be day 0 (MORNING), which is 5 days later

	scheduleType, nextDay := nextWorkingDay(5)

	// The result is actually correct
	if scheduleType != "du matin" || nextDay != "dans 5 jours" {
		t.Errorf("Expected 'du matin' and 'dans 5 jours', got '%s' and '%s'",
			scheduleType, nextDay)
	}

	// However, the logic is flawed because the loop checked schedule[10]
	// instead of wrapping around to schedule[0].
	// We can demonstrate this by verifying what the schedule map actually contains:

	// This test passes but demonstrates the fragility:
	// If we access a key that doesn't exist, we get the zero value
	value := schedule[10]
	if value != "" {
		t.Errorf("Expected zero value (empty string) for schedule[10], got '%s'", value)
	}

	// The loop relies on this zero value not equaling FREE
	if value == FREE {
		t.Error("BUG EXPOSED: If the zero value somehow equals FREE, the loop would be incorrect")
	}

	t.Log("This test passes, but the implementation accesses map keys 10+ which don't exist.")
	t.Log("The code works by accident because zero value '' != 'FREE'")
}

// TestNextWorkingDayWithModifiedScheduleType demonstrates what would happen
// if we had a different scenario. This test creates a scenario that shows
// the loop's behavior more clearly.
func TestNextWorkingDayLoopIterations(t *testing.T) {
	// When starting from day 5, the loop will iterate:
	// i=1: check schedule[6] = FREE, continue
	// i=2: check schedule[7] = FREE, continue
	// i=3: check schedule[8] = FREE, continue
	// i=4: check schedule[9] = FREE, continue
	// i=5: check schedule[10] = "" (OUT OF BOUNDS!), "" != FREE, exit

	// The correct implementation should be:
	// i=5: check schedule[(5+5)%10] = schedule[0] = MORNING, exit

	// Let's verify the schedule values that the loop checks:
	testCases := []struct {
		key   int
		value ScheduleType
		note  string
	}{
		{6, FREE, "Valid key, FREE"},
		{7, FREE, "Valid key, FREE"},
		{8, FREE, "Valid key, FREE"},
		{9, FREE, "Valid key, FREE"},
		{10, "", "INVALID KEY - returns zero value"},
	}

	for _, tc := range testCases {
		actual := schedule[tc.key]
		if actual != tc.value {
			t.Errorf("schedule[%d] = '%s', expected '%s' (%s)",
				tc.key, actual, tc.value, tc.note)
		}
		if tc.key >= 10 {
			t.Logf("WARNING: Accessing schedule[%d] which is outside the valid range [0-9]", tc.key)
		}
	}

	// Despite accessing invalid keys, the function returns the correct result
	scheduleType, nextDay := nextWorkingDay(5)
	if scheduleType != "du matin" || nextDay != "dans 5 jours" {
		t.Errorf("Function still works but for the wrong reasons")
	}
}

// TestNextWorkingDayBoundaryCase tests the edge case at day 9
func TestNextWorkingDayAtDay9(t *testing.T) {
	// Day 9 is FREE
	// Next working day is day 0 (MORNING), which is 1 day later
	// The loop will check:
	// i=1: schedule[10] = "" (OUT OF BOUNDS!), "" != FREE, exit immediately

	scheduleType, nextDay := nextWorkingDay(9)

	// The result is correct
	if scheduleType != "du matin" || nextDay != "demain" {
		t.Errorf("Expected 'du matin' and 'demain', got '%s' and '%s'",
			scheduleType, nextDay)
	}

	// But again, we accessed schedule[10] which doesn't exist
	t.Log("The loop checked schedule[10] instead of schedule[0]")
	t.Log("This works only because schedule[10] returns '' which != 'FREE'")
}

// TestNextWorkingDayWithStrictMapAccess would fail if we made map access stricter.
// This test demonstrates what SHOULD be checked vs what IS checked.
func TestNextWorkingDayWithStrictMapAccess(t *testing.T) {
	// If we were to make the code more strict and panic on out-of-bounds access,
	// the current implementation would fail.

	// Starting from day 5, the loop checks these keys:
	actualKeysChecked := []int{6, 7, 8, 9, 10} // 10 is OUT OF BOUNDS!

	// With proper modulo, it should check these keys:
	expectedKeysChecked := []int{6, 7, 8, 9, 0} // 0 is the wrapped-around day

	// The actual implementation checks key 10 instead of key 0
	if actualKeysChecked[4] == 10 && expectedKeysChecked[4] == 0 {
		t.Log("BUG CONFIRMED: Loop checks schedule[10] instead of schedule[0]")
		t.Log("Current code: for schedule[day+i] == FREE")
		t.Log("Should be:    for schedule[(day+i)%10] == FREE")
	}

	// If we had a way to track map accesses, we'd see that schedule[10] is accessed
	// but it shouldn't be - we should wrap around to schedule[0] instead.
}

// TestNextWorkingDayIfZeroValueWereFREE demonstrates what would happen
// if the zero value for ScheduleType was somehow "FREE" instead of "".
// This test is hypothetical but shows the fragility.
func TestNextWorkingDayIfZeroValueWereFREE(t *testing.T) {
	// In the current implementation, when we access schedule[10], we get ""
	// The loop continues only if schedule[day+i] == FREE
	// So the loop stops because "" != "FREE"

	// But imagine if ScheduleType had a different zero value, or if FREE was ""
	// Then schedule[10] would equal FREE, and the loop would continue:
	// i=5: schedule[10] = FREE, continue
	// i=6: schedule[11] = FREE, continue
	// ... infinite loop!

	// This demonstrates why the current implementation is fragile.
	// The correct implementation should use modulo to stay within bounds:
	// for schedule[(day+i)%10] == FREE

	t.Log("Current implementation relies on zero value '' not equaling 'FREE'")
	t.Log("If zero value changed to match FREE, we'd have an infinite loop")
	t.Log("Using modulo in the loop would prevent this fragility")
}
