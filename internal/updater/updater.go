package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/logger"
)

// UpdateInfo represents information about an available update
type UpdateInfo struct {
	Available       bool      `json:"available"`
	CurrentVersion  string    `json:"currentVersion"`
	LatestVersion   string    `json:"latestVersion"`
	ReleaseNotes    string    `json:"releaseNotes,omitempty"`
	DownloadURL     string    `json:"downloadUrl,omitempty"`
	PublishedAt     time.Time `json:"publishedAt,omitempty"`
	IsCritical      bool      `json:"isCritical"`
	MinimumRequired string    `json:"minimumRequired,omitempty"`
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Config holds update checker configuration
type Config struct {
	CurrentVersion string
	GitHubOwner    string
	GitHubRepo     string
	CheckInterval  time.Duration
	HTTPTimeout    time.Duration
}

// Checker handles update checking
type Checker struct {
	config Config
	client *http.Client
}

// NewChecker creates a new update checker
func NewChecker(config Config) *Checker {
	// Default values
	if config.GitHubOwner == "" {
		config.GitHubOwner = "patrickvassell"
	}
	if config.GitHubRepo == "" {
		config.GitHubRepo = "cks-weight-room"
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 24 * time.Hour
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 10 * time.Second
	}

	return &Checker{
		config: config,
		client: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

// CheckForUpdates checks GitHub for the latest release
func (c *Checker) CheckForUpdates() (*UpdateInfo, error) {
	logger.Info("Checking for updates (current version: %s)", c.config.CurrentVersion)

	// Fetch latest release from GitHub
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		c.config.GitHubOwner, c.config.GitHubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create update check request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent (GitHub API requires it)
	req.Header.Set("User-Agent", fmt.Sprintf("CKS-Weight-Room/%s", c.config.CurrentVersion))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Warn("Failed to check for updates (network error): %v", err)
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: c.config.CurrentVersion,
		}, nil // Return no update available on network error
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		logger.Debug("No releases found on GitHub")
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: c.config.CurrentVersion,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warn("GitHub API returned status %d", resp.StatusCode)
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: c.config.CurrentVersion,
		}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read update response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		logger.Error("Failed to parse GitHub release: %v", err)
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Compare versions
	latestVersion := normalizeVersion(release.TagName)
	currentVersion := normalizeVersion(c.config.CurrentVersion)

	updateAvailable := isNewerVersion(latestVersion, currentVersion)

	updateInfo := &UpdateInfo{
		Available:      updateAvailable,
		CurrentVersion: c.config.CurrentVersion,
		LatestVersion:  release.TagName,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
	}

	// Find download URL for current platform
	downloadURL := c.findDownloadURL(release)
	if downloadURL != "" {
		updateInfo.DownloadURL = downloadURL
	}

	// Check if update is critical (look for [CRITICAL] in release notes)
	if strings.Contains(strings.ToUpper(release.Body), "[CRITICAL]") ||
		strings.Contains(strings.ToUpper(release.Body), "SECURITY") {
		updateInfo.IsCritical = true
	}

	if updateAvailable {
		logger.Info("Update available: %s -> %s", c.config.CurrentVersion, release.TagName)
	} else {
		logger.Info("Application is up to date")
	}

	return updateInfo, nil
}

// findDownloadURL finds the appropriate download URL for the current platform
func (c *Checker) findDownloadURL(release GitHubRelease) string {
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Common naming patterns
	patterns := []string{
		fmt.Sprintf("cks-weight-room-%s-%s", platform, arch),
		fmt.Sprintf("cks-weight-room_%s_%s", platform, arch),
		fmt.Sprintf("%s-%s", platform, arch),
	}

	for _, asset := range release.Assets {
		assetName := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			if strings.Contains(assetName, pattern) {
				return asset.BrowserDownloadURL
			}
		}
	}

	return ""
}

// normalizeVersion removes 'v' prefix and cleans version string
func normalizeVersion(version string) string {
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimSpace(version)
	return version
}

// isNewerVersion compares two version strings
// Returns true if latest is newer than current
func isNewerVersion(latest, current string) bool {
	// Handle "dev" version
	if current == "dev" {
		return latest != "dev" && latest != ""
	}

	// Simple semantic version comparison
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Pad to same length
	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for len(latestParts) < maxLen {
		latestParts = append(latestParts, "0")
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}

	// Compare each part
	for i := 0; i < maxLen; i++ {
		var latestNum, currentNum int
		fmt.Sscanf(latestParts[i], "%d", &latestNum)
		fmt.Sscanf(currentParts[i], "%d", &currentNum)

		if latestNum > currentNum {
			return true
		}
		if latestNum < currentNum {
			return false
		}
	}

	return false
}

// GetInstallInstructions returns platform-specific install instructions
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return `To update via Homebrew:
  brew upgrade cks-weight-room

To update manually:
  1. Download the latest release from the URL above
  2. Replace the existing binary in /usr/local/bin/
  3. Restart the application`

	case "linux":
		return `To update:
  1. Download the latest release from the URL above
  2. Extract the archive
  3. Replace the existing binary
  4. Restart the application`

	case "windows":
		return `To update:
  1. Download the latest release from the URL above
  2. Close CKS Weight Room
  3. Replace the existing .exe file
  4. Start the application`

	default:
		return "Download the latest release from the URL above and replace your current installation"
	}
}
