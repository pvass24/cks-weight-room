package api

import (
	"encoding/json"
	"net/http"

	"github.com/patrickvassell/cks-weight-room/internal/logger"
	"github.com/patrickvassell/cks-weight-room/internal/updater"
)

// UpdateCheckResponse represents the response from update check
type UpdateCheckResponse struct {
	Available         bool   `json:"available"`
	CurrentVersion    string `json:"currentVersion"`
	LatestVersion     string `json:"latestVersion,omitempty"`
	ReleaseNotes      string `json:"releaseNotes,omitempty"`
	DownloadURL       string `json:"downloadUrl,omitempty"`
	PublishedAt       string `json:"publishedAt,omitempty"`
	IsCritical        bool   `json:"isCritical"`
	InstallInstructions string `json:"installInstructions,omitempty"`
	Error             string `json:"error,omitempty"`
}

// CheckForUpdates handles GET /api/update/check
func CheckForUpdates(w http.ResponseWriter, r *http.Request, version string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logger.Info("Update check requested")

	// Create update checker
	checker := updater.NewChecker(updater.Config{
		CurrentVersion: version,
	})

	// Check for updates
	updateInfo, err := checker.CheckForUpdates()
	if err != nil {
		logger.Error("Update check failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UpdateCheckResponse{
			Available:      false,
			CurrentVersion: version,
			Error:          "Failed to check for updates",
		})
		return
	}

	// Build response
	response := UpdateCheckResponse{
		Available:      updateInfo.Available,
		CurrentVersion: updateInfo.CurrentVersion,
		LatestVersion:  updateInfo.LatestVersion,
		ReleaseNotes:   updateInfo.ReleaseNotes,
		DownloadURL:    updateInfo.DownloadURL,
		IsCritical:     updateInfo.IsCritical,
	}

	if !updateInfo.PublishedAt.IsZero() {
		response.PublishedAt = updateInfo.PublishedAt.Format("2006-01-02 15:04:05")
	}

	if updateInfo.Available {
		response.InstallInstructions = updater.GetInstallInstructions()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
