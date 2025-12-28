package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/patrickvassell/cks-weight-room/internal/logger"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// DB is the global database connection
var DB *sql.DB

// Config holds database configuration
type Config struct {
	Path string
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Code    string
	Message string
	Err     error
}

func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeInitFailed    = "DB_INIT_FAILED"
	ErrCodeConnectFailed = "DB_CONNECT_FAILED"
	ErrCodeQueryFailed   = "DB_QUERY_FAILED"
	ErrCodeDirFailed     = "DB_DIR_FAILED"
)

// Initialize creates and initializes the SQLite database
func Initialize(cfg Config) error {
	logger.Info("Initializing database at: %s", cfg.Path)

	// Ensure data directory exists
	dbDir := filepath.Dir(cfg.Path)
	logger.Debug("Creating database directory: %s", dbDir)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logger.Error("Failed to create database directory: %v", err)
		return &DatabaseError{
			Code:    ErrCodeDirFailed,
			Message: "Failed to create database directory",
			Err:     err,
		}
	}

	// Open database connection
	logger.Debug("Opening database connection")
	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		logger.Error("Failed to open database connection: %v", err)
		return &DatabaseError{
			Code:    ErrCodeConnectFailed,
			Message: "Failed to open database connection",
			Err:     err,
		}
	}

	// Enable WAL mode for better concurrency
	logger.Debug("Enabling WAL mode")
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		logger.Error("Failed to enable WAL mode: %v", err)
		db.Close()
		return &DatabaseError{
			Code:    ErrCodeInitFailed,
			Message: "Failed to enable WAL mode",
			Err:     err,
		}
	}

	// Enable foreign keys
	logger.Debug("Enabling foreign keys")
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		logger.Error("Failed to enable foreign keys: %v", err)
		db.Close()
		return &DatabaseError{
			Code:    ErrCodeInitFailed,
			Message: "Failed to enable foreign keys",
			Err:     err,
		}
	}

	// Execute schema
	logger.Debug("Executing database schema")
	if _, err := db.Exec(schemaSQL); err != nil {
		logger.Error("Failed to initialize database schema: %v", err)
		db.Close()
		return &DatabaseError{
			Code:    ErrCodeInitFailed,
			Message: "Failed to initialize database schema",
			Err:     err,
		}
	}

	logger.Info("Database initialized successfully")

	// Set global DB connection
	DB = db

	return nil
}

// Connect opens a connection to an existing database
func Connect(cfg Config) error {
	// Open database connection
	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return &DatabaseError{
			Code:    ErrCodeConnectFailed,
			Message: "Failed to open database connection",
			Err:     err,
		}
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return &DatabaseError{
			Code:    ErrCodeConnectFailed,
			Message: "Failed to enable WAL mode",
			Err:     err,
		}
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return &DatabaseError{
			Code:    ErrCodeConnectFailed,
			Message: "Failed to enable foreign keys",
			Err:     err,
		}
	}

	// Set global DB connection
	DB = db

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// IsInitialized checks if the database has been initialized
func IsInitialized(path string) bool {
	// Check if database file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	// Try to open and query config table
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return false
	}
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM config WHERE key = 'db_initialized'").Scan(&value)
	return err == nil && value == "true"
}

// GetConfig retrieves a configuration value
func GetConfig(key string) (string, error) {
	if DB == nil {
		return "", &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	var value string
	err := DB.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to get config value",
			Err:     err,
		}
	}

	return value, nil
}

// SetConfig sets a configuration value
func SetConfig(key, value string) error {
	if DB == nil {
		return &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	_, err := DB.Exec(`
		INSERT INTO config (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
	`, key, value, value)

	if err != nil {
		return &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to set config value",
			Err:     err,
		}
	}

	return nil
}

// defaultPath can be overridden for testing
var defaultPath string

// GetDefaultPath returns the default database file path
func GetDefaultPath() string {
	if defaultPath != "" {
		return defaultPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "./data/cks-weight-room.db"
	}
	return filepath.Join(home, ".cks-weight-room", "data", "cks-weight-room.db")
}

// SetDefaultPathForTesting sets the default database path for testing
// This should only be used in tests
func SetDefaultPathForTesting(path string) {
	defaultPath = path
}
