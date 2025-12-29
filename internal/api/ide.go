package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/security"
)

// IDESession represents a code-server session for a node
type IDESession struct {
	NodeName      string
	Port          int       // Unique port for this code-server instance
	ContainerIP   string    // Container's IP address
	ProcessPID    int       // PID of code-server process
	StartedAt     time.Time
	LastAccess    time.Time
}

// IDEHandler manages code-server sessions
type IDEHandler struct {
	sessions      map[string]*IDESession // key: "slug-nodeName"
	mu            sync.RWMutex
	commandFilter *security.CommandFilter
	nextPort      int       // Next available port (starts at 8081)
}

// NewIDEHandler creates a new IDE handler
func NewIDEHandler() *IDEHandler {
	handler := &IDEHandler{
		sessions:      make(map[string]*IDESession),
		commandFilter: security.NewCommandFilter(),
		nextPort:      8081, // Start allocating ports from 8081
	}

	// Start cleanup goroutine for idle sessions
	go handler.cleanupIdleSessions()

	return handler
}

// allocatePort returns the next available port and increments the counter
// NOTE: Caller must already hold h.mu lock
func (h *IDEHandler) allocatePort() int {
	port := h.nextPort
	h.nextPort++
	return port
}

// HandleIDEProxy proxies HTTP requests to code-server in KIND node
// Route: /api/ide/{slug}?node={nodeName}
func (h *IDEHandler) HandleIDEProxy(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path
	pathAfterPrefix := r.URL.Path[len("/api/ide/"):]

	// Split path to get slug (first segment after /api/ide/)
	var slug string
	if idx := strings.Index(pathAfterPrefix, "/"); idx != -1 {
		slug = pathAfterPrefix[:idx]
	} else {
		slug = pathAfterPrefix
	}

	if slug == "" {
		http.Error(w, "Exercise slug required", http.StatusBadRequest)
		return
	}

	// Get node from query param (optional for resource requests)
	nodeName := r.URL.Query().Get("node")

	// If node not specified, try to find existing session for this slug
	if nodeName == "" {
		h.mu.RLock()
		// Look for any session that starts with "{slug}-"
		prefix := slug + "-"
		for key, sess := range h.sessions {
			if strings.HasPrefix(key, prefix) {
				nodeName = sess.NodeName
				log.Printf("Found existing session for slug=%s, using node=%s", slug, nodeName)
				break
			}
		}
		h.mu.RUnlock()

		// If still no node found, this is an error
		if nodeName == "" {
			http.Error(w, "node parameter required (no existing session found)", http.StatusBadRequest)
			return
		}
	}

	log.Printf("IDE proxy request for slug=%s, node=%s, path=%s", slug, nodeName, r.URL.Path)

	// Get or create session
	session, err := h.getOrCreateSession(slug, nodeName)
	if err != nil {
		log.Printf("Failed to start IDE session: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start IDE session: %v", err), http.StatusInternalServerError)
		return
	}

	// Update last access time
	h.mu.Lock()
	session.LastAccess = time.Now()
	h.mu.Unlock()

	// Proxy request to code-server
	h.proxyToCodeServer(w, r, session, slug)
}

// getOrCreateSession gets existing session or starts new code-server
func (h *IDEHandler) getOrCreateSession(slug, nodeName string) (*IDESession, error) {
	sessionKey := fmt.Sprintf("%s-%s", slug, nodeName)

	h.mu.RLock()
	session, exists := h.sessions[sessionKey]
	h.mu.RUnlock()

	if exists && h.isSessionHealthy(session) {
		log.Printf("Reusing existing IDE session for %s", sessionKey)
		return session, nil
	}

	// Create new session
	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock
	if session, exists := h.sessions[sessionKey]; exists && h.isSessionHealthy(session) {
		return session, nil
	}

	log.Printf("Creating new IDE session for %s", sessionKey)

	// Start code-server in KIND node
	log.Printf("About to call startCodeServer for slug=%s, node=%s", slug, nodeName)
	session, err := h.startCodeServer(slug, nodeName)
	if err != nil {
		log.Printf("startCodeServer failed: %v", err)
		return nil, err
	}
	log.Printf("startCodeServer succeeded, got session with port %d", session.Port)

	h.sessions[sessionKey] = session
	return session, nil
}

