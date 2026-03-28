# Team Selector Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a team selector dropdown allowing users to view schedules for 5 teams sharing the same 10-day rotation, each offset by 2 days.

**Architecture:** Add a `team` parameter to `CalculateSchedule()` that offsets the epoch date by `(team-1)*2` days. Handler resolves team from URL param > cookie > default(1), with redirect logic to keep URLs canonical. Template gets a dropdown that navigates between teams.

**Tech Stack:** Go 1.24, Gin, html/template, cookies via `http.SetCookie`

**Spec:** `docs/superpowers/specs/2026-03-28-team-selector-design.md`

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `pkg/team.go` | Create | Team constants, labels, validation, TeamInfo struct |
| `pkg/team_test.go` | Create | Tests for team validation and labels |
| `pkg/schedulerService.go` | Modify | Add team param to `CalculateSchedule()`, offset epoch |
| `pkg/schedulerService_test.go` | Modify | Update all calls, add team offset tests |
| `pkg/template.go` | Modify | Thread `schedule.Team` into `GenerateCalendarData` -> `CalculateSchedule` |
| `pkg/template_test.go` | Modify | Update `CalculateSchedule` calls to include team |
| `pkg/template_bug_test.go` | Modify | Update `CalculateSchedule` calls to include team |
| `main.go` | Modify | Team resolution, cookies, redirects, template data |
| `main_test.go` | Modify | Update existing tests, add team handler tests |
| `templates/miaroSchedule.tmpl` | Modify | Dropdown in all 3 themes, cookie migration JS |

---

### Task 1: Team types and validation (`pkg/team.go`)

**Files:**
- Create: `pkg/team.go`
- Create: `pkg/team_test.go`

- [ ] **Step 1: Write failing tests for team validation**

In `pkg/team_test.go`:

```go
package pkg

import (
	"testing"
)

func TestValidateTeam(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"valid team 1", 1, 1},
		{"valid team 5", 5, 5},
		{"zero defaults to 1", 0, 1},
		{"negative defaults to 1", -1, 1},
		{"too high defaults to 1", 6, 1},
		{"way too high defaults to 1", 100, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateTeam(tc.input)
			if result != tc.expected {
				t.Errorf("ValidateTeam(%d) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestTeamLabel(t *testing.T) {
	tests := []struct {
		team     int
		expected string
	}{
		{1, "Équipe 1 (Miaro)"},
		{2, "Équipe 2"},
		{3, "Équipe 3"},
		{4, "Équipe 4"},
		{5, "Équipe 5"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := TeamLabel(tc.team)
			if result != tc.expected {
				t.Errorf("TeamLabel(%d) = %q, want %q", tc.team, result, tc.expected)
			}
		})
	}
}

func TestBuildTeamList(t *testing.T) {
	teams := BuildTeamList(3)
	if len(teams) != TeamCount {
		t.Fatalf("expected %d teams, got %d", TeamCount, len(teams))
	}
	if !teams[2].Selected {
		t.Error("team 3 should be selected")
	}
	if teams[0].Selected {
		t.Error("team 1 should not be selected")
	}
	if teams[0].Label != "Équipe 1 (Miaro)" {
		t.Errorf("team 1 label = %q, want 'Équipe 1 (Miaro)'", teams[0].Label)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v -run "TestValidateTeam|TestTeamLabel|TestBuildTeamList" ./pkg`
Expected: compilation error — `ValidateTeam`, `TeamLabel`, `BuildTeamList` not defined

- [ ] **Step 3: Implement team.go**

In `pkg/team.go`:

