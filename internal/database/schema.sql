-- CKS Weight Room Database Schema
-- SQLite with WAL mode for better concurrency

-- Exercises/Challenges table
CREATE TABLE IF NOT EXISTS exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL, -- e.g., 'cluster-setup', 'cluster-hardening', 'system-hardening', 'minimize-microservice-vulnerabilities', 'supply-chain-security', 'monitoring-logging-runtime-security'
    difficulty TEXT NOT NULL CHECK(difficulty IN ('easy', 'medium', 'hard')),
    points INTEGER NOT NULL DEFAULT 10,
    estimated_minutes INTEGER,
    prerequisites TEXT, -- JSON array of exercise slugs
    hints TEXT, -- JSON array of hints
    solution TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- User progress tracking
CREATE TABLE IF NOT EXISTS progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exercise_id INTEGER NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('not-started', 'in-progress', 'completed', 'skipped')),
    started_at DATETIME,
    completed_at DATETIME,
    attempts INTEGER DEFAULT 0,
    time_spent_seconds INTEGER DEFAULT 0,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE
);

-- Kubernetes cluster state
CREATE TABLE IF NOT EXISTS clusters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    kind_config TEXT, -- JSON configuration
    status TEXT NOT NULL CHECK(status IN ('creating', 'running', 'stopped', 'error')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Application configuration
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Database version tracking for migrations
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_exercises_category ON exercises(category);
CREATE INDEX IF NOT EXISTS idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX IF NOT EXISTS idx_progress_exercise_id ON progress(exercise_id);
CREATE INDEX IF NOT EXISTS idx_progress_status ON progress(status);
CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status);

-- Triggers for updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_exercises_timestamp
    AFTER UPDATE ON exercises
    BEGIN
        UPDATE exercises SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_progress_timestamp
    AFTER UPDATE ON progress
    BEGIN
        UPDATE progress SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_clusters_timestamp
    AFTER UPDATE ON clusters
    BEGIN
        UPDATE clusters SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_config_timestamp
    AFTER UPDATE ON config
    BEGIN
        UPDATE config SET updated_at = CURRENT_TIMESTAMP WHERE key = NEW.key;
    END;

-- Initial schema version
INSERT OR IGNORE INTO schema_version (version) VALUES (1);

-- Initial configuration
INSERT OR IGNORE INTO config (key, value) VALUES ('db_initialized', 'true');
INSERT OR IGNORE INTO config (key, value) VALUES ('first_launch_completed', 'false');
