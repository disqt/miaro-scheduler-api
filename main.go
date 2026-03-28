package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"miaro-scheduler-api/pkg"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

// parseTeamString parses and validates a team number string. Returns (team, valid).
func parseTeamString(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	team, err := strconv.Atoi(s)
	if err != nil || pkg.ValidateTeam(team) != team {
		return 0, false
	}
	return team, true
}

func parseTeamParam(c *gin.Context) (int, bool) {
	return parseTeamString(c.Query("team"))
}

func readTeamCookie(c *gin.Context) (int, bool) {
	cookie, err := c.Cookie("miaro-team")
	if err != nil {
		return 0, false
	}
	return parseTeamString(cookie)
}

func setTeamCookie(c *gin.Context, team int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "miaro-team",
		Value:    strconv.Itoa(team),
		Path:     "/miaro",
		MaxAge:   365 * 24 * 60 * 60,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
}

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

// SchedulerHandler returns a Gin handler for the HTML schedule endpoint.
func SchedulerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		teamParam, hasValidParam := parseTeamParam(c)
		teamCookie, hasValidCookie := readTeamCookie(c)

		// Canonicalize: ?team=1 -> redirect to /miaro (preserve month)
		if hasValidParam && teamParam == pkg.MiaroTeam {
			setTeamCookie(c, pkg.MiaroTeam)
			redirect := "/miaro"
			if m := c.Query("month"); m != "" {
				redirect = fmt.Sprintf("/miaro?month=%s", m)
			}
			c.Redirect(http.StatusFound, redirect)
			return
		}

		if hasValidParam {
			// Valid ?team=N (2-5): render
			setTeamCookie(c, teamParam)
			renderSchedule(c, teamParam)
			return
		}

		// No valid param — check cookie (preserve month)
		if hasValidCookie && teamCookie != pkg.MiaroTeam {
			redirect := fmt.Sprintf("/miaro?team=%d", teamCookie)
			if m := c.Query("month"); m != "" {
				redirect = fmt.Sprintf("/miaro?team=%d&month=%s", teamCookie, m)
			}
			c.Redirect(http.StatusFound, redirect)
			return
		}

		// Default: team 1
		setTeamCookie(c, pkg.MiaroTeam)
		renderSchedule(c, pkg.MiaroTeam)
	}
}

// SchedulerJSONHandler returns a Gin handler for the JSON schedule endpoint.
func SchedulerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		team := pkg.MiaroTeam
		if t, valid := parseTeamParam(c); valid {
			team = t
		}

		schedule := pkg.CalculateSchedule(time.Now(), team)
		now := time.Now().In(pkg.ParisLoc())
		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		scheduleBeautified := pkg.FormatScheduleBeautified(schedule, currentMonth)

		c.JSON(http.StatusOK, gin.H{
			"schedule":                  scheduleBeautified.Schedule,
			"is_working":                scheduleBeautified.IsWorking,
			"next_working_day":          scheduleBeautified.NextWorkingDay,
			"schedule_next_working_day": scheduleBeautified.ScheduleNextWorkingDay,
			"raw_schedule":              schedule,
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