```go
package pkg

import "fmt"

const (
	MiaroTeam = 1
	TeamCount = 5
)

// TeamInfo holds display data for the team selector dropdown.
type TeamInfo struct {
	Number   int
	Label    string
	Selected bool
}

// ValidateTeam returns team if it's in [1, TeamCount], otherwise returns MiaroTeam.
func ValidateTeam(team int) int {
	if team < 1 || team > TeamCount {
		return MiaroTeam
	}
	return team
}

// TeamLabel returns the display label for a team number.
func TeamLabel(team int) string {
	if team == MiaroTeam {
		return fmt.Sprintf("Équipe %d (Miaro)", team)
	}
	return fmt.Sprintf("Équipe %d", team)
}

// BuildTeamList returns a slice of TeamInfo for rendering the dropdown.
func BuildTeamList(selectedTeam int) []TeamInfo {
	teams := make([]TeamInfo, TeamCount)
	for i := 0; i < TeamCount; i++ {
		num := i + 1
		teams[i] = TeamInfo{
			Number:   num,
			Label:    TeamLabel(num),
			Selected: num == selectedTeam,
		}
	}
	return teams
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v -run "TestValidateTeam|TestTeamLabel|TestBuildTeamList" ./pkg`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/team.go pkg/team_test.go
git commit -m "feat: add team types, validation, and label helpers"
```

---

### Task 2: Add team parameter to `CalculateSchedule()`

**Files:**
- Modify: `pkg/schedulerService.go`
- Modify: `pkg/schedulerService_test.go`

- [ ] **Step 1: Write failing tests for team offset**

Add to `pkg/schedulerService_test.go`:

```go
func TestCalculateSchedule_TeamOffset(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")

	// On Sep 2, 2024:
	// Team 1 (epoch Aug 31): diffDays=2, 2%10=2 -> AFTERNOON
	// Team 2 (epoch Sep 2):  diffDays=0, 0%10=0 -> MORNING
	date := time.Date(2024, time.September, 2, 12, 0, 0, 0, loc)

	tests := []struct {
		name         string
		team         int
		expectedDay  int
		expectedType ScheduleType
	}{
		{"team 1 on Sep 2", 1, 2, AFTERNOON},
		{"team 2 on Sep 2", 2, 0, MORNING},
		{"team 3 on Sep 2", 3, 8, FREE}, // epoch Sep 4, (-2%10+10)%10 = 8 -> FREE
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

func TestCalculateSchedule_InvalidTeamDefaultsTo1(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")
	date := time.Date(2024, time.September, 2, 12, 0, 0, 0, loc)

	result := CalculateSchedule(date, 0)
	expected := CalculateSchedule(date, 1)

	if result.DayInSchedule != expected.DayInSchedule {
		t.Errorf("invalid team should default to team 1")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v -run "TestCalculateSchedule_TeamOffset|TestCalculateSchedule_InvalidTeam" ./pkg`
Expected: compilation error — `CalculateSchedule` doesn't accept team param, `Schedule` has no `Team` field

- [ ] **Step 3: Update CalculateSchedule signature and implementation**

Replace the entire `CalculateSchedule` function and `Schedule` struct in `pkg/schedulerService.go`:

Add `Team` field to the struct:
```go
type Schedule struct {
	TimeRequested time.Time    `json:"time_requested"`
	ScheduleType  ScheduleType `json:"schedule_type"`
	DayInSchedule int          `json:"day_in_schedule"`
	Team          int          `json:"team"`
}
```

New `CalculateSchedule` signature — change from variadic `time.Time` to `(date time.Time, team int)`:

```go
func CalculateSchedule(date time.Time, team int) Schedule {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		panic("Failed to load Europe/Paris timezone: " + err.Error())
	}

	date = date.In(loc)
	team = ValidateTeam(team)

	// Offset epoch forward by (team-1)*2 days
	initialDate := time.Date(2024, time.August, 31, 0, 0, 0, 0, loc)
	initialDate = initialDate.AddDate(0, 0, (team-1)*2)

	diffDays := int(date.Sub(initialDate).Hours() / 24)

	// Positive modulo to handle dates before a team's epoch
	dayInSchedule := ((diffDays % 10) + 10) % 10

	return Schedule{
		TimeRequested: date,
		ScheduleType:  schedule[dayInSchedule],
		DayInSchedule: dayInSchedule,
		Team:          team,
	}
}
```

- [ ] **Step 4: Fix all existing callers and tests**

The signature change from variadic to `(time.Time, int)` will break all existing callers. Update each:

In `pkg/schedulerService_test.go`, update every `CalculateSchedule(...)` call:
- `CalculateSchedule()` -> `CalculateSchedule(time.Now(), MiaroTeam)`
- `CalculateSchedule(tc.date)` -> `CalculateSchedule(tc.date, MiaroTeam)`
- `CalculateSchedule(utc)` -> `CalculateSchedule(utc, MiaroTeam)`
- `CalculateSchedule(date1, date2)` — delete `TestCalculateSchedule_PanicOnMultipleDates` entirely (no longer applicable)

In `pkg/template.go` line 140, update the call inside `GenerateCalendarData`:
- `CalculateSchedule(currentDate)` -> `CalculateSchedule(currentDate, currentSchedule.Team)`

In `pkg/template_test.go` line 235:
- `CalculateSchedule(utc)` -> `CalculateSchedule(utc, MiaroTeam)`

In `pkg/template_bug_test.go` — no direct calls to `CalculateSchedule`, only to `nextWorkingDay` which doesn't call it. No changes needed.

In `main.go` (temporary — Task 3 rewrites these handlers, but this keeps the build green):
- `pkg.CalculateSchedule()` -> `pkg.CalculateSchedule(time.Now(), pkg.MiaroTeam)` (both handlers, lines 28 and 44)

- [ ] **Step 5: Run all tests to verify everything passes**

Run: `go test -v ./...`
Expected: all PASS (including the new team offset tests)

- [ ] **Step 6: Commit**

```bash
git add pkg/schedulerService.go pkg/schedulerService_test.go pkg/template.go pkg/template_test.go main.go
git commit -m "feat: add team parameter to CalculateSchedule with epoch offset"
```

---

### Task 3: Handler team resolution, cookies, and redirects

**Files:**
- Modify: `main.go`
- Modify: `main_test.go`

- [ ] **Step 1: Write failing tests for team resolution and redirects**

Add to `main_test.go`:

```go
func TestSchedulerHandler_TeamFromQueryParam(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?team=3", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check cookie was set
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
	if teamCookie.Value != "3" {
		t.Errorf("Expected cookie value '3', got '%s'", teamCookie.Value)
	}

	// Check dropdown is in the HTML
	body := w.Body.String()
	if !contains(body, "Équipe 3") {
		t.Error("Expected response to contain 'Équipe 3'")
	}
}

func TestSchedulerHandler_Team1RedirectsToCleanURL(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?team=1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected 302 redirect, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/miaro" {
		t.Errorf("Expected redirect to /miaro, got %s", w.Header().Get("Location"))
	}
}

func TestSchedulerHandler_CookieRedirect(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro", nil)
	req.AddCookie(&http.Cookie{Name: "miaro-team", Value: "4"})
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected 302 redirect, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/miaro?team=4" {
		t.Errorf("Expected redirect to /miaro?team=4, got %s", w.Header().Get("Location"))
	}
}

func TestSchedulerHandler_CookieTeam1NoRedirect(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro", nil)
	req.AddCookie(&http.Cookie{Name: "miaro-team", Value: "1"})
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (no redirect for team 1 cookie), got %d", w.Code)
	}
}

