package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// AnalyticsData represents comprehensive analytics data
type AnalyticsData struct {
	TotalPracticeSeconds   int                    `json:"totalPracticeSeconds"`
	ScenariosCompleted     int                    `json:"scenariosCompleted"`
	TotalScenarios         int                    `json:"totalScenarios"`
	AverageCompletionTime  int                    `json:"averageCompletionTime"` // in seconds
	AverageScore           float64                `json:"averageScore"`
	PersonalBestsSet       int                    `json:"personalBestsSet"`
	MockExamsTaken         int                    `json:"mockExamsTaken"`
	MockExamsPassed        int                    `json:"mockExamsPassed"`
	ProgressByDomain       []DetailedDomain       `json:"progressByDomain"`
	PersonalBests          []PersonalBest         `json:"personalBests"`
	PracticeTimeBreakdown  PracticeTimeBreakdown  `json:"practiceTimeBreakdown"`
}

// DetailedDomain represents progress for a domain with individual scenarios
type DetailedDomain struct {
	Domain              string             `json:"domain"`
	DisplayName         string             `json:"displayName"`
	Weight              int                `json:"weight"`
	CompletedCount      int                `json:"completedCount"`
	TotalCount          int                `json:"totalCount"`
	CompletionPercentage float64           `json:"completionPercentage"`
	Scenarios           []ScenarioProgress `json:"scenarios"`
}

// ScenarioProgress represents progress for a single scenario
type ScenarioProgress struct {
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	Difficulty      string `json:"difficulty"`
	PersonalBest    int    `json:"personalBest"` // in seconds, 0 if not started
	Attempts        int    `json:"attempts"`
	LastPracticed   string `json:"lastPracticed"` // ISO timestamp or empty
	Status          string `json:"status"` // "not-started", "in-progress", "completed"
}

// PersonalBest represents a personal best time for a scenario
type PersonalBest struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Domain        string `json:"domain"`
	DomainDisplay string `json:"domainDisplay"`
	Difficulty    string `json:"difficulty"`
	PersonalBest  int    `json:"personalBest"` // in seconds
	Attempts      int    `json:"attempts"`
	LastPracticed string `json:"lastPracticed"`
}

// PracticeTimeBreakdown represents practice time statistics
type PracticeTimeBreakdown struct {
	ThisWeekSeconds     int `json:"thisWeekSeconds"`
	ThisMonthSeconds    int `json:"thisMonthSeconds"`
	AllTimeSeconds      int `json:"allTimeSeconds"`
	AverageSessionTime  int `json:"averageSessionTime"` // in seconds
	LongestSessionTime  int `json:"longestSessionTime"` // in seconds
}

