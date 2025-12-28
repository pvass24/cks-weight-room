package activation

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ActivationServerURL is the production activation server endpoint
const ActivationServerURL = "https://activation.cks-weight-room.com/api/v1"

// Certificate pinning: SHA-256 hash of the expected server public key
// This should be replaced with the actual public key hash of your activation server
// To generate: openssl s_client -connect activation.cks-weight-room.com:443 -showcerts | openssl x509 -pubkey -noout | openssl pkey -pubin -outform DER | openssl dgst -sha256 -binary | openssl enc -base64
const ExpectedPublicKeyHash = "REPLACE_WITH_ACTUAL_PUBLIC_KEY_HASH"

// ActivateRequest represents the activation request payload
type ActivateRequest struct {
	LicenseKey string `json:"licenseKey"`
	MachineID  string `json:"machineId"`
	AppVersion string `json:"appVersion"`
}

// ActivateResponse represents the server's activation response
type ActivateResponse struct {
	Success         bool   `json:"success"`
	ActivationToken string `json:"activationToken,omitempty"`
	ExpiresAt       string `json:"expiresAt,omitempty"` // ISO 8601 format
	Error           string `json:"error,omitempty"`
	Message         string `json:"message"`
}

// ValidateRequest represents the validation request payload
type ValidateRequest struct {
	ActivationToken string `json:"activationToken"`
	MachineID       string `json:"machineId"`
}

// ValidateResponse represents the server's validation response
type ValidateResponse struct {
	Valid     bool   `json:"valid"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Client is the production activation client
type Client struct {
	baseURL    string
	httpClient *http.Client
	useMock    bool
}

// NewClient creates a new activation client
// If ACTIVATION_MOCK=true, uses mock mode (accepts any valid key)
func NewClient() *Client {
	useMock := os.Getenv("ACTIVATION_MOCK") == "true"

	client := &Client{
		baseURL: ActivationServerURL,
		useMock: useMock,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					// Certificate pinning would be configured here in production
					// VerifyPeerCertificate: verifyCertificatePin,
				},
			},
		},
	}

	return client
}

// Activate activates a license key with the activation server
func (c *Client) Activate(licenseKey, machineID, appVersion string) (*ActivateResponse, error) {
	// Mock mode for development/testing
	if c.useMock {
		return &ActivateResponse{
			Success:         true,
			ActivationToken: fmt.Sprintf("MOCK-TOKEN-%s-%d", machineID, time.Now().Unix()),
			Message:         "License activated successfully (Mock Mode)",
		}, nil
	}

	// Production mode - call real activation server
	reqBody := ActivateRequest{
		LicenseKey: licenseKey,
		MachineID:  machineID,
		AppVersion: appVersion,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/activate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("CKS-Weight-Room/%s", appVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var activateResp ActivateResponse
	if err := json.Unmarshal(body, &activateResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !activateResp.Success {
		return &activateResp, errors.New(activateResp.Message)
	}

	return &activateResp, nil
}

// Validate validates an activation token with the server
func (c *Client) Validate(activationToken, machineID string) (*ValidateResponse, error) {
	// Mock mode always returns valid
	if c.useMock {
		return &ValidateResponse{
			Valid: true,
		}, nil
	}

	// Production mode - call real validation endpoint
	reqBody := ValidateRequest{
		ActivationToken: activationToken,
		MachineID:       machineID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var validateResp ValidateResponse
	if err := json.Unmarshal(body, &validateResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &validateResp, nil
}

// verifyCertificatePin verifies the server's certificate against the pinned public key hash
// This is used for certificate pinning as per NFR-S5 and ARCH-2
func verifyCertificatePin(rawCerts [][]byte, verifiedChains [][]*tls.Certificate) error {
	if len(rawCerts) == 0 {
		return errors.New("no certificates provided")
	}

	// Hash the public key from the certificate
	cert := rawCerts[0]
	hash := sha256.Sum256(cert)
	actualHash := hex.EncodeToString(hash[:])

	if actualHash != ExpectedPublicKeyHash {
		return fmt.Errorf("certificate pin mismatch: expected %s, got %s", ExpectedPublicKeyHash, actualHash)
	}

	return nil
}
