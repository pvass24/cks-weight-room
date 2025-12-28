package api

import (
	"encoding/json"
	"net/http"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// ResetProgressRequest represents a reset request with confirmation
type ResetProgressRequest struct {
	Confirmation string `json:"confirmation"`
}

// ResetProgressResponse represents the response from a reset operation
type ResetProgressResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetResetStats handles GET /api/reset/stats
func GetResetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	stats := struct {
		AttemptsCount      int `json:"attemptsCount"`
		PersonalBestsCount int `json:"personalBestsCount"`
		MockExamsCount     int `json:"mockExamsCount"`
	}{}

	// Get counts for confirmation message
	database.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&stats.AttemptsCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM progress WHERE personal_best_seconds IS NOT NULL").Scan(&stats.PersonalBestsCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM mock_exams").Scan(&stats.MockExamsCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ResetProgress handles POST /api/reset
func ResetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	var req ResetProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify confirmation text
	if req.Confirmation != "DELETE" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResetProgressResponse{
			Success: false,
			Message: "Confirmation text must be 'DELETE'",
		})
		return
	}

	// Begin transaction for atomic operation
	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Delete all progress data
	_, err = tx.Exec("DELETE FROM attempts")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete attempts", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM progress")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete progress", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM mock_exams")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete mock exams", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResetProgressResponse{
		Success: true,
		Message: "All progress data has been reset.",
	})
}