// GetAnalytics handles GET /api/analytics
func GetAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	data := AnalyticsData{
		ProgressByDomain: []DetailedDomain{},
		PersonalBests:    []PersonalBest{},
	}

	// Get total scenarios count
	database.DB.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&data.TotalScenarios)

	// Get completed scenarios count
	database.DB.QueryRow("SELECT COUNT(*) FROM progress WHERE status = 'completed'").Scan(&data.ScenariosCompleted)

	// Get total practice time (sum of all attempts)
	var totalSeconds sql.NullInt64
	database.DB.QueryRow("SELECT COALESCE(SUM(duration_seconds), 0) FROM attempts").Scan(&totalSeconds)
	if totalSeconds.Valid {
		data.TotalPracticeSeconds = int(totalSeconds.Int64)
	}

	// Get average completion time (only for passed attempts)
	var avgTime sql.NullFloat64
	database.DB.QueryRow(`
		SELECT AVG(duration_seconds)
		FROM attempts
		WHERE passed = 1 AND duration_seconds > 0
	`).Scan(&avgTime)
	if avgTime.Valid {
		data.AverageCompletionTime = int(avgTime.Float64)
	}

	// Get average score
	var avgScore sql.NullFloat64
	database.DB.QueryRow(`
		SELECT AVG(CAST(score AS FLOAT) / CAST(max_score AS FLOAT) * 100)
		FROM attempts
		WHERE max_score > 0
	`).Scan(&avgScore)
	if avgScore.Valid {
		data.AverageScore = avgScore.Float64
	}

	// Get personal bests count
	database.DB.QueryRow("SELECT COUNT(*) FROM progress WHERE personal_best_seconds IS NOT NULL").Scan(&data.PersonalBestsSet)

	// Get mock exams stats
	database.DB.QueryRow("SELECT COUNT(*) FROM mock_exams").Scan(&data.MockExamsTaken)
	database.DB.QueryRow("SELECT COUNT(*) FROM mock_exams WHERE passed = 1").Scan(&data.MockExamsPassed)

	// Get detailed progress by domain
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
		detailedDomain := DetailedDomain{
			Domain:      domain,
			DisplayName: info.displayName,
			Weight:      info.weight,
			Scenarios:   []ScenarioProgress{},
		}

		// Get all scenarios for this domain
		rows, err := database.DB.Query(`
			SELECT
				e.slug,
				e.title,
				e.difficulty,
				COALESCE(p.personal_best_seconds, 0) as personal_best,
				COALESCE(p.attempts, 0) as attempts,
				COALESCE(p.completed_at, '') as last_practiced,
				COALESCE(p.status, 'not-started') as status
			FROM exercises e
			LEFT JOIN progress p ON e.id = p.exercise_id
			WHERE e.category = ?
			ORDER BY e.id
		`, domain)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var scenario ScenarioProgress
				var lastPracticed sql.NullString
				rows.Scan(
					&scenario.Slug,
					&scenario.Title,
					&scenario.Difficulty,
					&scenario.PersonalBest,
					&scenario.Attempts,
					&lastPracticed,
					&scenario.Status,
				)
				if lastPracticed.Valid {
					scenario.LastPracticed = lastPracticed.String
				}
				detailedDomain.Scenarios = append(detailedDomain.Scenarios, scenario)
				detailedDomain.TotalCount++
				if scenario.Status == "completed" {
					detailedDomain.CompletedCount++
				}
			}
		}

		if detailedDomain.TotalCount > 0 {
			detailedDomain.CompletionPercentage = float64(detailedDomain.CompletedCount) / float64(detailedDomain.TotalCount) * 100
		}

		data.ProgressByDomain = append(data.ProgressByDomain, detailedDomain)
	}

	// Get personal bests
	rows, err := database.DB.Query(`
		SELECT
			e.slug,
			e.title,
			e.category,
			e.difficulty,
			p.personal_best_seconds,
			p.attempts,
			COALESCE(p.completed_at, '') as last_practiced
		FROM progress p
		JOIN exercises e ON p.exercise_id = e.id
		WHERE p.personal_best_seconds IS NOT NULL
		ORDER BY p.personal_best_seconds ASC
	`)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pb PersonalBest
			var lastPracticed sql.NullString
			var domain string
			rows.Scan(
				&pb.Slug,
				&pb.Title,
				&domain,
				&pb.Difficulty,
				&pb.PersonalBest,
				&pb.Attempts,
				&lastPracticed,
			)
			pb.Domain = domain
			if info, ok := domains[domain]; ok {
				pb.DomainDisplay = info.displayName
			}
			if lastPracticed.Valid {
				pb.LastPracticed = lastPracticed.String
			}
			data.PersonalBests = append(data.PersonalBests, pb)
		}
	}

	// Practice time breakdown
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	oneMonthAgo := now.AddDate(0, -1, 0)

	// This week
	var thisWeek sql.NullInt64
	database.DB.QueryRow(`
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM attempts
		WHERE datetime(completed_at) >= datetime(?)
	`, oneWeekAgo.Format("2006-01-02 15:04:05")).Scan(&thisWeek)
	if thisWeek.Valid {
		data.PracticeTimeBreakdown.ThisWeekSeconds = int(thisWeek.Int64)
	}

	// This month
	var thisMonth sql.NullInt64
	database.DB.QueryRow(`
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM attempts
		WHERE datetime(completed_at) >= datetime(?)
	`, oneMonthAgo.Format("2006-01-02 15:04:05")).Scan(&thisMonth)
	if thisMonth.Valid {
		data.PracticeTimeBreakdown.ThisMonthSeconds = int(thisMonth.Int64)
	}

	// All time
	data.PracticeTimeBreakdown.AllTimeSeconds = data.TotalPracticeSeconds

	// Average session time
	var avgSession sql.NullFloat64
	database.DB.QueryRow("SELECT AVG(duration_seconds) FROM attempts WHERE duration_seconds > 0").Scan(&avgSession)
	if avgSession.Valid {
		data.PracticeTimeBreakdown.AverageSessionTime = int(avgSession.Float64)
	}

	// Longest session
	var longest sql.NullInt64
	database.DB.QueryRow("SELECT MAX(duration_seconds) FROM attempts").Scan(&longest)
	if longest.Valid {
		data.PracticeTimeBreakdown.LongestSessionTime = int(longest.Int64)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
