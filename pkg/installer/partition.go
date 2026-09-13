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

// formatDiskMacOS formats disk on macOS using diskutil with UNIBOOT dual-partition layout (Data Partition + ESP)
func formatDiskMacOS(ctx context.Context, targetDisk string) (string, error) {
	// Normalize disk device path (e.g., /dev/disk2 -> disk2)
	diskNode := filepath.Base(targetDisk)

	// Dual-Partition Command: Partition 1 ExFAT UNIBOOT (rest of disk), Partition 2 FAT32 VTOYEFI (64MB ESP)
	cmd := execCommand("diskutil", "partitionDisk", diskNode, "MBRFormat", "ExFAT", "UNIBOOT", "0b", "FAT32", "VTOYEFI", "64M")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to eraseDisk if partitionDisk fails on specific hardware
		cmdFallback := execCommand("diskutil", "eraseDisk", "ExFAT", "UNIBOOT", "MBRFormat", diskNode)
		if fbOut, fbErr := cmdFallback.CombinedOutput(); fbErr != nil {
			return "", fmt.Errorf("diskutil partitionDisk failed (%v): %s (fallback failed: %s)", err, string(output), string(fbOut))
		}
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

// FormatDiskModeA formats the target physical disk for Mode A (Hybrid Pro Mode - Ventoy + UniBoot)
// with the specified file system (exFAT, NTFS, FAT32, ext4) and volume label "VENTOY".
// Returns the resolved volume mount point (e.g. /Volumes/VENTOY, E:\, /mnt/VENTOY).
func FormatDiskModeA(ctx context.Context, targetDisk string, fsType string) (string, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return "", fmt.Errorf("disk validation failed: %w", err)
	}

	if fsType == "" {
		fsType = "exFAT"
	}

	// Dry-run mode for tests or safe simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		tempMount, err := os.MkdirTemp("", "uniboot-dryrun-modea-*")
		if err != nil {
			return "", fmt.Errorf("failed to create dry-run mount point for Mode A: %w", err)
		}
		return tempMount, nil
	}

	switch runtime.GOOS {
	case "darwin":
		return formatDiskModeAMacOS(ctx, targetDisk, fsType)
	case "windows":
		return formatDiskModeAWindows(ctx, targetDisk, fsType)
	case "linux":
		return formatDiskModeALinux(ctx, targetDisk, fsType)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func formatDiskModeAMacOS(ctx context.Context, targetDisk string, fsType string) (string, error) {
	diskNode := filepath.Base(targetDisk)
	fsFormat := strings.ToUpper(fsType)
	if fsFormat == "EXFAT" {
		fsFormat = "ExFAT"
	} else if fsFormat == "FAT32" {
		fsFormat = "FAT32"
	}

	// Dual Partition: Partition 1 Data (fsFormat UNIBOOT), Partition 2 ESP (FAT32 VTOYEFI 64M)
	cmd := execCommand("diskutil", "partitionDisk", diskNode, "MBRFormat", fsFormat, "UNIBOOT", "0b", "FAT32", "VTOYEFI", "64M")
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmdFallback := execCommand("diskutil", "eraseDisk", fsFormat, "UNIBOOT", "MBRFormat", diskNode)
		if fbOut, fbErr := cmdFallback.CombinedOutput(); fbErr != nil {
			return "", fmt.Errorf("diskutil partitionDisk for Mode A failed (%v): %s (fallback failed: %s)", err, string(output), string(fbOut))
		}
	}

	mountPoint := "/Volumes/UNIBOOT"
	if info, err := os.Stat(mountPoint); err == nil && info.IsDir() {
		return mountPoint, nil
	}

	return ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
}

func formatDiskModeAWindows(ctx context.Context, targetDisk string, fsType string) (string, error) {
	diskIndex := targetDisk
	diskIndex = strings.TrimPrefix(diskIndex, `\\.\PhysicalDrive`)
	diskIndex = strings.TrimPrefix(diskIndex, `disk`)
	diskIndex = strings.TrimPrefix(diskIndex, `Disk`)

	fsFormat := strings.ToLower(fsType)
	if fsFormat == "" {
		fsFormat = "exfat"
	}

	scriptContent := fmt.Sprintf("select disk %s\nclean\ncreate partition primary\nactive\nformat fs=%s label=\"VENTOY\" quick\nassign\n", diskIndex, fsFormat)
	tmpFile, err := os.CreateTemp("", "diskpart-modea-*.txt")
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
		return "", fmt.Errorf("diskpart for Mode A failed (%v): %s", err, string(output))
	}

	return ResolveMountPointWithLabel(targetDisk, "VENTOY")
}

