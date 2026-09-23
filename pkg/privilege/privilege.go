// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	elevationMutex   sync.RWMutex
	isCachedElevated bool
	cachedEvaluated  bool
)

// IsElevated checks whether the current process is running with root or administrator privileges,
// or possesses an active, authenticated privileged worker connection.
func IsElevated() bool {
	if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
		return true
	}

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

	// Unix elevation is the process EUID only. A cached `sudo -n` ticket must
	// not be treated as root, or later sudo -n dd/mount paths run without a prompt.
	return os.Geteuid() == 0
}

// RunElevated runs a command line with administrator/root privileges across operating systems.
// On macOS, it invokes AppleScript 'with administrator privileges'.
// On Linux, it leverages 'pkexec'.
// On Windows, it invokes PowerShell with 'RunAs' verb.
func splitElevatedCommand(cmdLine string) ([]string, error) {
	trimmed := strings.TrimSpace(cmdLine)
	if trimmed == "" {
		return nil, fmt.Errorf("command is empty")
	}
	for _, ch := range trimmed {
		if ch < 32 || ch == 127 {
			return nil, fmt.Errorf("command contains control characters")
		}
	}
	if strings.ContainsAny(trimmed, ";|`$><") {
		return nil, fmt.Errorf("command contains shell metacharacters that are not allowed")
	}
	if strings.ContainsAny(trimmed, "&()[]{}\\\"") {
		return nil, fmt.Errorf("command contains unsupported shell syntax")
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return nil, fmt.Errorf("command is empty")
	}
	if err := ValidateCommandName(fields[0]); err != nil {
		return nil, err
	}
	for _, arg := range fields[1:] {
		if err := ValidateCommandArgument(arg); err != nil {
			return nil, fmt.Errorf("unsafe argument for %q: %w", fields[0], err)
		}
	}
	return fields, nil
}

func validateElevatedCommand(cmdLine string) error {
	_, err := splitElevatedCommand(cmdLine)
	return err
}

// ValidateCommandName enforces a narrow allowlist for system commands that can be invoked
// by the application. This prevents arbitrary shell execution from spreading across the codebase.
func ValidateCommandName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("command name is empty")
	}

	base := filepath.Base(trimmed)
	base = strings.TrimSpace(base)
	if base == "" {
		return fmt.Errorf("command name is empty")
	}
	if strings.ContainsAny(base, "\x00\r\n;|`$&<> ") {
		return fmt.Errorf("command name contains unsafe characters: %q", base)
	}

	allowlist := map[string]struct{}{
		"diskutil": {}, "lsblk": {}, "umount": {}, "udisksctl": {}, "mount": {}, "mount_msdos": {},
		"chmod": {}, "dd": {}, "sudo": {}, "pkexec": {}, "osascript": {}, "net": {}, "true": {},
		"cmd.exe": {}, "powershell": {}, "powershell.exe": {}, "wmic": {}, "open": {}, "diskpart": {},
		"cmd": {}, "echo": {}, "chown": {}, "parted": {}, "mkfs.vfat": {}, "mkfs.ext4": {}, "mkfs.exfat": {}, "mkfs.ntfs": {},
	}
	lower := strings.ToLower(base)
	if _, ok := allowlist[lower]; !ok {
		return fmt.Errorf("command %q is not in the allowlist", base)
	}
	return nil
}

// ValidateCommandArgument ensures argument values do not embed shell syntax or path traversal patterns.
func ValidateCommandArgument(arg string) error {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return fmt.Errorf("command argument is empty")
	}
	if strings.ContainsAny(trimmed, "\x00\r\n") {
		return fmt.Errorf("command argument contains control characters")
	}
	if strings.Contains(trimmed, "&&") || strings.Contains(trimmed, "||") || strings.Contains(trimmed, "`") || strings.Contains(trimmed, "$(") || strings.Contains(trimmed, "${") {
		return fmt.Errorf("command argument contains unsafe shell substitution")
	}
	if strings.Contains(trimmed, ";") {
		return fmt.Errorf("command argument contains unsafe command chaining")
	}
	if strings.Contains(trimmed, "..") {
		return fmt.Errorf("command argument contains path traversal")
	}
	return nil
}

