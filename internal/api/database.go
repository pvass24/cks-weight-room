package api

import (
	"encoding/json"
	"net/http"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// InitializeResponse represents the API response for database initialization
type InitializeResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	ErrorCode string `json:"errorCode,omitempty"`
}

// InitializeDatabase handles the /api/setup/initialize endpoint
func InitializeDatabase(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get default database path
	dbPath := database.GetDefaultPath()

	// Check if already initialized
	if database.IsInitialized(dbPath) {
		response := InitializeResponse{
			Success: true,
			Message: "Database already initialized",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Initialize database
	cfg := database.Config{
		Path: dbPath,
	}

	err := database.Initialize(cfg)
	if err != nil {
		response := InitializeResponse{
			Success: false,
		}

		if dbErr, ok := err.(*database.DatabaseError); ok {
			response.ErrorCode = dbErr.Code
			response.Message = dbErr.Message
		} else {
			response.ErrorCode = "UNKNOWN_ERROR"
			response.Message = err.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Mark first launch as completed
	database.SetConfig("first_launch_completed", "true")

	// Automatically seed exercises on first initialization
	if err := database.SeedExercises(); err != nil {
		// Log error but don't fail initialization
		// Exercises can be seeded later via /api/admin/seed
	}

	response := InitializeResponse{
		Success: true,
		Message: "Database initialized successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDatabaseStatus handles the /api/setup/db-status endpoint
func GetDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbPath := database.GetDefaultPath()
	initialized := database.IsInitialized(dbPath)

	response := map[string]interface{}{
		"initialized": initialized,
		"path":        dbPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
