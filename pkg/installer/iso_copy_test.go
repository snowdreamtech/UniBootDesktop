// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateIsoPath(t *testing.T) {
	tmpDir := t.TempDir()

	validExtensions := []string{".iso", ".wim", ".img", ".vhd", ".vhdx", ".vti", ".efi", ".bin", ".xz", ".gz", ".raw"}
	for _, ext := range validExtensions {
		filePath := filepath.Join(tmpDir, "test_system"+ext)
		if err := os.WriteFile(filePath, []byte("dummy image content"), 0644); err != nil {
			t.Fatalf("failed to create dummy file: %v", err)
		}
		if !ValidateIsoPath(filePath) {
			t.Errorf("expected ValidateIsoPath to return true for extension %s", ext)
		}
	}

	invalidFilePath := filepath.Join(tmpDir, "test_file.txt")
	if err := os.WriteFile(invalidFilePath, []byte("text content"), 0644); err != nil {
		t.Fatalf("failed to create dummy text file: %v", err)
	}
	if ValidateIsoPath(invalidFilePath) {
		t.Errorf("expected ValidateIsoPath to return false for .txt file")
	}
}

func TestCopyIsoFilesToDisk(t *testing.T) {
	tmpSrcDir := t.TempDir()
	tmpMountDir := t.TempDir()

	iso1 := filepath.Join(tmpSrcDir, "ubuntu-24.04-desktop-amd64.iso")
	iso2 := filepath.Join(tmpSrcDir, "win11_install.wim")

	content1 := []byte("fake ubuntu iso content 12345")
	content2 := []byte("fake windows wim content 67890")

	if err := os.WriteFile(iso1, content1, 0644); err != nil {
		t.Fatalf("failed to write dummy iso1: %v", err)
	}
	if err := os.WriteFile(iso2, content2, 0644); err != nil {
		t.Fatalf("failed to write dummy iso2: %v", err)
	}

	var reportedProgress []IsoCopyProgress
	cb := func(p IsoCopyProgress) {
		reportedProgress = append(reportedProgress, p)
	}

	err := CopyIsoFilesToDisk(tmpMountDir, []string{iso1, iso2}, cb)
	if err != nil {
		t.Fatalf("CopyIsoFilesToDisk failed: %v", err)
	}

	// Verify target iso/ directory creation and contents
	destIso1 := filepath.Join(tmpMountDir, "iso", "ubuntu-24.04-desktop-amd64.iso")
	destIso2 := filepath.Join(tmpMountDir, "iso", "win11_install.wim")

	b1, err := os.ReadFile(destIso1)
	if err != nil || string(b1) != string(content1) {
		t.Errorf("destIso1 content mismatch or read error: %v", err)
	}

	b2, err := os.ReadFile(destIso2)
	if err != nil || string(b2) != string(content2) {
		t.Errorf("destIso2 content mismatch or read error: %v", err)
	}

	if len(reportedProgress) == 0 {
		t.Errorf("expected progress callback to be triggered")
	}
}
