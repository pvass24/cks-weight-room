package database

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/002_add_attempts_and_mock_exams.sql
var migration002 string

//go:embed migrations/003_add_activation_table.sql
var migration003 string

// ApplyMigrations applies any pending database migrations
func ApplyMigrations() error {
	if DB == nil {
		return &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	// Get current schema version
	var currentVersion int
	err := DB.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to get current schema version",
			Err:     err,
		}
	}

	// Apply migrations in order
	migrations := []struct {
		version int
		sql     string
	}{
		{2, migration002},
		{3, migration003},
	}

	for _, migration := range migrations {
		if migration.version > currentVersion {
			// Apply migration in a transaction
			tx, err := DB.Begin()
			if err != nil {
				return &DatabaseError{
					Code:    ErrCodeQueryFailed,
					Message: fmt.Sprintf("Failed to start transaction for migration %d", migration.version),
					Err:     err,
				}
			}

			// Execute migration SQL
			_, err = tx.Exec(migration.sql)
			if err != nil {
				tx.Rollback()
				return &DatabaseError{
					Code:    ErrCodeQueryFailed,
					Message: fmt.Sprintf("Failed to apply migration %d", migration.version),
					Err:     err,
				}
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return &DatabaseError{
					Code:    ErrCodeQueryFailed,
					Message: fmt.Sprintf("Failed to commit migration %d", migration.version),
					Err:     err,
				}
			}

			fmt.Printf("Applied migration %d\n", migration.version)
		}
	}

	return nil
}

// GetCurrentSchemaVersion returns the current database schema version
func GetCurrentSchemaVersion() (int, error) {
	if DB == nil {
		return 0, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Database not initialized",
		}
	}

	var version int
	err := DB.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, &DatabaseError{
			Code:    ErrCodeQueryFailed,
			Message: "Failed to get schema version",
			Err:     err,
		}
	}

	return version, nil
}
