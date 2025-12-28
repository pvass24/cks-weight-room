package api

import (
	"encoding/json"
	"net/http"

	"github.com/patrickvassell/cks-weight-room/internal/bugreport"
	"github.com/patrickvassell/cks-weight-room/internal/logger"
)

// BugReportRequest represents a bug report submission request
type BugReportRequest struct {
	Description      string `json:"description"`
	ExpectedBehavior string `json:"expectedBehavior,omitempty"`
	ActualBehavior   string `json:"actualBehavior,omitempty"`
	StepsToReproduce string `json:"stepsToReproduce,omitempty"`
	Email            string `json:"email,omitempty"`
	IncludeLogs      bool   `json:"includeLogs"`
	IncludeDBStats   bool   `json:"includeDbStats"`
}

// BugReportResponse represents the response from bug report generation
type BugReportResponse struct {
	Success  bool   `json:"success"`
	FilePath string `json:"filePath,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// SubmitBugReport handles POST /api/bugreport/submit
func SubmitBugReport(w http.ResponseWriter, r *http.Request, version string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BugReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Invalid bug report request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BugReportResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate description is provided
	if req.Description == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BugReportResponse{
			Success: false,
			Error:   "Description is required",
		})
		return
	}

	logger.Info("Bug report submission received: %s", truncate(req.Description, 50))

	// Generate bug report
	config := bugreport.Config{
		Version: version,
		UserReport: bugreport.UserReport{
			Description:      req.Description,
			ExpectedBehavior: req.ExpectedBehavior,
			ActualBehavior:   req.ActualBehavior,
			StepsToReproduce: req.StepsToReproduce,
			Email:            req.Email,
		},
		MaxLogLines:    1000,
		IncludeDBStats: req.IncludeDBStats,
	}

	filePath, err := bugreport.Generate(config)
	if err != nil {
		logger.Error("Failed to generate bug report: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BugReportResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	logger.Info("Bug report generated successfully: %s", filePath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BugReportResponse{
		Success:  true,
		FilePath: filePath,
		Message:  "Bug report generated successfully. Please send this file to support@cks-weight-room.com",
	})
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
