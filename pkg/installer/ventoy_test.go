// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeployVentoyVhdBoot_Fresh(t *testing.T) {
	tempDir := t.TempDir()
	ventoyDir := filepath.Join(tempDir, "ventoy")

	if err := DeployVentoyVhdBoot(ventoyDir); err != nil {
		t.Fatalf("DeployVentoyVhdBoot failed on fresh directory: %v", err)
	}

	targetFile := filepath.Join(ventoyDir, "ventoy_vhdboot.img")
	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Expected ventoy_vhdboot.img to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("Expected ventoy_vhdboot.img to be non-empty")
	}
}

func TestDeployVentoyVhdBoot_Idempotent(t *testing.T) {
	tempDir := t.TempDir()
	ventoyDir := filepath.Join(tempDir, "ventoy")

	// First deployment
	if err := DeployVentoyVhdBoot(ventoyDir); err != nil {
		t.Fatalf("Initial DeployVentoyVhdBoot failed: %v", err)
	}

	targetFile := filepath.Join(ventoyDir, "ventoy_vhdboot.img")
	info1, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Second deployment (idempotent)
	if err := DeployVentoyVhdBoot(ventoyDir); err != nil {
		t.Fatalf("Second DeployVentoyVhdBoot failed: %v", err)
	}

	info2, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info1.ModTime() != info2.ModTime() {
		t.Errorf("Expected file not to be overwritten if already present and non-empty")
	}
}

func TestDeployVentoyVhdBoot_PreserveCustom(t *testing.T) {
	tempDir := t.TempDir()
	ventoyDir := filepath.Join(tempDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	targetFile := filepath.Join(ventoyDir, "ventoy_vhdboot.img")
	customContent := []byte("custom-vhdboot-content-for-testing")
	if err := os.WriteFile(targetFile, customContent, 0644); err != nil {
		t.Fatalf("Failed to write custom file: %v", err)
	}

	// Deploy should recognize existing non-empty file and preserve it
	if err := DeployVentoyVhdBoot(ventoyDir); err != nil {
		t.Fatalf("DeployVentoyVhdBoot failed: %v", err)
	}

	readData, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(readData) != string(customContent) {
		t.Fatalf("Expected custom content to be preserved, got: %s", string(readData))
	}
}

func TestDeployVentoyVhdBoot_RecoverCorruptedZeroByte(t *testing.T) {
	tempDir := t.TempDir()
	ventoyDir := filepath.Join(tempDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	targetFile := filepath.Join(ventoyDir, "ventoy_vhdboot.img")
	// Create a corrupted 0-byte file
	if err := os.WriteFile(targetFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create 0-byte file: %v", err)
	}

	// Deploy should detect 0-byte file and replace it with embedded image
	if err := DeployVentoyVhdBoot(ventoyDir); err != nil {
		t.Fatalf("DeployVentoyVhdBoot failed on corrupted file: %v", err)
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("Expected 0-byte file to be replaced with valid image")
	}
}

func TestDeployVentoyWimBoot_Fresh(t *testing.T) {
	tempDir := t.TempDir()
	ventoyDir := filepath.Join(tempDir, "ventoy")

	if err := DeployVentoyWimBoot(ventoyDir); err != nil {
		t.Fatalf("DeployVentoyWimBoot failed on fresh directory: %v", err)
	}

	targetFile := filepath.Join(ventoyDir, "ventoy_wimboot.img")
	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Expected ventoy_wimboot.img to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("Expected ventoy_wimboot.img to be non-empty")
	}
}

func TestDeployVentoyWimBoot_Idempotent(t *testing.T) {
	tempDir := t.TempDir()
	ventoyDir := filepath.Join(tempDir, "ventoy")

	if err := DeployVentoyWimBoot(ventoyDir); err != nil {
		t.Fatalf("Initial DeployVentoyWimBoot failed: %v", err)
	}

	targetFile := filepath.Join(ventoyDir, "ventoy_wimboot.img")
	info1, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if err := DeployVentoyWimBoot(ventoyDir); err != nil {
		t.Fatalf("Second DeployVentoyWimBoot failed: %v", err)
	}

	info2, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info1.ModTime() != info2.ModTime() {
		t.Errorf("Expected ventoy_wimboot.img not to be overwritten if already present and non-empty")
	}
}

func TestWriteVentoyConfig_IncludesVhdAndWimBoot(t *testing.T) {
	mountDir := t.TempDir()

	if err := WriteVentoyConfig(mountDir); err != nil {
		t.Fatalf("WriteVentoyConfig failed: %v", err)
	}

	vhdBootPath := filepath.Join(mountDir, "ventoy", "ventoy_vhdboot.img")
	infoVhd, err := os.Stat(vhdBootPath)
	if err != nil {
		t.Fatalf("Expected ventoy_vhdboot.img to exist in mountDir: %v", err)
	}
	if infoVhd.Size() == 0 {
		t.Fatalf("Expected ventoy_vhdboot.img to have non-zero size")
	}

	wimBootPath := filepath.Join(mountDir, "ventoy", "ventoy_wimboot.img")
	infoWim, err := os.Stat(wimBootPath)
	if err != nil {
		t.Fatalf("Expected ventoy_wimboot.img to exist in mountDir: %v", err)
	}
	if infoWim.Size() == 0 {
		t.Fatalf("Expected ventoy_wimboot.img to have non-zero size")
	}
}
