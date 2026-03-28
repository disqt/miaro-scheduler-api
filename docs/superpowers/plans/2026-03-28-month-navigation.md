# Month Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add month navigation to the calendar so users can browse past and future months, with full-week grids showing adjacent-month days.

**Architecture:** Add `?month=YYYY-MM` query param to the HTML handler. Refactor `GenerateCalendarData` to accept a target month and fill leading/trailing days from adjacent months. Add prev/next navigation arrows to the calendar header in all 3 themes.

**Tech Stack:** Go/Gin, HTML/CSS templates, Playwright for browser testing

**Spec:** `docs/superpowers/specs/2026-03-28-month-navigation-design.md`

---

### Task 1: Update CalendarDay struct and GenerateCalendarData

**Files:**
- Modify: `pkg/template.go:17-25` (CalendarDay struct)
- Modify: `pkg/template.go:8-15` (ScheduleBeautified struct)
- Modify: `pkg/template.go:104-163` (GenerateCalendarData function)
- Modify: `pkg/template.go:30-46` (FormatScheduleBeautified function)
- Test: `pkg/template_test.go`

- [ ] **Step 1: Write failing test for GenerateCalendarData with full weeks**

Add to `pkg/template_test.go`:

```go
func TestGenerateCalendarData_FullWeeks(t *testing.T) {
	// March 2026: starts on Sunday, ends on Tuesday
	// Week 1 should have Mon Feb 23 - Sat Feb 28 as other-month days
	// Last week should have Wed Apr 1 - Sun Apr 5 as other-month days
	now := time.Date(2026, time.March, 15, 12, 0, 0, 0, parisLoc)
	targetMonth := time.Date(2026, time.March, 1, 0, 0, 0, 0, parisLoc)
	schedule := CalculateSchedule(now, MiaroTeam)

	days := GenerateCalendarData(schedule, targetMonth, 1)

	// Grid should be 5 weeks * 7 = 35 days
	assert.Equal(t, 35, len(days))

	// First day should be Monday Feb 23 (other-month)
	assert.Equal(t, 23, days[0].DayNumber)
	assert.Equal(t, true, days[0].IsOtherMonth)
	assert.NotEqual(t, "", days[0].ShiftClass)

	// 7th day should be Sunday March 1 (this month)
	assert.Equal(t, 1, days[6].DayNumber)
	assert.Equal(t, false, days[6].IsOtherMonth)

	// Day 15 should be marked as today
	// Feb has 6 leading days, so March 15 is at index 6+14 = 20
	assert.Equal(t, 15, days[20].DayNumber)
	assert.Equal(t, true, days[20].IsToday)

	// Last day should be Sunday Apr 5 (other-month)
	assert.Equal(t, 5, days[34].DayNumber)
	assert.Equal(t, true, days[34].IsOtherMonth)
}

func TestGenerateCalendarData_OtherMonthNoToday(t *testing.T) {
	// Viewing April 2026 while today is March 15 -- no day should be IsToday
	now := time.Date(2026, time.March, 15, 12, 0, 0, 0, parisLoc)
	targetMonth := time.Date(2026, time.April, 1, 0, 0, 0, 0, parisLoc)
	schedule := CalculateSchedule(now, MiaroTeam)

	days := GenerateCalendarData(schedule, targetMonth, 1)

	for _, day := range days {
		if day.IsToday {
			t.Errorf("No day should be IsToday when viewing a different month, but day %d is", day.DayNumber)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestGenerateCalendarData_FullWeeks ./pkg/`
Expected: FAIL -- wrong number of arguments to `GenerateCalendarData`

- [ ] **Step 3: Update CalendarDay struct**

In `pkg/template.go`, replace the `CalendarDay` struct:

```go
// CalendarDay represents a single day in the calendar view
type CalendarDay struct {
	DayNumber    int    // Day of the month (1-31)
	ShiftType    string // "Matin", "Après-midi", "Nuit", "Libre"
	IsToday      bool   // True if this is today
	IsNextWork   bool   // True if this is the next working day
	IsOtherMonth bool   // True for days from adjacent months
	ShiftClass   string // CSS class: "morning", "afternoon", "night", "free"
}
```

