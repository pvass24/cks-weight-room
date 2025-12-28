package main

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedFilesServe(t *testing.T) {
	// Create test server with embedded files
	staticFS, err := createStaticFS()
	if err != nil {
		t.Fatalf("Failed to create static FS: %v", err)
	}

	handler := http.FileServer(staticFS)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test serving index.html
	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to fetch index: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Verify content contains expected title
	if !strings.Contains(string(body), "<title>CKS Weight Room</title>") {
		t.Error("Expected index.html to contain CKS Weight Room title")
	}
}

// Helper function to create static FS (same logic as main)
func createStaticFS() (http.FileSystem, error) {
	staticFS, err := fs.Sub(webFS, "web/out")
	if err != nil {
		return nil, err
	}
	return http.FS(staticFS), nil
}
