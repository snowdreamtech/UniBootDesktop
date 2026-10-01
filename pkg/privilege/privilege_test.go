// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestIsElevatedIgnoresCachedSudoTicket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows elevation uses Administrator token, not sudo")
	}
	if os.Geteuid() == 0 {
		t.Skip("process is already root")
	}

	ResetElevationCache()
	if IsElevated() {
		t.Fatal("non-root process must not report elevation from a sudo timestamp")
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
	fields, err := splitElevatedCommand("net session")
	if err != nil {
		t.Fatalf("expected a simple approved command to pass validation: %v", err)
	}
	if len(fields) != 2 || fields[0] != "net" || fields[1] != "session" {
		t.Fatalf("unexpected fields: %v", fields)
	}
}

func TestRunElevatedRejectsEmptyCommand(t *testing.T) {
	_, err := RunElevated("prompt", "   ")
	if err == nil {
		t.Fatal("expected empty command string to be rejected")
	}
}

func TestValidateRawDevicePathRejectsUnsafeInputs(t *testing.T) {
	for _, blocked := range []string{"/dev/sda", "/dev/nvme0n1", "/dev/disk0", `\\.\PhysicalDrive0`, "PhysicalDrive0", "C:"} {
		if err := ValidateRawDevicePath(blocked); err == nil {
			t.Fatalf("expected system disk path %q to be rejected", blocked)
		}
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
	if err := ValidateCommandName("chown"); err != nil {
		t.Fatal("expected chown to be allowlisted for ownership restore")
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

	infoBefore, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	origPerm := infoBefore.Mode().Perm()

	restore := RelaxRawDiskPermissionsTemporarily(path)
	infoAfter, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if infoAfter.Mode().Perm() != origPerm {
		t.Fatalf("non-device path must not be chmod'd, want %o, got %o", origPerm, infoAfter.Mode().Perm())
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
