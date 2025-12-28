package bugreport

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/database"
	"github.com/patrickvassell/cks-weight-room/internal/logger"
)

// BugReport represents a complete bug report
type BugReport struct {
	GeneratedAt   string                 `json:"generatedAt"`
	Version       string                 `json:"version"`
	SystemInfo    SystemInfo             `json:"systemInfo"`
	UserReport    UserReport             `json:"userReport"`
	RecentLogs    []string               `json:"recentLogs"`
	DatabaseStats map[string]interface{} `json:"databaseStats,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"goVersion"`
	NumCPU       int    `json:"numCpu"`
	DockerStatus string `json:"dockerStatus"`
	KindStatus   string `json:"kindStatus"`
	DiskSpace    string `json:"diskSpace"`
}

// UserReport contains user-provided information
type UserReport struct {
	Description      string `json:"description"`
	ExpectedBehavior string `json:"expectedBehavior,omitempty"`
	ActualBehavior   string `json:"actualBehavior,omitempty"`
	StepsToReproduce string `json:"stepsToReproduce,omitempty"`
	Email            string `json:"email,omitempty"`
}

// Config holds bug report configuration
type Config struct {
	Version         string
	UserReport      UserReport
	MaxLogLines     int
	IncludeDBStats  bool
	OutputDir       string
}

// Generate creates a bug report and saves it as a zip file
func Generate(cfg Config) (string, error) {
	logger.Info("Generating bug report")

	// Default values
	if cfg.MaxLogLines == 0 {
		cfg.MaxLogLines = 1000
	}
	if cfg.OutputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.OutputDir = filepath.Join(homeDir, "Downloads")
	}

	// Collect bug report data
	report := BugReport{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Version:     cfg.Version,
		SystemInfo:  collectSystemInfo(),
		UserReport:  cfg.UserReport,
		RecentLogs:  collectRecentLogs(cfg.MaxLogLines),
	}

	if cfg.IncludeDBStats && database.DB != nil {
		report.DatabaseStats = collectDatabaseStats()
	}

	// Create output file
	timestamp := time.Now().Format("20060102-150405")
	reportName := fmt.Sprintf("cks-weight-room-bugreport-%s.zip", timestamp)
	reportPath := filepath.Join(cfg.OutputDir, reportName)

	logger.Info("Creating bug report file: %s", reportPath)

	// Create zip file
	zipFile, err := os.Create(reportPath)
	if err != nil {
		logger.Error("Failed to create bug report file: %v", err)
		return "", fmt.Errorf("failed to create bug report file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add bug report JSON
	if err := addJSONToZip(zipWriter, "bug-report.json", report); err != nil {
		return "", fmt.Errorf("failed to add bug report JSON: %w", err)
	}

	// Add full log file if it exists
	logPath := filepath.Join(logger.GetLogDir(), "cks-weight-room.log")
	if _, err := os.Stat(logPath); err == nil {
		if err := addFileToZip(zipWriter, "logs/cks-weight-room.log", logPath); err != nil {
			logger.Warn("Failed to add log file to report: %v", err)
		}
	}

	// Add rotated log files
	rotatedLogs, _ := filepath.Glob(filepath.Join(logger.GetLogDir(), "cks-weight-room-*.log"))
	for _, logFile := range rotatedLogs {
		fileName := filepath.Base(logFile)
		if err := addFileToZip(zipWriter, "logs/"+fileName, logFile); err != nil {
			logger.Warn("Failed to add rotated log file %s: %v", fileName, err)
		}
	}

	logger.Info("Bug report generated successfully: %s", reportPath)
	return reportPath, nil
}

// collectSystemInfo gathers system information
func collectSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
	}

	// Check Docker status
	dockerCmd := exec.Command("docker", "info")
	if err := dockerCmd.Run(); err != nil {
		info.DockerStatus = fmt.Sprintf("Not running or not installed: %v", err)
	} else {
		info.DockerStatus = "Running"
	}

	// Check KIND status
	kindCmd := exec.Command("kind", "version")
	kindOutput, err := kindCmd.CombinedOutput()
	if err != nil {
		info.KindStatus = "Not installed"
	} else {
		info.KindStatus = strings.TrimSpace(string(kindOutput))
	}

	// Get disk space
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		dfCmd := exec.Command("df", "-h", "/")
		dfOutput, err := dfCmd.CombinedOutput()
		if err == nil {
			lines := strings.Split(string(dfOutput), "\n")
			if len(lines) > 1 {
				info.DiskSpace = strings.TrimSpace(lines[1])
			}
		}
	}

	return info
}

// collectRecentLogs reads the last N lines from the log file
func collectRecentLogs(maxLines int) []string {
	logPath := filepath.Join(logger.GetLogDir(), "cks-weight-room.log")

	file, err := os.Open(logPath)
	if err != nil {
		logger.Warn("Failed to open log file for bug report: %v", err)
		return []string{"Log file not available"}
	}
	defer file.Close()

	// Read all lines into a slice
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Return last N lines
	if len(lines) <= maxLines {
		return lines
	}
	return lines[len(lines)-maxLines:]
}

// collectDatabaseStats gathers database statistics
func collectDatabaseStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count exercises
	var exerciseCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&exerciseCount); err == nil {
		stats["exerciseCount"] = exerciseCount
	}

	// Count attempts
	var attemptCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&attemptCount); err == nil {
		stats["attemptCount"] = attemptCount
	}

	// Check activation status
	var activationCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM activation").Scan(&activationCount); err == nil {
		stats["isActivated"] = activationCount > 0
	}

	// Get database file size
	dbPath := database.GetDefaultPath()
	if info, err := os.Stat(dbPath); err == nil {
		stats["databaseSizeBytes"] = info.Size()
	}

	return stats
}

// addJSONToZip adds a JSON object to the zip file
func addJSONToZip(zipWriter *zip.Writer, filename string, data interface{}) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// addFileToZip adds a file to the zip archive
func addFileToZip(zipWriter *zip.Writer, zipPath string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// GetDefaultOutputDir returns the default output directory for bug reports
func GetDefaultOutputDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(homeDir, "Downloads")
}
