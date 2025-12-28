package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/patrickvassell/cks-weight-room/internal/cluster"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost only
		return true
	},
}

// TerminalMessage represents messages sent/received over WebSocket
type TerminalMessage struct {
	Type string `json:"type"` // "input", "resize"
	Data string `json:"data,omitempty"`
	Rows int    `json:"rows,omitempty"`
	Cols int    `json:"cols,omitempty"`
}

// HandleTerminal manages WebSocket connections for interactive terminal sessions
func HandleTerminal(w http.ResponseWriter, r *http.Request) {
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

	// Start shell session with PTY
	cmd := exec.Command("/bin/bash")
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"KUBECONFIG="+os.Getenv("HOME")+"/.kube/config",
	)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start PTY: %v", err)
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

	// Send initial commands to set up kubectl context
	initCommands := "alias k=kubectl\n" +
		"kubectl config use-context " + kubectxContext + " 2>/dev/null\n" +
		"clear\n" +
		"echo 'Connected to CKS practice environment'\n" +
		"echo 'Cluster: " + clusterName + "'\n" +
		"echo ''\n" +
		"kubectl get nodes 2>/dev/null || echo 'Cluster is starting up...'\n" +
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

	// Copy from WebSocket to PTY
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
			if _, err := ptmx.Write([]byte(msg.Data)); err != nil {
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

// setWinsize sets the size of the given PTY
func setWinsize(fd uintptr, w, h uint16) error {
	ws := &struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}{
		Row: h,
		Col: w,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