// SafeExecCommandContext runs a system command only after validating the command name and arguments.
func SafeExecCommandContext(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	if err := ValidateCommandName(name); err != nil {
		return nil, err
	}
	for _, arg := range args {
		if err := ValidateCommandArgument(arg); err != nil {
			return nil, fmt.Errorf("unsafe argument for %q: %w", name, err)
		}
	}
	return exec.CommandContext(ctx, name, args...), nil
}

const (
	// temporaryRawDiskPerm is owner-only access after the node is chown'd to the
	// current user. World-writable 0666 is never applied. Callers MUST restore
	// the original owner and mode.
	temporaryRawDiskPerm os.FileMode = 0600
	defaultRawDiskPerm   os.FileMode = 0640
)

func chmodPath(path string, mode os.FileMode) error {
	if err := os.Chmod(path, mode); err == nil {
		return nil
	}
	octal := fmt.Sprintf("%o", mode.Perm())
	_, err := RunElevated("Restore or adjust raw disk permissions", "chmod "+octal+" "+path)
	return err
}

// RelaxRawDiskPermissionsTemporarily sets a temporary mode on validated raw
// device nodes and returns a restore function that re-applies the original
// permissions exactly once. Restore is a no-op when no mode change occurred.
func RelaxRawDiskPermissionsTemporarily(paths ...string) func() {
	if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
		var validPaths []string
		for _, raw := range paths {
			path := strings.TrimSpace(raw)
			if path != "" && ValidateRawDevicePath(path) == nil {
				validPaths = append(validPaths, path)
			}
		}
		if len(validPaths) > 0 {
			if err := worker.AcquireDiskAccess(validPaths); err == nil {
				var once sync.Once
				return func() {
					once.Do(func() {
						_ = worker.ReleaseDiskAccess(validPaths)
					})
				}
			}
		}
	}

	type snapshot struct {
		path      string
		perm      os.FileMode
		uid       int
		gid       int
		haveOwner bool
	}
	var changed []snapshot
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		if err := ValidateRawDevicePath(path); err != nil {
			continue
		}
		info, err := os.Stat(path)
		orig := defaultRawDiskPerm
		uid, gid := 0, 0
		haveOwner := false
		if err == nil {
			orig = info.Mode().Perm()
			uid, gid, haveOwner = snapshotOwner(info)
		}

		// Fail closed: never fall back to world-writable modes. If we cannot
		// assign the node to the current user, skip the path.
		if err := chownPath(path, os.Getuid(), os.Getgid()); err != nil {
			continue
		}
		if err := chmodPath(path, temporaryRawDiskPerm); err != nil {
			if haveOwner {
				_ = chownPath(path, uid, gid)
			}
			continue
		}
		changed = append(changed, snapshot{path: path, perm: orig, uid: uid, gid: gid, haveOwner: haveOwner})
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			for _, item := range changed {
				if item.haveOwner {
					_ = chownPath(item.path, item.uid, item.gid)
				}
				_ = chmodPath(item.path, item.perm)
			}
		})
	}
}

