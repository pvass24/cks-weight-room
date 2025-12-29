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

	// Get node parameter from query string (optional)
	nodeName := r.URL.Query().Get("node")

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get cluster context for this exercise
	clusterName := cluster.GetClusterName(slug)

	// Use docker exec for ALL nodes (control-plane and workers)
	// This provides better isolation - each node only sees its own cluster context
	if nodeName == "" {
		// If no node specified, default to control plane
		// Find the control plane node name
		nodes, err := cluster.GetClusterNodes(r.Context(), clusterName)
		if err != nil {
			log.Printf("Failed to get cluster nodes: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to get cluster nodes: %v\r\n", err)))
			return
		}
		for _, node := range nodes {
			if node.Role == "control-plane" {
				nodeName = node.Name
				break
			}
		}
		if nodeName == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("No control plane node found\r\n"))
			return
		}
	}

	// Connect to the specified node (control-plane or worker)
	log.Printf("Connecting to node: %s", nodeName)
	h.handleWorkerNodeTerminal(conn, nodeName, slug)
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
		"--network", "host",            // Use host network so localhost works for KIND clusters
		"--memory", maxMemoryCLI,       // Memory limit
		"--cpus", maxCPUsCLI,           // CPU limit
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=100m",
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

// handleWorkerNodeTerminal connects to a worker node's KIND container directly
func (h *SecureTerminalCLIHandler) handleWorkerNodeTerminal(conn *websocket.Conn, nodeName, slug string) {
	log.Printf("Attempting to connect to worker node container: %s", nodeName)

	// First check if the container exists
	checkCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", nodeName), "--format", "{{.Names}}")
	output, err := checkCmd.Output()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		errMsg := fmt.Sprintf("KIND node container '%s' not found. Make sure the cluster is running.\r\n", nodeName)
		log.Printf("Container check failed: %v (output: %s)", err, string(output))
		conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return
	}
	log.Printf("Container found: %s", strings.TrimSpace(string(output)))

	// Execute interactive bash directly in the KIND node container
	// Use -it flags to allocate a proper TTY inside the container
	// This enables readline (history/up arrow) and proper terminal behavior
	cmd := exec.Command("docker", "exec", "-it", "-e", "TERM=xterm-256color", nodeName, "/bin/bash")
	cmd.Env = os.Environ()

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start PTY in node %s: %v", nodeName, err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to connect to node %s: %v\r\n", nodeName, err)))
		return
	}
	log.Printf("Successfully started PTY for node %s", nodeName)
	defer func() {
		ptmx.Close()
		cmd.Process.Kill()
	}()

	// Set initial terminal size
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})

	// Start copying from PTY to WebSocket BEFORE sending init commands
	// so we don't miss any output
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

	// Wait for bash to be fully ready
	time.Sleep(500 * time.Millisecond)

	// Send init commands in stages to ensure they're processed
	// First, disable all echo/verbose modes
	ptmx.Write([]byte("set +v +x +o verbose +o xtrace 2>/dev/null\n"))
	time.Sleep(100 * time.Millisecond)

	// Then set up aliases and prompt
	ptmx.Write([]byte("shopt -s expand_aliases; alias k=kubectl; export PS1='\\u@\\h:\\w\\$ '\n"))
	time.Sleep(100 * time.Millisecond)

	// Finally, clear the screen to hide init output
	ptmx.Write([]byte("clear\n"))
	time.Sleep(100 * time.Millisecond)

	// Copy from WebSocket to PTY (with command filtering)
	cmdBuffer := ""
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
					// Validate command (same filtering as secure container)
					if valid, reason := h.commandFilter.ValidateCommand(cmd); !valid {
						// Send newline to PTY so prompt advances
						ptmx.Write([]byte("\r\n"))
						// Show warning to user
						warningMsg := fmt.Sprintf("\033[31m⚠  Command blocked: %s\033[0m\r\n", reason)
						conn.WriteMessage(websocket.TextMessage, []byte(warningMsg))
						log.Printf("Blocked command on node %s for %s: %s (reason: %s)", nodeName, slug, cmd, reason)
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
