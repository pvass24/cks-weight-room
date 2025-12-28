package cluster

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/logger"
)

// ClusterStatus represents the current state of a cluster
type ClusterStatus string

const (
	StatusProvisioning ClusterStatus = "provisioning"
	StatusReady        ClusterStatus = "ready"
	StatusError        ClusterStatus = "error"
	StatusNotFound     ClusterStatus = "not_found"
)

// Cluster represents a KIND cluster for an exercise
type Cluster struct {
	Name          string        `json:"name"`
	ExerciseSlug  string        `json:"exerciseSlug"`
	Status        ClusterStatus `json:"status"`
	CreatedAt     time.Time     `json:"createdAt"`
	ErrorMessage  string        `json:"errorMessage,omitempty"`
	KubeconfigCtx string        `json:"kubeconfigContext,omitempty"`
}

// ClusterError represents a cluster operation error
type ClusterError struct {
	Code    string
	Message string
	Err     error
}

func (e *ClusterError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeDockerNotRunning = "DOCKER_NOT_RUNNING"
	ErrCodeKindNotInstalled = "KIND_NOT_INSTALLED"
	ErrCodeProvisionFailed  = "PROVISION_FAILED"
	ErrCodeDeleteFailed     = "DELETE_FAILED"
	ErrCodeGetStatusFailed  = "GET_STATUS_FAILED"
)

// CheckDocker verifies Docker is running
func CheckDocker(ctx context.Context) error {
	logger.Debug("Checking Docker Desktop status...")
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		logger.Warn("Docker Desktop is not running: %v", err)
		return &ClusterError{
			Code:    ErrCodeDockerNotRunning,
			Message: "Docker Desktop is not running. Please start Docker Desktop and try again.",
			Err:     err,
		}
	}
	logger.Debug("Docker Desktop is running")
	return nil
}

// CheckKind verifies KIND is installed
func CheckKind(ctx context.Context) error {
	logger.Debug("Checking KIND installation...")
	cmd := exec.CommandContext(ctx, "kind", "version")
	if err := cmd.Run(); err != nil {
		logger.Warn("KIND is not installed: %v", err)
		return &ClusterError{
			Code:    ErrCodeKindNotInstalled,
			Message: "KIND is not installed. Install with: brew install kind (macOS) or see https://kind.sigs.k8s.io/",
			Err:     err,
		}
	}
	logger.Debug("KIND is installed")
	return nil
}

// GetClusterName generates a cluster name for an exercise
func GetClusterName(exerciseSlug string) string {
	return fmt.Sprintf("cks-%s", exerciseSlug)
}

// ClusterExists checks if a KIND cluster exists
func ClusterExists(ctx context.Context, clusterName string) (bool, error) {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return false, &ClusterError{
			Code:    ErrCodeGetStatusFailed,
			Message: "Failed to get cluster list",
			Err:     err,
		}
	}

	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if cluster == clusterName {
			return true, nil
		}
	}
	return false, nil
}

// ProvisionCluster creates a new KIND cluster for an exercise
// This is a simplified version - in production would use KIND's Go API
func ProvisionCluster(ctx context.Context, exerciseSlug string, progressChan chan<- string) (*Cluster, error) {
	clusterName := GetClusterName(exerciseSlug)
	logger.Info("Starting cluster provisioning for exercise: %s (cluster: %s)", exerciseSlug, clusterName)

	cluster := &Cluster{
		Name:         clusterName,
		ExerciseSlug: exerciseSlug,
		Status:       StatusProvisioning,
		CreatedAt:    time.Now(),
	}

	// Check prerequisites
	if progressChan != nil {
		progressChan <- "Checking Docker Desktop status..."
	}
	if err := CheckDocker(ctx); err != nil {
		cluster.Status = StatusError
		cluster.ErrorMessage = err.Error()
		return cluster, err
	}

	if progressChan != nil {
		progressChan <- "Checking KIND installation..."
	}
	if err := CheckKind(ctx); err != nil {
		cluster.Status = StatusError
		cluster.ErrorMessage = err.Error()
		return cluster, err
	}

	// Check if cluster already exists
	logger.Debug("Checking if cluster already exists: %s", clusterName)
	exists, err := ClusterExists(ctx, clusterName)
	if err != nil {
		logger.Error("Failed to check cluster existence: %v", err)
		cluster.Status = StatusError
		cluster.ErrorMessage = err.Error()
		return cluster, err
	}

	if exists {
		logger.Info("Cluster %s already exists, reusing existing cluster", clusterName)
		if progressChan != nil {
			progressChan <- fmt.Sprintf("Cluster %s already exists, using existing cluster...", clusterName)
		}
		cluster.Status = StatusReady
		cluster.KubeconfigCtx = fmt.Sprintf("kind-%s", clusterName)
		return cluster, nil
	}

	// Create cluster
	logger.Info("Creating new KIND cluster: %s", clusterName)
	if progressChan != nil {
		progressChan <- fmt.Sprintf("Creating KIND cluster (%s)...", clusterName)
	}

	// Create cluster with 1 control plane + 2 workers (CKS exam environment)
	cmd := exec.CommandContext(ctx, "kind", "create", "cluster",
		"--name", clusterName,
		"--config", "-",
	)

	// KIND cluster config matching CKS exam environment
	kindConfig := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
`
	cmd.Stdin = strings.NewReader(kindConfig)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to create cluster %s: %v (output: %s)", clusterName, err, string(output))
		cluster.Status = StatusError
		cluster.ErrorMessage = fmt.Sprintf("Failed to create cluster: %s", string(output))
		return cluster, &ClusterError{
			Code:    ErrCodeProvisionFailed,
			Message: string(output),
			Err:     err,
		}
	}

	logger.Info("Successfully created cluster: %s", clusterName)
	if progressChan != nil {
		progressChan <- "Cluster created successfully!"
	}

	cluster.Status = StatusReady
	cluster.KubeconfigCtx = fmt.Sprintf("kind-%s", clusterName)

	return cluster, nil
}

// DeleteCluster removes a KIND cluster
func DeleteCluster(ctx context.Context, clusterName string) error {
	logger.Info("Deleting cluster: %s", clusterName)
	cmd := exec.CommandContext(ctx, "kind", "delete", "cluster", "--name", clusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to delete cluster %s: %v (output: %s)", clusterName, err, string(output))
		return &ClusterError{
			Code:    ErrCodeDeleteFailed,
			Message: fmt.Sprintf("Failed to delete cluster: %s", string(output)),
			Err:     err,
		}
	}
	logger.Info("Successfully deleted cluster: %s", clusterName)
	return nil
}

// GetClusterStatus gets the current status of a cluster
func GetClusterStatus(ctx context.Context, clusterName string) (*Cluster, error) {
	exists, err := ClusterExists(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	if !exists {
		return &Cluster{
			Name:   clusterName,
			Status: StatusNotFound,
		}, nil
	}

	// TODO: Could add more detailed health checks here
	// For now, if cluster exists, assume it's ready
	return &Cluster{
		Name:          clusterName,
		Status:        StatusReady,
		KubeconfigCtx: fmt.Sprintf("kind-%s", clusterName),
	}, nil
}
