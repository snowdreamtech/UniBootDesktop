// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowdreamtech/unibootdesktop/pkg/config"
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

	jsonPath := filepath.Join(mountDir, "ventoy", "ventoy.json")
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("Expected ventoy.json to exist: %v", err)
	}
	if !strings.Contains(string(jsonData), "VTOY_VHD_NO_WARNING") {
		t.Errorf("Expected ventoy.json to contain VTOY_VHD_NO_WARNING setting")
	}
}

func TestBuildVentoyConfigData_PreserveCustomSections(t *testing.T) {
	existingJSON := []byte(`{
    "persistence": [
        {
            "image": "/ISO/ubuntu-24.04.iso",
            "backend": "/persistence/ubuntu.dat"
        }
    ],
    "auto_install": [
        {
            "image": "/ISO/win11.iso",
            "template": "/ventoy/script/unattend.xml"
        }
    ],
    "control": [
        { "VTOY_DEFAULT_SEARCH_ROOT": "/my_iso_dir" }
    ],
    "image_alias": [
        {
            "image": "/ISO/custom.iso",
            "alias": "My Custom Image"
        }
    ]
}`)

	mergedBytes, err := BuildVentoyConfigData(existingJSON, nil, t.TempDir())
	if err != nil {
		t.Fatalf("BuildVentoyConfigData failed: %v", err)
	}

	mergedStr := string(mergedBytes)

	// Verify user-defined sections are preserved intact
	if !strings.Contains(mergedStr, "/persistence/ubuntu.dat") {
		t.Errorf("Expected persistence section to be preserved in merged config")
	}
	if !strings.Contains(mergedStr, "/ventoy/script/unattend.xml") {
		t.Errorf("Expected auto_install section to be preserved in merged config")
	}
	if !strings.Contains(mergedStr, "/my_iso_dir") {
		t.Errorf("Expected custom VTOY_DEFAULT_SEARCH_ROOT control to be preserved")
	}
	if !strings.Contains(mergedStr, "My Custom Image") {
		t.Errorf("Expected custom image_alias to be preserved")
	}

	// Verify UniBoot settings are woven in
	if !strings.Contains(mergedStr, "VTOY_VHD_NO_WARNING") {
		t.Errorf("Expected VTOY_VHD_NO_WARNING to be woven in")
	}
	if !strings.Contains(mergedStr, "themes/uniboot/theme.txt") {
		t.Errorf("Expected UniBoot theme to be woven in")
	}
	if !strings.Contains(mergedStr, "UniBoot Network & Local Installation System") {
		t.Errorf("Expected UniBoot alias to be woven in")
	}
}