// startCodeServer launches code-server inside KIND node container
func (h *IDEHandler) startCodeServer(slug, nodeName string) (*IDESession, error) {
	log.Printf("[DEBUG] startCodeServer called for slug=%s, node=%s", slug, nodeName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Allocate a unique port for this code-server instance
	port := h.allocatePort()
	bindAddr := fmt.Sprintf("0.0.0.0:%d", port)

	log.Printf("Starting code-server in node: %s on port %d", nodeName, port)

	// First check if the container exists
	checkCmd := exec.CommandContext(ctx, "docker", "ps", "--filter", fmt.Sprintf("name=%s", nodeName), "--format", "{{.Names}}")
	output, err := checkCmd.Output()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return nil, fmt.Errorf("KIND node container '%s' not found", nodeName)
	}

	// Start code-server in background inside container with unique port
	// Each session gets its own port (8081, 8082, etc.)
	startCmd := exec.CommandContext(ctx, "docker", "exec", "-d", nodeName,
		"code-server",
		"--bind-addr", bindAddr,
		"--auth", "none",
		"--disable-telemetry",
		"/root")

	output, err = startCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start code-server: %w - %s", err, string(output))
	}

	log.Printf("code-server started on port %d, output: %s", port, string(output))

	// Wait for code-server to be ready
	log.Printf("Waiting for code-server to be ready on port %d...", port)
	testURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		testCmd := exec.CommandContext(ctx, "docker", "exec", nodeName, "curl", "-s", testURL)
		if testCmd.Run() == nil {
			log.Printf("code-server is ready on port %d!", port)
			break
		}
	}

	// Get container's IP address
	log.Printf("Getting container IP address for %s", nodeName)
	ipCmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", nodeName)
	ipOutput, err := ipCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container IP: %w", err)
	}
	containerIP := strings.TrimSpace(string(ipOutput))
	if containerIP == "" {
		return nil, fmt.Errorf("container IP address is empty")
	}
	log.Printf("Container %s IP: %s", nodeName, containerIP)

	// On macOS, container IPs aren't accessible from host
	// Start a lightweight proxy container to bridge the connection
	localPort := 9000 + (port - 8080) // Map 8081->9001, 8082->9002, etc.
	proxyName := fmt.Sprintf("ide-proxy-%s-%d", slug, port)
	target := fmt.Sprintf("%s:%d", containerIP, port)

	log.Printf("Starting proxy container %s: localhost:%d -> %s", proxyName, localPort, target)

	// Remove existing proxy container if it exists (from previous session)
	exec.Command("docker", "rm", "-f", proxyName).Run() // Ignore errors if container doesn't exist

	// Use alpine/socat to proxy from host port to container IP:port
	proxyCmd := exec.CommandContext(context.Background(), "docker", "run", "-d",
		"--name", proxyName,
		"--rm",
		"--network", "kind",
		"-p", fmt.Sprintf("127.0.0.1:%d:%d", localPort, localPort),
		"alpine/socat",
		fmt.Sprintf("TCP-LISTEN:%d,fork,reuseaddr", localPort),
		fmt.Sprintf("TCP:%s", target))

	output, err = proxyCmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: failed to start proxy container: %v - %s", err, string(output))
		log.Printf("Attempting direct connection to %s", target)
	} else {
		log.Printf("Proxy container started: %s", strings.TrimSpace(string(output)))
		time.Sleep(500 * time.Millisecond) // Give proxy time to start
	}

	session := &IDESession{
		NodeName:    nodeName,
		Port:        localPort, // Use local port exposed by proxy container
		ContainerIP: "127.0.0.1",
		StartedAt:   time.Now(),
		LastAccess:  time.Now(),
	}

	return session, nil
}

// isSessionHealthy checks if code-server process is still running
func (h *IDEHandler) isSessionHealthy(session *IDESession) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if code-server process is running
	checkCmd := exec.CommandContext(ctx, "docker", "exec", session.NodeName, "pgrep", "-f", "code-server")
	if err := checkCmd.Run(); err != nil {
		log.Printf("Session unhealthy for %s: code-server not running", session.NodeName)
		return false
	}

	return true
}

