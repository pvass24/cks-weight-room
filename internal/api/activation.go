package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/patrickvassell/cks-weight-room/internal/activation"
	"github.com/patrickvassell/cks-weight-room/internal/crypto"
	"github.com/patrickvassell/cks-weight-room/internal/database"
	"github.com/patrickvassell/cks-weight-room/internal/logger"
)

// ActivationRequest represents a license activation request
type ActivationRequest struct {
	LicenseKey string `json:"licenseKey"`
}

// ActivationResponse represents the response from activation
type ActivationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ActivationStatusResponse represents the current activation status
type ActivationStatusResponse struct {
	IsActivated       bool   `json:"isActivated"`
	LicenseKey        string `json:"licenseKey,omitempty"` // Last 5 chars only
	MachineID         string `json:"machineId"`
	ActivatedAt       string `json:"activatedAt,omitempty"`
	ExpiresAt         string `json:"expiresAt,omitempty"`
	DaysRemaining     int    `json:"daysRemaining,omitempty"`
	InGracePeriod     bool   `json:"inGracePeriod"`
	GraceDaysLeft     int    `json:"graceDaysLeft,omitempty"`
	NeedsValidation   bool   `json:"needsValidation"`   // True if >7 days since last validation
	ValidationExpired bool   `json:"validationExpired"` // True if grace period expired
	LastValidatedAt   string `json:"lastValidatedAt,omitempty"`
}

// ValidateActivationResponse represents the response from periodic validation
type ValidateActivationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// OfflineActivationFile represents the structure of an offline activation file
type OfflineActivationFile struct {
	LicenseKey      string `json:"licenseKey"`
	MachineID       string `json:"machineId"`
	ActivationToken string `json:"activationToken"`
	IssuedAt        string `json:"issuedAt"`
	ExpiresAt       string `json:"expiresAt,omitempty"`
	Signature       string `json:"signature"` // Digital signature for verification
}

// MachineIDResponse represents the machine ID response
type MachineIDResponse struct {
	MachineID string `json:"machineId"`
}

