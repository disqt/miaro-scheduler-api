# miaro-scheduler-api

A Go API to track a 10-day rotating work schedule. Provides both a modern HTML dashboard and JSON endpoints to check the current work status and schedule.

## Features

- 10-day rotating work schedule (Morning, Afternoon, Night shifts + Free days)
- Modern UI with Bootswatch Lux theme and visual calendar view
- Timezone-aware calculations (Europe/Paris)
- JSON and HTML endpoints
- Health check endpoint for monitoring
- CORS support (configurable)
- Structured JSON logging
- Systemd service support
- Comprehensive test coverage
- Graceful shutdown

## Schedule Pattern

The 10-day cycle repeats as follows:

| Day | Shift Type | Working Hours |
|-----|------------|---------------|
| 0-1 | Morning | 06:00 - 14:00 |
| 2-3 | Afternoon | 14:00 - 22:00 |
| 4-5 | Night | 22:00 - 06:00 |
| 6-9 | Free | Off duty |

The cycle started on August 31, 2024.

## Quick Start

### Prerequisites

- Go 1.21+

### Run Locally

```bash
# Clone the repository
git clone https://github.com/disqt/miaro-scheduler-api.git
cd miaro-scheduler-api

# Run directly
go run .

# Or build and run
go build -o miaro-scheduler-api .
./miaro-scheduler-api
```

The API will be available at `http://localhost:8081/miaro`

## Deployment

The service runs as a systemd unit on the VPS.

### Deploy Script

A deployment script is provided for convenience:

```bash
# Make the script executable
chmod +x scripts/deploy.sh

# Run tests
./scripts/deploy.sh test

# Build binary
./scripts/deploy.sh build

# Full deploy (test + build + install + restart service)
./scripts/deploy.sh deploy

# Other commands
./scripts/deploy.sh restart   # Restart service
./scripts/deploy.sh stop      # Stop service
./scripts/deploy.sh logs      # View logs (journalctl)
./scripts/deploy.sh status    # Check service status and health
./scripts/deploy.sh help      # Show all commands
```

### Manual Deployment

```bash
# Build the binary
CGO_ENABLED=0 go build -o miaro-scheduler-api .

# Copy to server
sudo mkdir -p /opt/miaro-scheduler-api
sudo cp miaro-scheduler-api /opt/miaro-scheduler-api/
sudo cp -r templates /opt/miaro-scheduler-api/

# Install systemd service (first time only)
sudo cp scripts/miaro-scheduler-api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable miaro-scheduler-api

# Start/restart service
sudo systemctl restart miaro-scheduler-api
```

### Systemd Service

The service file is located at `scripts/miaro-scheduler-api.service`. Install it with:

```bash
sudo cp scripts/miaro-scheduler-api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable miaro-scheduler-api
sudo systemctl start miaro-scheduler-api
```

Useful commands:

```bash
# Check status
sudo systemctl status miaro-scheduler-api

# View logs
sudo journalctl -u miaro-scheduler-api -f

# Restart
sudo systemctl restart miaro-scheduler-api
```

## API Endpoints

### GET /health

Health check endpoint for monitoring.

```json
{
  "status": "ok",
  "service": "miaro-scheduler-api"
}
```

### GET /miaro

Returns an HTML dashboard displaying the current schedule with:
- Current work status (working/not working)
- Current shift details
- Visual calendar view of the month
- Next working day indicator

### GET /miaro/json

Returns schedule information in JSON format.

```json
{
  "schedule": "du matin",
  "is_working": "est au travail",
  "next_working_day": "demain",
  "schedule_next_working_day": "du matin",
  "raw_schedule": {
    "time_requested": "2025-01-18T10:00:00+01:00",
    "schedule_type": "MORNING",
    "day_in_schedule": 3
  }
}
```

## Configuration

Environment variables (set in the systemd service file):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | Server port |
| `TIMEZONE` | `Europe/Paris` | Timezone for schedule calculations |
| `ENABLE_CORS` | `false` | Enable CORS for cross-origin requests |

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run with coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# View coverage report
go tool cover -html=coverage.txt
```

## Project Structure

```
.
├── main.go                    # Application entry point and HTTP handlers
├── main_test.go               # HTTP handler tests
├── pkg/
│   ├── config.go              # Configuration management
│   ├── schedulerService.go    # Core schedule calculation logic
│   ├── template.go            # Template formatting functions
│   └── *_test.go              # Unit tests
├── templates/
│   └── miaroSchedule.tmpl     # HTML template (Bootswatch Lux theme)
├── scripts/
│   ├── deploy.sh              # Build and deploy script
│   └── miaro-scheduler-api.service  # Systemd service file
├── .github/
│   └── workflows/
│       └── ci.yml             # GitHub Actions CI/CD
└── README.md
```

## UI Theme

The HTML dashboard uses the [Bootswatch Lux](https://bootswatch.com/lux/) theme with:
- Dark gradient background
- Clean card-based layout
- Status indicators with glow effects
- Visual calendar grid with shift indicators
- Responsive design for mobile devices

## CI/CD

GitHub Actions workflow runs on push/PR to main and dev branches:

- Runs tests with race detection
- Generates coverage reports (Codecov)
- Runs linting (golangci-lint)
- Builds the application

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Ensure tests pass: `go test ./...`
5. Ensure code is formatted: `go fmt ./...`
6. Submit a pull request

## License

MIT License - See [LICENSE](LICENSE) for details.

---

**[View on GitHub](https://github.com/disqt/miaro-scheduler-api)**