func formatDiskModeALinux(ctx context.Context, targetDisk string, fsType string) (string, error) {
	cmd := execCommand("parted", "-s", targetDisk, "mklabel", "msdos")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mklabel failed (%v): %s", err, string(output))
	}

	cmd = execCommand("parted", "-s", targetDisk, "mkpart", "primary", "1MiB", "100%")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mkpart failed (%v): %s", err, string(output))
	}

	partition := targetDisk + "1"
	if strings.Contains(targetDisk, "nvme") || strings.Contains(targetDisk, "mmcblk") {
		partition = targetDisk + "p1"
	}

	mkfsCmd := "mkfs.exfat"
	switch strings.ToLower(fsType) {
	case "fat32":
		mkfsCmd = "mkfs.vfat"
	case "ntfs":
		mkfsCmd = "mkfs.ntfs"
	case "ext4":
		mkfsCmd = "mkfs.ext4"
	}

	cmd = execCommand(mkfsCmd, "-n", "VENTOY", partition)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s failed (%v): %s", mkfsCmd, err, string(output))
	}

	mountPoint := "/mnt/VENTOY"
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
	return ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
}

// ResolveMountPointWithLabel resolves the active mount point for a specified volume label on the system.
func ResolveMountPointWithLabel(targetDisk string, label string) (string, error) {
	// macOS standard mount location
	if runtime.GOOS == "darwin" {
		macPath := filepath.Join("/Volumes", label)
		if info, err := os.Stat(macPath); err == nil && info.IsDir() {
			return macPath, nil
		}
	}

	// Fallback to checking existing system mounts or dry-run temp directory
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		return os.TempDir(), nil
	}

	return "", fmt.Errorf("could not resolve mount point for label %s on target disk %s", label, targetDisk)
}

// MountAndResolveEFIPartition resolves or automatically mounts Partition 2 (VTOYEFI / ESP) for existing Ventoy drives.
func MountAndResolveEFIPartition(targetDisk string) (string, error) {
	if runtime.GOOS == "darwin" {
		vtoyEfiPath := "/Volumes/VTOYEFI"
		if info, err := os.Stat(vtoyEfiPath); err == nil && info.IsDir() {
			return vtoyEfiPath, nil
		}
		diskNode := filepath.Base(targetDisk)
		if strings.HasPrefix(diskNode, "disk") {
			part2 := diskNode + "s2"
			cmd := execCommand("diskutil", "mount", part2)
			if err := cmd.Run(); err == nil {
				if info, err := os.Stat(vtoyEfiPath); err == nil && info.IsDir() {
					return vtoyEfiPath, nil
				}
			}
		}
	}

	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		return os.TempDir(), nil
	}

	return "", fmt.Errorf("could not resolve EFI boot partition for target disk %s", targetDisk)
}

// UpdateVolumeLabel non-destructively renames the data partition volume label to newLabel.
// Returns the updated active mount point path if changed.
func UpdateVolumeLabel(targetDisk string, mountPoint string, newLabel string) string {
	if newLabel == "" || mountPoint == "" {
		return mountPoint
	}
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		return mountPoint
	}

	if runtime.GOOS == "darwin" {
		cmd := execCommand("diskutil", "rename", mountPoint, newLabel)
		if err := cmd.Run(); err == nil {
			newMount := filepath.Join("/Volumes", newLabel)
			if info, statErr := os.Stat(newMount); statErr == nil && info.IsDir() {
				return newMount
			}
		}
		return mountPoint
	}

	if runtime.GOOS == "windows" {
		driveLetter := strings.TrimSuffix(mountPoint, "\\")
		driveLetter = strings.TrimSuffix(driveLetter, "/")
		if len(driveLetter) >= 2 && driveLetter[1] == ':' {
			cmd := execCommand("cmd", "/c", "label", driveLetter, newLabel)
			_ = cmd.Run()
		}
		return mountPoint
	}

	if runtime.GOOS == "linux" {
		part1 := targetDisk + "1"
		if strings.Contains(targetDisk, "nvme") || strings.Contains(targetDisk, "mmcblk") {
			part1 = targetDisk + "p1"
		}
		cmd := execCommand("fatlabel", part1, newLabel)
		if err := cmd.Run(); err != nil {
			cmd2 := execCommand("exfatlabel", part1, newLabel)
			_ = cmd2.Run()
		}
		return mountPoint
	}

	return mountPoint
}

