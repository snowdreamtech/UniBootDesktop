// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package qemu

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/privilege"
)

// QEMUStatus contains detection metadata for QEMU installation.
type QEMUStatus struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Version   string `json:"version"`
}

// Detect checks if QEMU executable is available on PATH or common installation directories.
func Detect() *QEMUStatus {
	candidates := []string{
		"qemu-system-x86_64",
		"qemu-system-aarch64",
		"qemu-system-i386",
	}

	commonPaths := []string{
		"/opt/local/bin", // MacPorts (macOS)
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/bin",
		`C:\Program Files\qemu`,
		`C:\Program Files (x86)\qemu`,
	}

	// 1. Try finding candidates via system PATH
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return &QEMUStatus{
				Installed: true,
				Path:      path,
				Version:   fmt.Sprintf("QEMU (%s)", name),
			}
		}
	}

	// 2. Search common installation directories directly
	for _, dir := range commonPaths {
		for _, name := range candidates {
			exeName := name
			if runtime.GOOS == "windows" {
				exeName += ".exe"
			}
			fullPath := filepath.Join(dir, exeName)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return &QEMUStatus{
					Installed: true,
					Path:      fullPath,
					Version:   fmt.Sprintf("QEMU (%s)", name),
				}
			}
		}
	}

	// Support dry-run or mock mode when UNIBOOT_DRY_RUN is set
	if os.Getenv("UNIBOOT_DRY_RUN") != "" {
		return &QEMUStatus{
			Installed: true,
			Path:      "/usr/local/bin/qemu-system-x86_64 (Dry-Run)",
			Version:   "QEMU system x86_64 (Mock)",
		}
	}

	return &QEMUStatus{
		Installed: false,
		Path:      "",
		Version:   "Not Installed",
	}
}

