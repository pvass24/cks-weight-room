-- Migration 002: Add attempts and mock_exams tables for progress tracking
-- This supports Story 4.3: Overall Progress Statistics

-- Individual attempts table
CREATE TABLE IF NOT EXISTS attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exercise_id INTEGER NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    duration_seconds INTEGER, -- Calculated: completed_at - started_at
    score INTEGER NOT NULL DEFAULT 0, -- Points earned (0 to exercise.points)
    max_score INTEGER NOT NULL, -- Maximum possible points for this exercise
    passed BOOLEAN NOT NULL DEFAULT 0, -- 1 if passed, 0 if failed
    feedback TEXT, -- Grading feedback
    details TEXT, -- JSON array of check details
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE
);

-- Mock exam results table
CREATE TABLE IF NOT EXISTS mock_exams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exam_type TEXT NOT NULL, -- 'quick-practice' or 'full-mock-exam'
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    total_duration_seconds INTEGER,
    overall_score INTEGER NOT NULL DEFAULT 0, -- Total points earned
    max_score INTEGER NOT NULL, -- Total possible points
    passed BOOLEAN NOT NULL DEFAULT 0, -- 1 if passed, 0 if failed
    exercises_completed INTEGER NOT NULL DEFAULT 0, -- Count of exercises completed
    exercises_total INTEGER NOT NULL, -- Total exercises in exam
    results TEXT, -- JSON array of {exercise_id, score, max_score, passed}
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Update progress table to add personal_best_seconds
ALTER TABLE progress ADD COLUMN personal_best_seconds INTEGER;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_attempts_exercise_id ON attempts(exercise_id);
CREATE INDEX IF NOT EXISTS idx_attempts_completed_at ON attempts(completed_at);
CREATE INDEX IF NOT EXISTS idx_mock_exams_exam_type ON mock_exams(exam_type);
CREATE INDEX IF NOT EXISTS idx_mock_exams_completed_at ON mock_exams(completed_at);

-- Update schema version
INSERT INTO schema_version (version) VALUES (2);
