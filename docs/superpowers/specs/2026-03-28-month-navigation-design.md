# Month Navigation Design

Add month navigation to the calendar view so users can browse past and future months.

## Current Behavior

- Calendar shows the current month only
- Leading days before the 1st are empty cells
- No trailing days after the last day of the month
- No way to view other months

## New Behavior

### Query Parameter

`?month=YYYY-MM` selects which month to display. Missing or invalid defaults to current month. Combines with existing `?team=N` param.

Examples:
- `/miaro` -- current month, default team
- `/miaro?month=2026-04` -- April 2026, default team
- `/miaro?month=2026-04&team=3` -- April 2026, team 3

### Full-Week Calendar Grid

Instead of empty cells at the start/end of the month, show actual dates from adjacent months:

- **Leading days**: If March 1 is a Sunday, show Mon Feb 23 through Sat Feb 28 before it, greyed out with their shift types
- **Trailing days**: If March 31 is a Tuesday, show Wed Apr 1 through Sun Apr 5 after it, greyed out with their shift types

Adjacent-month days get shift coloring (faint) but are visually distinct via reduced opacity.

### Navigation Header

Replace the static month/year header with `< Mars 2026 >` where `<` and `>` are clickable links:

- `<` links to `?month=2026-02` (preserving `?team=` if set)
- `>` links to `?month=2026-04` (preserving `?team=` if set)
- No bounds -- schedule is deterministic for any date
- Month name in French (existing `title` template function)

### Status Cards Unchanged

The shift status, working status, and next-working-day cards always reflect **today** regardless of which month the calendar shows. Only the calendar grid changes.

## Data Model Changes

### CalendarDay struct

Add field:
- `IsOtherMonth bool` -- true for leading/trailing days from adjacent months

Remove field:
- `IsEmpty bool` -- no longer needed (replaced by `IsOtherMonth`)

### ScheduleBeautified struct

Add fields:
- `CalendarMonthLabel string` -- display label, e.g. "Mars 2026"
- `PrevMonth string` -- `YYYY-MM` for prev link
- `NextMonth string` -- `YYYY-MM` for next link
- `IsCurrentMonth bool` -- true when viewing the current month

### GenerateCalendarData changes

New signature: accept an explicit `targetMonth time.Time` parameter instead of deriving from `schedule.TimeRequested`.

Logic:
1. Compute first/last day of target month
2. Fill leading days from previous month (actual dates, `IsOtherMonth: true`)
3. Fill days of the month (as today, with `IsToday`/`IsNextWork` only when `IsCurrentMonth`)
4. Fill trailing days from next month to complete the last week row

### Handler changes (main.go)

1. Parse `?month=YYYY-MM` from query string
2. Validate format; default to current month on invalid/missing
3. Compute `targetMonth` as first-of-month in `parisLoc`
4. Pass `targetMonth` to calendar generation
5. Compute `CalendarMonthLabel`, `PrevMonth`, `NextMonth`
6. Pass all new fields to template data

## Template Changes

All 3 themes (neumorphism, bento, terminal):

1. Calendar header: replace static text with `< {{ .CalendarMonthLabel }} >` navigation
2. Calendar day: add `{{ if .IsOtherMonth }}other-month{{ end }}` class
3. Remove `{{ if .IsEmpty }}empty{{ end }}` class usage (replaced by `IsOtherMonth`)

## CSS Changes

All 3 themes:

1. `.calendar-day.other-month` -- reduced opacity (~0.35-0.4), no hover effects
2. Remove `.calendar-day.empty` rules (replace with `.other-month`)
3. Navigation arrows: styled per-theme as subtle clickable elements

## JSON Endpoint

`/miaro/json` does not include calendar data today and is not affected.

## Testing

- Unit test `GenerateCalendarData` with a known month to verify leading/trailing days are correct
- Test `?month=` param parsing (valid, invalid, missing)
- Test that `IsOtherMonth` days have correct shift types
- Test month label and prev/next values
- Integration test that the HTML contains nav links
