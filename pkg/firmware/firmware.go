// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package firmware

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
)

//go:embed assets/*
var embeddedAssets embed.FS

// FirmwareMapping defines the mapping between a UniBoot release asset name and its UEFI/BIOS standard target path.
type FirmwareMapping struct {
	ReleaseName string `json:"releaseName"` // Original Release asset filename (e.g. ipxe-x86_64.efi, undionly.kpxe)
	TargetPath  string `json:"targetPath"`  // Standard target path in UNIBOOTEFI / U-disk (e.g. EFI/BOOT/BOOTX64.EFI, undionly.kpxe)
	Description string `json:"description"` // Architecture / target description
	IsReserved  bool   `json:"isReserved"`  // Whether this is a reserved / optional firmware module (e.g. undionly.kpxe)
}

// UniBootReleaseAsset represents a file asset attached to a UniBoot GitHub release.
type UniBootReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UniBootReleaseInfo represents release metadata fetched from GitHub API for UniBoot repository.
type UniBootReleaseInfo struct {
	TagName     string                `json:"tagName"`
	Name        string                `json:"name"`
	PublishedAt string                `json:"publishedAt"`
	Body        string                `json:"body"`
	Assets      []UniBootReleaseAsset `json:"assets"`
	LocalTag    string                `json:"localTag"`
	HasUpdate   bool                  `json:"hasUpdate"`
}

// StandardFirmwareMappings defines the full matrix of UniBoot firmware files to be deployed.
var StandardFirmwareMappings = []FirmwareMapping{
	// UEFI Architectures
	{ReleaseName: "ipxe-x86_64.efi", TargetPath: "EFI/BOOT/BOOTX64.EFI", Description: "UEFI x86_64 (Intel/AMD 64-bit)", IsReserved: false},
	{ReleaseName: "ipxe-arm64.efi", TargetPath: "EFI/BOOT/BOOTAA64.EFI", Description: "UEFI ARM64 (Apple Silicon Mac / ARM Server)", IsReserved: false},
	{ReleaseName: "ipxe-arm.efi", TargetPath: "EFI/BOOT/BOOTARM.EFI", Description: "UEFI ARM32 (32-bit ARM)", IsReserved: false},
	{ReleaseName: "ipxe-i386.efi", TargetPath: "EFI/BOOT/BOOTIA32.EFI", Description: "UEFI IA32 (32-bit x86 Tablets/Board)", IsReserved: false},
	{ReleaseName: "ipxe-loongarch64.efi", TargetPath: "EFI/BOOT/BOOTLOONGARCH64.EFI", Description: "UEFI LoongArch64 (Loongson 64-bit)", IsReserved: false},
	{ReleaseName: "ipxe-riscv64.efi", TargetPath: "EFI/BOOT/BOOTRISCV64.EFI", Description: "UEFI RISC-V 64-bit", IsReserved: false},
	{ReleaseName: "ipxe-riscv32.efi", TargetPath: "EFI/BOOT/BOOTRISCV32.EFI", Description: "UEFI RISC-V 32-bit", IsReserved: false},

	// Legacy BIOS / MBR Boot Images
	{ReleaseName: "ipxe.lkrn", TargetPath: "ipxe.lkrn", Description: "Legacy BIOS U-disk MBR Kernel Boot Image (x86)", IsReserved: false},
	{ReleaseName: "ipxe-riscv64.lkrn", TargetPath: "ipxe-riscv64.lkrn", Description: "Legacy MBR Kernel Boot Image (RISC-V 64-bit)", IsReserved: false},
	{ReleaseName: "ipxe-riscv32.lkrn", TargetPath: "ipxe-riscv32.lkrn", Description: "Legacy MBR Kernel Boot Image (RISC-V 32-bit)", IsReserved: false},
	{ReleaseName: "undionly.kpxe", TargetPath: "undionly.kpxe", Description: "Legacy BIOS UNDI PXE Network Boot Firmware", IsReserved: true},

	// Entry Scripts & Theme Background Assets
	{ReleaseName: "boot.ipxe", TargetPath: "boot.ipxe", Description: "iPXE Global Entry Script", IsReserved: false},
	{ReleaseName: "uniboot.ipxe", TargetPath: "uniboot.ipxe", Description: "UniBoot Main Interactive Menu Script", IsReserved: false},
	{ReleaseName: "background.png", TargetPath: "background.png", Description: "UniBoot 1:1 Ventoy Theme Background Image (PNG)", IsReserved: false},
	{ReleaseName: "UniBoot.iso", TargetPath: "iso/UniBoot.iso", Description: "UniBoot Full UEFI/BIOS Hybrid Boot ISO Image", IsReserved: true},
}