// proxyToCodeServer proxies HTTP/WebSocket requests to code-server
func (h *IDEHandler) proxyToCodeServer(w http.ResponseWriter, r *http.Request, session *IDESession, slug string) {
	// Construct target URL - proxy directly to container IP:port
	// Each code-server runs on its own unique port
	targetURL := fmt.Sprintf("http://%s:%d", session.ContainerIP, session.Port)
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Failed to parse target URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Proxying IDE request to %s (node: %s, port: %d)", targetURL, session.NodeName, session.Port)

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize director to preserve WebSocket headers and strip path prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Strip /api/ide/{slug} prefix from path before forwarding to code-server
		// code-server expects requests at root path
		prefix := "/api/ide/" + slug
		if strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
		}

		// Set required headers for code-server proxy
		req.Header.Set("Host", req.Host)

		// Preserve WebSocket upgrade headers
		if upgrade := r.Header.Get("Upgrade"); upgrade != "" {
			req.Header.Set("Upgrade", upgrade)
		}
		if connection := r.Header.Get("Connection"); connection != "" {
			req.Header.Set("Connection", connection)
		}
		if wsKey := r.Header.Get("Sec-WebSocket-Key"); wsKey != "" {
			req.Header.Set("Sec-WebSocket-Key", wsKey)
		}
		if wsVersion := r.Header.Get("Sec-WebSocket-Version"); wsVersion != "" {
			req.Header.Set("Sec-WebSocket-Version", wsVersion)
		}
		if wsProtocol := r.Header.Get("Sec-WebSocket-Protocol"); wsProtocol != "" {
			req.Header.Set("Sec-WebSocket-Protocol", wsProtocol)
		}
	}

	// Rewrite redirect Location headers to preserve the /api/ide/{slug} prefix
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove restrictive CSP headers that prevent code-server from loading in iframe
		if csp := resp.Header.Get("Content-Security-Policy"); csp != "" {
			log.Printf("Stripping CSP header: %s", csp)
			resp.Header.Del("Content-Security-Policy")
		}
		resp.Header.Del("Content-Security-Policy-Report-Only")
		resp.Header.Del("X-Frame-Options") // Allow iframe embedding

		// For HTML responses, inject a <base> tag to fix absolute path resolution
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				resp.Body.Close()

				// Inject <base href="/api/ide/{slug}/?node={nodeName}"> after <head> tag
				// Include node parameter so all resource requests inherit it
				baseTag := fmt.Sprintf(`<base href="/api/ide/%s/?node=%s">`, slug, session.NodeName)
				modifiedBody := strings.Replace(string(body), "<head>", "<head>"+baseTag, 1)

				// Update response body and Content-Length
				resp.Body = io.NopCloser(strings.NewReader(modifiedBody))
				resp.ContentLength = int64(len(modifiedBody))
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

				log.Printf("Injected base tag with node parameter into HTML response")
			}
		}

		// Check if this is a redirect
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			if location := resp.Header.Get("Location"); location != "" {
				// If location is relative (starts with . or doesn't have http://), rewrite it
				if strings.HasPrefix(location, "./") || strings.HasPrefix(location, "/") || !strings.HasPrefix(location, "http") {
					// Prepend /api/ide/{slug} to make it absolute
					prefix := "/api/ide/" + slug
					if strings.HasPrefix(location, "./") {
						// ./foo -> /api/ide/{slug}/foo
						location = prefix + "/" + strings.TrimPrefix(location, "./")
					} else if strings.HasPrefix(location, "/") {
						// /foo -> /api/ide/{slug}/foo
						location = prefix + location
					} else {
						// foo -> /api/ide/{slug}/foo
						location = prefix + "/" + location
					}
					resp.Header.Set("Location", location)
					log.Printf("Rewrote redirect location to: %s", location)
				}
			}
		}
		return nil
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", session.NodeName, err)
		http.Error(w, fmt.Sprintf("Failed to proxy to code-server: %v", err), http.StatusBadGateway)
	}

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

// stopSession terminates code-server process in container
func (h *IDEHandler) stopSession(session *IDESession) {
	log.Printf("Stopping IDE session for %s on port %d", session.NodeName, session.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kill the specific code-server instance running on this port
	killCmd := fmt.Sprintf("pkill -f 'code-server.*:%d'", session.Port)
	exec.CommandContext(ctx, "docker", "exec", session.NodeName, "sh", "-c", killCmd).Run()
}

// cleanupIdleSessions removes sessions idle for > 30 minutes
func (h *IDEHandler) cleanupIdleSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		for key, session := range h.sessions {
			if time.Since(session.LastAccess) > 30*time.Minute {
				log.Printf("Cleaning up idle session: %s", key)
				h.stopSession(session)
				delete(h.sessions, key)
			}
		}
		h.mu.Unlock()
	}
}

// CleanupClusterSessions removes all sessions for a specific cluster (slug)
func (h *IDEHandler) CleanupClusterSessions(slug string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	prefix := slug + "-"
	for key, session := range h.sessions {
		if strings.HasPrefix(key, prefix) {
			log.Printf("Cleaning up session for deleted cluster: %s", key)
			h.stopSession(session)
			delete(h.sessions, key)
		}
	}
}