- [ ] **Step 4: Update ScheduleBeautified struct**

In `pkg/template.go`, replace the `ScheduleBeautified` struct:

```go
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
```

- [ ] **Step 5: Add French month names helper**

In `pkg/template.go`, add after the struct definitions:

```go
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
```

- [ ] **Step 6: Rewrite GenerateCalendarData**

Replace `GenerateCalendarData` in `pkg/template.go`:

```go
// GenerateCalendarData generates calendar data for the target month with full weeks.
// Leading/trailing days from adjacent months are included with IsOtherMonth=true.
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
```

- [ ] **Step 7: Update FormatScheduleBeautified**

Replace `FormatScheduleBeautified` in `pkg/template.go`:

```go
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
```

- [ ] **Step 8: Fix any references to IsEmpty in template.go**

Search for `IsEmpty` in `pkg/template.go` and remove any remaining references. The old empty-cell loop in `GenerateCalendarData` is already gone from the rewrite above.

- [ ] **Step 9: Run tests to verify they pass**

Run: `go test -v -run TestGenerateCalendarData ./pkg/`
Expected: PASS for both `TestGenerateCalendarData_FullWeeks` and `TestGenerateCalendarData_OtherMonthNoToday`

- [ ] **Step 10: Commit**

```bash
git add pkg/template.go pkg/template_test.go
git commit -m "feat: add full-week calendar grid with month navigation data model"
```

---

### Task 2: Update handler to parse ?month= and pass new data

**Files:**
- Modify: `main.go:62-75` (renderSchedule function)
- Test: `main_test.go`

- [ ] **Step 1: Write failing tests for month param**

Add to `main_test.go`:

```go
func TestSchedulerHandler_MonthParam(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?month=2026-04", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Avril 2026") {
		t.Error("Expected response to contain 'Avril 2026' for month=2026-04")
	}
}

func TestSchedulerHandler_InvalidMonthDefaultsToCurrent(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?month=invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSchedulerHandler_MonthAndTeamParams(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?month=2026-05&team=3", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Mai 2026") {
		t.Error("Expected response to contain 'Mai 2026'")
	}

	// Check cookie was set for team
	cookies := w.Result().Cookies()
	var teamCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "miaro-team" {
			teamCookie = c
		}
	}
	if teamCookie == nil {
		t.Fatal("Expected miaro-team cookie to be set")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v -run TestSchedulerHandler_MonthParam -run TestSchedulerHandler_InvalidMonth -run TestSchedulerHandler_MonthAndTeam .`
Expected: FAIL -- compilation errors since `FormatScheduleBeautified` signature changed

- [ ] **Step 3: Add month parsing helper to main.go**

Add to `main.go` near the other parse helpers:

```go
func parseMonthParam(c *gin.Context) time.Time {
	monthStr := c.Query("month")
	if monthStr != "" {
		t, err := time.Parse("2006-01", monthStr)
		if err == nil {
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		}
	}
	now := time.Now().In(pkg.ParisLoc())
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}
```

- [ ] **Step 4: Expose parisLoc from pkg**

Add to `pkg/schedulerService.go`:

```go
// ParisLoc returns the Europe/Paris timezone location.
func ParisLoc() *time.Location {
	return parisLoc
}
```

- [ ] **Step 5: Update renderSchedule to pass month and new template data**

Replace `renderSchedule` in `main.go`:

