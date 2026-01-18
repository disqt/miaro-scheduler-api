# miaro-scheduler-api

A simple Go API to track an unusual 10-day rotating work schedule. The API provides both HTML and JSON endpoints to check the current work status and schedule.

## Features

- 🔄 10-day rotating work schedule (Morning, Afternoon, Night shifts + Free days)
- 🌍 Timezone-aware (Europe/Paris)
- 📊 JSON and HTML endpoints
- 🏥 Health check endpoint for monitoring
- 🔒 CORS support (configurable)
- 📝 Structured JSON logging
- 🐳 Docker support
- ✅ Comprehensive test coverage
- 🚀 Graceful shutdown

## Schedule Pattern

The 10-day cycle repeats as follows:

| Day | Shift Type | Working Hours |
|-----|------------|---------------|
| 0-1 | Morning | 06:00 - 14:00 |
| 2-3 | Afternoon | 14:00 - 22:00 |
| 4-5 | Night | 22:00 - 06:00 |
| 6-9 | Free | Off duty |

The cycle started on August 31, 2024.

## API Endpoints

### GET /health

Health check endpoint for monitoring and load balancers.

**Response:**
```json
{
  "status": "ok",
  "service": "miaro-scheduler-api"
}
```

### GET /miaro

Returns an HTML page displaying the current schedule in French.

**Response:** HTML page with schedule information

### GET /miaro/json

Returns schedule information in JSON format.

**Response:**
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

The application can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | Server port |
| `TIMEZONE` | `Europe/Paris` | Timezone for schedule calculations |
| `ENABLE_CORS` | `false` | Enable CORS for cross-origin requests |

## Building

### Local Build

```bash
go build ./main.go
```

### Running Locally

```bash
go run ./main.go
```

The API will listen on `http://localhost:8081` by default.

### With Custom Configuration

```bash
PORT=3000 ENABLE_CORS=true go run ./main.go
```

## Docker

### Build Docker Image

```bash
docker build -t miaro-scheduler-api .
```

### Run Docker Container

```bash
docker run -p 8081:8081 miaro-scheduler-api
```

### With Environment Variables

```bash
docker run -p 3000:3000 \
  -e PORT=3000 \
  -e ENABLE_CORS=true \
  miaro-scheduler-api
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
```

### View Coverage Report

```bash
go tool cover -html=coverage.txt
```

## Development

### Project Structure

```
.
├── main.go                    # Main application entry point
├── pkg/
│   ├── config.go             # Configuration management
│   ├── schedulerService.go   # Core schedule calculation logic
│   ├── template.go           # Template formatting functions
│   ├── schedulerService_test.go
│   ├── template_test.go
│   └── template_bug_test.go  # Bug demonstration tests
├── templates/
│   └── miaroSchedule.tmpl    # HTML template
├── main_test.go              # HTTP handler tests
├── Dockerfile                # Docker configuration
├── .github/
│   └── workflows/
│       └── ci.yml           # GitHub Actions CI/CD
└── README.md

```

### Key Components

- **CalculateSchedule**: Determines the current day in the 10-day cycle
- **FormatScheduleBeautified**: Converts schedule data to human-readable French text
- **Config**: Manages application configuration from environment variables
- **Middleware**: Structured logging and CORS support

## CI/CD

The project includes a GitHub Actions workflow that:

- ✅ Runs tests with race detection
- 📊 Generates coverage reports
- 🔍 Runs linting (golangci-lint)
- 🏗️ Builds the application
- 🐳 Builds Docker images on push to main/dev

## Recent Improvements

- ✅ Fixed critical bug in `nextWorkingDay()` function (modulo wrap-around)
- ✅ Fixed timezone inconsistency in schedule calculations
- ✅ Added proper error handling for timezone loading
- ✅ Exported Schedule struct fields for better API usability
- ✅ Added comprehensive configuration management
- ✅ Implemented structured logging with slog
- ✅ Added health check and JSON API endpoints
- ✅ Implemented CORS support
- ✅ Added graceful shutdown
- ✅ Updated all dependencies and fixed security vulnerabilities
- ✅ Added comprehensive test coverage (including HTTP handler tests)
- ✅ Created Docker support with health checks
- ✅ Set up GitHub Actions CI/CD pipeline

## License

This project was created as a learning exercise to explore Go features such as datetime handling and templates.

## Contributing

Issues and pull requests are welcome! Please ensure:

1. All tests pass: `go test ./...`
2. Code is formatted: `go fmt ./...`
3. Linting passes: `golangci-lint run`

---

**[View on GitHub](https://github.com/disqt/miaro-scheduler-api)**
