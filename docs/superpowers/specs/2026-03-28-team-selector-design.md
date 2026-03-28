# Team Selector Design

Add a dropdown to select between 5 work teams sharing the same 10-day rotating schedule, each offset by 2 days. Team 1 is Miaro's team (the existing default).

## Teams

| Team | Label | Epoch offset |
|------|-------|-------------|
| 1 | Equipe 1 (Miaro) | Aug 31, 2024 (original) |
| 2 | Equipe 2 | Sep 2, 2024 (+2 days) |
| 3 | Equipe 3 | Sep 4, 2024 (+4 days) |
| 4 | Equipe 4 | Sep 6, 2024 (+6 days) |
| 5 | Equipe 5 | Sep 8, 2024 (+8 days) |

Formula: `epochDate + (team - 1) * 2 days`

### Concrete example

On Sep 2, 2024:
- **Team 1** (epoch Aug 31): `diffDays = 2`, `2 % 10 = 2` -> AFTERNOON (day index 2)
- **Team 2** (epoch Sep 2): `diffDays = 0`, `0 % 10 = 0` -> MORNING (day index 0)

So Team 2 starts their MORNING cycle 2 days after Team 1 did. Higher team numbers start later.

## Schedule calculation

`CalculateSchedule()` gains a `team int` parameter (1-5). The epoch date is offset forward by `(team-1)*2` days before computing `daysSinceEpoch % 10`. Invalid team values default to 1.

A `MiaroTeam = 1` constant identifies which team is Miaro's.

The `Schedule` struct gains a `Team int` field.

### Call chain propagation

The team parameter must be threaded through the full call chain:
- `CalculateSchedule(time, team)` — core calculation with offset epoch
- `FormatScheduleBeautified(schedule)` — no change needed (reads `Schedule.Team` from struct if needed)
- `GenerateCalendarData(schedule, nextWorkDays)` — calls `CalculateSchedule()` internally for each day of the month. Must pass the team from `schedule.Team` so calendar days use the correct offset.

## Request flow

### HTML endpoint (`GET /miaro`, `GET /miaro?team=N`)

Resolution order for team:
1. `?team=` query parameter (if present and valid 1-5)
2. `miaro-team` cookie (if set and valid)
3. Default: 1

An invalid `?team=` value (non-numeric, out of range) is treated as absent for both resolution and redirect purposes.

Redirect and canonicalization logic (based on resolved team, not raw query string):
- If `?team=1` is present: **302 redirect** to `/miaro` (strip the param), set cookie to 1
- If resolved team != 1 but URL has no valid `?team=` param: **302 redirect** to `/miaro?team=N`
- If resolved team is 1 and no `?team=` param: render Team 1 at `/miaro` (no redirect)
- Otherwise (valid `?team=N` where N is 2-5): render the requested team, set cookie

On every non-redirect request, set `miaro-team` cookie (max-age 1 year, `Path=/miaro`, `SameSite=Lax`) to the resolved team number.

### JSON endpoint (`GET /miaro/json`, `GET /miaro/json?team=N`)

Reads `?team=` query parameter, defaults to 1 if absent or invalid. No cookie logic, no redirects. The JSON response includes the `team` field from `raw_schedule`.

## Template changes

### Dropdown

A `<select>` element in the header/title area of each theme (Neumorphism, Bento, Terminal). Displays:
- "Equipe 1 (Miaro)"
- "Equipe 2"
- "Equipe 3"
- "Equipe 4"
- "Equipe 5"

The current team is pre-selected via template data.

On change: navigates to `/miaro?team=N` for all teams (the server handles canonicalizing team 1 to `/miaro`).

### Template data

Handler passes to the template:
- `Team` (int) — the current team number
- `Teams` (slice of structs with `Number`, `Label`, `Selected`) — for rendering the dropdown
- All existing fields unchanged

The page title stays "Horaire de Miaro" regardless of team — Miaro is the app name, not just a team name.

### Cookie migration

Theme storage moves from `localStorage` to a `miaro-theme` cookie (max-age 1 year, `Path=/miaro`, `SameSite=Lax`). The JS checks for a `localStorage` value on load, migrates it to the cookie if found, then deletes the `localStorage` entry. New reads/writes use the cookie only. Theme is still applied client-side only (same flash-of-default behavior as current localStorage approach). Default theme remains neumorphism.

## Styling

The dropdown is styled per-theme to match existing design language:
- **Neumorphism:** neumorphic inset select with soft shadows
- **Bento:** clean bordered select matching bento card style
- **Terminal:** monospace, dark background, green/cyan border

## Validation

- Team parameter must be integer 1-5
- Any invalid value (out of range, non-numeric, empty) defaults to team 1
- No error responses — silent fallback