```go
func renderSchedule(c *gin.Context, team int) {
	schedule := pkg.CalculateSchedule(time.Now(), team)
	targetMonth := parseMonthParam(c)
	scheduleBeautified := pkg.FormatScheduleBeautified(schedule, targetMonth)

	// Build month nav query string preserving team param
	monthQuery := ""
	if team > 1 {
		monthQuery = fmt.Sprintf("&team=%d", team)
	}

	c.HTML(http.StatusOK, "miaroSchedule.tmpl", gin.H{
		"Schedule":               scheduleBeautified.Schedule,
		"IsWorking":              scheduleBeautified.IsWorking,
		"NextWorkingDay":         scheduleBeautified.NextWorkingDay,
		"ScheduleNextWorkingDay": scheduleBeautified.ScheduleNextWorkingDay,
		"CalendarDays":           scheduleBeautified.CalendarDays,
		"CalendarMonthLabel":     scheduleBeautified.CalendarMonthLabel,
		"PrevMonth":              scheduleBeautified.PrevMonth,
		"NextMonth":              scheduleBeautified.NextMonth,
		"IsCurrentMonth":         scheduleBeautified.IsCurrentMonth,
		"MonthQuery":             monthQuery,
		"Team":                   team,
		"Teams":                  pkg.BuildTeamList(team),
	})
}
```

- [ ] **Step 6: Run all tests**

Run: `go test -v ./... `
Expected: PASS -- all existing and new tests pass

- [ ] **Step 7: Commit**

```bash
git add main.go main_test.go pkg/schedulerService.go
git commit -m "feat: parse ?month= param and pass calendar nav data to template"
```

---

### Task 3: Update template -- navigation header and other-month styling

**Files:**
- Modify: `templates/miaroSchedule.tmpl` (CSS + HTML for all 3 themes)

This task uses the **frontend-design** skill for UI implementation.

- [ ] **Step 1: Add other-month CSS rules for neumorphism theme**

After the `.calendar-day.free` rule (around line 105), add:

```css
body.theme-neumorphism .calendar-day.other-month { opacity: 0.35; pointer-events: none; }
body.theme-neumorphism .calendar-day.other-month.today,
body.theme-neumorphism .calendar-day.other-month.next-work { opacity: 0.35; }
```

Replace the `.calendar-day.empty` rule:
```css
/* Remove: body.theme-neumorphism .calendar-day.empty { box-shadow: none; background: transparent; } */
```

Add nav arrow styles after the `.calendar-header` styles:

```css
body.theme-neumorphism .calendar-nav { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
body.theme-neumorphism .calendar-nav h2 { font-size: 1.125rem; font-weight: 700; }
body.theme-neumorphism .calendar-nav a {
    width: 36px; height: 36px; display: flex; align-items: center; justify-content: center;
    border-radius: 50%; color: var(--text); text-decoration: none; font-size: 1.125rem;
    background: var(--bg);
    box-shadow: 3px 3px 6px var(--shadow-dark), -3px -3px 6px var(--shadow-light);
    transition: all 0.2s;
}
body.theme-neumorphism .calendar-nav a:hover {
    box-shadow: 1px 1px 3px var(--shadow-dark), -1px -1px 3px var(--shadow-light);
}
```

- [ ] **Step 2: Add other-month CSS rules for bento theme**

After the `.calendar-day.free` rule in bento section, add:

```css
body.theme-bento .calendar-day.other-month { opacity: 0.35; pointer-events: none; }
```

Remove the `.calendar-day.empty` rule for bento.

Add nav styles:

```css
body.theme-bento .calendar-nav { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
body.theme-bento .calendar-nav h2 { font-size: 1rem; font-weight: 600; }
body.theme-bento .calendar-nav a {
    width: 32px; height: 32px; display: flex; align-items: center; justify-content: center;
    border-radius: 10px; color: var(--text); text-decoration: none; font-size: 1rem;
    background: var(--bg); border: 1px solid var(--border); transition: all 0.15s;
}
body.theme-bento .calendar-nav a:hover { background: #e2e8f0; }
```

- [ ] **Step 3: Add other-month CSS rules for terminal theme**

After the `.calendar-day.free` rule in terminal section, add:

```css
body.theme-terminal .calendar-day.other-month { opacity: 0.3; pointer-events: none; }
```

Remove the `.calendar-day.empty` rule for terminal.

Add nav styles:

```css
body.theme-terminal .calendar-nav { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; }
body.theme-terminal .calendar-nav span { color: var(--cyan); font-weight: 600; }
body.theme-terminal .calendar-nav a { color: var(--green); text-decoration: none; font-size: 0.875rem; }
body.theme-terminal .calendar-nav a:hover { text-decoration: underline; }
```

