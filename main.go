package main

import (
	"context"
	"embed"
	"html/template"
	"log/slog"
	"miaro-scheduler-api/pkg"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Embed the templates directory
//
//go:embed templates/*
var templatesFS embed.FS

// teamNames maps team numbers (1-5) to display names.
var teamNames = map[int]string{
	1: "Miaro",
	2: "Équipe 2",
	3: "Équipe 3",
	4: "Équipe 4",
	5: "Équipe 5",
}

// parseTeam extracts the team number (1-5) from the query string, defaulting to 1.
func parseTeam(c *gin.Context) int {
	teamStr := c.DefaultQuery("team", "1")
	team, err := strconv.Atoi(teamStr)
	if err != nil || team < 1 || team > 5 {
		return 1
	}
	return team
}

// SchedulerHandler returns a Gin handler for the HTML schedule endpoint.
func SchedulerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		team := parseTeam(c)
		teamOffset := (team - 1) * 2
		schedule := pkg.CalculateScheduleForTeam(teamOffset)
		scheduleBeautified := pkg.FormatScheduleBeautified(schedule)

		c.HTML(http.StatusOK, "miaroSchedule.tmpl", gin.H{
			"Schedule":               scheduleBeautified.Schedule,
			"IsWorking":              scheduleBeautified.IsWorking,
			"NextWorkingDay":         scheduleBeautified.NextWorkingDay,
			"ScheduleNextWorkingDay": scheduleBeautified.ScheduleNextWorkingDay,
			"CalendarDays":           scheduleBeautified.CalendarDays,
			"Team":                   team,
			"TeamName":               teamNames[team],
			"TeamNames":              teamNames,
		})
	}
}

// SchedulerJSONHandler returns a Gin handler for the JSON schedule endpoint.
func SchedulerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		team := parseTeam(c)
		teamOffset := (team - 1) * 2
		schedule := pkg.CalculateScheduleForTeam(teamOffset)
		scheduleBeautified := pkg.FormatScheduleBeautified(schedule)

		c.JSON(http.StatusOK, gin.H{
			"schedule":                  scheduleBeautified.Schedule,
			"is_working":                scheduleBeautified.IsWorking,
			"next_working_day":          scheduleBeautified.NextWorkingDay,
			"schedule_next_working_day": scheduleBeautified.ScheduleNextWorkingDay,
			"raw_schedule":              schedule,
			"team":                      team,
			"team_name":                 teamNames[team],
		})
	}
}

// HealthCheckHandler returns a Gin handler for health checks.
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "miaro-scheduler-api",
		})
	}
}

func setupRouter(config *pkg.Config, logger *slog.Logger) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))

	// Configure CORS if enabled
	if config.EnableCORS {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		router.Use(cors.New(corsConfig))
		logger.Info("CORS enabled")
	}

	// Create custom template functions
	funcMap := template.FuncMap{
		"title": func(s string) string {
			// Title case function for French text
			if s == "" {
				return s
			}
			// Convert first character to uppercase
			runes := []rune(s)
			runes[0] = unicode.ToUpper(runes[0])
			return string(runes)
		},
		"lower": strings.ToLower,
	}

	// Parse the templates from the embedded filesystem with custom functions
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.tmpl"))
	router.SetHTMLTemplate(tmpl)

	// Define routes
	router.GET("/health", HealthCheckHandler())
	router.GET("/miaro", SchedulerHandler())
	router.GET("/miaro/json", SchedulerJSONHandler())

	return router
}

// LoggerMiddleware creates a Gin middleware for structured logging.
func LoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency,
			"client_ip", clientIP,
		)
	}
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting Miaro Scheduler API")

	// Load configuration
	config, err := pkg.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		panic(err)
	}

	logger.Info("Configuration loaded",
		"port", config.Port,
		"timezone", config.Timezone,
		"cors_enabled", config.EnableCORS,
	)

	router := setupRouter(config, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", "port", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			panic(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		panic(err)
	}

	logger.Info("Server exited")
}