// ValidateLicenseKeyFormat checks if the license key matches the expected format
// Expected format: CKSWT-XXXXX-XXXXX-XXXXX-XXXXX
func ValidateLicenseKeyFormat(key string) bool {
	pattern := `^CKSWT-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

// GetMachineID handles GET /api/activation/machine-id
func GetMachineID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	machineID, err := crypto.GetMachineID()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get machine ID: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MachineIDResponse{
		MachineID: machineID,
	})
}

// GetActivationStatus handles GET /api/activation/status
func GetActivationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	machineID, err := crypto.GetMachineID()
	if err != nil {
		http.Error(w, "Failed to get machine ID", http.StatusInternalServerError)
		return
	}

	response := ActivationStatusResponse{
		IsActivated:   false,
		MachineID:     machineID,
		InGracePeriod: false,
	}

	// Check if activation exists
	var encryptedLicenseKey, encryptedToken, nonce, activatedAt string
	var expiresAt sql.NullString
	var gracePeriodStartedAt sql.NullString
	var lastValidatedAt string

	err = database.DB.QueryRow(`
		SELECT license_key, activation_token, encryption_nonce, activated_at, expires_at,
		       last_validated_at, grace_period_started_at
		FROM activation
		WHERE machine_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, machineID).Scan(&encryptedLicenseKey, &encryptedToken, &nonce, &activatedAt, &expiresAt, &lastValidatedAt, &gracePeriodStartedAt)

	if err == sql.ErrNoRows {
		// No activation found
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if err != nil {
		http.Error(w, "Failed to check activation", http.StatusInternalServerError)
		return
	}

	// Decrypt license key to show last 5 chars
	machineIDForEncryption, err := crypto.GetMachineIDForEncryption()
	if err != nil {
		http.Error(w, "Failed to derive encryption key", http.StatusInternalServerError)
		return
	}

	key := crypto.DeriveKey(machineIDForEncryption)
	decryptedLicenseKey, err := crypto.Decrypt(encryptedLicenseKey, nonce, key)
	if err != nil {
		http.Error(w, "Failed to decrypt license key", http.StatusInternalServerError)
		return
	}

	// Show only last 5 characters
	if len(decryptedLicenseKey) >= 5 {
		response.LicenseKey = "..." + decryptedLicenseKey[len(decryptedLicenseKey)-5:]
	}

	response.IsActivated = true
	response.ActivatedAt = activatedAt

	// Check expiration
	if expiresAt.Valid {
		response.ExpiresAt = expiresAt.String
		expiryTime, err := time.Parse("2006-01-02 15:04:05", expiresAt.String)
		if err == nil {
			daysRemaining := int(time.Until(expiryTime).Hours() / 24)
			response.DaysRemaining = daysRemaining
		}
	}

	// Check if periodic validation is needed (every 7 days)
	response.LastValidatedAt = lastValidatedAt
	lastValidated, err := time.Parse("2006-01-02 15:04:05", lastValidatedAt)
	if err == nil {
		daysSinceValidation := int(time.Since(lastValidated).Hours() / 24)
		if daysSinceValidation >= 7 {
			response.NeedsValidation = true
		}
	}

	// Check grace period
	if gracePeriodStartedAt.Valid {
		response.InGracePeriod = true
		gracePeriodStart, err := time.Parse("2006-01-02 15:04:05", gracePeriodStartedAt.String)
		if err == nil {
			gracePeriodEnd := gracePeriodStart.Add(30 * 24 * time.Hour)
			graceDaysLeft := int(time.Until(gracePeriodEnd).Hours() / 24)
			if graceDaysLeft < 0 {
				// Grace period expired
				response.ValidationExpired = true
				response.GraceDaysLeft = 0
			} else {
				response.GraceDaysLeft = graceDaysLeft
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ActivateLicense handles POST /api/activation/activate
func ActivateLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	var req ActivationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Info("License activation request received")

	// Validate license key format
	if !ValidateLicenseKeyFormat(req.LicenseKey) {
		logger.Warn("Invalid license key format provided")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ActivationResponse{
			Success: false,
			Error:   "Invalid license key format. Expected format: CKSWT-XXXXX-XXXXX-XXXXX-XXXXX",
		})
		return
	}

	// Get machine ID
	machineID, err := crypto.GetMachineID()
	if err != nil {
		logger.Error("Failed to get machine ID: %v", err)
		http.Error(w, "Failed to get machine ID", http.StatusInternalServerError)
		return
	}

	logger.Debug("Machine ID: %s", machineID)

	// Get machine ID for encryption
	machineIDForEncryption, err := crypto.GetMachineIDForEncryption()
	if err != nil {
		http.Error(w, "Failed to derive encryption key", http.StatusInternalServerError)
		return
	}

	// Call activation server (uses mock mode if ACTIVATION_MOCK=true)
	logger.Debug("Contacting activation server...")
	activationClient := activation.NewClient()
	activateResp, err := activationClient.Activate(req.LicenseKey, machineID, "0.1.0")
	if err != nil {
		logger.Error("Activation server returned error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ActivationResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	logger.Info("License validated successfully by activation server")
	activationToken := activateResp.ActivationToken

	// Derive encryption key from machine ID
	key := crypto.DeriveKey(machineIDForEncryption)

	// Encrypt license key
	encryptedLicenseKey, nonce, err := crypto.Encrypt(req.LicenseKey, key)
	if err != nil {
		http.Error(w, "Failed to encrypt license key", http.StatusInternalServerError)
		return
	}

	// Encrypt activation token
	encryptedToken, _, err := crypto.Encrypt(activationToken, key)
	if err != nil {
		http.Error(w, "Failed to encrypt activation token", http.StatusInternalServerError)
		return
	}

	// Store activation in database
	logger.Debug("Storing encrypted activation data in database")
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(`
		INSERT INTO activation (license_key, activation_token, machine_id, activated_at, last_validated_at, encryption_nonce)
		VALUES (?, ?, ?, ?, ?, ?)
	`, encryptedLicenseKey, encryptedToken, machineID, now, now, nonce)

	if err != nil {
		logger.Error("Failed to store activation in database: %v", err)
		http.Error(w, fmt.Sprintf("Failed to store activation: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info("License activation completed successfully for machine: %s", machineID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ActivationResponse{
		Success: true,
		Message: "License activated successfully! Welcome to CKS Weight Room.",
	})
}

// ValidateActivation handles POST /api/activation/validate
// Attempts periodic validation with the activation server
func ValidateActivation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Get machine ID
	machineID, err := crypto.GetMachineID()
	if err != nil {
		http.Error(w, "Failed to get machine ID", http.StatusInternalServerError)
		return
	}

	// Get current activation
	var encryptedToken, nonce string
	err = database.DB.QueryRow(`
		SELECT activation_token, encryption_nonce
		FROM activation
		WHERE machine_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, machineID).Scan(&encryptedToken, &nonce)

	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidateActivationResponse{
			Success: false,
			Error:   "No activation found",
		})
		return
	}

	if err != nil {
		http.Error(w, "Failed to get activation", http.StatusInternalServerError)
		return
	}

	// Decrypt activation token
	machineIDForEncryption, err := crypto.GetMachineIDForEncryption()
	if err != nil {
		http.Error(w, "Failed to derive encryption key", http.StatusInternalServerError)
		return
	}

	key := crypto.DeriveKey(machineIDForEncryption)
	activationToken, err := crypto.Decrypt(encryptedToken, nonce, key)
	if err != nil {
		http.Error(w, "Failed to decrypt activation token", http.StatusInternalServerError)
		return
	}

	// Attempt validation with activation server
	activationClient := activation.NewClient()
	validateResp, err := activationClient.Validate(activationToken, machineID)

	if err != nil {
		// Network error - enter grace period
		now := time.Now().Format("2006-01-02 15:04:05")
		_, updateErr := database.DB.Exec(`
			UPDATE activation
			SET grace_period_started_at = COALESCE(grace_period_started_at, ?)
			WHERE machine_id = ?
		`, now, machineID)

		if updateErr != nil {
			http.Error(w, "Failed to update grace period", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidateActivationResponse{
			Success: false,
			Message: "Unable to validate license. You can continue practicing for 30 days without internet.",
		})
		return
	}

	if !validateResp.Valid {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidateActivationResponse{
			Success: false,
			Error:   "License validation failed: " + validateResp.Error,
		})
		return
	}

	// Validation succeeded - update last_validated_at and clear grace period
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(`
		UPDATE activation
		SET last_validated_at = ?, grace_period_started_at = NULL
		WHERE machine_id = ?
	`, now, machineID)

	if err != nil {
		http.Error(w, "Failed to update validation timestamp", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ValidateActivationResponse{
		Success: true,
		Message: "License validated successfully",
	})
}

// ActivateOffline handles POST /api/activation/activate-offline
// Activates license using an offline activation file
func ActivateOffline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if database.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Parse the activation file from request body
	var activationFile OfflineActivationFile
	if err := json.NewDecoder(r.Body).Decode(&activationFile); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ActivationResponse{
			Success: false,
			Error:   "Invalid activation file format",
		})
		return
	}

	// Get current machine ID
	machineID, err := crypto.GetMachineID()
	if err != nil {
		http.Error(w, "Failed to get machine ID", http.StatusInternalServerError)
		return
	}

	// Verify machine ID matches
	if activationFile.MachineID != machineID {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ActivationResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid activation file: This activation file is for a different machine. Machine ID mismatch. Expected: %s, Found: %s", machineID, activationFile.MachineID),
		})
		return
	}

	// Validate license key format
	if !ValidateLicenseKeyFormat(activationFile.LicenseKey) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ActivationResponse{
			Success: false,
			Error:   "Invalid license key format in activation file",
		})
		return
	}

	// TODO: Verify digital signature in production
	// For now, we accept the activation file if machine ID matches
	// In production, verify activationFile.Signature using hardcoded public key

	// Get machine ID for encryption
	machineIDForEncryption, err := crypto.GetMachineIDForEncryption()
	if err != nil {
		http.Error(w, "Failed to derive encryption key", http.StatusInternalServerError)
		return
	}

	// Derive encryption key from machine ID
	key := crypto.DeriveKey(machineIDForEncryption)

	// Encrypt license key
	encryptedLicenseKey, nonce, err := crypto.Encrypt(activationFile.LicenseKey, key)
	if err != nil {
		http.Error(w, "Failed to encrypt license key", http.StatusInternalServerError)
		return
	}

	// Encrypt activation token
	encryptedToken, _, err := crypto.Encrypt(activationFile.ActivationToken, key)
	if err != nil {
		http.Error(w, "Failed to encrypt activation token", http.StatusInternalServerError)
		return
	}

	// Store activation in database
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(`
		INSERT INTO activation (license_key, activation_token, machine_id, activated_at, last_validated_at, encryption_nonce)
		VALUES (?, ?, ?, ?, ?, ?)
	`, encryptedLicenseKey, encryptedToken, machineID, now, now, nonce)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store activation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ActivationResponse{
		Success: true,
		Message: "License activated successfully! (Offline Mode)",
	})
}
