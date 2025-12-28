package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/patrickvassell/cks-weight-room/internal/cluster"
	"github.com/patrickvassell/cks-weight-room/internal/security"
)

const (
	terminalImageCLI   = "cks-weight-room/terminal:latest"
	maxMemoryCLI       = "512m"
	maxCPUsCLI         = "1.0"
	terminalTimeoutCLI = 2 * time.Hour
)

// SecureTerminalCLIHandler manages containerized terminal sessions using Docker CLI
type SecureTerminalCLIHandler struct {
	commandFilter *security.CommandFilter
}

// NewSecureTerminalCLIHandler creates a new secure terminal handler using Docker CLI
func NewSecureTerminalCLIHandler() (*SecureTerminalCLIHandler, error) {
	// Check if Docker is available
	if err := exec.Command("docker", "version").Run(); err != nil {
		return nil, fmt.Errorf("Docker is not available: %w", err)
	}

	// Check if terminal image exists
	checkCmd := exec.Command("docker", "images", "-q", terminalImageCLI)
	output, err := checkCmd.Output()
	if err != nil || len(output) == 0 {
		return nil, fmt.Errorf("terminal image not found - run: ./scripts/build-terminal-image.sh")
	}

	return &SecureTerminalCLIHandler{
		commandFilter: security.NewCommandFilter(),
	}, nil
}

// HandleSecureTerminalCLI manages WebSocket connections with containerized terminals
func (h *SecureTerminalCLIHandler) HandleSecureTerminalCLI(w http.ResponseWriter, r *http.Request) {
	// Extract exercise slug from path
	slug := r.URL.Path[len("/api/terminal/"):]
	if slug == "" {
		http.Error(w, "Exercise slug required", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get cluster context for this exercise
	clusterName := cluster.GetClusterName(slug)
	kubectxContext := "kind-" + clusterName

	// Create and start container
	containerID, err := h.createAndStartContainer(slug, kubectxContext)
	if err != nil {
		log.Printf("Failed to create container: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to start secure terminal: %v\r\n", err)))
		return
	}
	defer h.cleanupContainer(containerID)

	// Execute bash in container with PTY
	cmd := exec.Command("docker", "exec", "-it", containerID, "/bin/bash")
	cmd.Env = os.Environ()

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start PTY in container: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to start terminal session\r\n"))
		return
	}
	defer func() {
		ptmx.Close()
		cmd.Process.Kill()
	}()

	// Set initial terminal size
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})

	// Send initial commands
	initCommands := "alias k=kubectl\n" +
		"kubectl config use-context " + kubectxContext + " 2>/dev/null\n" +
		"clear\n" +
		"echo '\033[32m✓ Connected to CKS practice environment (Secure Mode)\033[0m'\n" +
		"echo 'Cluster: " + clusterName + "'\n" +
		"echo '\033[33mCommand filtering and resource limits active\033[0m'\n" +
		"echo ''\n"
	ptmx.Write([]byte(initCommands))

	// Copy from PTY to WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from PTY: %v", err)
				}
				return
			}
			if n > 0 {
				if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					return
				}
			}
		}
	}()

	// Buffer for command assembly
	cmdBuffer := ""

	// Copy from WebSocket to PTY with command filtering
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch msg.Type {
		case "input":
			// Sanitize input
			sanitized := h.commandFilter.SanitizeInput(msg.Data)

			// Add to buffer
			cmdBuffer += sanitized

			// Check for command execution (newline/return)
			if strings.Contains(sanitized, "\n") || strings.Contains(sanitized, "\r") {
				// Extract command (remove newline)
				cmd := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(cmdBuffer, "\n", ""), "\r", ""))
				cmdBuffer = "" // Reset buffer

				if cmd != "" {
					// Validate command
					if valid, reason := h.commandFilter.ValidateCommand(cmd); !valid {
						warningMsg := fmt.Sprintf("\r\n\033[31m⚠  Command blocked: %s\033[0m\r\n", reason)
						conn.WriteMessage(websocket.TextMessage, []byte(warningMsg))
						log.Printf("Blocked command for %s: %s (reason: %s)", slug, cmd, reason)
						continue
					}
				}
			}

			// Write to PTY
			if _, err := ptmx.Write([]byte(sanitized)); err != nil {
				log.Printf("Error writing to PTY: %v", err)
				return
			}

		case "resize":
			if msg.Rows > 0 && msg.Cols > 0 {
				ws := &pty.Winsize{
					Rows: uint16(msg.Rows),
					Cols: uint16(msg.Cols),
				}
				if err := pty.Setsize(ptmx, ws); err != nil {
					log.Printf("Error resizing PTY: %v", err)
				}
			}
		}
	}
}

// createAndStartContainer creates and starts a container with security constraints
func (h *SecureTerminalCLIHandler) createAndStartContainer(slug, kubectxContext string) (string, error) {
	// Get kubeconfig path
	kubeconfigPath := os.Getenv("HOME") + "/.kube/config"

	// Container name
	containerName := fmt.Sprintf("cks-terminal-%s-%d", slug, time.Now().Unix())

	// Docker run command with security options
	args := []string{
		"run",
		"-d",                           // Detached
		"--name", containerName,        // Container name
		"--rm",                         // Auto-remove
		"--memory", maxMemoryCLI,       // Memory limit
		"--cpus", maxCPUsCLI,           // CPU limit
		"--read-only",                  // Read-only root filesystem
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=100m",
		"--tmpfs", "/home/cksuser:rw,noexec,nosuid,size=50m",
		"--security-opt", "no-new-privileges:true", // No privilege escalation
		"--cap-drop", "ALL",            // Drop all capabilities
		"--cap-add", "NET_RAW",         // Add only ping capability
		"-v", kubeconfigPath + ":/tmp/.kube/config:ro", // Mount kubeconfig read-only
		"-e", "TERM=xterm-256color",
		"-e", "KUBECONFIG=/tmp/.kube/config",
		"-e", "KUBECTL_CONTEXT=" + kubectxContext,
		"-w", "/home/cksuser",
		terminalImageCLI,
		"sleep", "infinity", // Keep container running
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w - %s", err, string(output))
	}

	// Get container ID from output
	containerID := strings.TrimSpace(string(output))

	// Wait for container to be running
	time.Sleep(500 * time.Millisecond)

	return containerID, nil
}

// cleanupContainer stops and removes the container
func (h *SecureTerminalCLIHandler) cleanupContainer(containerID string) {
	if containerID == "" {
		return
	}

	// Stop container
	stopCmd := exec.Command("docker", "stop", "-t", "5", containerID)
	stopCmd.Run() // Ignore errors, container might already be stopped

	// Remove container (if not auto-removed)
	rmCmd := exec.Command("docker", "rm", "-f", containerID)
	rmCmd.Run() // Ignore errors, container might already be removed
}

// checkDockerAvailable checks if Docker is installed and running
func checkDockerAvailable() error {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}
	return nil
}

// checkTerminalImage checks if the terminal image exists
func checkTerminalImage() error {
	cmd := exec.Command("docker", "images", "-q", terminalImageCLI)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check for terminal image: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		return fmt.Errorf("terminal image not found - run: ./scripts/build-terminal-image.sh")
	}

	return nil
}
