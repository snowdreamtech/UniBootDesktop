// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"strings"
	"testing"
)

func TestIsElevated(t *testing.T) {
	// Call IsElevated to ensure no crash and valid boolean return
	elevated := IsElevated()
	t.Logf("IsElevated returned: %v", elevated)

	// ResetElevationCache test
	ResetElevationCache()
	elevated2 := IsElevated()
	if elevated != elevated2 {
		t.Fatalf("elevation status changed after cache reset: %v != %v", elevated, elevated2)
	}
}

func TestReadSector_EmptyPath(t *testing.T) {
	_, err := ReadSector("", 512)
	if err == nil {
		t.Fatal("expected error for empty device path, got nil")
	}
}

func TestMountHiddenESP_EmptyDevice(t *testing.T) {
	_, cleanup, err := MountHiddenESP("")
	defer cleanup()
	if err == nil {
		t.Fatal("expected error for empty partition device, got nil")
	}
}

func TestRunElevatedRejectsUnsafeShellInput(t *testing.T) {
	_, err := RunElevated("prompt", "echo ok; rm -rf /")
	if err == nil {
		t.Fatal("expected unsafe command string to be rejected")
	}
}

func TestRunElevatedAllowsSimpleCommand(t *testing.T) {
	_, err := RunElevated("prompt", "net session")
	if err != nil && strings.Contains(err.Error(), "unsafe elevated command") {
		t.Fatal("expected a simple approved command to pass validation")
	}
}

func TestRunElevatedRejectsEmptyCommand(t *testing.T) {
	_, err := RunElevated("prompt", "   ")
	if err == nil {
		t.Fatal("expected empty command string to be rejected")
	}
}

func TestValidateRawDevicePathRejectsUnsafeInputs(t *testing.T) {
	if err := ValidateRawDevicePath("/dev/sda"); err == nil {
		t.Fatal("expected system disk path to be rejected")
	}
	if err := ValidateRawDevicePath("/dev/sdb"); err != nil {
		t.Fatal("expected valid removable device path to be accepted")
	}
	if err := ValidateRawDevicePath("/dev/sdb;rm -rf /"); err == nil {
		t.Fatal("expected injected command string to be rejected")
	}
}

func TestValidateCommandNameRejectsDangerousInput(t *testing.T) {
	if err := ValidateCommandName("sh"); err == nil {
		t.Fatal("expected shell command to be rejected by policy")
	}
	if err := ValidateCommandName("diskutil"); err != nil {
		t.Fatal("expected allowed system command to pass validation")
	}
}
