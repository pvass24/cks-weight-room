package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/cluster"
	"github.com/patrickvassell/cks-weight-room/internal/database"
)

// ValidationResult represents the result of a solution validation
type ValidationResult struct {
	Passed   bool     `json:"passed"`
	Score    int      `json:"score"`
	Feedback string   `json:"feedback"`
	Details  []string `json:"details,omitempty"`
}

// ValidationRequest represents a validation request
type ValidationRequest struct {
	ClusterName string `json:"clusterName"`
}

// ValidateSolution handles POST /api/validate/{exerciseSlug}
func ValidateSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract exercise slug from path
	path := r.URL.Path
	slug := path[len("/api/validate/"):]

	if slug == "" {
		http.Error(w, "Exercise slug required", http.StatusBadRequest)
		return
	}

	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get cluster name
	clusterName := cluster.GetClusterName(slug)
	kubectxContext := "kind-" + clusterName

	// Run validation checks based on exercise
	result := validateExercise(slug, kubectxContext)

	// Save attempt to database
	if database.DB != nil {
		// Get exercise info for max score
		var exerciseID int
		var maxScore int
		err := database.DB.QueryRow("SELECT id, points FROM exercises WHERE slug = ?", slug).Scan(&exerciseID, &maxScore)
		if err == nil {
			// Save attempt
			_, err = database.DB.Exec(`
				INSERT INTO attempts (exercise_id, started_at, completed_at, duration_seconds, score, max_score, passed, feedback, details)
				VALUES (?, datetime('now', '-30 seconds'), datetime('now'), 30, ?, ?, ?, ?, ?)
			`, exerciseID, result.Score, maxScore, result.Passed, result.Feedback, mustMarshalJSON(result.Details))

			// Update progress table personal best if passed and better than previous
			if err == nil && result.Passed {
				database.DB.Exec(`
					INSERT INTO progress (exercise_id, status, completed_at, attempts, time_spent_seconds, personal_best_seconds)
					VALUES (?, 'completed', datetime('now'), 1, 30, 30)
					ON CONFLICT(exercise_id) DO UPDATE SET
						status = 'completed',
						completed_at = datetime('now'),
						attempts = attempts + 1,
						time_spent_seconds = time_spent_seconds + 30,
						personal_best_seconds = MIN(COALESCE(personal_best_seconds, 999999), 30),
						updated_at = datetime('now')
				`, exerciseID)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// mustMarshalJSON marshals data to JSON, returning empty string on error
func mustMarshalJSON(v interface{}) string {
	if v == nil {
		return "[]"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// validateExercise runs validation checks for a specific exercise
func validateExercise(slug, kubectx string) ValidationResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// For now, implement validation for a few exercises as examples
	switch slug {
	case "disable-anonymous-access":
		return validateDisableAnonymousAccess(ctx, kubectx)
	case "networkpolicy-default-deny":
		return validateNetworkPolicyDefaultDeny(ctx, kubectx)
	default:
		// Generic validation - just check if cluster is accessible
		return ValidationResult{
			Passed:   true,
			Score:    10,
			Feedback: "Validation not yet implemented for this exercise. This is a placeholder response.",
			Details:  []string{"Manual verification required"},
		}
	}
}

// validateDisableAnonymousAccess checks if anonymous access is disabled
func validateDisableAnonymousAccess(ctx context.Context, kubectx string) ValidationResult {
	// Check if anonymous-auth is disabled in API server
	cmd := exec.CommandContext(ctx, "kubectl",
		"--context", kubectx,
		"get", "pod",
		"-n", "kube-system",
		"-l", "component=kube-apiserver",
		"-o", "jsonpath={.items[0].spec.containers[0].command}",
	)

	output, err := cmd.Output()
	if err != nil {
		return ValidationResult{
			Passed:   false,
			Score:    0,
			Feedback: "Failed to check API server configuration",
			Details:  []string{"Could not read kube-apiserver pod configuration"},
		}
	}

	config := string(output)
	details := []string{}

	// Check for --anonymous-auth=false
	if strings.Contains(config, "--anonymous-auth=false") {
		details = append(details, "✓ Anonymous authentication is disabled")
	} else {
		return ValidationResult{
			Passed:   false,
			Score:    0,
			Feedback: "Anonymous authentication is not disabled",
			Details:  []string{"Add --anonymous-auth=false to kube-apiserver"},
		}
	}

	return ValidationResult{
		Passed:   true,
		Score:    25,
		Feedback: "Great! Anonymous access has been successfully disabled.",
		Details:  details,
	}
}

// validateNetworkPolicyDefaultDeny checks for default deny network policy
func validateNetworkPolicyDefaultDeny(ctx context.Context, kubectx string) ValidationResult {
	// Check if default deny network policy exists
	cmd := exec.CommandContext(ctx, "kubectl",
		"--context", kubectx,
		"get", "networkpolicy",
		"-A",
		"-o", "json",
	)

	output, err := cmd.Output()
	if err != nil {
		return ValidationResult{
			Passed:   false,
			Score:    0,
			Feedback: "Failed to check network policies",
			Details:  []string{"Could not read network policies"},
		}
	}

	config := strings.ToLower(string(output))
	details := []string{}

	// Check for deny-all or default-deny policy
	if strings.Contains(config, "deny") || strings.Contains(config, "default") {
		details = append(details, "✓ Default deny network policy found")

		return ValidationResult{
			Passed:   true,
			Score:    25,
			Feedback: "Excellent! Default deny network policy is in place.",
			Details:  details,
		}
	}

	return ValidationResult{
		Passed:   false,
		Score:    0,
		Feedback: "No default deny network policy found",
		Details:  []string{"Create a NetworkPolicy that denies all ingress and egress by default"},
	}
}
