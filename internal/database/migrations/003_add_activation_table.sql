-- Migration 003: Add activation table for license management
-- This table stores encrypted license activation data

CREATE TABLE IF NOT EXISTS activation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    license_key TEXT NOT NULL, -- Encrypted with AES-256-GCM
    activation_token TEXT, -- Encrypted activation token from server
    machine_id TEXT NOT NULL, -- Hardware fingerprint
    activated_at DATETIME NOT NULL,
    expires_at DATETIME, -- NULL for perpetual licenses
    last_validated_at DATETIME NOT NULL,
    grace_period_started_at DATETIME, -- When grace period began (if any)
    encryption_nonce TEXT NOT NULL, -- Nonce for AES-256-GCM decryption
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for quick activation status checks
CREATE INDEX IF NOT EXISTS idx_activation_machine_id ON activation(machine_id);

-- Trigger for updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_activation_timestamp
    AFTER UPDATE ON activation
    BEGIN
        UPDATE activation SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Insert schema version
INSERT INTO schema_version (version) VALUES (3);
