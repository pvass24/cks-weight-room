package database

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed seed_exercises.json
var seedExercisesJSON []byte

// Exercise represents a CKS exercise/challenge
type Exercise struct {
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	Difficulty       string   `json:"difficulty"`
	Points           int      `json:"points"`
	EstimatedMinutes int      `json:"estimatedMinutes"`
	Prerequisites    []string `json:"prerequisites"`
	Hints            []string `json:"hints"`
	Solution         string   `json:"solution"`
}

// SeedExercises populates the database with initial CKS exercises
func SeedExercises() error {
	if DB == nil {
		return &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	// Parse seed data
	var exercises []Exercise
	if err := json.Unmarshal(seedExercisesJSON, &exercises); err != nil {
		return &DatabaseError{
			Code:    "SEED_PARSE_FAILED",
			Message: "Failed to parse seed data",
			Err:     err,
		}
	}

	// Check if exercises already exist
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count)
	if err != nil {
		return &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to check existing exercises",
			Err:     err,
		}
	}

	if count > 0 {
		return nil // Already seeded
	}

	// Begin transaction
	tx, err := DB.Begin()
	if err != nil {
		return &DatabaseError{
			Code:    "SEED_TRANSACTION_FAILED",
			Message: "Failed to begin transaction",
			Err:     err,
		}
	}
	defer tx.Rollback()

	// Insert exercises
	stmt, err := tx.Prepare(`
		INSERT INTO exercises (
			slug, title, description, category, difficulty,
			points, estimated_minutes, prerequisites, hints, solution
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return &DatabaseError{
			Code:    "SEED_PREPARE_FAILED",
			Message: "Failed to prepare insert statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	for _, ex := range exercises {
		// Convert slices to JSON strings for storage
		prerequisitesJSON, _ := json.Marshal(ex.Prerequisites)
		hintsJSON, _ := json.Marshal(ex.Hints)

		_, err := stmt.Exec(
			ex.Slug,
			ex.Title,
			ex.Description,
			ex.Category,
			ex.Difficulty,
			ex.Points,
			ex.EstimatedMinutes,
			string(prerequisitesJSON),
			string(hintsJSON),
			ex.Solution,
		)
		if err != nil {
			return &DatabaseError{
				Code:    "SEED_INSERT_FAILED",
				Message: fmt.Sprintf("Failed to insert exercise: %s", ex.Slug),
				Err:     err,
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return &DatabaseError{
			Code:    "SEED_COMMIT_FAILED",
			Message: "Failed to commit seed transaction",
			Err:     err,
		}
	}

	// Update config to mark seeding as complete
	return SetConfig("exercises_seeded", "true")
}

// GetExercises retrieves all exercises from the database
func GetExercises() ([]Exercise, error) {
	if DB == nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	rows, err := DB.Query(`
		SELECT slug, title, description, category, difficulty,
		       points, estimated_minutes, prerequisites, hints, solution
		FROM exercises
		ORDER BY category, difficulty, points
	`)
	if err != nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to query exercises",
			Err:     err,
		}
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var ex Exercise
		var prerequisitesJSON, hintsJSON string

		err := rows.Scan(
			&ex.Slug,
			&ex.Title,
			&ex.Description,
			&ex.Category,
			&ex.Difficulty,
			&ex.Points,
			&ex.EstimatedMinutes,
			&prerequisitesJSON,
			&hintsJSON,
			&ex.Solution,
		)
		if err != nil {
			return nil, &DatabaseError{
				Code:    ErrCodeQueryFailed,
				Message: "Failed to scan exercise row",
				Err:     err,
			}
		}

		// Parse JSON fields
		json.Unmarshal([]byte(prerequisitesJSON), &ex.Prerequisites)
		json.Unmarshal([]byte(hintsJSON), &ex.Hints)

		exercises = append(exercises, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Error iterating exercise rows",
			Err:     err,
		}
	}

	return exercises, nil
}

// GetExerciseBySlug retrieves a single exercise by its slug
func GetExerciseBySlug(slug string) (*Exercise, error) {
	if DB == nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	var ex Exercise
	var prerequisitesJSON, hintsJSON string

	err := DB.QueryRow(`
		SELECT slug, title, description, category, difficulty,
		       points, estimated_minutes, prerequisites, hints, solution
		FROM exercises
		WHERE slug = ?
	`, slug).Scan(
		&ex.Slug,
		&ex.Title,
		&ex.Description,
		&ex.Category,
		&ex.Difficulty,
		&ex.Points,
		&ex.EstimatedMinutes,
		&prerequisitesJSON,
		&hintsJSON,
		&ex.Solution,
	)

	if err != nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: fmt.Sprintf("Exercise not found: %s", slug),
			Err:     err,
		}
	}

	// Parse JSON fields
	json.Unmarshal([]byte(prerequisitesJSON), &ex.Prerequisites)
	json.Unmarshal([]byte(hintsJSON), &ex.Hints)

	return &ex, nil
}

// GetExercisesByCategory retrieves exercises filtered by category
func GetExercisesByCategory(category string) ([]Exercise, error) {
	if DB == nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	rows, err := DB.Query(`
		SELECT slug, title, description, category, difficulty,
		       points, estimated_minutes, prerequisites, hints, solution
		FROM exercises
		WHERE category = ?
		ORDER BY difficulty, points
	`, category)
	if err != nil {
		return nil, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to query exercises by category",
			Err:     err,
		}
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var ex Exercise
		var prerequisitesJSON, hintsJSON string

		err := rows.Scan(
			&ex.Slug,
			&ex.Title,
			&ex.Description,
			&ex.Category,
			&ex.Difficulty,
			&ex.Points,
			&ex.EstimatedMinutes,
			&prerequisitesJSON,
			&hintsJSON,
			&ex.Solution,
		)
		if err != nil {
			return nil, &DatabaseError{
				Code:    ErrCodeQueryFailed,
				Message: "Failed to scan exercise row",
				Err:     err,
			}
		}

		// Parse JSON fields
		json.Unmarshal([]byte(prerequisitesJSON), &ex.Prerequisites)
		json.Unmarshal([]byte(hintsJSON), &ex.Hints)

		exercises = append(exercises, ex)
	}

	return exercises, nil
}
