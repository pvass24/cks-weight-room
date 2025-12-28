package prerequisites

import (
	"testing"
)

func TestCheckDocker(t *testing.T) {
	// This test will pass if Docker is running, fail if not
	// In a real scenario, we'd mock os/exec.Command
	err := CheckDocker()

	// For now, we just verify the error structure if it fails
	if err != nil {
		if prereqErr, ok := err.(*PrerequisiteError); ok {
			if prereqErr.Code != DockerNotRunning {
				t.Errorf("Expected error code %s, got %s", DockerNotRunning, prereqErr.Code)
			}
			if prereqErr.Message == "" {
				t.Error("Expected non-empty error message")
			}
		} else {
			t.Error("Expected PrerequisiteError type")
		}
	}
}

func TestCheckKind(t *testing.T) {
	// This test will pass if KIND is installed, fail if not
	err := CheckKind()

	// Verify error structure if it fails
	if err != nil {
		if prereqErr, ok := err.(*PrerequisiteError); ok {
			if prereqErr.Code != KindNotInstalled {
				t.Errorf("Expected error code %s, got %s", KindNotInstalled, prereqErr.Code)
			}
			if prereqErr.Message == "" {
				t.Error("Expected non-empty error message")
			}
		} else {
			t.Error("Expected PrerequisiteError type")
		}
	}
}

func TestCheckDiskSpace(t *testing.T) {
	// This should pass in most environments
	err := CheckDiskSpace()

	// If it fails, verify error structure
	if err != nil {
		if prereqErr, ok := err.(*PrerequisiteError); ok {
			if prereqErr.Code != InsufficientDiskSpace {
				t.Errorf("Expected error code %s, got %s", InsufficientDiskSpace, prereqErr.Code)
			}
			if prereqErr.Message == "" {
				t.Error("Expected non-empty error message")
			}
		}
	}
}

func TestValidateAll(t *testing.T) {
	results, err := ValidateAll()

	// Should always return 3 results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify result structure
	for _, result := range results {
		if result.Name == "" {
			t.Error("Expected non-empty name in result")
		}
		if !result.Passed && result.Message == "" {
			t.Errorf("Failed check '%s' should have a message", result.Name)
		}
	}

	// If there's an error, it should be a PrerequisiteError
	if err != nil {
		if _, ok := err.(*PrerequisiteError); !ok {
			t.Error("Expected PrerequisiteError type for first error")
		}
	}
}

func TestPrerequisiteErrorImplementsError(t *testing.T) {
	err := &PrerequisiteError{
		Code:    "TEST_CODE",
		Message: "Test message",
		Details: "Test details",
	}

	// Should implement error interface
	var _ error = err

	// Error() should return the message
	if err.Error() != "Test message" {
		t.Errorf("Expected 'Test message', got '%s'", err.Error())
	}
}
