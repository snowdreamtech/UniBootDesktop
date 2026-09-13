// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unigodesktop/pkg/disk"
)

var (
	// execCommand allows overriding exec.Command in tests
	execCommand = exec.Command
)

// FormatDiskModeB formats the target physical disk to FAT32 with MBR partition table
// and volume label "UNIBOOT" for Mode B (1-sec Cloud Pure Mode).
// Returns the resolved volume mount point (e.g. /Volumes/UNIBOOT, E:\, /mnt/UNIBOOT).
func FormatDiskModeB(ctx context.Context, targetDisk string) (string, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return "", fmt.Errorf("disk validation failed: %w", err)
	}

	// Dry-run mode for tests or safe simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		tempMount, err := os.MkdirTemp("", "uniboot-dryrun-mount-*")
		if err != nil {
			return "", fmt.Errorf("failed to create dry-run mount point: %w", err)
		}
		return tempMount, nil
	}

	switch runtime.GOOS {
	case "darwin":
		return formatDiskMacOS(ctx, targetDisk)
	case "windows":
		return formatDiskWindows(ctx, targetDisk)
	case "linux":
		return formatDiskLinux(ctx, targetDisk)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// formatDiskMacOS formats disk on macOS using diskutil
func formatDiskMacOS(ctx context.Context, targetDisk string) (string, error) {
	// Normalize disk device path (e.g., /dev/disk2 -> disk2)
	diskNode := filepath.Base(targetDisk)

	// Command: diskutil eraseDisk FAT32 UNIBOOT MBRFormat diskN
	cmd := execCommand("diskutil", "eraseDisk", "FAT32", "UNIBOOT", "MBRFormat", diskNode)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("diskutil eraseDisk failed (%v): %s", err, string(output))
	}

	mountPoint := "/Volumes/UNIBOOT"
	if info, err := os.Stat(mountPoint); err == nil && info.IsDir() {
		return mountPoint, nil
	}

	return ResolveMountPoint(targetDisk)
}

// formatDiskWindows formats disk on Windows using diskpart
func formatDiskWindows(ctx context.Context, targetDisk string) (string, error) {
	// Extract disk index if targetDisk is like "disk1" or "\\.\PhysicalDrive1" or "1"
	diskIndex := targetDisk
	diskIndex = strings.TrimPrefix(diskIndex, `\\.\PhysicalDrive`)
	diskIndex = strings.TrimPrefix(diskIndex, `disk`)
	diskIndex = strings.TrimPrefix(diskIndex, `Disk`)

	scriptContent := fmt.Sprintf("select disk %s\nclean\ncreate partition primary\nactive\nformat fs=fat32 label=\"UNIBOOT\" quick\nassign\n", diskIndex)
	tmpFile, err := os.CreateTemp("", "diskpart-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create diskpart script: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write diskpart script: %w", err)
	}
	tmpFile.Close()

	cmd := execCommand("diskpart", "/s", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("diskpart failed (%v): %s", err, string(output))
	}

	return ResolveMountPoint(targetDisk)
}

// formatDiskLinux formats disk on Linux using parted & mkfs.vfat
func formatDiskLinux(ctx context.Context, targetDisk string) (string, error) {
	// 1. Create MBR partition table
	cmd := execCommand("parted", "-s", targetDisk, "mklabel", "msdos")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mklabel failed (%v): %s", err, string(output))
	}

	// 2. Create Primary FAT32 partition
	cmd = execCommand("parted", "-s", targetDisk, "mkpart", "primary", "fat32", "1MiB", "100%")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mkpart failed (%v): %s", err, string(output))
	}

	// Partition naming convention (/dev/sdb -> /dev/sdb1, /dev/nvme0n1 -> /dev/nvme0n1p1)
	partition := targetDisk + "1"
	if strings.Contains(targetDisk, "nvme") || strings.Contains(targetDisk, "mmcblk") {
		partition = targetDisk + "p1"
	}

	// 3. Format as FAT32 with label UNIBOOT
	cmd = execCommand("mkfs.vfat", "-F", "32", "-n", "UNIBOOT", partition)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mkfs.vfat failed (%v): %s", err, string(output))
	}

	// 4. Create mount directory and mount
	mountPoint := "/mnt/UNIBOOT"
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return "", fmt.Errorf("failed to create mount dir %s: %w", mountPoint, err)
	}

	cmd = execCommand("mount", partition, mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount partition failed (%v): %s", err, string(output))
	}

	return mountPoint, nil
}

// ResolveMountPoint resolves the active mount point for volume label "UNIBOOT" on the system.
func ResolveMountPoint(targetDisk string) (string, error) {
	// macOS standard mount location
	if runtime.GOOS == "darwin" {
		macPath := "/Volumes/UNIBOOT"
		if info, err := os.Stat(macPath); err == nil && info.IsDir() {
			return macPath, nil
		}
	}

	// Fallback to checking existing system mounts or dry-run temp directory
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		return os.TempDir(), nil
	}

	return "", fmt.Errorf("could not resolve mount point for target disk %s", targetDisk)
}