// GetFirmwareMappings returns a copy of all standard firmware mappings.
func GetFirmwareMappings() []FirmwareMapping {
	mappings := make([]FirmwareMapping, len(StandardFirmwareMappings))
	copy(mappings, StandardFirmwareMappings)
	return mappings
}

// GetMappingByReleaseName looks up a firmware mapping by its release asset name.
func GetMappingByReleaseName(name string) (FirmwareMapping, bool) {
	for _, m := range StandardFirmwareMappings {
		if m.ReleaseName == name {
			return m, true
		}
	}
	return FirmwareMapping{}, false
}

// TargetPathForReleaseAsset returns the target path for a given release asset filename,
// or empty string if not recognized.
func TargetPathForReleaseAsset(name string) string {
	if m, ok := GetMappingByReleaseName(name); ok {
		return m.TargetPath
	}
	return ""
}

// GetLocalUniBootVersion returns current local/cached UniBoot version tag.
func GetLocalUniBootVersion() string {
	versionFile := filepath.Join(env.GetDataDir(), "firmware", "version.json")
	if data, err := os.ReadFile(versionFile); err == nil {
		var ver struct {
			TagName string `json:"tagName"`
		}
		if err := json.Unmarshal(data, &ver); err == nil && ver.TagName != "" {
			return ver.TagName
		}
	}
	return "v1.0.0 (Embedded)"
}

