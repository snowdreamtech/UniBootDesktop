// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package firmware

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStandardFirmwareMappings(t *testing.T) {
	mappings := GetFirmwareMappings()
	if len(mappings) == 0 {
		t.Fatalf("expected non-empty firmware mappings")
	}

	// Verify undionly.kpxe presence and reserved status
	foundUndionly := false
	for _, m := range mappings {
		if m.ReleaseName == "undionly.kpxe" {
			foundUndionly = true
			if m.TargetPath != "undionly.kpxe" {
				t.Errorf("expected TargetPath 'undionly.kpxe', got '%s'", m.TargetPath)
			}
			if !m.IsReserved {
				t.Errorf("expected IsReserved to be true for undionly.kpxe")
			}
		}
	}

	if !foundUndionly {
		t.Errorf("undionly.kpxe mapping missing from StandardFirmwareMappings")
	}
}

func TestGetMappingByReleaseName(t *testing.T) {
	mapping, ok := GetMappingByReleaseName("undionly.kpxe")
	if !ok {
		t.Fatalf("expected undionly.kpxe mapping to be found")
	}
	if mapping.TargetPath != "undionly.kpxe" {
		t.Errorf("expected target path undionly.kpxe, got %s", mapping.TargetPath)
	}

	_, okNonExistent := GetMappingByReleaseName("non-existent-firmware.bin")
	if okNonExistent {
		t.Errorf("expected non-existent firmware mapping to return false")
	}
}

func TestTargetPathForReleaseAsset(t *testing.T) {
	tests := []struct {
		releaseName string
		expected    string
	}{
		{"ipxe-x86_64.efi", "EFI/BOOT/BOOTX64.EFI"},
		{"ipxe-arm64.efi", "EFI/BOOT/BOOTAA64.EFI"},
		{"ipxe.lkrn", "ipxe.lkrn"},
		{"undionly.kpxe", "undionly.kpxe"},
		{"uniboot.ipxe", "uniboot.ipxe"},
		{"unknown.efi", ""},
	}

	for _, tt := range tests {
		got := TargetPathForReleaseAsset(tt.releaseName)
		if got != tt.expected {
			t.Errorf("TargetPathForReleaseAsset(%q) = %q; want %q", tt.releaseName, got, tt.expected)
		}
	}
}

func TestExtractFirmwareToDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := ExtractFirmwareToDir(tmpDir)
	if err != nil {
		t.Fatalf("ExtractFirmwareToDir failed: %v", err)
	}

	// Check that critical extracted files exist
	expectedFiles := []string{
		"EFI/BOOT/BOOTX64.EFI",
		"EFI/BOOT/BOOTAA64.EFI",
		"ipxe.lkrn",
		"undionly.kpxe",
		"boot.ipxe",
		"uniboot.ipxe",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(tmpDir, filepath.FromSlash(f))
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected extracted file missing: %s", fullPath)
		}
	}

	// Error test empty dir
	if err := ExtractFirmwareToDir(""); err == nil {
		t.Errorf("expected error when targetDir is empty")
	}
}

func TestExtractFirmwareModeA(t *testing.T) {
	tmpDir := t.TempDir()
	if err := ExtractFirmwareModeA(tmpDir); err != nil {
		t.Fatalf("ExtractFirmwareModeA failed: %v", err)
	}

	expectedFiles := []string{
		"ipxe/ipxe-x86_64.efi",
		"ipxe/boot.ipxe",
		"ipxe/background.png",
		"iso/UniBoot.iso",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(tmpDir, filepath.FromSlash(f))
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Mode A expected extracted file missing: %s", fullPath)
		}
	}
}

func TestExtractFirmwareModeB(t *testing.T) {
	tmpDir := t.TempDir()
	if err := ExtractFirmwareModeB(tmpDir); err != nil {
		t.Fatalf("ExtractFirmwareModeB failed: %v", err)
	}

	expectedFiles := []string{
		"EFI/BOOT/BOOTX64.EFI",
		"EFI/BOOT/BOOTAA64.EFI",
		"boot.ipxe",
		"uniboot.ipxe",
		"background.png",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(tmpDir, filepath.FromSlash(f))
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Mode B expected extracted file missing: %s", fullPath)
		}
	}

	// Mode B MUST NOT contain UniBoot.iso (saving ESP partition space)
	isoPath := filepath.Join(tmpDir, "iso", "UniBoot.iso")
	if _, err := os.Stat(isoPath); !os.IsNotExist(err) {
		t.Errorf("Mode B should NOT extract UniBoot.iso into ESP partition")
	}
}

func TestGetFirmwareData_Priority(t *testing.T) {
	// Test embedded fallback default
	data, src, err := GetFirmwareData("boot.ipxe")
	if err != nil {
		t.Fatalf("GetFirmwareData failed for boot.ipxe: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("expected non-empty firmware data")
	}
	if src == "" {
		t.Errorf("expected non-empty source label")
	}

	// Test local downloaded file override
	tmpDataDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", tmpDataDir)

	localFwDir := filepath.Join(tmpDataDir, "firmware")
	if err := os.MkdirAll(localFwDir, 0755); err != nil {
		t.Fatalf("failed to create local firmware dir: %v", err)
	}

	overrideContent := []byte("# Custom downloaded boot.ipxe override")
	if err := os.WriteFile(filepath.Join(localFwDir, "boot.ipxe"), overrideContent, 0644); err != nil {
		t.Fatalf("failed to write override file: %v", err)
	}

	overriddenData, overrideSrc, err := GetFirmwareData("boot.ipxe")
	if err != nil {
		t.Fatalf("GetFirmwareData with override failed: %v", err)
	}
	if string(overriddenData) != string(overrideContent) {
		t.Errorf("expected overridden content %q, got %q", string(overrideContent), string(overriddenData))
	}
	if overrideSrc == "" {
		t.Errorf("expected non-empty override source string")
	}
}

func TestGetLocalUniBootVersion(t *testing.T) {
	ver := GetLocalUniBootVersion()
	if ver == "" {
		t.Errorf("expected non-empty version string")
	}
}

func TestCleanDirectoryContents(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subfolder")
	dummyFile := filepath.Join(tmpDir, "stale_file.txt")

	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(dummyFile, []byte("stale content"), 0644)

	if err := CleanDirectoryContents(tmpDir); err != nil {
		t.Fatalf("CleanDirectoryContents failed: %v", err)
	}

	entries, _ := os.ReadDir(tmpDir)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after cleaning, got %d", len(entries))
	}
}

func TestFetchLatestUniBootRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	ctx := context.Background()
	info, err := FetchLatestUniBootRelease(ctx, "")
	if err != nil {
		t.Logf("FetchLatestUniBootRelease network query warning (acceptable in offline env): %v", err)
		return
	}
	if info.TagName == "" {
		t.Errorf("expected non-empty TagName from GitHub API")
	}
}

func TestSyncUniBootFirmware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	tmpDataDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", tmpDataDir)

	ctx := context.Background()
	info, err := SyncUniBootFirmware(ctx, "")
	if err != nil {
		t.Fatalf("SyncUniBootFirmware failed: %v", err)
	}
	if info.TagName == "" {
		t.Errorf("expected TagName to be set")
	}
}