func TestSchedulerHandler_InvalidTeamDefaultsTo1(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?team=99", nil)
	router.ServeHTTP(w, req)

	// Invalid team treated as absent, no cookie -> default team 1, render
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSchedulerHandler_InvalidTeamWithCookieRedirects(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?team=abc", nil)
	req.AddCookie(&http.Cookie{Name: "miaro-team", Value: "2"})
	router.ServeHTTP(w, req)

	// Invalid ?team= treated as absent, cookie says 2 -> redirect
	if w.Code != http.StatusFound {
		t.Errorf("Expected 302 redirect, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/miaro?team=2" {
		t.Errorf("Expected redirect to /miaro?team=2, got %s", w.Header().Get("Location"))
	}
}

func TestSchedulerJSONHandler_WithTeam(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro/json?team=2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	rawSchedule := response["raw_schedule"].(map[string]interface{})
	team := rawSchedule["team"].(float64)
	if int(team) != 2 {
		t.Errorf("Expected team 2 in raw_schedule, got %v", team)
	}
}

func TestSchedulerJSONHandler_DefaultTeam(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro/json", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	rawSchedule := response["raw_schedule"].(map[string]interface{})
	team := rawSchedule["team"].(float64)
	if int(team) != 1 {
		t.Errorf("Expected team 1 in raw_schedule, got %v", team)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v -run "TestSchedulerHandler_Team|TestSchedulerHandler_Cookie|TestSchedulerJSONHandler_WithTeam|TestSchedulerJSONHandler_DefaultTeam" .`
Expected: FAIL — handlers don't handle `?team=` param, no cookie logic, no redirects

- [ ] **Step 3: Implement handler team resolution**

Add a `resolveTeam` helper and a `setTeamCookie` helper in `main.go`. Then update `SchedulerHandler`:

```go
// parseTeamParam parses the ?team= query parameter. Returns (team, valid).
func parseTeamParam(c *gin.Context) (int, bool) {
	teamStr := c.Query("team")
	if teamStr == "" {
		return 0, false
	}
	team, err := strconv.Atoi(teamStr)
	if err != nil || team < 1 || team > pkg.TeamCount {
		return 0, false
	}
	return team, true
}

// readTeamCookie reads the miaro-team cookie. Returns (team, valid).
func readTeamCookie(c *gin.Context) (int, bool) {
	cookie, err := c.Cookie("miaro-team")
	if err != nil {
		return 0, false
	}
	team, err := strconv.Atoi(cookie)
	if err != nil || team < 1 || team > pkg.TeamCount {
		return 0, false
	}
	return team, true
}

func setTeamCookie(c *gin.Context, team int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "miaro-team",
		Value:    strconv.Itoa(team),
		Path:     "/miaro",
		MaxAge:   365 * 24 * 60 * 60,
		SameSite: http.SameSiteLaxMode,
	})
}
```

Update `SchedulerHandler`:

```go
func SchedulerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Resolve team: URL param > cookie > default
		teamParam, hasValidParam := parseTeamParam(c)
		teamCookie, hasValidCookie := readTeamCookie(c)

		// Canonicalize: ?team=1 -> redirect to /miaro
		if hasValidParam && teamParam == pkg.MiaroTeam {
			setTeamCookie(c, pkg.MiaroTeam)
			c.Redirect(http.StatusFound, "/miaro")
			return
		}

		if hasValidParam {
			// Valid ?team=N (2-5): render
			team := teamParam
			setTeamCookie(c, team)
			renderSchedule(c, team)
			return
		}

		// No valid param — check cookie
		if hasValidCookie && teamCookie != pkg.MiaroTeam {
			c.Redirect(http.StatusFound, fmt.Sprintf("/miaro?team=%d", teamCookie))
			return
		}

		// Default: team 1
		setTeamCookie(c, pkg.MiaroTeam)
		renderSchedule(c, pkg.MiaroTeam)
	}
}

func renderSchedule(c *gin.Context, team int) {
	schedule := pkg.CalculateSchedule(time.Now(), team)
	scheduleBeautified := pkg.FormatScheduleBeautified(schedule)

	c.HTML(http.StatusOK, "miaroSchedule.tmpl", gin.H{
		"Schedule":               scheduleBeautified.Schedule,
		"IsWorking":              scheduleBeautified.IsWorking,
		"NextWorkingDay":         scheduleBeautified.NextWorkingDay,
		"ScheduleNextWorkingDay": scheduleBeautified.ScheduleNextWorkingDay,
		"CalendarDays":           scheduleBeautified.CalendarDays,
		"Team":                   team,
		"Teams":                  pkg.BuildTeamList(team),
	})
}
```

Update `SchedulerJSONHandler`:

```go
func SchedulerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		team := pkg.MiaroTeam
		if t, valid := parseTeamParam(c); valid {
			team = t
		}

		schedule := pkg.CalculateSchedule(time.Now(), team)
		scheduleBeautified := pkg.FormatScheduleBeautified(schedule)

		c.JSON(http.StatusOK, gin.H{
			"schedule":                  scheduleBeautified.Schedule,
			"is_working":                scheduleBeautified.IsWorking,
			"next_working_day":          scheduleBeautified.NextWorkingDay,
			"schedule_next_working_day": scheduleBeautified.ScheduleNextWorkingDay,
			"raw_schedule":              schedule,
		})
	}
}
```

Add imports to `main.go`: `"fmt"`, `"strconv"`, `"net/http"` (already present).

- [ ] **Step 4: Run all tests**

Run: `go test -v ./...`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add main.go main_test.go
git commit -m "feat: add team resolution with URL params, cookies, and redirects"
```