- [ ] **Step 4: Update neumorphism calendar header HTML**

Replace the neumorphism calendar header (around line 590-593):

```html
<div class="calendar-nav">
    <a href="/miaro?month={{ .PrevMonth }}{{ .MonthQuery }}">&larr;</a>
    <h2>{{ .CalendarMonthLabel }}</h2>
    <a href="/miaro?month={{ .NextMonth }}{{ .MonthQuery }}">&rarr;</a>
</div>
```

- [ ] **Step 5: Update neumorphism calendar day classes**

Replace the calendar day div class in neumorphism (around line 595):

```html
<div class="calendar-day {{ .ShiftClass }} {{ if .IsOtherMonth }}other-month{{ end }} {{ if .IsToday }}today{{ end }} {{ if .IsNextWork }}next-work{{ end }}">
    {{ if not .IsOtherMonth }}
    <div class="day-number">{{ .DayNumber }}</div>
    <div class="shift-indicator">{{ .ShiftType }}</div>
    {{ else }}
    <div class="day-number">{{ .DayNumber }}</div>
    {{ end }}
</div>
```

- [ ] **Step 6: Update bento calendar header HTML**

Replace the bento calendar title section (around line 680-683):

```html
<div class="calendar-nav">
    <a href="/miaro?month={{ .PrevMonth }}{{ .MonthQuery }}">&larr;</a>
    <h2>{{ .CalendarMonthLabel }}</h2>
    <a href="/miaro?month={{ .NextMonth }}{{ .MonthQuery }}">&rarr;</a>
</div>
```

- [ ] **Step 7: Update bento calendar day classes**

Replace the calendar day div class in bento (around line 690):

```html
<div class="calendar-day {{ .ShiftClass }} {{ if .IsOtherMonth }}other-month{{ end }} {{ if .IsToday }}today{{ end }} {{ if .IsNextWork }}next-work{{ end }}">
    {{ if not .IsOtherMonth }}
    <div class="day-number">{{ .DayNumber }}</div>
    <div class="shift-indicator">{{ .ShiftType }}</div>
    {{ else }}
    <div class="day-number">{{ .DayNumber }}</div>
    {{ end }}
</div>
```

- [ ] **Step 8: Update terminal calendar header HTML**

Replace the terminal section title (around line 785):

```html
<div class="calendar-nav">
    <a href="/miaro?month={{ .PrevMonth }}{{ .MonthQuery }}">[prev]</a>
    <span>$ cal --schedule {{ .CalendarMonthLabel }}</span>
    <a href="/miaro?month={{ .NextMonth }}{{ .MonthQuery }}">[next]</a>
</div>
```

- [ ] **Step 9: Update terminal calendar day classes**

Replace the calendar day div class in terminal (around line 795):

```html
<div class="calendar-day {{ .ShiftClass }} {{ if .IsOtherMonth }}other-month{{ end }} {{ if .IsToday }}today{{ end }} {{ if .IsNextWork }}next-work{{ end }}">
    {{ if not .IsOtherMonth }}
    <div class="day-number">{{ .DayNumber }}</div>
    <div class="shift-indicator">{{ .ShiftType }}</div>
    {{ else }}
    <div class="day-number">{{ .DayNumber }}</div>
    {{ end }}
</div>
```

- [ ] **Step 10: Remove all IsEmpty references from template**

Search the template for `IsEmpty` and `empty` class references. Remove:
- All `{{ if .IsEmpty }}empty{{ end }}` from calendar day classes
- The `{{ if not .IsEmpty }}` conditionals (replaced by `{{ if not .IsOtherMonth }}` above)
- CSS rules for `.calendar-day.empty` in all three themes

- [ ] **Step 11: Update selectTeam JavaScript to preserve month param**

In the JavaScript section (around line 830), update `selectTeam`:

