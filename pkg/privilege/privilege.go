// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	elevationMutex sync.RWMutex
	isCachedElevated bool
	cachedEvaluated  bool
)

// IsElevated checks whether the current process is running with root or administrator privileges.
func IsElevated() bool {
	elevationMutex.RLock()
	if cachedEvaluated {
		elevated := isCachedElevated
		elevationMutex.RUnlock()
		return elevated
	}
	elevationMutex.RUnlock()

	elevationMutex.Lock()
	defer elevationMutex.Unlock()

	if cachedEvaluated {
		return isCachedElevated
	}

	isCachedElevated = checkIsElevated()
	cachedEvaluated = true
	return isCachedElevated
}

// ResetElevationCache invalidates the cached privilege status.
func ResetElevationCache() {
	elevationMutex.Lock()
	defer elevationMutex.Unlock()
	cachedEvaluated = false
}

func checkIsElevated() bool {
	if runtime.GOOS == "windows" {
		// On Windows, 'net session' exits with 0 only if running as Administrator
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	}

	// On Unix-like systems (macOS, Linux), EUID == 0 indicates root privilege
	if os.Geteuid() == 0 {
		return true
	}

	// Check if active non-interactive sudo session exists
	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run() == nil
}

// RunElevated runs a command line with administrator/root privileges across operating systems.
// On macOS, it invokes AppleScript 'with administrator privileges'.
// On Linux, it leverages 'pkexec'.
// On Windows, it invokes PowerShell with 'RunAs' verb.
func RunElevated(prompt string, cmdLine string) (string, error) {
	if IsElevated() {
		// Already elevated, run directly via shell
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd.exe", "/C", cmdLine)
		} else {
			cmd = exec.Command("sh", "-c", cmdLine)
		}
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	if prompt == "" {
		prompt = "UniGoDesktop requires administrator privileges to access raw storage devices and verify boot partitions."
	}

	switch runtime.GOOS {
	case "darwin":
		// Escape double quotes and backslashes for AppleScript
		escapedCmd := strings.ReplaceAll(cmdLine, `\`, `\\`)
		escapedCmd = strings.ReplaceAll(escapedCmd, `"`, `\"`)
		escapedPrompt := strings.ReplaceAll(prompt, `"`, `\"`)

		appleScript := fmt.Sprintf(`do shell script "%s" with prompt "%s" with administrator privileges`, escapedCmd, escapedPrompt)
		cmd := exec.Command("osascript", "-e", appleScript)
		out, err := cmd.CombinedOutput()
		return string(out), err

	case "linux":
		// Leverage pkexec on Linux desktops
		cmd := exec.Command("pkexec", "sh", "-c", cmdLine)
		out, err := cmd.CombinedOutput()
		return string(out), err

	case "windows":
		// Run via PowerShell Start-Process -Verb RunAs
		psCmd := fmt.Sprintf(`Start-Process cmd.exe -ArgumentList '/c %s' -Verb RunAs -Wait`, cmdLine)
		cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", psCmd)
		out, err := cmd.CombinedOutput()
		return string(out), err

	default:
		return "", fmt.Errorf("unsupported operating system for privilege elevation: %s", runtime.GOOS)
	}
}

// ReadSector reads the first 'numBytes' (typically 512) directly from a raw physical disk device.
// Executes direct raw read when privileges permit, returning an error without blocking on interactive prompts.
func ReadSector(devicePath string, numBytes int) ([]byte, error) {
	if devicePath == "" {
		return nil, fmt.Errorf("empty device path")
	}
	if numBytes <= 0 {
		numBytes = 512
	}

	rawDevice := devicePath
	if runtime.GOOS == "darwin" && strings.HasPrefix(devicePath, "/dev/disk") && !strings.HasPrefix(devicePath, "/dev/rdisk") {
		rawDevice = "/dev/r" + strings.TrimPrefix(devicePath, "/dev/")
	}

	buf := make([]byte, numBytes)
	f, err := os.Open(rawDevice)
	if err != nil {
		f, err = os.Open(devicePath)
	}
	if err == nil {
		defer f.Close()
		n, readErr := f.Read(buf)
		if readErr == nil && n >= numBytes {
			return buf, nil
		}
	}

	// If direct open failed but process has elevation or sudo access, try sudo dd
	if IsElevated() && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		out, errDd := exec.Command("sudo", "-n", "dd", fmt.Sprintf("if=%s", rawDevice), fmt.Sprintf("bs=%d", numBytes), "count=1").Output()
		if errDd == nil && len(out) >= numBytes {
			return out[:numBytes], nil
		}
	}

	return nil, fmt.Errorf("raw sector read failed: %w", err)
}

// MountHiddenESP safely mounts an unmounted EFI / ESP / 0xEF partition to a temporary directory in read-only mode,
// returning the mounted directory path and a cleanup function.
// Uses unprivileged read-only mount when possible, avoiding unexpected GUI popups during passive scans.
func MountHiddenESP(partitionDevice string) (string, func(), error) {
	if partitionDevice == "" {
		return "", func() {}, fmt.Errorf("empty partition device")
	}

	tempDir, err := os.MkdirTemp("", "uniboot_esp_*")
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temporary mount directory: %w", err)
	}

	cleanup := func() {
		switch runtime.GOOS {
		case "darwin":
			_ = exec.Command("diskutil", "unmount", tempDir).Run()
			if os.Geteuid() == 0 {
				_ = exec.Command("umount", "-f", tempDir).Run()
			} else {
				_ = exec.Command("sudo", "-n", "umount", "-f", tempDir).Run()
			}
		case "linux":
			if os.Geteuid() == 0 {
				_ = exec.Command("umount", "-f", tempDir).Run()
			} else {
				_ = exec.Command("sudo", "-n", "umount", "-f", tempDir).Run()
			}
		}
		_ = os.RemoveAll(tempDir)
	}

	switch runtime.GOOS {
	case "darwin":
		// On macOS, attempt read-only mount via diskutil
		outDiskutil, errDiskutil := exec.Command("diskutil", "mount", "readOnly", "-mountPoint", tempDir, partitionDevice).CombinedOutput()
		if errDiskutil == nil && strings.Contains(string(outDiskutil), "mounted") {
			return tempDir, cleanup, nil
		}

		// If running with root/elevated privilege, use mount_msdos directly or via sudo
		if IsElevated() {
			var cmd *exec.Cmd
			if os.Geteuid() == 0 {
				cmd = exec.Command("mount_msdos", "-o", "rdonly", partitionDevice, tempDir)
			} else {
				cmd = exec.Command("sudo", "-n", "mount_msdos", "-o", "rdonly", partitionDevice, tempDir)
			}
			outMount, errMount := cmd.CombinedOutput()
			if errMount == nil {
				return tempDir, cleanup, nil
			}
			cleanup()
			return "", func() {}, fmt.Errorf("elevated mount failed: %s", string(outMount))
		}

	case "linux":
		// On Linux, attempt standard mount if elevated
		if IsElevated() {
			var cmd *exec.Cmd
			if os.Geteuid() == 0 {
				cmd = exec.Command("mount", "-o", "ro", partitionDevice, tempDir)
			} else {
				cmd = exec.Command("sudo", "-n", "mount", "-o", "ro", partitionDevice, tempDir)
			}
			if errMount := cmd.Run(); errMount == nil {
				return tempDir, cleanup, nil
			}
		}

	default:
		cleanup()
		return "", func() {}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	cleanup()
	return "", func() {}, fmt.Errorf("unprivileged mount unavailable for %s", partitionDevice)
}

