package cluster

import (
	"context"
	"fmt"
	"os"
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

	// Install code-server and bashrc in all nodes
	if progressChan != nil {
		progressChan <- "Installing code-server and configuring nodes..."
	}

	nodes, err := GetClusterNodes(ctx, clusterName)
	if err != nil {
		logger.Warn("Failed to get cluster nodes: %v", err)
	} else {
		for _, node := range nodes {
			// Install code-server
			if err := InstallCodeServerInNode(ctx, node.Name); err != nil {
				logger.Warn("Failed to install code-server in %s: %v", node.Name, err)
				// Don't fail provisioning if code-server install fails
			} else {
				logger.Info("Successfully installed code-server in %s", node.Name)
			}

			// Install CKS-style .bashrc
			if err := InstallBashrcInNode(ctx, node.Name); err != nil {
				logger.Warn("Failed to install .bashrc in %s: %v", node.Name, err)
			} else {
				logger.Info("Successfully installed .bashrc in %s", node.Name)
			}
		}
	}

	if progressChan != nil {
		progressChan <- "Code-server installation complete!"
	}

	// Run exercise-specific setup
	if progressChan != nil {
		progressChan <- "Setting up exercise environment..."
	}
	if err := SetupExercise(ctx, exerciseSlug, clusterName); err != nil {
		logger.Warn("Failed to setup exercise environment: %v", err)
		// Don't fail provisioning if exercise setup fails
	} else {
		logger.Info("Exercise environment setup complete")
	}

	if progressChan != nil {
		progressChan <- "Exercise setup complete!"
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

// Node represents a node in the cluster
type Node struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// GetClusterNodes returns the list of nodes in a cluster
func GetClusterNodes(ctx context.Context, clusterName string) ([]Node, error) {
	kubectxContext := fmt.Sprintf("kind-%s", clusterName)
	cmd := exec.CommandContext(ctx, "kubectl", "get", "nodes",
		"--context", kubectxContext,
		"--no-headers",
		"-o", "custom-columns=NAME:.metadata.name",
	)

	output, err := cmd.Output()
	if err != nil {
		logger.Error("Failed to get nodes for cluster %s: %v", clusterName, err)
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var nodes []Node
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		nodeName := strings.TrimSpace(line)
		if nodeName == "" {
			continue
		}

		// Determine role from node name (KIND naming convention)
		role := "worker"
		if strings.Contains(nodeName, "control-plane") {
			role = "control-plane"
		}

		nodes = append(nodes, Node{
			Name: nodeName,
			Role: role,
		})
	}

	logger.Info("Found %d nodes in cluster %s", len(nodes), clusterName)
	for _, node := range nodes {
		logger.Debug("Node: %s (role: %s)", node.Name, node.Role)
	}

	return nodes, nil
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

// InstallCodeServerInNode installs code-server in a KIND node container
func InstallCodeServerInNode(ctx context.Context, nodeName string) error {
	logger.Info("Installing code-server in node: %s", nodeName)

	// Check if already installed
	checkCmd := exec.CommandContext(ctx, "docker", "exec", nodeName, "which", "code-server")
	if checkCmd.Run() == nil {
		logger.Debug("code-server already installed in %s", nodeName)
		return nil
	}

	// Install code-server
	logger.Info("Installing code-server (this may take 1-2 minutes)...")
	installScript := `
		curl -fsSL https://code-server.dev/install.sh | sh -s -- --version=4.22.1 && \
		mkdir -p /root/.config/code-server
	`

	cmd := exec.CommandContext(ctx, "docker", "exec", nodeName, "bash", "-c", installScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install code-server: %w - %s", err, string(output))
	}

	logger.Debug("code-server install output: %s", string(output))

	// Install socat for port forwarding
	logger.Info("Installing socat in %s...", nodeName)
	socatCmd := exec.CommandContext(ctx, "docker", "exec", nodeName, "bash", "-c",
		"apt-get update -qq && apt-get install -y -qq socat")
	socatOutput, err := socatCmd.CombinedOutput()
	if err != nil {
		logger.Warn("Failed to install socat in %s: %v - %s", nodeName, err, string(socatOutput))
		// Don't fail here - socat might already be installed
	}

	// Install curl if not present (needed for health checks)
	curlCmd := exec.CommandContext(ctx, "docker", "exec", nodeName, "bash", "-c",
		"which curl || (apt-get update -qq && apt-get install -y -qq curl)")
	curlCmd.Run() // Ignore errors

	logger.Info("Successfully installed code-server in %s", nodeName)
	return nil
}

// SetupExercise runs exercise-specific setup scripts and manifests
func SetupExercise(ctx context.Context, exerciseSlug, clusterName string) error {
	setupDir := fmt.Sprintf("internal/exercises/setups/%s", exerciseSlug)
	
	// Check if setup directory exists
	if _, err := os.Stat(setupDir); os.IsNotExist(err) {
		logger.Debug("No setup directory found for exercise: %s", exerciseSlug)
		return nil // Not an error - exercise might not need setup
	}

	logger.Info("Running setup for exercise: %s", exerciseSlug)
	kubectxContext := fmt.Sprintf("kind-%s", clusterName)

	// Apply Kubernetes manifests if they exist
	manifestPath := fmt.Sprintf("%s/deployments.yaml", setupDir)
	if _, err := os.Stat(manifestPath); err == nil {
		logger.Info("Applying Kubernetes manifests...")
		cmd := exec.CommandContext(ctx, "kubectl", "apply",
			"-f", manifestPath,
			"--context", kubectxContext)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to apply manifests: %w - %s", err, string(output))
		}
		logger.Debug("Manifest apply output: %s", string(output))
	}

	// Run setup script if it exists
	setupScript := fmt.Sprintf("%s/setup.sh", setupDir)
	if _, err := os.Stat(setupScript); err == nil {
		logger.Info("Running setup script on all nodes...")

		// Get all cluster nodes
		nodes, err := GetClusterNodes(ctx, clusterName)
		if err != nil {
			return fmt.Errorf("failed to get cluster nodes: %w", err)
		}

		// Make script executable
		os.Chmod(setupScript, 0755)

		// Run setup script on each node (for tools like Falco that need to run on all nodes)
		for _, node := range nodes {
			logger.Info("Running setup script on node: %s (%s)", node.Name, node.Role)

			cmd := exec.CommandContext(ctx, "bash", setupScript, node.Name)
			cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s/.kube/config", os.Getenv("HOME")))
			output, err := cmd.CombinedOutput()
			if err != nil {
				logger.Warn("Setup script failed on node %s: %v - %s", node.Name, err, string(output))
				// Continue with other nodes even if one fails
				continue
			}
			logger.Debug("Setup script output for %s: %s", node.Name, string(output))
		}
	}

	logger.Info("Exercise setup completed successfully")
	return nil
}

// InstallBashrcInNode installs CKS exam-like .bashrc in a KIND node
func InstallBashrcInNode(ctx context.Context, nodeName string) error {
	logger.Info("Installing bash-completion and .bashrc in node: %s", nodeName)

	// Install bash-completion package
	logger.Debug("Installing bash-completion package...")
	installCmd := exec.CommandContext(ctx, "docker", "exec", nodeName, "bash", "-c",
		"apt-get update -qq && apt-get install -y -qq bash-completion")
	installOutput, err := installCmd.CombinedOutput()
	if err != nil {
		logger.Warn("Failed to install bash-completion in %s: %v - %s", nodeName, err, string(installOutput))
		// Continue anyway - might already be installed
	}

	// Read bashrc template
	bashrcContent, err := os.ReadFile("internal/cluster/bashrc-template.sh")
	if err != nil {
		return fmt.Errorf("failed to read bashrc template: %w", err)
	}

	// Copy bashrc to node
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", nodeName, "bash", "-c", "cat > /root/.bashrc")
	cmd.Stdin = strings.NewReader(string(bashrcContent))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install bashrc: %w - %s", err, string(output))
	}

	logger.Debug("Successfully installed bash-completion and .bashrc in %s", nodeName)
	return nil
}
