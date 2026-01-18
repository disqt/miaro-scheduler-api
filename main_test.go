package main

import (
	"encoding/json"
	"log/slog"
	"miaro-schedule-api/pkg"
	"net/http"
	"net/http/httptest"
	"os"
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
		"Miaro est",
		"il",
	}

	for _, expected := range expectedStrings {
		if !contains(body, expected) {
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

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
