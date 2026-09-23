// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"os"
	"path/filepath"
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

func TestValidateCommandArgumentRejectsShellMetacharacters(t *testing.T) {
	for _, bad := range []string{"a; rm -rf /", "a&&b", "$(id)", "`whoami`", "../etc/passwd"} {
		if err := ValidateCommandArgument(bad); err == nil {
			t.Fatalf("expected shell injection-like argument %q to be rejected", bad)
		}
	}
}

func TestSplitElevatedCommandRejectsUnsupportedShellSyntax(t *testing.T) {
	if _, err := splitElevatedCommand("diskutil list; echo pwned"); err == nil {
		t.Fatal("expected semicolon-based command chaining to be rejected")
	}
	if _, err := splitElevatedCommand("diskutil list"); err != nil {
		t.Fatal("expected a simple approved command to pass validation")
	}
}

func TestRelaxRawDiskPermissionsTemporarilyRestoresOwnedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "disk-node")
	if err := os.WriteFile(path, []byte("x"), 0640); err != nil {
		t.Fatal(err)
	}

	restore := RelaxRawDiskPermissionsTemporarily(path)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Fatalf("non-device path must not be chmod'd, got %o", info.Mode().Perm())
	}
	restore()
}

func TestRelaxRawDiskPermissionsTemporarilyIgnoresEmptyPath(t *testing.T) {
	restore := RelaxRawDiskPermissionsTemporarily("", "   ")
	restore()
	restore()
}

func TestBuildPowerShellStartProcessCommandEscapesQuotes(t *testing.T) {
	cmd := buildPowerShellStartProcessCommand("net", []string{"session", "user'admin", "value with spaces"})
	if !strings.Contains(cmd, "user''admin") {
		t.Fatalf("expected apostrophes to be escaped in PowerShell arguments, got %q", cmd)
	}
	if !strings.Contains(cmd, "-ArgumentList @('") {
		t.Fatalf("expected PowerShell argument list array format, got %q", cmd)
	}
}