// DetectOVMF searches common system paths for edk2 / OVMF UEFI firmware image across macOS, Linux & Windows.
func DetectOVMF() string {
	searchPaths := []string{
		// macOS
		"/opt/local/share/qemu/edk2-x86_64-code.fd",
		"/opt/homebrew/share/qemu/edk2-x86_64-code.fd",
		// Linux
		"/usr/share/OVMF/OVMF_CODE.fd",
		"/usr/share/ovmf/OVMF.fd",
		"/usr/share/qemu/ovmf-x86_64-code.bin",
		"/usr/share/edk2/ovmf/OVMF_CODE.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_CODE.fd",
		// Windows
		`C:\Program Files\qemu\share\edk2-x86_64-code.fd`,
	}

	for _, path := range searchPaths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// extractPlistString parses a string value for a given key from a plist string.
func extractPlistString(plistStr string, key string) string {
	keyPattern := fmt.Sprintf("<key>%s</key>", key)
	idx := strings.Index(plistStr, keyPattern)
	if idx == -1 {
		return ""
	}
	rest := plistStr[idx+len(keyPattern):]
	startStr := strings.Index(rest, "<string>")
	if startStr == -1 {
		return ""
	}
	rest = rest[startStr+len("<string>"):]
	endStr := strings.Index(rest, "</string>")
	if endStr == -1 {
		return ""
	}
	return strings.TrimSpace(rest[:endStr])
}

func safeExecCommand(name string, args ...string) (*exec.Cmd, error) {
	return privilege.SafeExecCommandContext(context.Background(), name, args...)
}

// ResolveRawDiskDevice resolves volume mount paths (e.g. /Volumes/Ventoy, E:\, /mnt/UNIBOOT)
// or partition paths (e.g. /dev/disk2s1, /dev/sdb1) to raw block device paths suitable for QEMU across macOS, Linux, and Windows.
func ResolveRawDiskDevice(diskPath string) string {
	diskPath = strings.TrimSpace(diskPath)
	if diskPath == "" {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		// Case 1: Already a raw disk node (e.g. /dev/rdisk2)
		if strings.HasPrefix(diskPath, "/dev/rdisk") {
			return diskPath
		}

		// Case 2: Standard block disk node (e.g. /dev/disk2 or /dev/disk2s1)
		if strings.HasPrefix(diskPath, "/dev/disk") || strings.HasPrefix(diskPath, "/dev/rdisk") {
			baseNode := disk.NormalizeDarwinDiskNode(diskPath)
			return "/dev/r" + baseNode
		}

		// Case 3: Volume mount path (e.g. /Volumes/Ventoy, /Volumes/UNIBOOT)
		if err := privilege.ValidateRawDevicePath(diskPath); err == nil {
			diskPath = strings.TrimSpace(diskPath)
		}
		cmd, cmdErr := safeExecCommand("diskutil", "info", "-plist", diskPath)
		if cmdErr != nil {
			return diskPath
		}
		output, err := cmd.Output()
		if err == nil {
			plistStr := string(output)
			if parentDisk := extractPlistString(plistStr, "ParentWholeDisk"); parentDisk != "" {
				return "/dev/r" + parentDisk
			}
			if devNode := extractPlistString(plistStr, "DeviceNode"); devNode != "" {
				node := disk.NormalizeDarwinDiskNode(devNode)
				return "/dev/r" + node
			}
		}

		if strings.HasPrefix(diskPath, "disk") || strings.HasPrefix(filepath.Base(diskPath), "disk") {
			node := disk.NormalizeDarwinDiskNode(diskPath)
			return "/dev/r" + node
		}

	case "linux":
		// Case 1: Whole disk device like /dev/sdb or /dev/nvme0n1
		if strings.HasPrefix(diskPath, "/dev/") {
			base := diskPath
			if strings.Contains(base, "nvme") || strings.Contains(base, "mmcblk") {
				if idx := strings.LastIndex(base, "p"); idx != -1 && idx > len("/dev/nvme") {
					base = base[:idx]
				}
			} else {
				base = strings.TrimRight(base, "0123456789")
			}
			return base
		}

	case "windows":
		// Format Windows drive letter (e.g. "E:", "E:\") -> "\\.\E:" or PhysicalDrive
		cleanDrive := strings.TrimRight(diskPath, `\`)
		if len(cleanDrive) == 2 && cleanDrive[1] == ':' {
			return fmt.Sprintf(`\\.\%s`, cleanDrive)
		}
		if strings.HasPrefix(diskPath, `disk`) || strings.HasPrefix(diskPath, `Disk`) {
			diskIdx := strings.TrimPrefix(strings.TrimPrefix(diskPath, "disk"), "Disk")
			return fmt.Sprintf(`\\.\PhysicalDrive%s`, diskIdx)
		}
	}

	return diskPath
}

// ensureDiskPermissions ensures the current GUI user has read/write permissions on the target raw disk node.
func ensureDiskPermissions(targetPath string) {
	if targetPath == "" {
		return
	}
	if err := privilege.ValidateRawDevicePath(targetPath); err != nil {
		logger.Warn("Rejecting unsafe disk permission change target", "targetPath", targetPath, "error", err)
		return
	}
	f, err := os.OpenFile(targetPath, os.O_RDWR, 0)
	if err == nil {
		_ = f.Close()
		return
	}

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		rawNode := "r" + diskNode

		logger.Info("Elevating disk node permissions for QEMU GUI session via osascript", "diskNode", diskNode)
		script := fmt.Sprintf(`do shell script "chmod 666 /dev/%s /dev/%s" with administrator privileges`, rawNode, diskNode)
		cmd, err := safeExecCommand("osascript", "-e", script)
		if err == nil {
			if err := cmd.Run(); err != nil {
				logger.Warn("Failed to elevate disk node permissions via osascript", "error", err)
			}
		} else {
			logger.Warn("Rejected unsafe osascript permission escalation", "error", err)
		}
	} else if runtime.GOOS == "linux" {
		cmd, err := safeExecCommand("pkexec", "chmod", "666", targetPath)
		if err == nil {
			_ = cmd.Run()
		}
	}
}

// LaunchTest executes a non-blocking QEMU preview test instance on the target disk drive safely across macOS, Windows and Linux.
func LaunchTest(ctx context.Context, diskPath string) error {
	if diskPath == "" {
		return fmt.Errorf("please select target disk drive first")
	}

	// Dry-run mode for tests or simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(diskPath, "dummy") || strings.HasPrefix(diskPath, "test") {
		logger.Info("Dry-run QEMU simulation test executed", "disk", diskPath)
		return nil
	}

	status := Detect()
	if !status.Installed {
		logger.Error("QEMU simulator not detected on host system")
		return fmt.Errorf("QEMU simulator not detected! Please install QEMU first (e.g. via brew install qemu).")
	}

	// Resolve volume/mount path to underlying raw block device across all OSes
	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	logger.Info("Executing QEMU preview simulation test", "disk", targetPath, "qemuPath", status.Path)

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		// 1. Force unmount target disk volumes to release macOS disk arbitration lock
		unmountCmd, err := safeExecCommand("diskutil", "unmountDisk", "force", fmt.Sprintf("/dev/%s", diskNode))
		if err == nil {
			_ = unmountCmd.Run()
		}
		time.Sleep(300 * time.Millisecond)
	} else if runtime.GOOS == "linux" {
		// Try udisksctl unmount for Linux volume partitions
		unmountCmd, err := safeExecCommand("udisksctl", "unmount", "-b", diskPath)
		if err == nil {
			_ = unmountCmd.Run()
		}
	}

	// Ensure target disk node has read/write permissions for current user GUI process
	ensureDiskPermissions(targetPath)

	ovmfFw := DetectOVMF()

	// Build optimized hardware acceleration parameters
	args := []string{
		"-machine", "q35",
		"-m", "2048",
		"-device", "virtio-vga,xres=1280,yres=800",
		"-netdev", "user,id=net0",
		"-device", "e1000,netdev=net0",
	}

	// Host OS specific display and KVM acceleration
	if runtime.GOOS == "darwin" {
		args = append(args, "-display", "cocoa,zoom-to-fit=on")
	} else if runtime.GOOS == "linux" {
		if _, err := os.Stat("/dev/kvm"); err == nil {
			args = append(args, "-enable-kvm")
		}
	}

	if ovmfFw != "" {
		args = append(args, "-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", ovmfFw))
	}

	args = append(args, "-drive", fmt.Sprintf("file=%s,format=raw", targetPath))

	cmd := exec.Command(status.Path, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Detach QEMU process group from parent Wails app so the window stays alive independently
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start QEMU process: %w", err)
	}

	// Wait up to 800ms to catch startup errors (e.g. missing files or invalid args)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			errOutput := strings.TrimSpace(stderr.String())
			if errOutput != "" {
				return fmt.Errorf("QEMU launch message: %s", errOutput)
			}
			return fmt.Errorf("QEMU exited unexpectedly: %w", err)
		}
		return nil
	case <-time.After(800 * time.Millisecond):
		return nil
	}
}
