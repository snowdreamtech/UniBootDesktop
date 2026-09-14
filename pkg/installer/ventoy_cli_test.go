// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateVentoyCliEmptyPath(t *testing.T) {
	res := ValidateVentoyCli("")
	if res.Valid {
		t.Errorf("Expected invalid result for empty path, got valid")
	}
}

func TestValidateVentoyCliNonExistentPath(t *testing.T) {
	res := ValidateVentoyCli("/path/to/nonexistent/ventoy")
	if res.Valid {
		t.Errorf("Expected invalid result for nonexistent path, got valid")
	}
}

func TestExtractVentoyVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ventoy2Disk.sh v1.0.99", "v1.0.99"},
		{"Ventoy2Disk version 1.0.88", "v1.0.88"},
		{"v1.0.95 (c) Ventoy", "v1.0.95"},
		{"No version number here", ""},
	}

	for _, tt := range tests {
		got := extractVentoyVersion(tt.input)
		if got != tt.expected {
			t.Errorf("extractVentoyVersion(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestValidateVentoyCliDummyScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping bash script test on Windows")
	}
	t.Setenv("UNIBOOT_DRY_RUN", "1")

	tmpDir, err := os.MkdirTemp("", "ventoy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "Ventoy2Disk.sh")
	scriptContent := "#!/bin/sh\necho \"Ventoy2Disk.sh v1.0.99\"\nexit 0\n"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	res := ValidateVentoyCli(tmpDir)
	if !res.Valid {
		t.Errorf("Expected valid result for mock script dir, got invalid: %s", res.Message)
	}
	if res.Version != "v1.0.99" && res.Version != "v1.0.99 (Dry-Run)" {
		t.Errorf("Expected version v1.0.99 or v1.0.99 (Dry-Run), got %s", res.Version)
	}
}
