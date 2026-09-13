// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatDiskModeB_SafetyValidation(t *testing.T) {
	ctx := context.Background()

	// Test safety block against system drive
	systemDrives := []string{"/", "/dev/sda", "C:", "/dev/nvme0n1"}
	for _, drive := range systemDrives {
		_, err := FormatDiskModeB(ctx, drive)
		if err == nil {
			t.Errorf("Expected safety validation error for system drive %s, got nil", drive)
		}
	}
}

func TestFormatDiskModeB_DryRun(t *testing.T) {
	ctx := context.Background()
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mountPoint, err := FormatDiskModeB(ctx, "dummy_usb_disk")
	if err != nil {
		t.Fatalf("FormatDiskModeB dry-run failed: %v", err)
	}

	if mountPoint == "" {
		t.Fatalf("Expected non-empty mount point in dry-run mode")
	}

	info, err := os.Stat(mountPoint)
	if err != nil || !info.IsDir() {
		t.Fatalf("Expected mount point %s to be a valid directory", mountPoint)
	}
}

func TestResolveMountPoint_DryRun(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mount, err := ResolveMountPoint("dummy_disk")
	if err != nil {
		t.Fatalf("ResolveMountPoint failed in dry-run mode: %v", err)
	}
	if mount == "" {
		t.Errorf("Expected non-empty mount point, got empty")
	}
}

func TestFormatDiskModeB_ExtractionIntegration(t *testing.T) {
	ctx := context.Background()
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mountPoint, err := FormatDiskModeB(ctx, "test_usb_disk")
	if err != nil {
		t.Fatalf("FormatDiskModeB failed: %v", err)
	}

	// Verify that target directory exists and can be written to
	testFile := filepath.Join(mountPoint, "test.txt")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		t.Fatalf("Failed to write to dry-run mount point %s: %v", mountPoint, err)
	}
}