// FetchLatestUniBootRelease queries https://api.github.com/repos/snowdreamtech/UniBoot/releases/latest.
func FetchLatestUniBootRelease(ctx context.Context, proxyPrefix string) (*UniBootReleaseInfo, error) {
	apiURL := "https://api.github.com/repos/snowdreamtech/UniBoot/releases/latest"
	finalURL := updater.BuildProxyURL(apiURL, proxyPrefix)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "UniBootDesktop/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)

	// Fallback to direct URL if proxy request fails
	if (err != nil || resp.StatusCode != http.StatusOK) && finalURL != apiURL {
		if resp != nil {
			resp.Body.Close()
		}
		fallbackReq, fallbackErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if fallbackErr == nil {
			fallbackReq.Header.Set("Accept", "application/vnd.github.v3+json")
			fallbackReq.Header.Set("User-Agent", "UniBootDesktop/1.0")
			resp, err = client.Do(fallbackReq)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query UniBoot latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status code %d for %s", resp.StatusCode, finalURL)
	}

	var ghRelease struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
		Body        string `json:"body"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release response: %w", err)
	}

	localTag := GetLocalUniBootVersion()
	hasUpdate := localTag != ghRelease.TagName

	releaseInfo := &UniBootReleaseInfo{
		TagName:     ghRelease.TagName,
		Name:        ghRelease.Name,
		PublishedAt: ghRelease.PublishedAt,
		Body:        ghRelease.Body,
		LocalTag:    localTag,
		HasUpdate:   hasUpdate,
	}

	for _, a := range ghRelease.Assets {
		releaseInfo.Assets = append(releaseInfo.Assets, UniBootReleaseAsset{
			Name:               a.Name,
			BrowserDownloadURL: a.BrowserDownloadURL,
			Size:               a.Size,
		})
	}

	return releaseInfo, nil
}

// SyncUniBootFirmware downloads release assets from latest UniBoot release into GetDataDir()/firmware/<releaseName>.
func SyncUniBootFirmware(ctx context.Context, proxyPrefix string) (*UniBootReleaseInfo, error) {
	rel, err := FetchLatestUniBootRelease(ctx, proxyPrefix)
	if err != nil {
		return nil, err
	}

	firmwareDir := filepath.Join(env.GetDataDir(), "firmware")
	if err := os.MkdirAll(firmwareDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create firmware cache dir: %w", err)
	}

	expectedAssets := make(map[string]bool)
	for _, m := range StandardFirmwareMappings {
		expectedAssets[m.ReleaseName] = true
	}

	downloadedCount := 0
	var firstErr error

	for _, asset := range rel.Assets {
		if expectedAssets[asset.Name] {
			destPath := filepath.Join(firmwareDir, asset.Name)
			if err := updater.DownloadFileWithProxy(ctx, asset.BrowserDownloadURL, destPath, proxyPrefix); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			downloadedCount++
		}
	}

	if downloadedCount == 0 && firstErr != nil {
		return nil, fmt.Errorf("firmware upgrade failed: %w", firstErr)
	}

	versionFile := filepath.Join(firmwareDir, "version.json")
	versionData, _ := json.MarshalIndent(map[string]interface{}{
		"tagName":   rel.TagName,
		"updatedAt": time.Now().Format(time.RFC3339),
		"count":     downloadedCount,
	}, "", "  ")
	_ = os.WriteFile(versionFile, versionData, 0644)

	rel.LocalTag = rel.TagName
	rel.HasUpdate = false

	return rel, nil
}

// GetFirmwareData retrieves binary data for a firmware asset based on priority:
// Priority 1: User downloaded / cached firmware in GetDataDir()/firmware/<releaseName>
// Priority 2: Built-in embedded binary (embed.FS)
func GetFirmwareData(releaseName string) ([]byte, string, error) {
	// Check user data directory for manually downloaded / updated firmware
	localPath := filepath.Join(env.GetDataDir(), "firmware", releaseName)
	if info, err := os.Stat(localPath); err == nil && info.Size() > 0 {
		data, err := os.ReadFile(localPath)
		if err == nil {
			return data, fmt.Sprintf("Local Cache (%s)", localPath), nil
		}
	}

	// Fallback: Read from built-in embedded binary
	embeddedPath := "assets/" + releaseName
	data, err := embeddedAssets.ReadFile(embeddedPath)
	if err != nil {
		return nil, "", fmt.Errorf("firmware asset %s not found: %w", releaseName, err)
	}
	return data, "Embedded Binary (embed.FS)", nil
}

// ExtractFirmwareToDir extracts firmware files to the specified target directory using priority selection.
// It populates standard UEFI/BIOS paths (EFI/BOOT/...) as well as UniBoot structure (ipxe/, iso/).
func ExtractFirmwareToDir(targetDir string) error {
	if targetDir == "" {
		return fmt.Errorf("target directory cannot be empty")
	}

	for _, mapping := range StandardFirmwareMappings {
		data, _, err := GetFirmwareData(mapping.ReleaseName)
		if err != nil {
			return fmt.Errorf("failed to load firmware asset %s: %w", mapping.ReleaseName, err)
		}

		// 1. Primary target path (e.g. EFI/BOOT/BOOTX64.EFI or root)
		destPath := filepath.Join(targetDir, filepath.FromSlash(mapping.TargetPath))
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to extract firmware asset to %s: %w", destPath, err)
		}

		// 2. Extra UniBoot structure sync: populate ipxe/ for Ventoy menu integration
		if strings.HasSuffix(mapping.ReleaseName, ".efi") ||
			strings.HasSuffix(mapping.ReleaseName, ".lkrn") ||
			strings.HasSuffix(mapping.ReleaseName, ".kpxe") ||
			strings.HasSuffix(mapping.ReleaseName, ".png") ||
			strings.HasSuffix(mapping.ReleaseName, ".ipxe") {
			ipxePath := filepath.Join(targetDir, "ipxe", mapping.ReleaseName)
			if err := os.MkdirAll(filepath.Dir(ipxePath), 0755); err == nil {
				_ = os.WriteFile(ipxePath, data, 0644)
			}
		}
	}

	return nil
}

// ExtractFirmwareModeA extracts Mode A assets (ipxe/ & iso/) directly to the main data partition (Partition 1),
// without redundantly polluting the data partition with EFI/BOOT/ files (handled by Ventoy Partition 2).
func ExtractFirmwareModeA(dataMountDir string) error {
	if dataMountDir == "" {
		return fmt.Errorf("data mount directory cannot be empty")
	}

	for _, mapping := range StandardFirmwareMappings {
		data, _, err := GetFirmwareData(mapping.ReleaseName)
		if err != nil {
			return fmt.Errorf("failed to load firmware asset %s: %w", mapping.ReleaseName, err)
		}

		// 1. Populate ipxe/ directory for Ventoy F6 custom menu integration
		if strings.HasSuffix(mapping.ReleaseName, ".efi") ||
			strings.HasSuffix(mapping.ReleaseName, ".lkrn") ||
			strings.HasSuffix(mapping.ReleaseName, ".kpxe") ||
			strings.HasSuffix(mapping.ReleaseName, ".png") ||
			strings.HasSuffix(mapping.ReleaseName, ".ipxe") {
			ipxePath := filepath.Join(dataMountDir, "ipxe", mapping.ReleaseName)
			if err := os.MkdirAll(filepath.Dir(ipxePath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", ipxePath, err)
			}
			if err := os.WriteFile(ipxePath, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", ipxePath, err)
			}
		}

		// 2. Populate iso/ directory for UniBoot ISO placement
		if mapping.ReleaseName == "UniBoot.iso" {
			isoPath := filepath.Join(dataMountDir, "iso", "UniBoot.iso")
			if err := os.MkdirAll(filepath.Dir(isoPath), 0755); err == nil {
				_ = os.WriteFile(isoPath, data, 0644)
			}
		}
	}
	return nil
}

// CleanDirectoryContents removes all files and subdirectories inside dirPath without deleting dirPath itself.
func CleanDirectoryContents(dirPath string) error {
	if dirPath == "" || dirPath == "/" || dirPath == "." {
		return fmt.Errorf("refusing to clean unsafe root directory: %q", dirPath)
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())
		_ = os.RemoveAll(path)
	}
	return nil
}

// ExtractFirmwareModeB extracts Mode B assets (EFI/BOOT/ & root scripts & background image) directly to ESP partition (Partition 2),
// providing 100% native iPXE cloud boot matching 1:1 Ventoy theme design.
func ExtractFirmwareModeB(efiMountDir string) error {
	if efiMountDir == "" {
		return fmt.Errorf("EFI mount directory cannot be empty")
	}

	// Wipe all stale files/directories inside ESP partition before extracting fresh Mode B firmware
	_ = CleanDirectoryContents(efiMountDir)

	for _, mapping := range StandardFirmwareMappings {
		// Mode B executes directly from ESP partition (64MB) and does NOT need 18.5MB UniBoot.iso
		if mapping.ReleaseName == "UniBoot.iso" || (mapping.IsReserved && strings.HasSuffix(mapping.ReleaseName, ".iso")) {
			continue
		}

		data, _, err := GetFirmwareData(mapping.ReleaseName)
		if err != nil {
			return fmt.Errorf("failed to load firmware asset %s: %w", mapping.ReleaseName, err)
		}

		// Mode B writes EFI/BOOT/ BOOTX64.EFI, BOOTAA64.EFI, boot.ipxe, uniboot.ipxe, background.png to ESP partition
		destPath := filepath.Join(efiMountDir, filepath.FromSlash(mapping.TargetPath))
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to extract firmware asset to %s: %w", destPath, err)
		}
	}
	return nil
}