---

### Task 4: Template dropdown and cookie migration

**Files:**
- Modify: `templates/miaroSchedule.tmpl`

- [ ] **Step 1: Add team selector dropdown CSS**

Add these styles before the `/* THEME SWITCHER */` section (around line 313). Add styles for each theme:

```css
/* ============================================
   TEAM SELECTOR (All themes)
   ============================================ */

/* Neumorphism team selector */
body.theme-neumorphism .team-selector {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    margin-top: 1rem;
}
body.theme-neumorphism .team-selector label {
    font-size: 0.875rem;
    color: var(--text-muted);
    font-weight: 600;
}
body.theme-neumorphism .team-selector select {
    appearance: none;
    -webkit-appearance: none;
    background: var(--bg);
    border: none;
    border-radius: 12px;
    padding: 0.625rem 2rem 0.625rem 1rem;
    font-family: inherit;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text);
    cursor: pointer;
    box-shadow: inset 4px 4px 8px var(--shadow-dark), inset -4px -4px 8px var(--shadow-light);
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23718096' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.75rem center;
}

/* Bento team selector */
body.theme-bento .team-selector {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.75rem;
}
body.theme-bento .team-selector label {
    font-size: 0.8125rem;
    color: var(--text-muted);
    font-weight: 500;
}
body.theme-bento .team-selector select {
    appearance: none;
    -webkit-appearance: none;
    background: white;
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 0.5rem 2rem 0.5rem 0.75rem;
    font-family: inherit;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--text);
    cursor: pointer;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%236b7280' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.75rem center;
}
body.theme-bento .team-selector select:focus {
    outline: none;
    border-color: var(--accent);
}

/* Terminal team selector */
body.theme-terminal .team-selector {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.75rem;
}
body.theme-terminal .team-selector label {
    font-size: 0.75rem;
    color: var(--text-muted);
}
body.theme-terminal .team-selector select {
    appearance: none;
    -webkit-appearance: none;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 0;
    padding: 0.375rem 1.75rem 0.375rem 0.5rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    color: var(--green);
    cursor: pointer;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 10 10'%3E%3Cpath fill='%233fb950' d='M5 7L1 3h8z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.5rem center;
}
body.theme-terminal .team-selector select:focus {
    outline: none;
    border-color: var(--cyan);
}
```

