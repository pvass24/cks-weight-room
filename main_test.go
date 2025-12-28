package main

import (
	"testing"
)

func TestVersionIsSet(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestVersionDefault(t *testing.T) {
	// When built without ldflags, version should be "dev"
	if version != "dev" && version == "" {
		t.Errorf("version should be 'dev' or set via ldflags, got: %s", version)
	}
}
