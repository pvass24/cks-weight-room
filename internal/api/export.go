package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// ExportData represents all exportable progress data
type ExportData struct {
	ExportDate              string              `json:"export_date"`
	TotalPracticeTimeMinutes int                 `json:"total_practice_time_minutes"`
	ScenariosCompleted      int                 `json:"scenarios_completed"`
	Attempts                []ExportAttempt     `json:"attempts"`
	PersonalBests           []ExportPersonalBest `json:"personal_bests"`
	MockExams               []ExportMockExam    `json:"mock_exams"`
}

// ExportAttempt represents a single attempt for export
type ExportAttempt struct {
	AttemptID             int     `json:"attempt_id"`
	ScenarioID            int     `json:"scenario_id"`
	ScenarioName          string  `json:"scenario_name"`
	Timestamp             string  `json:"timestamp"`
	CompletionTimeSeconds int     `json:"completion_time_seconds"`
	Score                 float64 `json:"score"`
	MaxScore              int     `json:"max_score"`
	Status                string  `json:"status"`
	Feedback              string  `json:"feedback,omitempty"`
}

// ExportPersonalBest represents a personal best for export
type ExportPersonalBest struct {
	ScenarioID      int    `json:"scenario_id"`
	ScenarioName    string `json:"scenario_name"`
	BestTimeSeconds int    `json:"best_time_seconds"`
	AchievedAt      string `json:"achieved_at"`
}

// ExportMockExam represents a mock exam for export
type ExportMockExam struct {
	ExamID           int     `json:"exam_id"`
	ExamType         string  `json:"exam_type"`
	Timestamp        string  `json:"timestamp"`
	TotalTimeSeconds int     `json:"total_time_seconds"`
	OverallScore     float64 `json:"overall_score"`
	Result           string  `json:"result"`
}

// GetExportData handles GET /api/export
func GetExportData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	exportData := ExportData{
		ExportDate: time.Now().UTC().Format(time.RFC3339),
		Attempts:   []ExportAttempt{},
		PersonalBests: []ExportPersonalBest{},
		MockExams:  []ExportMockExam{},
	}

	// Get total practice time
	var totalSeconds sql.NullInt64
	database.DB.QueryRow("SELECT COALESCE(SUM(duration_seconds), 0) FROM attempts").Scan(&totalSeconds)
	if totalSeconds.Valid {
		exportData.TotalPracticeTimeMinutes = int(totalSeconds.Int64 / 60)
	}

	// Get scenarios completed
	database.DB.QueryRow("SELECT COUNT(*) FROM progress WHERE status = 'completed'").Scan(&exportData.ScenariosCompleted)

	// Get all attempts
	rows, err := database.DB.Query(`
		SELECT
			a.id,
			a.exercise_id,
			e.title,
			COALESCE(a.completed_at, a.started_at) as timestamp,
			COALESCE(a.duration_seconds, 0) as duration,
			CAST(a.score AS FLOAT) / CAST(a.max_score AS FLOAT) as score_ratio,
			a.max_score,
			CASE WHEN a.passed = 1 THEN 'completed' ELSE 'failed' END as status,
			COALESCE(a.feedback, '') as feedback
		FROM attempts a
		JOIN exercises e ON a.exercise_id = e.id
		ORDER BY a.id
	`)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var attempt ExportAttempt
			var timestamp sql.NullString
			rows.Scan(
				&attempt.AttemptID,
				&attempt.ScenarioID,
				&attempt.ScenarioName,
				&timestamp,
				&attempt.CompletionTimeSeconds,
				&attempt.Score,
				&attempt.MaxScore,
				&attempt.Status,
				&attempt.Feedback,
			)
			if timestamp.Valid {
				attempt.Timestamp = timestamp.String
			}
			exportData.Attempts = append(exportData.Attempts, attempt)
		}
	}

	// Get personal bests
	rows, err = database.DB.Query(`
		SELECT
			p.exercise_id,
			e.title,
			p.personal_best_seconds,
			COALESCE(p.completed_at, '') as achieved_at
		FROM progress p
		JOIN exercises e ON p.exercise_id = e.id
		WHERE p.personal_best_seconds IS NOT NULL
		ORDER BY p.personal_best_seconds
	`)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pb ExportPersonalBest
			var achievedAt sql.NullString
			rows.Scan(
				&pb.ScenarioID,
				&pb.ScenarioName,
				&pb.BestTimeSeconds,
				&achievedAt,
			)
			if achievedAt.Valid {
				pb.AchievedAt = achievedAt.String
			}
			exportData.PersonalBests = append(exportData.PersonalBests, pb)
		}
	}

	// Get mock exams
	rows, err = database.DB.Query(`
		SELECT
			id,
			exam_type,
			COALESCE(completed_at, started_at) as timestamp,
			COALESCE(total_duration_seconds, 0) as duration,
			CAST(overall_score AS FLOAT) / CAST(max_score AS FLOAT) as score_ratio,
			CASE WHEN passed = 1 THEN 'passed' ELSE 'failed' END as result
		FROM mock_exams
		ORDER BY id
	`)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var exam ExportMockExam
			var timestamp sql.NullString
			rows.Scan(
				&exam.ExamID,
				&exam.ExamType,
				&timestamp,
				&exam.TotalTimeSeconds,
				&exam.OverallScore,
				&exam.Result,
			)
			if timestamp.Valid {
				exam.Timestamp = timestamp.String
			}
			exportData.MockExams = append(exportData.MockExams, exam)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exportData)
}