- [ ] **Step 2: Add dropdown HTML to each theme**

**Neumorphism theme** — add inside `<header class="neu">` after the subtitle `<p>`:

```html
<div class="team-selector">
    <label for="team-select-neu">Équipe :</label>
    <select id="team-select-neu" class="team-select" onchange="selectTeam(this.value)">
        {{ range .Teams }}
        <option value="{{ .Number }}" {{ if .Selected }}selected{{ end }}>{{ .Label }}</option>
        {{ end }}
    </select>
</div>
```

**Bento theme** — add inside `<div class="bento-item bento-title">` after the `<p>`:

```html
<div class="team-selector">
    <label for="team-select-bento">Équipe :</label>
    <select id="team-select-bento" class="team-select" onchange="selectTeam(this.value)">
        {{ range .Teams }}
        <option value="{{ .Number }}" {{ if .Selected }}selected{{ end }}>{{ .Label }}</option>
        {{ end }}
    </select>
</div>
```

**Terminal theme** — add after the first `<div class="output">` (the comment line `# Vérification du statut...`), as a new line block:

```html
<div class="line">
    <span class="prompt">miaro@schedule:~$</span> <span class="command">export TEAM=</span>
    <div class="team-selector" style="display: inline-flex; margin-left: 0;">
        <select id="team-select-term" class="team-select" onchange="selectTeam(this.value)">
            {{ range .Teams }}
            <option value="{{ .Number }}" {{ if .Selected }}selected{{ end }}>{{ .Label }}</option>
            {{ end }}
        </select>
    </div>
</div>
```

