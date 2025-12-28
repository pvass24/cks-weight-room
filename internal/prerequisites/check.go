package prerequisites

import (
	"fmt"
	"os/exec"
	"syscall"
)

// Error codes for prerequisite failures
const (
	DockerNotRunning      = "DOCKER_NOT_RUNNING"
	KindNotInstalled      = "KIND_NOT_INSTALLED"
	InsufficientDiskSpace = "INSUFFICIENT_DISK_SPACE"
)

// PrerequisiteError represents a structured error with code and details
type PrerequisiteError struct {
	Code    string
	Message string
	Details string
}

func (e *PrerequisiteError) Error() string {
	return e.Message
}

// CheckDocker verifies that Docker Desktop is running
func CheckDocker() error {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		return &PrerequisiteError{
			Code:    DockerNotRunning,
			Message: "Docker Desktop is not running. Please start Docker Desktop and try again.",
			Details: "https://www.docker.com/products/docker-desktop",
		}
	}
	return nil
}

// CheckKind verifies that KIND is installed
func CheckKind() error {
	cmd := exec.Command("kind", "version")
	if err := cmd.Run(); err != nil {
		return &PrerequisiteError{
			Code:    KindNotInstalled,
			Message: "KIND is not installed. Run: brew install kind (Mac) or curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64",
			Details: "https://kind.sigs.k8s.io/docs/user/quick-start/",
		}
	}
	return nil
}

// CheckDiskSpace verifies that at least 10GB of disk space is available
func CheckDiskSpace() error {
	var stat syscall.Statfs_t

	// Check disk space for current directory
	if err := syscall.Statfs(".", &stat); err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	// Calculate available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	// 10GB minimum requirement
	minRequired := uint64(10 * 1024 * 1024 * 1024)

	if availableBytes < minRequired {
		availableGB := float64(availableBytes) / (1024 * 1024 * 1024)
		return &PrerequisiteError{
			Code:    InsufficientDiskSpace,
			Message: fmt.Sprintf("Insufficient disk space. CKS Weight Room requires at least 10GB free. You have %.1fGB available.", availableGB),
			Details: "",
		}
	}

	return nil
}

// CheckResult represents the result of a single prerequisite check
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

// ValidateAll runs all prerequisite checks and returns results
func ValidateAll() ([]CheckResult, error) {
	results := []CheckResult{}
	var firstError error

	// Check Docker
	if err := CheckDocker(); err != nil {
		results = append(results, CheckResult{
			Name:    "Docker",
			Passed:  false,
			Message: err.Error(),
		})
		if firstError == nil {
			firstError = err
		}
	} else {
		results = append(results, CheckResult{
			Name:   "Docker",
			Passed: true,
		})
	}

	// Check KIND
	if err := CheckKind(); err != nil {
		results = append(results, CheckResult{
			Name:    "KIND",
			Passed:  false,
			Message: err.Error(),
		})
		if firstError == nil {
			firstError = err
		}
	} else {
		results = append(results, CheckResult{
			Name:   "KIND",
			Passed: true,
		})
	}

	// Check Disk Space
	if err := CheckDiskSpace(); err != nil {
		results = append(results, CheckResult{
			Name:    "Disk Space",
			Passed:  false,
			Message: err.Error(),
		})
		if firstError == nil {
			firstError = err
		}
	} else {
		results = append(results, CheckResult{
			Name:   "Disk Space",
			Passed: true,
		})
	}

	return results, firstError
}
