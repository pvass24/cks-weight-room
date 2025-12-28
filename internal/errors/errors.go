package errors

import "fmt"

// Error codes for different failure scenarios
const (
	// Docker errors
	ErrDockerNotRunning     = "DOCKER_NOT_RUNNING"
	ErrDockerPermission     = "DOCKER_PERMISSION_DENIED"
	ErrDockerNotInstalled   = "DOCKER_NOT_INSTALLED"

	// KIND/Cluster errors
	ErrClusterProvisionFailed = "CLUSTER_PROVISION_FAILED"
	ErrClusterDeleteFailed    = "CLUSTER_DELETE_FAILED"
	ErrClusterNotFound        = "CLUSTER_NOT_FOUND"
	ErrInsufficientDiskSpace  = "INSUFFICIENT_DISK_SPACE"

	// Activation errors
	ErrActivationNetworkError = "ACTIVATION_NETWORK_ERROR"
	ErrActivationInvalidKey   = "ACTIVATION_INVALID_KEY"
	ErrActivationExpired      = "ACTIVATION_EXPIRED"

	// Database errors
	ErrDatabaseCorrupted    = "DATABASE_CORRUPTED"
	ErrDatabaseLocked       = "DATABASE_LOCKED"
	ErrDatabaseWriteFailed  = "DATABASE_WRITE_FAILED"

	// Validation errors
	ErrValidationFailed   = "VALIDATION_FAILED"
	ErrValidationTimeout  = "VALIDATION_TIMEOUT"

	// WebSocket errors
	ErrWebSocketDisconnected = "WEBSOCKET_DISCONNECTED"
	ErrWebSocketFailed       = "WEBSOCKET_CONNECTION_FAILED"

	// Generic errors
	ErrNetworkTimeout   = "NETWORK_TIMEOUT"
	ErrInternalError    = "INTERNAL_ERROR"
	ErrOperationFailed  = "OPERATION_FAILED"
)

// ActionableError represents an error with actionable information for users
type ActionableError struct {
	Code        string            `json:"code"`
	What        string            `json:"what"`
	Why         string            `json:"why"`
	HowToFix    []string          `json:"howToFix"`
	Retryable   bool              `json:"retryable"`
	Context     map[string]string `json:"context,omitempty"`
	InternalErr error             `json:"-"`
}

// Error implements the error interface
func (e *ActionableError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.What)
}

// Unwrap returns the underlying error
func (e *ActionableError) Unwrap() error {
	return e.InternalErr
}

// NewActionableError creates a new actionable error
func NewActionableError(code, what, why string, howToFix []string, retryable bool) *ActionableError {
	return &ActionableError{
		Code:      code,
		What:      what,
		Why:       why,
		HowToFix:  howToFix,
		Retryable: retryable,
		Context:   make(map[string]string),
	}
}

// WithContext adds context to the error
func (e *ActionableError) WithContext(key, value string) *ActionableError {
	e.Context[key] = value
	return e
}

// WithInternalError adds the underlying error
func (e *ActionableError) WithInternalError(err error) *ActionableError {
	e.InternalErr = err
	return e
}

// Common pre-defined errors

// NewDockerNotRunningError creates a Docker not running error
func NewDockerNotRunningError() *ActionableError {
	return NewActionableError(
		ErrDockerNotRunning,
		"Cannot provision Kubernetes cluster because Docker Desktop is not running.",
		"CKS Weight Room uses Docker to create KIND clusters for practice scenarios.",
		[]string{
			"Start Docker Desktop application",
			"Wait for Docker to finish starting (check system tray icon)",
			"Click 'Retry' below",
		},
		true,
	)
}

// NewClusterProvisionFailedError creates a cluster provision failed error
func NewClusterProvisionFailedError(reason string) *ActionableError {
	return NewActionableError(
		ErrClusterProvisionFailed,
		"Unable to create KIND cluster.",
		reason,
		[]string{
			"Check Docker Desktop is running",
			"Ensure you have sufficient disk space (5GB minimum)",
			"Click 'Retry' to provision the cluster again",
		},
		true,
	)
}

// NewInsufficientDiskSpaceError creates an insufficient disk space error
func NewInsufficientDiskSpaceError(available, required string) *ActionableError {
	return NewActionableError(
		ErrInsufficientDiskSpace,
		"Unable to create KIND cluster due to insufficient disk space.",
		"KIND requires at least 5GB free space to download container images and create a cluster.",
		[]string{
			fmt.Sprintf("Free up at least %s of disk space", required),
			"Click 'Retry' to provision the cluster again",
		},
		true,
	).WithContext("available", available).WithContext("required", required)
}

// NewActivationNetworkError creates an activation network error
func NewActivationNetworkError() *ActionableError {
	return NewActionableError(
		ErrActivationNetworkError,
		"Unable to validate license with activation server.",
		"Network connection failed (timeout after 10 seconds).",
		[]string{
			"Check your internet connection",
			"Verify firewall allows HTTPS to activation.cks-weight-room.com",
			"Try again, or continue practicing (30-day grace period)",
		},
		true,
	)
}

// NewDatabaseCorruptedError creates a database corruption error
func NewDatabaseCorruptedError(backupPath string) *ActionableError {
	return NewActionableError(
		ErrDatabaseCorrupted,
		"Database corruption detected.",
		"The SQLite database file has integrity issues.",
		[]string{
			fmt.Sprintf("A backup has been created at: %s", backupPath),
			"The database has been reinitialized",
			"Your progress data may be lost",
			"Contact support if this persists",
		},
		false,
	).WithContext("backupPath", backupPath)
}

// NewWebSocketDisconnectedError creates a WebSocket disconnection error
func NewWebSocketDisconnectedError() *ActionableError {
	return NewActionableError(
		ErrWebSocketDisconnected,
		"Connection to terminal lost.",
		"WebSocket connection was closed unexpectedly.",
		[]string{
			"Reconnecting automatically...",
		},
		true,
	)
}
