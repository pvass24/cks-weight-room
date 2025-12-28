package api

import (
	"encoding/json"
	"net/http"

	"github.com/patrickvassell/cks-weight-room/internal/prerequisites"
)

// ValidationResponse represents the API response for prerequisite validation
type ValidationResponse struct {
	Success   bool                        `json:"success"`
	Checks    []prerequisites.CheckResult `json:"checks"`
	ErrorCode string                      `json:"errorCode,omitempty"`
	Message   string                      `json:"message,omitempty"`
}

// ValidatePrerequisites handles the /api/setup/validate endpoint
func ValidatePrerequisites(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Run all prerequisite checks
	checks, err := prerequisites.ValidateAll()

	response := ValidationResponse{
		Success: err == nil,
		Checks:  checks,
	}

	// If there's an error, add error details
	if err != nil {
		if prereqErr, ok := err.(*prerequisites.PrerequisiteError); ok {
			response.ErrorCode = prereqErr.Code
			response.Message = prereqErr.Message
		} else {
			response.ErrorCode = "UNKNOWN_ERROR"
			response.Message = err.Error()
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