// ValidateRawDevicePath rejects unsafe or clearly system-owned device paths before OS access.
func ValidateRawDevicePath(devicePath string) error {
	trimmed := strings.TrimSpace(devicePath)
	if trimmed == "" {
		return fmt.Errorf("device path is empty")
	}
	if strings.ContainsAny(trimmed, "\x00\r\n;|`$&<>") {
		return fmt.Errorf("device path contains unsafe shell syntax")
	}

	blocked := map[string]struct{}{
		"/": {}, "C:": {}, "C:\\": {},
		"/dev/sda": {}, "/dev/nvme0n1": {}, "/dev/mmcblk0": {}, "/dev/vda": {},
		"/dev/disk0": {}, "disk0": {},
		`\\.\PhysicalDrive0`: {}, "PhysicalDrive0": {},
	}
	if _, ok := blocked[trimmed]; ok {
		return fmt.Errorf("device path %q is blocked as a system-owned path", trimmed)
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^/dev/(?:r?disk\d+(?:s\d+)?|sd[a-z]+|nvme\d+n\d+|mmcblk\d+|vd[a-z]+)$`),
		regexp.MustCompile(`^/Volumes/[A-Za-z0-9_\.\-\s]+$`),
		regexp.MustCompile(`^\\\\\.\\PhysicalDrive\d+$`),
		regexp.MustCompile(`^PhysicalDrive\d+$`),
		regexp.MustCompile(`^disk\d+$`),
		regexp.MustCompile(`^[A-Za-z]:$`),
	}
	for _, pattern := range patterns {
		if pattern.MatchString(trimmed) {
			return nil
		}
	}
	return fmt.Errorf("unsupported raw device path: %q", trimmed)
}

func escapePowerShellSingleQuotedString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func buildPowerShellStartProcessCommand(filePath string, args []string) string {
	escapedPath := escapePowerShellSingleQuotedString(filePath)
	quotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, "'"+escapePowerShellSingleQuotedString(arg)+"'")
	}
	if len(quotedArgs) == 0 {
		return fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs -Wait", escapedPath)
	}
	return fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList @(%s) -Verb RunAs -Wait", escapedPath, strings.Join(quotedArgs, ", "))
}

func RunElevated(prompt string, cmdLine string) (string, error) {
	fields, err := splitElevatedCommand(cmdLine)
	if err != nil {
		return "", fmt.Errorf("unsafe elevated command: %w", err)
	}
	cmdName := fields[0]
	args := fields[1:]

	if (runtime.GOOS == "windows" && checkIsElevated()) || os.Geteuid() == 0 {
		cmd := exec.Command(cmdName, args...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
		return worker.RunCommand(cmdName, args...)
	}

	if prompt == "" {
		prompt = "UniGoDesktop requires administrator privileges to access raw storage devices and verify boot partitions."
	}

	switch runtime.GOOS {
	case "darwin":
		quotedArgs := make([]string, 0, len(fields))
		for _, field := range fields {
			quotedArgs = append(quotedArgs, "'"+strings.ReplaceAll(field, "'", "'\"'\"'")+"'")
		}
		safeCommand := strings.Join(append([]string{cmdName}, quotedArgs[1:]...), " ")
		escapedPrompt := strings.ReplaceAll(prompt, `"`, `\"`)
		appleScript := fmt.Sprintf(`do shell script "%s" with prompt "%s" with administrator privileges`, safeCommand, escapedPrompt)
		cmd := exec.Command("osascript", "-e", appleScript)
		out, err := cmd.CombinedOutput()
		return string(out), err

	case "linux":
		cmd := exec.Command("pkexec", append([]string{cmdName}, args...)...)
		out, err := cmd.CombinedOutput()
		return string(out), err

	case "windows":
		psCmd := buildPowerShellStartProcessCommand(cmdName, args)
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
	if err := ValidateRawDevicePath(devicePath); err != nil {
		return nil, fmt.Errorf("unsafe raw device path: %w", err)
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

	// If privileged worker is active, delegate directly to worker (zero sudo dependency)
	if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
		workerBuf, workerErr := worker.ReadSector(rawDevice, numBytes)
		if workerErr == nil && len(workerBuf) >= numBytes {
			return workerBuf[:numBytes], nil
		}
	}

	// If the process is actually root, retry via sudo dd as a last resort.
	// The dd arguments are tightly structured and validated so they remain usable for
	// legitimate raw-disk reads without allowing shell command injection.
	if IsElevated() && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		cmd, err := SafeExecCommandContext(context.Background(), "sudo", "-n", "dd",
			fmt.Sprintf("if=%s", rawDevice),
			fmt.Sprintf("bs=%d", numBytes),
			"count=1",
		)
		if err != nil {
			return nil, fmt.Errorf("unsafe sudo dd invocation: %w", err)
		}
		out, errDd := cmd.Output()
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
	if err := ValidateRawDevicePath(partitionDevice); err != nil {
		return "", func() {}, fmt.Errorf("unsafe partition device: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "uniboot_esp_*")
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temporary mount directory: %w", err)
	}

	cleanup := func() {
		switch runtime.GOOS {
		case "darwin":
			if cmd, err := SafeExecCommandContext(context.Background(), "diskutil", "unmount", tempDir); err == nil {
				_ = cmd.Run()
			}
			if os.Geteuid() == 0 {
				if cmd, err := SafeExecCommandContext(context.Background(), "umount", "-f", tempDir); err == nil {
					_ = cmd.Run()
				}
			} else if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
				_, _ = worker.RunCommand("umount", "-f", tempDir)
			} else {
				if cmd, err := SafeExecCommandContext(context.Background(), "sudo", "-n", "umount", "-f", tempDir); err == nil {
					_ = cmd.Run()
				}
			}
		case "linux":
			if os.Geteuid() == 0 {
				if cmd, err := SafeExecCommandContext(context.Background(), "umount", "-f", tempDir); err == nil {
					_ = cmd.Run()
				}
			} else if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
				_, _ = worker.RunCommand("umount", "-f", tempDir)
			} else {
				if cmd, err := SafeExecCommandContext(context.Background(), "sudo", "-n", "umount", "-f", tempDir); err == nil {
					_ = cmd.Run()
				}
			}
		}
		_ = os.RemoveAll(tempDir)
	}

	switch runtime.GOOS {
	case "darwin":
		// On macOS, attempt read-only mount via diskutil
		cmd, err := SafeExecCommandContext(context.Background(), "diskutil", "mount", "readOnly", "-mountPoint", tempDir, partitionDevice)
		if err == nil {
			outDiskutil, errDiskutil := cmd.CombinedOutput()
			if errDiskutil == nil && strings.Contains(string(outDiskutil), "mounted") {
				return tempDir, cleanup, nil
			}
		}

		// If running with root/elevated privilege, use mount_msdos directly or via worker
		if IsElevated() {
			if os.Geteuid() == 0 {
				mountCmd, err := SafeExecCommandContext(context.Background(), "mount_msdos", "-o", "rdonly", partitionDevice, tempDir)
				if err == nil {
					if _, errMount := mountCmd.CombinedOutput(); errMount == nil {
						return tempDir, cleanup, nil
					}
				}
			} else if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
				outMount, errMount := worker.RunCommand("mount_msdos", "-o", "rdonly", partitionDevice, tempDir)
				if errMount == nil {
					return tempDir, cleanup, nil
				}
				cleanup()
				return "", func() {}, fmt.Errorf("worker mount failed: %s", outMount)
			} else {
				mountCmd, err := SafeExecCommandContext(context.Background(), "sudo", "-n", "mount_msdos", "-o", "rdonly", partitionDevice, tempDir)
				if err == nil {
					outMount, errMount := mountCmd.CombinedOutput()
					if errMount == nil {
						return tempDir, cleanup, nil
					}
					cleanup()
					return "", func() {}, fmt.Errorf("elevated mount failed: %s", string(outMount))
				}
			}
		}

	case "linux":
		// On Linux, attempt standard mount if elevated
		if IsElevated() {
			if os.Geteuid() == 0 {
				mountCmd, err := SafeExecCommandContext(context.Background(), "mount", "-o", "ro", partitionDevice, tempDir)
				if err == nil && mountCmd.Run() == nil {
					return tempDir, cleanup, nil
				}
			} else if worker := GetActiveWorkerClient(); worker != nil && worker.IsAlive() {
				if _, errMount := worker.RunCommand("mount", "-o", "ro", partitionDevice, tempDir); errMount == nil {
					return tempDir, cleanup, nil
				}
			} else {
				mountCmd, err := SafeExecCommandContext(context.Background(), "sudo", "-n", "mount", "-o", "ro", partitionDevice, tempDir)
				if err == nil && mountCmd.Run() == nil {
					return tempDir, cleanup, nil
				}
			}
		}

	default:
		cleanup()
		return "", func() {}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	cleanup()
	return "", func() {}, fmt.Errorf("unprivileged mount unavailable for %s", partitionDevice)
}
