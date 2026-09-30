// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPlistStringValue(t *testing.T) {
	plistSample := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>MountPoint</key>
	<string>/Volumes/UNIBOOT</string>
	<key>VolumeName</key>
	<string>UNIBOOT</string>
</dict>
</plist>`

	assert.Equal(t, "/Volumes/UNIBOOT", extractPlistStringValue(plistSample, "MountPoint"))
	assert.Equal(t, "UNIBOOT", extractPlistStringValue(plistSample, "VolumeName"))
	assert.Equal(t, "", extractPlistStringValue(plistSample, "NonExistentKey"))
}

func TestGetVolumeLabel_DryRun(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	// 1. Target already labeled as UNIBOOT
	labelUniBoot := GetVolumeLabel("dummy_disk_already_uniboot", "/Volumes/UNIBOOT")
	assert.Equal(t, "UNIBOOT", labelUniBoot)

	// 2. Target labeled as Ventoy
	labelVentoy := GetVolumeLabel("dummy_disk_ventoy", "/Volumes/Ventoy")
	assert.Equal(t, "Ventoy", labelVentoy)

	// 3. Unknown target
	labelEmpty := GetVolumeLabel("dummy_disk", "")
	assert.Equal(t, "", labelEmpty)
}

func TestUpdateVolumeLabel_SkipWhenAlreadyUniBoot(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	// Target volume is already named UNIBOOT
	mountPoint := "/Volumes/UNIBOOT"
	targetDisk := "dummy_usb_already_uniboot"

	// Should skip renaming and return original mount point immediately
	resMount := UpdateVolumeLabel(targetDisk, mountPoint, "UNIBOOT")
	assert.Equal(t, mountPoint, resMount)

	// Lowercase uniboot check
	resMountLower := UpdateVolumeLabel(targetDisk, mountPoint, "uniboot")
	assert.Equal(t, mountPoint, resMountLower)
}

func TestUpdateVolumeLabel_RenamesWhenDifferent(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	// Target disk currently named Ventoy
	mountPoint := filepath.Join(os.TempDir(), "uniboot_test_mount")
	_ = os.MkdirAll(mountPoint, 0755)
	defer os.RemoveAll(mountPoint)

	targetDisk := "dummy_usb_ventoy"

	// Calling UpdateVolumeLabel with target UNIBOOT
	resMount := UpdateVolumeLabel(targetDisk, mountPoint, "UNIBOOT")
	// In dry-run mode, when it proceeds past the skip check, it returns mountPoint
	assert.Equal(t, mountPoint, resMount)
}

func TestUpdateVolumeLabel_EmptyParams(t *testing.T) {
	assert.Equal(t, "", UpdateVolumeLabel("", "", ""))
	assert.Equal(t, "/test/path", UpdateVolumeLabel("dummy", "/test/path", ""))
	assert.Equal(t, "", UpdateVolumeLabel("", "", "UNIBOOT"))
}
