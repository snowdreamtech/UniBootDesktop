// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatDiskCloudMode_SafetyValidation(t *testing.T) {
	ctx := context.Background()

	// Test safety block against system drive
	systemDrives := []string{"/", "/dev/sda", "C:", "/dev/nvme0n1"}
	for _, drive := range systemDrives {
		_, err := FormatDiskCloudMode(ctx, drive)
		if err == nil {
			t.Errorf("Expected safety validation error for system drive %s, got nil", drive)
		}
	}
}

func TestFormatDiskCloudMode_DryRun(t *testing.T) {
	ctx := context.Background()
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mountPoint, err := FormatDiskCloudMode(ctx, "dummy_usb_disk")
	if err != nil {
		t.Fatalf("FormatDiskCloudMode dry-run failed: %v", err)
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

func TestFormatDiskCloudMode_ExtractionIntegration(t *testing.T) {
	ctx := context.Background()
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mountPoint, err := FormatDiskCloudMode(ctx, "test_usb_disk")
	if err != nil {
		t.Fatalf("FormatDiskCloudMode failed: %v", err)
	}

	// Verify that target directory exists and can be written to
	testFile := filepath.Join(mountPoint, "test.txt")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		t.Fatalf("Failed to write to dry-run mount point %s: %v", mountPoint, err)
	}
}

func TestFormatDiskHybridMode_SafetyValidation(t *testing.T) {
	ctx := context.Background()

	systemDrives := []string{"/", "/dev/sda", "C:", "/dev/nvme0n1"}
	for _, drive := range systemDrives {
		_, err := FormatDiskHybridMode(ctx, drive, "exFAT")
		if err == nil {
			t.Errorf("Expected safety validation error for system drive %s in Hybrid Mode, got nil", drive)
		}
	}
}

func TestFormatDiskHybridMode_DryRun(t *testing.T) {
	ctx := context.Background()
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mountPoint, err := FormatDiskHybridMode(ctx, "dummy_usb_disk", "exFAT")
	if err != nil {
		t.Fatalf("FormatDiskHybridMode dry-run failed: %v", err)
	}

	if mountPoint == "" {
		t.Fatalf("Expected non-empty mount point in dry-run mode for Hybrid Mode")
	}

	info, err := os.Stat(mountPoint)
	if err != nil || !info.IsDir() {
		t.Fatalf("Expected mount point %s to be a valid directory", mountPoint)
	}

	// Verify WriteVentoyConfig in dry-run directory
	if err := WriteVentoyConfig(mountPoint); err != nil {
		t.Fatalf("WriteVentoyConfig failed: %v", err)
	}

	jsonPath := filepath.Join(mountPoint, "ventoy", "ventoy.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("Expected ventoy.json to exist at %s", jsonPath)
	}

	grubPath := filepath.Join(mountPoint, "ventoy", "ventoy_grub.cfg")
	if _, err := os.Stat(grubPath); err != nil {
		t.Errorf("Expected ventoy_grub.cfg to exist at %s", grubPath)
	}

	themeTxtPath := filepath.Join(mountPoint, "ventoy", "themes", "uniboot", "theme.txt")
	if _, err := os.Stat(themeTxtPath); err != nil {
		t.Errorf("Expected theme.txt to exist at %s", themeTxtPath)
	}

	bgPath := filepath.Join(mountPoint, "ventoy", "themes", "uniboot", "background.png")
	if _, err := os.Stat(bgPath); err != nil {
		t.Errorf("Expected background.png to exist at %s", bgPath)
	}
}