```javascript
function selectTeam(teamNum) {
    const url = new URL(window.location);
    if (teamNum === '1' || teamNum === 1) {
        url.searchParams.delete('team');
    } else {
        url.searchParams.set('team', teamNum);
    }
    window.location.href = url.pathname + '?' + url.searchParams.toString();
}
```

- [ ] **Step 12: Run build and all tests**

Run: `go build -o /dev/null . && go test -v ./...`
Expected: PASS -- build succeeds, all tests pass

- [ ] **Step 13: Commit**

```bash
git add templates/miaroSchedule.tmpl
git commit -m "feat: add month navigation UI with full-week grid and other-month styling"
```

---

### Task 4: Update existing tests for new signatures

**Files:**
- Modify: `main_test.go` (fix any broken assertions)
- Modify: `pkg/template_test.go` (fix any broken assertions)

- [ ] **Step 1: Check for broken tests from IsEmpty removal**

Run: `go test -v ./...`

If there are failures related to `IsEmpty`, update test assertions to use `IsOtherMonth` instead.

- [ ] **Step 2: Update TestSchedulerHandler expected strings**

In `main_test.go`, the test `TestSchedulerHandler` checks for `"Planning du mois"`. This header text is now replaced by the dynamic month label. Update:

```go
expectedStrings := []string{
    "<!DOCTYPE html>",
    "Horaire de Miaro",
    "Statut",
}
```

The month label is dynamic so don't assert on a fixed string.

- [ ] **Step 3: Run all tests**

Run: `go test -v ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add main_test.go pkg/template_test.go
git commit -m "test: update tests for month navigation changes"
```

---

### Task 5: Visual testing with Playwright

**Files:**
- Test files in `/tmp/` (Playwright convention)

This task uses the **playwright-skill** for browser testing across all 3 themes.

- [ ] **Step 1: Start the dev server**

```bash
PORT=8099 go run . &
```

- [ ] **Step 2: Test neumorphism theme -- current month**

Navigate to `http://localhost:8099/miaro`. Verify:
- Calendar shows current month label (e.g., "Mars 2026")
- Prev/next arrows are visible
- Today is highlighted
- Other-month days are greyed out with day numbers
- Shift colors are visible on all days

Take a screenshot.

- [ ] **Step 3: Test month navigation -- next month**

Click the next arrow (or navigate to `?month=2026-04`). Verify:
- Calendar shows "Avril 2026"
- No day is highlighted as today
- Other-month days fill the first and last rows
- Shift colors are visible

Take a screenshot.

- [ ] **Step 4: Test month navigation -- previous month**

Click the prev arrow. Verify it returns to the current month. Click prev again for February. Verify:
- Calendar shows "Février 2026"
- Correct number of days (28)
- Full weeks rendered

- [ ] **Step 5: Test with team param preserved**

Navigate to `http://localhost:8099/miaro?team=3&month=2026-05`. Verify:
- Calendar shows "Mai 2026"
- Team 3 is selected in dropdown
- Clicking prev/next preserves team=3 in the URL

- [ ] **Step 6: Test bento theme**

Switch to Bento theme. Verify:
- Navigation arrows styled correctly
- Other-month days greyed out
- Shift colors visible

Take a screenshot.

- [ ] **Step 7: Test terminal theme**

Switch to Terminal theme. Verify:
- `[prev]` and `[next]` links styled correctly
- Header shows `$ cal --schedule Mars 2026`
- Other-month days greyed out

Take a screenshot.

- [ ] **Step 8: Test mobile viewport**

Resize to 375px width. Verify:
- Calendar grid still fits
- Navigation arrows accessible
- No horizontal scrolling

- [ ] **Step 9: Stop dev server**

```bash
kill %1 2>/dev/null
```

---

### Task 6: Final cleanup and commit

- [ ] **Step 1: Run full test suite**

```bash
go test -v -race ./...
```

Expected: PASS with no race conditions

- [ ] **Step 2: Format code**

```bash
go fmt ./...
```

- [ ] **Step 3: Verify build**

```bash
go build -o /dev/null .
```

- [ ] **Step 4: Review changes**

```bash
git diff main --stat
```

Verify only expected files changed.
