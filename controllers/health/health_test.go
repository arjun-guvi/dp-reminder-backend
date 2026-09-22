package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Create controller
	ctrl := New()

	// Register route
	router.GET("/health", ctrl.Health)

	// Create test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response body contains expected fields
	body := w.Body.String()
	if body == "" {
		t.Error("Response body should not be empty")
	}
}

func TestReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	ctrl := New()
	router.GET("/health/ready", ctrl.Ready)

	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestLive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	ctrl := New()
	router.GET("/health/live", ctrl.Live)

	req, _ := http.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