- [ ] **Step 3: Update JavaScript — add team navigation and cookie migration**

Replace the entire `<script>` block at the bottom of the template:

```html
<script>
    (function() {
        // ---- Cookie helpers ----
        function getCookie(name) {
            var match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'));
            return match ? decodeURIComponent(match[1]) : null;
        }
        function setCookie(name, value) {
            document.cookie = name + '=' + encodeURIComponent(value) + '; path=/miaro; max-age=31536000; SameSite=Lax';
        }

        // ---- Theme (migrate from localStorage to cookie) ----
        var THEME_KEY = 'miaro-theme';
        var DEFAULT_THEME = 'neumorphism';

        // One-time migration from localStorage
        var lsTheme = localStorage.getItem(THEME_KEY);
        if (lsTheme) {
            setCookie(THEME_KEY, lsTheme);
            localStorage.removeItem(THEME_KEY);
        }

        function getTheme() {
            return getCookie(THEME_KEY) || DEFAULT_THEME;
        }

        function setTheme(theme) {
            document.body.className = 'theme-' + theme;
            setCookie(THEME_KEY, theme);
            document.querySelectorAll('.theme-btn').forEach(function(btn) {
                btn.classList.toggle('active', btn.dataset.theme === theme);
            });
            // Sync all team dropdowns across themes
            document.querySelectorAll('.team-select').forEach(function(sel) {
                sel.value = '{{ .Team }}';
            });
        }

        setTheme(getTheme());

        document.querySelectorAll('.theme-btn').forEach(function(btn) {
            btn.addEventListener('click', function() {
                setTheme(this.dataset.theme);
            });
        });

        // ---- Team selector ----
        window.selectTeam = function(teamNum) {
            window.location.href = '/miaro?team=' + teamNum;
        };
    })();
</script>
```

- [ ] **Step 4: Run all tests**

Run: `go test -v ./...`
Expected: all PASS. The template tests in `main_test.go` should still pass since they check `/miaro` (team 1 default) and the HTML structure hasn't changed, just added elements.

- [ ] **Step 5: Manual smoke test**

Run: `go run .`
Then check in browser:
- `http://localhost:8081/miaro` — should show Team 1 with dropdown
- Select Team 3 from dropdown — should navigate to `/miaro?team=3`
- Visit `/miaro?team=1` — should redirect to `/miaro`
- Visit `/miaro/json?team=2` — should show team 2 in raw_schedule
- Switch themes — verify dropdown appears in all three
- Close and reopen browser — team selection should persist via cookie

- [ ] **Step 6: Commit**

```bash
git add templates/miaroSchedule.tmpl
git commit -m "feat: add team selector dropdown with cookie persistence and theme cookie migration"
```

---

### Task 5: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add team documentation to CLAUDE.md**

Add a brief section about teams to the architecture section, after the shift table:

```markdown
### Teams

5 teams share the same 10-day rotation, each offset by 2 days. Team 1 (Miaro) is the default. Formula: `epoch + (team-1)*2 days`. Team selection uses `?team=N` query param with cookie persistence. `MiaroTeam` constant in `pkg/team.go`.
```

Update the schedule calculation description to mention the team parameter.

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add team selector info to CLAUDE.md"
```
