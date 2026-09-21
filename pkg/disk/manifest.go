// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MagicUniBootDisk is the official unique magic string embedded in uniboot.json.
const MagicUniBootDisk = "UNIBOOT_DISK"

// UniBootManifest defines the structured specification for ipxe/uniboot.json.
type UniBootManifest struct {
	Magic     string      `json:"magic"`                // Must equal MagicUniBootDisk ("UNIBOOT_DISK")
	Version   string      `json:"version"`              // Manifest and firmware version, e.g. "1.0.0"
	Mode      string      `json:"mode"`                 // Deployment mode: "cloud" or "hybrid"
	Arch      []string    `json:"arch,omitempty"`       // Supported architectures: ["x86_64", "arm64", "ia32"]
	Engine    *EngineInfo `json:"engine,omitempty"`     // Core boot engine information
	CreatedAt int64       `json:"created_at"`           // Creation/upgrade timestamp in Unix seconds
	UUID      string      `json:"uuid,omitempty"`       // Unique installation instance UUID
}

// EngineInfo describes the underlying bootloader engine.
type EngineInfo struct {
	Name    string `json:"name"`    // "ipxe" or "ventoy+ipxe"
	Version string `json:"version"` // Engine version string
}

// generateRandomUUID generates a random RFC4122 v4 UUID without external dependencies.
func generateRandomUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// GetManifestPath returns the canonical path to uniboot.json relative to a partition mount point.
func GetManifestPath(mountPoint string) string {
	if mountPoint == "" {
		return ""
	}
	return filepath.Join(mountPoint, "ipxe", "uniboot.json")
}

// ReadUniBootManifest attempts to read and validate uniboot.json from the ipxe/ directory of a mount point.
func ReadUniBootManifest(mountPoint string) (*UniBootManifest, error) {
	manifestPath := GetManifestPath(mountPoint)
	if manifestPath == "" {
		return nil, fmt.Errorf("empty mount point")
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var m UniBootManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}

	if m.Magic != MagicUniBootDisk {
		return nil, fmt.Errorf("magic mismatch: expected %s, got %s", MagicUniBootDisk, m.Magic)
	}

	return &m, nil
}

// HasUniBootManifest checks whether a valid UniBoot manifest exists at the specified mount point.
func HasUniBootManifest(mountPoint string) bool {
	m, err := ReadUniBootManifest(mountPoint)
	return err == nil && m != nil && m.Magic == MagicUniBootDisk
}

// WriteUniBootManifest writes or overwrites uniboot.json in the ipxe/ directory of the target mount point.
func WriteUniBootManifest(mountPoint string, mode string, version string) error {
	if mountPoint == "" {
		return fmt.Errorf("cannot write manifest to empty mount point")
	}

	ipxeDir := filepath.Join(mountPoint, "ipxe")
	if err := os.MkdirAll(ipxeDir, 0755); err != nil {
		return fmt.Errorf("failed to create ipxe directory: %w", err)
	}

	if version == "" {
		version = "1.0.0"
	}

	engineName := "ipxe"
	engineVer := "1.21.1"
	if mode == "hybrid" {
		engineName = "ventoy+ipxe"
		engineVer = "1.0.99+ipxe"
	}

	manifest := UniBootManifest{
		Magic:     MagicUniBootDisk,
		Version:   version,
		Mode:      mode,
		Arch:      []string{"x86_64", "arm64", "ia32"},
		Engine: &EngineInfo{
			Name:    engineName,
			Version: engineVer,
		},
		CreatedAt: time.Now().Unix(),
		UUID:      generateRandomUUID(),
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestPath := filepath.Join(ipxeDir, "uniboot.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write uniboot.json: %w", err)
	}

	return nil
}

// GetDiskUniBootManifest inspects mounted partitions of targetDisk and returns the valid UniBoot manifest if present.
func GetDiskUniBootManifest(targetDisk string) *UniBootManifest {
	if targetDisk == "" {
		return nil
	}

	mounts := GetDiskMountPoints(targetDisk)
	if len(mounts) == 0 {
		mounts = []string{targetDisk}
	}

	for _, mp := range mounts {
		if m, err := ReadUniBootManifest(mp); err == nil && m != nil {
			return m
		}
	}

	return nil
}

