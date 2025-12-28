package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// ExercisesResponse represents the API response for exercises
type ExercisesResponse struct {
	Success   bool                `json:"success"`
	Exercises []database.Exercise `json:"exercises,omitempty"`
	Exercise  *database.Exercise  `json:"exercise,omitempty"`
	ErrorCode string              `json:"errorCode,omitempty"`
	Message   string              `json:"message,omitempty"`
}

// GetExercises handles the /api/exercises endpoint
func GetExercises(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for category filter
	category := r.URL.Query().Get("category")

	var exercises []database.Exercise
	var err error

	if category != "" {
		exercises, err = database.GetExercisesByCategory(category)
	} else {
		exercises, err = database.GetExercises()
	}

	if err != nil {
		response := ExercisesResponse{
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

	response := ExercisesResponse{
		Success:   true,
		Exercises: exercises,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExerciseBySlug handles the /api/exercises/{slug} endpoint
func GetExerciseBySlug(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract slug from path
	path := strings.TrimPrefix(r.URL.Path, "/api/exercises/")
	slug := strings.Split(path, "/")[0]

	if slug == "" {
		response := ExercisesResponse{
			Success:   false,
			ErrorCode: "INVALID_SLUG",
			Message:   "Exercise slug is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	exercise, err := database.GetExerciseBySlug(slug)
	if err != nil {
		response := ExercisesResponse{
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
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ExercisesResponse{
		Success:  true,
		Exercise: exercise,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SeedExercises handles the /api/admin/seed endpoint
func SeedExercises(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := database.SeedExercises()
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

	response := InitializeResponse{
		Success: true,
		Message: "Exercises seeded successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
