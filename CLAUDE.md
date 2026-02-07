# Miaro Scheduler API

Go/Gin API that tracks a 10-day rotating work schedule. Serves an HTML dashboard (with 3 switchable themes) and a JSON endpoint showing the current shift, work status, next working day, and a monthly calendar.

## Build / Run / Test

```bash
# Run locally (default port 8081)
go run .

# Build binary
go build -o miaro-scheduler-api .

# Run tests
go test ./...

# Tests with race detection and coverage (same as CI)
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Format
go fmt ./...

# Deploy (on VPS, runs tests + builds + installs + restarts service)
./scripts/deploy.sh deploy
```

## Architecture

```
Request -> Gin router -> Handler -> CalculateSchedule() -> FormatScheduleBeautified() -> Render
```

### Request flow

1. `main.go` sets up the Gin router with Recovery, structured JSON logging middleware, and optional CORS.
2. `GET /miaro` calls `SchedulerHandler` which computes the schedule and renders `miaroSchedule.tmpl` as HTML.
3. `GET /miaro/json` calls `SchedulerJSONHandler` which returns the same data as JSON.
4. `GET /health` returns `{"status":"ok","service":"miaro-scheduler-api"}`.

### Schedule calculation (`pkg/schedulerService.go`)

- Epoch: **August 31, 2024** at midnight in `Europe/Paris`.
- Computes `dayInSchedule = daysSinceEpoch % 10`.
- Maps day index to shift type:

| Day index | Shift | Hours |
|-----------|-------|-------|
| 0-1 | MORNING | 06:00 - 14:00 |
| 2-3 | AFTERNOON | 14:00 - 22:00 |
| 4-5 | NIGHT | 22:00 - 06:00 |
| 6-9 | FREE | Off duty |

### Formatting (`pkg/template.go`)

- Converts schedule to French text (`du matin`, `de l'apres-midi`, `de nuit`, `libre`).
- Determines real-time work status by checking hour against shift windows.
- Finds next working day (`demain` or `dans N jours`).
- Generates `CalendarDay` slice for the current month's calendar grid.

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | HTTP server, Gin router setup, handlers, structured logging middleware, graceful shutdown |
| `main_test.go` | Integration tests for all HTTP endpoints |
| `pkg/schedulerService.go` | Core 10-day modulo cycle calculation |
| `pkg/schedulerService_test.go` | Unit tests for schedule calculation |
| `pkg/config.go` | Environment variable loading with defaults |
| `pkg/template.go` | French text formatting, work status logic, calendar data generation |
| `pkg/template_test.go` | Unit tests for formatting |
| `pkg/template_bug_test.go` | Regression tests for template edge cases |
| `templates/miaroSchedule.tmpl` | Embedded HTML template with 3 themes (Neumorphism, Bento, Terminal) |
| `scripts/deploy.sh` | Build and deploy script (test/build/deploy/restart/stop/logs/status) |
| `scripts/miaro-scheduler-api.service` | Systemd unit file |
| `Dockerfile` | Multi-stage Docker build (golang:1.24-alpine -> alpine) |
| `.github/workflows/ci.yml` | GitHub Actions CI: test, lint (golangci-lint), build, Docker image |

## Configuration

Environment variables (defaults in parentheses):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | Server listen port |
| `TIMEZONE` | `Europe/Paris` | Timezone for schedule calculations |
| `ENABLE_CORS` | `false` | Enable CORS with allow-all-origins |

## Deployment

- **Service:** `miaro-scheduler-api` (systemd)
- **User:** `www-data` (per the service file; deploy script installs to `/opt/miaro-scheduler-api`)
- **Port:** 8081
- **Public URL:** `https://disqt.com/miaro` (nginx reverse proxy)
- **VPS access:** `ssh dev`
- Templates are embedded at compile time via `//go:embed templates/*`, but the deploy script also copies the `templates/` directory to the install path.

### Deploy commands

```bash
# Full deploy (on VPS)
./scripts/deploy.sh deploy

# Or manually
CGO_ENABLED=0 go build -o miaro-scheduler-api .
sudo cp miaro-scheduler-api /opt/miaro-scheduler-api/
sudo systemctl restart miaro-scheduler-api

# Check status
sudo systemctl status miaro-scheduler-api
curl http://localhost:8081/health
```

## CI/CD

GitHub Actions (`.github/workflows/ci.yml`) runs on push/PR to `main` and `dev`:

1. **Test** - `go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...` + Codecov upload
2. **Lint** - `golangci-lint` (continue-on-error)
3. **Build** - `go build -v ./...` (depends on test + lint)
4. **Docker** - Builds Docker image on push to main/dev only (no push to registry)

## Design Decisions

- **Embedded templates:** Uses `//go:embed` so the binary is fully self-contained. No external template files needed at runtime.
- **Modulo 10 cycle:** Simple `daysSinceEpoch % 10` calculation. No database needed -- the schedule is deterministic from the epoch date.
- **Europe/Paris timezone:** Hardcoded in `CalculateSchedule()` to handle CET/CEST (DST) transitions correctly. The `TIMEZONE` config var only affects the config struct's `ScheduleStart` field.
- **French localization:** All UI text is in French. The `title` template function provides basic French-compatible title casing.
- **Three themes:** Neumorphism (default), Bento, and Terminal. Theme selection persists in `localStorage`. All three layouts are rendered in the same template and toggled via CSS/JS.
- **Graceful shutdown:** Server listens for SIGINT/SIGTERM and gives 5 seconds for in-flight requests to complete.
- **Structured logging:** Uses `slog` with JSON handler for production-friendly log output.
