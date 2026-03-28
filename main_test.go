package main

import (
	"encoding/json"
	"log/slog"
	"miaro-scheduler-api/pkg"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func setupTestRouter() *gin.Engine {
	config := &pkg.Config{
		Port:       "8081",
		Timezone:   "Europe/Paris",
		EnableCORS: false,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))

	return setupRouter(config, logger)
}

func TestHealthCheckHandler(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}

	if response["service"] != "miaro-scheduler-api" {
		t.Errorf("Expected service name 'miaro-scheduler-api', got '%v'", response["service"])
	}
}

func TestSchedulerHandler(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type is HTML
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected content type 'text/html; charset=utf-8', got '%s'", contentType)
	}

	// Check that response body contains expected text
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty HTML response")
	}

	// Check for key HTML elements
	expectedStrings := []string{
		"<!DOCTYPE html>",
		"Horaire de Miaro",
		"Statut",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected response to contain '%s'", expected)
		}
	}
}

func TestSchedulerJSONHandler(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro/json", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type is JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected content type 'application/json; charset=utf-8', got '%s'", contentType)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check required fields
	requiredFields := []string{
		"schedule",
		"is_working",
		"next_working_day",
		"schedule_next_working_day",
		"raw_schedule",
	}

	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' in JSON response", field)
		}
	}

	// Check that raw_schedule contains expected fields
	rawSchedule, ok := response["raw_schedule"].(map[string]interface{})
	if !ok {
		t.Fatal("raw_schedule is not a map")
	}

	if _, exists := rawSchedule["time_requested"]; !exists {
		t.Error("Expected 'time_requested' in raw_schedule")
	}

	if _, exists := rawSchedule["schedule_type"]; !exists {
		t.Error("Expected 'schedule_type' in raw_schedule")
	}

	if _, exists := rawSchedule["day_in_schedule"]; !exists {
		t.Error("Expected 'day_in_schedule' in raw_schedule")
	}
}

func TestSchedulerJSONHandler_ScheduleTypes(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro/json", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	rawSchedule := response["raw_schedule"].(map[string]interface{})
	scheduleType := rawSchedule["schedule_type"].(string)

	// Verify schedule type is one of the valid types
	validTypes := map[string]bool{
		"MORNING":   true,
		"AFTERNOON": true,
		"NIGHT":     true,
		"FREE":      true,
	}

	if !validTypes[scheduleType] {
		t.Errorf("Invalid schedule type: %s", scheduleType)
	}
}

func TestNotFoundHandler(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

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

func TestSchedulerHandler_MonthParam(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/miaro?month=2026-04", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Template outputs CalendarMonthLabel once it's wired (Task 3).
	// For now, verify the page renders successfully.
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty HTML response for month=2026-04")
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

	// Template outputs CalendarMonthLabel once it's wired (Task 3).
	// For now, verify the page renders and the team cookie is set.
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty HTML response for month=2026-05&team=3")
	}

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