func TestWriteVentoyConfig_NonDestructiveExistingConfig(t *testing.T) {
	mountDir := t.TempDir()
	ventoyDir := filepath.Join(mountDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	existingJSON := []byte(`{
    "persistence": [
        {
            "image": "/ISO/kali.iso",
            "backend": "/persistence/kali.dat"
        }
    ]
}`)
	jsonPath := filepath.Join(ventoyDir, "ventoy.json")
	if err := os.WriteFile(jsonPath, existingJSON, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Upgrade/Rewrite
	if err := WriteVentoyConfig(mountDir); err != nil {
		t.Fatalf("WriteVentoyConfig failed on existing directory: %v", err)
	}

	readBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	readStr := string(readBytes)
	if !strings.Contains(readStr, "/persistence/kali.dat") {
		t.Errorf("Expected existing persistence configuration to be preserved")
	}
	if !strings.Contains(readStr, "themes/uniboot/theme.txt") {
		t.Errorf("Expected UniBoot theme to be injected")
	}
	if !strings.Contains(readStr, "VTOY_VHD_NO_WARNING") {
		t.Errorf("Expected VTOY_VHD_NO_WARNING to be injected")
	}
}

func TestBuildVentoyConfigData_StripBOM(t *testing.T) {
	// Simulate Windows Notepad UTF-8 with BOM (\xef\xbb\xbf)
	bomJSON := append([]byte("\xef\xbb\xbf"), []byte(`{
    "control": [
        { "VTOY_DEFAULT_SEARCH_ROOT": "/MY_ISO" }
    ]
}`)...)

	mergedBytes, err := BuildVentoyConfigData(bomJSON, nil, t.TempDir())
	if err != nil {
		t.Fatalf("BuildVentoyConfigData failed on BOM input: %v", err)
	}

	// Result must NOT start with UTF-8 BOM
	if bytes.HasPrefix(mergedBytes, []byte("\xef\xbb\xbf")) {
		t.Errorf("Output ventoy.json must not have UTF-8 BOM")
	}

	mergedStr := string(mergedBytes)
	if !strings.Contains(mergedStr, "/MY_ISO") {
		t.Errorf("Expected custom control from BOM file to be preserved")
	}
	if !strings.Contains(mergedStr, "VTOY_VHD_NO_WARNING") {
		t.Errorf("Expected UniBoot control to be woven in")
	}
}

func TestBuildVentoyConfigData_MalformedJSONError(t *testing.T) {
	brokenJSON := []byte(`{ "invalid_json": [ unclosed `)

	_, err := BuildVentoyConfigData(brokenJSON, nil, t.TempDir())
	if err == nil {
		t.Fatalf("Expected BuildVentoyConfigData to fail on malformed JSON")
	}
}

func TestWriteVentoyConfig_CorruptedConfigBackup(t *testing.T) {
	mountDir := t.TempDir()
	ventoyDir := filepath.Join(mountDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	brokenJSON := []byte(`{ "broken": [ missing close bracket `)
	jsonPath := filepath.Join(ventoyDir, "ventoy.json")
	if err := os.WriteFile(jsonPath, brokenJSON, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// WriteVentoyConfig should catch the corruption, back up to ventoy.json.corrupt.bak, and write a fresh ventoy.json
	if err := WriteVentoyConfig(mountDir); err != nil {
		t.Fatalf("WriteVentoyConfig should gracefully recover from corrupted config: %v", err)
	}

	backupPath := filepath.Join(ventoyDir, "ventoy.json.corrupt.bak")
	backupBytes, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Expected ventoy.json.corrupt.bak to exist: %v", err)
	}
	if string(backupBytes) != string(brokenJSON) {
		t.Errorf("Backup content does not match original corrupted JSON")
	}

	// New ventoy.json must be valid JSON
	newBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("ReadFile on new ventoy.json failed: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(newBytes, &parsed); err != nil {
		t.Fatalf("New ventoy.json is not valid JSON: %v", err)
	}
	if !strings.Contains(string(newBytes), "themes/uniboot/theme.txt") {
		t.Errorf("Expected new ventoy.json to contain UniBoot theme")
	}
}

func TestBuildVentoyConfigData_SecondaryMenuConfig(t *testing.T) {
	// 1. Default / False: VTOY_SECONDARY_BOOT_MENU should be "0" (direct boot)
	cfgDefault := &config.AppConfig{VentoySecondaryMenu: false}
	dataDirect, err := BuildVentoyConfigData(nil, cfgDefault, t.TempDir())
	if err != nil {
		t.Fatalf("BuildVentoyConfigData failed: %v", err)
	}
	if !strings.Contains(string(dataDirect), `"VTOY_SECONDARY_BOOT_MENU": "0"`) {
		t.Errorf("Expected VTOY_SECONDARY_BOOT_MENU to be '0' for direct boot, got: %s", string(dataDirect))
	}

	// 2. Enabled / True: VTOY_SECONDARY_BOOT_MENU should be "1" (secondary menu enabled)
	cfgSecondary := &config.AppConfig{VentoySecondaryMenu: true}
	dataSecondary, err := BuildVentoyConfigData(nil, cfgSecondary, t.TempDir())
	if err != nil {
		t.Fatalf("BuildVentoyConfigData failed: %v", err)
	}
	if !strings.Contains(string(dataSecondary), `"VTOY_SECONDARY_BOOT_MENU": "1"`) {
		t.Errorf("Expected VTOY_SECONDARY_BOOT_MENU to be '1' when enabled, got: %s", string(dataSecondary))
	}
}
