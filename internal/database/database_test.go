package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		Path: dbPath,
	}

	// Test initialization
	err := Initialize(cfg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Verify DB is not nil
	if DB == nil {
		t.Error("DB is nil after initialization")
	}

	// Clean up
	Close()
}

func TestIsInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Test with non-existent database
	if IsInitialized(dbPath) {
		t.Error("IsInitialized returned true for non-existent database")
	}

	// Initialize database
	cfg := Config{Path: dbPath}
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer Close()

	// Test with initialized database
	if !IsInitialized(dbPath) {
		t.Error("IsInitialized returned false for initialized database")
	}
}

func TestGetSetConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{Path: dbPath}
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer Close()

	// Test setting a config value
	key := "test_key"
	value := "test_value"

	err := SetConfig(key, value)
	if err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	// Test getting the config value
	retrievedValue, err := GetConfig(key)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if retrievedValue != value {
		t.Errorf("Expected value %s, got %s", value, retrievedValue)
	}

	// Test updating config value
	newValue := "new_test_value"
	err = SetConfig(key, newValue)
	if err != nil {
		t.Fatalf("SetConfig (update) failed: %v", err)
	}

	retrievedValue, err = GetConfig(key)
	if err != nil {
		t.Fatalf("GetConfig (after update) failed: %v", err)
	}

	if retrievedValue != newValue {
		t.Errorf("Expected updated value %s, got %s", newValue, retrievedValue)
	}
}

func TestGetConfigNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{Path: dbPath}
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer Close()

	// Test getting non-existent config value
	value, err := GetConfig("non_existent_key")
	if err != nil {
		t.Fatalf("GetConfig failed for non-existent key: %v", err)
	}

	if value != "" {
		t.Errorf("Expected empty string for non-existent key, got %s", value)
	}
}

func TestDatabaseSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{Path: dbPath}
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer Close()

	// Verify all tables exist
	tables := []string{"exercises", "progress", "clusters", "config", "schema_version"}

	for _, table := range tables {
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query table %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("Table %s does not exist", table)
		}
	}
}

func TestWALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{Path: dbPath}
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer Close()

	// Verify WAL mode is enabled
	var journalMode string
	err := DB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got %s", journalMode)
	}
}

func TestForeignKeys(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{Path: dbPath}
	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer Close()

	// Verify foreign keys are enabled
	var foreignKeys int
	err := DB.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("Failed to query foreign keys: %v", err)
	}

	if foreignKeys != 1 {
		t.Error("Foreign keys are not enabled")
	}
}
