package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "PUT request fails",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/setup/validate", nil)
			w := httptest.NewRecorder()

			ValidatePrerequisites(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse {
				var response ValidationResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Verify response structure
				if response.Checks == nil {
					t.Error("Expected checks array, got nil")
				}

				// Content-Type should be application/json
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

func TestValidatePrerequisitesResponseStructure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/setup/validate", nil)
	w := httptest.NewRecorder()

	ValidatePrerequisites(w, req)

	var response ValidationResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify all checks are present
	expectedChecks := []string{"Docker", "KIND", "Disk Space"}
	if len(response.Checks) != len(expectedChecks) {
		t.Errorf("Expected %d checks, got %d", len(expectedChecks), len(response.Checks))
	}

	// Verify check names
	checkNames := make(map[string]bool)
	for _, check := range response.Checks {
		checkNames[check.Name] = true
	}

	for _, expected := range expectedChecks {
		if !checkNames[expected] {
			t.Errorf("Expected check '%s' not found", expected)
		}
	}
}
