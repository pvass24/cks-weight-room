package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/patrickvassell/cks-weight-room/internal/api"
	"github.com/patrickvassell/cks-weight-room/internal/database"
	"github.com/patrickvassell/cks-weight-room/internal/logger"
)

// Version information (set via ldflags at build time)
var version = "dev"

//go:embed all:web/out
var webFS embed.FS

func main() {
	// Command line flags
	versionFlag := flag.Bool("version", false, "Display version information")
	portFlag := flag.String("port", "3000", "Server port (default: 3000)")
	flag.Parse()

	// Handle --version flag
	if *versionFlag {
		fmt.Printf("CKS Weight Room v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Initialize logger
	logLevel := logger.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.LevelDebug
	}
	if err := logger.Init(logger.Config{
		Level:    logLevel,
		MaxSize:  10, // 10MB
		MaxFiles: 5,
	}); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("CKS Weight Room v%s starting (%s/%s)", version, runtime.GOOS, runtime.GOARCH)
	logger.Info("Log directory: %s", logger.GetLogDir())

	// Connect to database if it exists
	dbPath := database.GetDefaultPath()
	if database.IsInitialized(dbPath) {
		logger.Debug("Database path: %s", dbPath)
		if err := database.Connect(database.Config{Path: dbPath}); err != nil {
			logger.Error("Failed to connect to database: %v", err)
		} else {
			logger.Info("Connected to existing database")

			// Apply any pending migrations
			if err := database.ApplyMigrations(); err != nil {
				logger.Error("Failed to apply migrations: %v", err)
			} else {
				logger.Debug("Database migrations applied successfully")
			}
		}
	} else {
		logger.Info("Database not yet initialized (will be created on first setup)")
	}

	// Serve embedded frontend
	staticFS, err := fs.Sub(webFS, "web/out")
	if err != nil {
		log.Fatalf("Failed to load embedded frontend: %v", err)
	}

	// Setup HTTP server
	// API routes
	http.HandleFunc("/api/setup/validate", api.ValidatePrerequisites)
	http.HandleFunc("/api/setup/initialize", api.InitializeDatabase)
	http.HandleFunc("/api/setup/db-status", api.GetDatabaseStatus)
	http.HandleFunc("/api/exercises", api.GetExercises)
	http.HandleFunc("/api/exercises/", api.GetExerciseBySlug)
	http.HandleFunc("/api/admin/seed", api.SeedExercises)

	// Cluster management routes
	http.HandleFunc("/api/cluster/provision", api.ProvisionCluster)
	http.HandleFunc("/api/cluster/status/", api.GetClusterStatus)
	http.HandleFunc("/api/cluster/", api.DeleteCluster)

	// Terminal WebSocket route - use secure mode if enabled
	if os.Getenv("SECURE_TERMINAL") == "true" {
		logger.Info("Secure terminal mode enabled")
		secureHandler, err := api.NewSecureTerminalCLIHandler()
		if err != nil {
			logger.Error("Failed to initialize secure terminal handler: %v", err)
			logger.Warn("Falling back to standard terminal mode")
			http.HandleFunc("/api/terminal/", api.HandleTerminal)
		} else {
			http.HandleFunc("/api/terminal/", secureHandler.HandleSecureTerminalCLI)
			logger.Info("Using containerized terminal sessions with command filtering")
		}
	} else {
		logger.Warn("Standard terminal mode (running directly on host)")
		logger.Warn("For better security, set SECURE_TERMINAL=true and build terminal image")
		http.HandleFunc("/api/terminal/", api.HandleTerminal)
	}

	// Validation route
	http.HandleFunc("/api/validate/", api.ValidateSolution)

	// Progress statistics route
	http.HandleFunc("/api/progress/stats", api.GetProgressStats)

	// Analytics route
	http.HandleFunc("/api/analytics", api.GetAnalytics)

	// Export route
	http.HandleFunc("/api/export", api.GetExportData)

	// Reset routes
	http.HandleFunc("/api/reset/stats", api.GetResetStats)
	http.HandleFunc("/api/reset", api.ResetProgress)

	// Activation routes
	http.HandleFunc("/api/activation/machine-id", api.GetMachineID)
	http.HandleFunc("/api/activation/status", api.GetActivationStatus)
	http.HandleFunc("/api/activation/activate", api.ActivateLicense)
	http.HandleFunc("/api/activation/activate-offline", api.ActivateOffline)
	http.HandleFunc("/api/activation/validate", api.ValidateActivation)

	// Bug report route
	http.HandleFunc("/api/bugreport/submit", func(w http.ResponseWriter, r *http.Request) {
		api.SubmitBugReport(w, r, version)
	})

	// Update check route
	http.HandleFunc("/api/update/check", func(w http.ResponseWriter, r *http.Request) {
		api.CheckForUpdates(w, r, version)
	})

	// Static file server (must be last)
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	addr := fmt.Sprintf("127.0.0.1:%s", *portFlag)
	fmt.Printf("CKS Weight Room v%s starting on http://%s\n", version, addr)
	fmt.Println("Press Ctrl+C to stop")

	logger.Info("Starting HTTP server on %s", addr)
	logger.Debug("Server bound to localhost only (NFR-S1)")

	// Start server (localhost-only binding as per NFR-S1)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error("Server failed: %v", err)
		log.Fatalf("Server failed: %v", err)
	}
}
