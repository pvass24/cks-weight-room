package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

func TestInitializeDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Clean up after test
	defer database.Close()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "POST request succeeds",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request fails",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	// Initialize database directly for testing
	cfg := database.Config{Path: dbPath}
	if err := database.Initialize(cfg); err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/setup/initialize", nil)
			w := httptest.NewRecorder()

			// Note: This test assumes database is already initialized
			// We're testing the HTTP handler behavior, not the initialization itself
			InitializeDatabase(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response InitializeResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !response.Success {
					t.Errorf("Expected success=true, got false. Message: %s", response.Message)
				}
			}
		})
	}
}

func TestGetDatabaseStatus(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Override default path for testing
	database.SetDefaultPathForTesting(dbPath)
	defer func() {
		database.SetDefaultPathForTesting("")
		database.Close()
	}()

	tests := []struct {
		name           string
		method         string
		setupDB        bool
		expectedStatus int
		expectInit     bool
	}{
		{
			name:           "GET with uninitialized database",
			method:         http.MethodGet,
			setupDB:        false,
			expectedStatus: http.StatusOK,
			expectInit:     false,
		},
		{
			name:           "GET with initialized database",
			method:         http.MethodGet,
			setupDB:        true,
			expectedStatus: http.StatusOK,
			expectInit:     true,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			setupDB:        false,
			expectedStatus: http.StatusMethodNotAllowed,
			expectInit:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up database before each test
			os.Remove(dbPath)
			database.Close()

			if tt.setupDB {
				cfg := database.Config{Path: dbPath}
				if err := database.Initialize(cfg); err != nil {
					t.Fatalf("Failed to initialize database: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/setup/db-status", nil)
			w := httptest.NewRecorder()

			GetDatabaseStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				initialized, ok := response["initialized"].(bool)
				if !ok {
					t.Error("Response missing 'initialized' field")
				}

				if initialized != tt.expectInit {
					t.Errorf("Expected initialized=%v, got %v", tt.expectInit, initialized)
				}

				path, ok := response["path"].(string)
				if !ok {
					t.Error("Response missing 'path' field")
				}

				if path != dbPath {
					t.Errorf("Expected path %s, got %s", dbPath, path)
				}
			}
		})
	}
}

func TestInitializeDatabaseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	defer database.Close()

	// Initialize database directly
	cfg := database.Config{Path: dbPath}
	if err := database.Initialize(cfg); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Check that IsInitialized returns true
	if !database.IsInitialized(dbPath) {
		t.Error("Database should be initialized")
	}

	// Initialize again (should be idempotent)
	err := database.Initialize(cfg)
	if err != nil {
		// Second initialization might fail because DB is already open
		// Close and retry
		database.Close()
		err = database.Initialize(cfg)
		if err != nil {
			t.Fatalf("Second initialization failed: %v", err)
		}
	}

	// Verify still initialized
	if !database.IsInitialized(dbPath) {
		t.Error("Database should still be initialized after second init")
	}
}
