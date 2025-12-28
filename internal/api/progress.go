package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// ProgressStats represents overall progress statistics
type ProgressStats struct {
	ScenariosCompleted   int               `json:"scenariosCompleted"`
	TotalScenarios       int               `json:"totalScenarios"`
	CompletionPercentage float64           `json:"completionPercentage"`
	TotalPracticeMinutes int               `json:"totalPracticeMinutes"`
	AverageScore         float64           `json:"averageScore"`
	MockExamsTaken       int               `json:"mockExamsTaken"`
	MockExamsPassed      int               `json:"mockExamsPassed"`
	ProgressByDomain     []DomainProgress  `json:"progressByDomain"`
	RecentActivity       []RecentAttempt   `json:"recentActivity"`
}

// DomainProgress represents progress for a single CKS domain
type DomainProgress struct {
	Domain              string  `json:"domain"`
	DisplayName         string  `json:"displayName"`
	Weight              int     `json:"weight"`
	CompletedCount      int     `json:"completedCount"`
	TotalCount          int     `json:"totalCount"`
	CompletionPercentage float64 `json:"completionPercentage"`
}

// RecentAttempt represents a recent practice attempt
type RecentAttempt struct {
	ExerciseSlug string `json:"exerciseSlug"`
	ExerciseTitle string `json:"exerciseTitle"`
	CompletedAt  string `json:"completedAt"`
	Duration     int    `json:"durationSeconds"`
	Score        int    `json:"score"`
	MaxScore     int    `json:"maxScore"`
	Passed       bool   `json:"passed"`
	IsPersonalBest bool  `json:"isPersonalBest"`
}

// GetProgressStats handles GET /api/progress/stats
func GetProgressStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	stats := ProgressStats{
		ProgressByDomain: []DomainProgress{},
		RecentActivity:   []RecentAttempt{},
	}

	// Get total scenarios count
	err := database.DB.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&stats.TotalScenarios)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to get total scenarios", http.StatusInternalServerError)
		return
	}

	// Get completed scenarios count (from progress table)
	err = database.DB.QueryRow("SELECT COUNT(*) FROM progress WHERE status = 'completed'").Scan(&stats.ScenariosCompleted)
	if err != nil && err != sql.ErrNoRows {
		stats.ScenariosCompleted = 0
	}

	// Calculate completion percentage
	if stats.TotalScenarios > 0 {
		stats.CompletionPercentage = float64(stats.ScenariosCompleted) / float64(stats.TotalScenarios) * 100
	}

	// Get total practice time (sum of all attempts)
	var totalSeconds sql.NullInt64
	err = database.DB.QueryRow("SELECT COALESCE(SUM(duration_seconds), 0) FROM attempts").Scan(&totalSeconds)
	if err == nil && totalSeconds.Valid {
		stats.TotalPracticeMinutes = int(totalSeconds.Int64 / 60)
	}

	// Get average score from attempts
	var avgScore sql.NullFloat64
	err = database.DB.QueryRow(`
		SELECT AVG(CAST(score AS FLOAT) / CAST(max_score AS FLOAT) * 100)
		FROM attempts
		WHERE max_score > 0
	`).Scan(&avgScore)
	if err == nil && avgScore.Valid {
		stats.AverageScore = avgScore.Float64
	}

	// Get mock exams stats
	database.DB.QueryRow("SELECT COUNT(*) FROM mock_exams").Scan(&stats.MockExamsTaken)
	database.DB.QueryRow("SELECT COUNT(*) FROM mock_exams WHERE passed = 1").Scan(&stats.MockExamsPassed)

	// Get progress by domain
	domains := map[string]struct {
		displayName string
		weight      int
	}{
		"cluster-setup":                       {"Cluster Setup", 10},
		"cluster-hardening":                   {"Cluster Hardening", 15},
		"system-hardening":                    {"System Hardening", 15},
		"minimize-microservice-vulnerabilities": {"Minimize Microservice Vulnerabilities", 20},
		"supply-chain-security":               {"Supply Chain Security", 20},
		"monitoring-logging-runtime-security": {"Monitoring, Logging & Runtime Security", 20},
	}

	for domain, info := range domains {
		var totalCount, completedCount int

		// Get total count for this domain
		database.DB.QueryRow("SELECT COUNT(*) FROM exercises WHERE category = ?", domain).Scan(&totalCount)

		// Get completed count for this domain
		database.DB.QueryRow(`
			SELECT COUNT(*)
			FROM progress p
			JOIN exercises e ON p.exercise_id = e.id
			WHERE e.category = ? AND p.status = 'completed'
		`, domain).Scan(&completedCount)

		percentage := 0.0
		if totalCount > 0 {
			percentage = float64(completedCount) / float64(totalCount) * 100
		}

		stats.ProgressByDomain = append(stats.ProgressByDomain, DomainProgress{
			Domain:              domain,
			DisplayName:         info.displayName,
			Weight:              info.weight,
			CompletedCount:      completedCount,
			TotalCount:          totalCount,
			CompletionPercentage: percentage,
		})
	}

	// Get recent activity (last 5 attempts)
	rows, err := database.DB.Query(`
		SELECT
			e.slug,
			e.title,
			a.completed_at,
			a.duration_seconds,
			a.score,
			a.max_score,
			a.passed,
			CASE
				WHEN a.duration_seconds = p.personal_best_seconds THEN 1
				ELSE 0
			END as is_personal_best
		FROM attempts a
		JOIN exercises e ON a.exercise_id = e.id
		LEFT JOIN progress p ON a.exercise_id = p.exercise_id
		WHERE a.completed_at IS NOT NULL
		ORDER BY a.completed_at DESC
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var attempt RecentAttempt
			var completedAt sql.NullString
			err := rows.Scan(
				&attempt.ExerciseSlug,
				&attempt.ExerciseTitle,
				&completedAt,
				&attempt.Duration,
				&attempt.Score,
				&attempt.MaxScore,
				&attempt.Passed,
				&attempt.IsPersonalBest,
			)
			if err == nil && completedAt.Valid {
				attempt.CompletedAt = completedAt.String
				stats.RecentActivity = append(stats.RecentActivity, attempt)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
