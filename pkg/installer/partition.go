// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
)

var (
	// execCommand allows overriding exec.Command in tests
	execCommand = exec.Command
)

// FormatDiskCloudMode formats the target physical disk to FAT32 with MBR partition table
// and volume label "UNIBOOT" for Cloud Mode (1-sec Cloud Pure Mode).
// Returns the resolved volume mount point (e.g. /Volumes/UNIBOOT, E:\, /mnt/UNIBOOT).
func FormatDiskCloudMode(ctx context.Context, targetDisk string) (string, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return "", fmt.Errorf("disk validation failed: %w", err)
	}

	// Dry-run mode for tests or safe simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		if strings.Contains(targetDisk, "fail") {
			return "", fmt.Errorf("simulated formatting failure for disk %s", targetDisk)
		}
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
	// Normalize disk device path (e.g., /dev/disk2 -> disk2, /dev/disk2s1 -> disk2)
	diskNode := disk.NormalizeDarwinDiskNode(targetDisk)

	// Strict validation: diskNode must match expected macOS pattern
	matched, err := regexp.MatchString(`^(r)?disk\d+$`, diskNode)
	if err != nil || !matched {
		return "", fmt.Errorf("invalid macOS disk node format: %s", diskNode)
	}

	logger.Info("Unmounting existing volumes on disk...", "diskNode", diskNode)
	_ = execCommand("diskutil", "unmountDisk", "force", diskNode).Run()

	logger.Info("Executing macOS diskutil partition command (MBR + ExFAT + 64MB FAT32 ESP)...", "diskNode", diskNode)

	// Dual-Partition Command: Partition 1 ExFAT Ventoy (rest of disk), Partition 2 FAT32 VTOYEFI (64MB ESP)
	cmd := execCommand("diskutil", "partitionDisk", diskNode, "2", "MBRFormat", "ExFAT", "Ventoy", "R", "FAT32", "VTOYEFI", "64M")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("diskutil partitionDisk command failed", "diskNode", diskNode, "error", string(output))
		return "", fmt.Errorf("diskutil partitionDisk failed (%v): %s", err, string(output))
	}

	logger.Info("Partitioning succeeded, mounting ESP boot partition (Partition 2)...", "diskNode", diskNode)

	// Mount Partition 2 (VTOYEFI) explicitly on macOS so dual partitions are visible in Finder/system
	part2Node := diskNode + "s2"
	_ = execCommand("diskutil", "mount", part2Node).Run()

	mountPoint := "/Volumes/Ventoy"
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

	// Strict validation: diskIndex must be numeric only
	matched, err := regexp.MatchString(`^\d+$`, diskIndex)
	if err != nil || !matched {
		return "", fmt.Errorf("invalid disk index format: %s (must be numeric)", diskIndex)
	}

	// Windows dual-partition setup using diskpart:
	// Partition 1: Primary FAT32 Ventoy (data)
	// Partition 2: Primary FAT32 VTOYEFI (64MB ESP partition at end of disk)
	scriptContent := fmt.Sprintf(
		"select disk %s\nclean\nconvert mbr\ncreate partition primary\nshrink desired=64\nactive\nformat fs=fat32 label=\"Ventoy\" quick\nassign\ncreate partition primary\nformat fs=fat32 label=\"VTOYEFI\" quick\nset id=ef\n",
		diskIndex,
	)
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
	// Strict validation: targetDisk must match expected Linux disk path pattern
	matched, err := regexp.MatchString(`^/dev/(sd[a-z]+|nvme\d+n\d+|mmcblk\d+|vd[a-z]+)$`, targetDisk)
	if err != nil || !matched {
		return "", fmt.Errorf("invalid Linux disk path format: %s", targetDisk)
	}

	// 1. Create MBR partition table
	cmd := execCommand("parted", "-s", targetDisk, "mklabel", "msdos")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mklabel failed (%v): %s", err, string(output))
	}

	// 2. Create Partition 1 (Primary Data) leaving 64MiB at the end
	cmd = execCommand("parted", "-s", targetDisk, "mkpart", "primary", "fat32", "1MiB", "-65MiB")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mkpart P1 failed (%v): %s", err, string(output))
	}
	_ = execCommand("parted", "-s", targetDisk, "set", "1", "boot", "on").Run()

	// 3. Create Partition 2 (64MiB ESP Partition)
	cmd = execCommand("parted", "-s", targetDisk, "mkpart", "primary", "fat32", "-64MiB", "100%")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mkpart P2 failed (%v): %s", err, string(output))
	}

	// Partition naming convention (/dev/sdb -> /dev/sdb1, /dev/nvme0n1 -> /dev/nvme0n1p1)
	part1 := targetDisk + "1"
	part2 := targetDisk + "2"
	if strings.Contains(targetDisk, "nvme") || strings.Contains(targetDisk, "mmcblk") {
		part1 = targetDisk + "p1"
		part2 = targetDisk + "p2"
	}

	// 4. Format Partition 1 as FAT32 with label Ventoy
	cmd = execCommand("mkfs.vfat", "-F", "32", "-n", "Ventoy", part1)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mkfs.vfat P1 failed (%v): %s", err, string(output))
	}

	// 5. Format Partition 2 as FAT32 with label VTOYEFI
	cmd = execCommand("mkfs.vfat", "-F", "32", "-n", "VTOYEFI", part2)
	if _, err := cmd.CombinedOutput(); err != nil {
		_ = execCommand("mkfs.vfat", "-F", "16", "-n", "VTOYEFI", part2).Run()
	}

	// 6. Create mount directory and mount Partition 1
	mountPoint := "/mnt/Ventoy"
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return "", fmt.Errorf("failed to create mount dir %s: %w", mountPoint, err)
	}

	cmd = execCommand("mount", part1, mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount partition failed (%v): %s", err, string(output))
	}

	// 7. Create mount directory and mount Partition 2 (VTOYEFI)
	efiMountPoint := "/mnt/VTOYEFI"
	_ = os.MkdirAll(efiMountPoint, 0755)
	_ = execCommand("mount", part2, efiMountPoint).Run()

	return mountPoint, nil
}

// FormatDiskHybridMode formats the target physical disk for Hybrid Mode (Hybrid Pro Mode - Ventoy + UniBoot)
// with the specified file system (exFAT, NTFS, FAT32, ext4) and volume label "UNIBOOT".
// Returns the resolved volume mount point (e.g. /Volumes/UNIBOOT, E:\, /mnt/UNIBOOT).
func FormatDiskHybridMode(ctx context.Context, targetDisk string, fsType string) (string, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return "", fmt.Errorf("disk validation failed: %w", err)
	}

	if fsType == "" {
		fsType = "exFAT"
	}

	// Dry-run mode for tests or safe simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		if strings.Contains(targetDisk, "fail") {
			return "", fmt.Errorf("simulated Hybrid Mode formatting failure for disk %s", targetDisk)
		}
		tempMount, err := os.MkdirTemp("", "uniboot-dryrun-hybridmode-*")
		if err != nil {
			return "", fmt.Errorf("failed to create dry-run mount point for Hybrid Mode: %w", err)
		}
		return tempMount, nil
	}

	switch runtime.GOOS {
	case "darwin":
		return formatDiskHybridModeMacOS(ctx, targetDisk, fsType)
	case "windows":
		return formatDiskHybridModeWindows(ctx, targetDisk, fsType)
	case "linux":
		return formatDiskHybridModeLinux(ctx, targetDisk, fsType)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func formatDiskHybridModeMacOS(ctx context.Context, targetDisk string, fsType string) (string, error) {
	diskNode := filepath.Base(targetDisk)
	fsFormat := strings.ToUpper(fsType)
	if fsFormat == "EXFAT" {
		fsFormat = "ExFAT"
	} else if fsFormat == "FAT32" {
		fsFormat = "FAT32"
	}

	// Dual Partition: Partition 1 Data (fsFormat Ventoy), Partition 2 ESP (FAT32 VTOYEFI 64M)
	cmd := execCommand("diskutil", "partitionDisk", diskNode, "2", "MBRFormat", fsFormat, "Ventoy", "R", "FAT32", "VTOYEFI", "64M")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("diskutil partitionDisk for Hybrid Mode failed (%v): %s", err, string(output))
	}

	// Mount Partition 2 (VTOYEFI) explicitly on macOS so dual partitions are visible in Finder/system
	part2Node := diskNode + "s2"
	_ = execCommand("diskutil", "mount", part2Node).Run()

	mountPoint := "/Volumes/Ventoy"
	if info, err := os.Stat(mountPoint); err == nil && info.IsDir() {
		return mountPoint, nil
	}

	return ResolveMountPointWithLabel(targetDisk, "Ventoy")
}

func formatDiskHybridModeWindows(ctx context.Context, targetDisk string, fsType string) (string, error) {
	diskIndex := targetDisk
	diskIndex = strings.TrimPrefix(diskIndex, `\\.\PhysicalDrive`)
	diskIndex = strings.TrimPrefix(diskIndex, `disk`)
	diskIndex = strings.TrimPrefix(diskIndex, `Disk`)

	fsFormat := strings.ToLower(fsType)
	if fsFormat == "" {
		fsFormat = "exfat"
	}

	// Windows dual-partition setup using diskpart:
	// Partition 1: Primary data partition (Ventoy)
	// Partition 2: Primary FAT32 VTOYEFI (64MB ESP partition at end of disk)
	scriptContent := fmt.Sprintf(
		"select disk %s\nclean\nconvert mbr\ncreate partition primary\nshrink desired=64\nactive\nformat fs=%s label=\"Ventoy\" quick\nassign\ncreate partition primary\nformat fs=fat32 label=\"VTOYEFI\" quick\nset id=ef\n",
		diskIndex,
		fsFormat,
	)
	tmpFile, err := os.CreateTemp("", "diskpart-hybridmode-*.txt")
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
		return "", fmt.Errorf("diskpart for Hybrid Mode failed (%v): %s", err, string(output))
	}

	return ResolveMountPointWithLabel(targetDisk, "Ventoy")
}

func formatDiskHybridModeLinux(ctx context.Context, targetDisk string, fsType string) (string, error) {
	cmd := execCommand("parted", "-s", targetDisk, "mklabel", "msdos")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mklabel failed (%v): %s", err, string(output))
	}

	cmd = execCommand("parted", "-s", targetDisk, "mkpart", "primary", "1MiB", "-65MiB")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mkpart P1 failed (%v): %s", err, string(output))
	}
	_ = execCommand("parted", "-s", targetDisk, "set", "1", "boot", "on").Run()

	cmd = execCommand("parted", "-s", targetDisk, "mkpart", "primary", "fat32", "-64MiB", "100%")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("parted mkpart P2 failed (%v): %s", err, string(output))
	}

	part1 := targetDisk + "1"
	part2 := targetDisk + "2"
	if strings.Contains(targetDisk, "nvme") || strings.Contains(targetDisk, "mmcblk") {
		part1 = targetDisk + "p1"
		part2 = targetDisk + "p2"
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

	cmd = execCommand(mkfsCmd, "-n", "Ventoy", part1)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s P1 failed (%v): %s", mkfsCmd, err, string(output))
	}

	cmd = execCommand("mkfs.vfat", "-F", "32", "-n", "VTOYEFI", part2)
	if _, err := cmd.CombinedOutput(); err != nil {
		_ = execCommand("mkfs.vfat", "-F", "16", "-n", "VTOYEFI", part2).Run()
	}

	mountPoint := "/mnt/Ventoy"
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return "", fmt.Errorf("failed to create mount dir %s: %w", mountPoint, err)
	}

	cmd = execCommand("mount", part1, mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount partition failed (%v): %s", err, string(output))
	}

	efiMountPoint := "/mnt/VTOYEFI"
	_ = os.MkdirAll(efiMountPoint, 0755)
	_ = execCommand("mount", part2, efiMountPoint).Run()

	return mountPoint, nil
}

// ResolveMountPoint resolves the active mount point for data partition on the system (UNIBOOT -> Ventoy -> VENTOY).
func ResolveMountPoint(targetDisk string) (string, error) {
	if m, err := ResolveMountPointWithLabel(targetDisk, "UNIBOOT"); err == nil && m != "" {
		return m, nil
	}
	if m, err := ResolveMountPointWithLabel(targetDisk, "Ventoy"); err == nil && m != "" {
		return m, nil
	}
	return ResolveMountPointWithLabel(targetDisk, "VENTOY")
}

// ResolveMountPointWithLabel resolves the active mount point for a specified volume label on the system.
func ResolveMountPointWithLabel(targetDisk string, label string) (string, error) {
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		if strings.Contains(targetDisk, "fail") {
			return "", fmt.Errorf("simulated mount resolution failure for disk %s", targetDisk)
		}
		return os.TempDir(), nil
	}

	// macOS target isolation
	if runtime.GOOS == "darwin" {
		diskNode := filepath.Base(targetDisk)
		if !strings.Contains(diskNode, "s") {
			diskNode = diskNode + "s1"
		}

		infoCmd := execCommand("diskutil", "info", "-plist", diskNode)
		infoOut, infoErr := infoCmd.Output()
		if infoErr == nil {
			mount := extractPlistStringValue(string(infoOut), "MountPoint")
			if mount != "" {
				if info, err := os.Stat(mount); err == nil && info.IsDir() {
					return mount, nil
				}
			}
		}

		macPath := filepath.Join("/Volumes", label)
		if info, err := os.Stat(macPath); err == nil && info.IsDir() {
			return macPath, nil
		}
	}

	return "", fmt.Errorf("could not resolve mount point for label %s on target disk %s", label, targetDisk)
}

// MountAndResolveEFIPartition resolves or automatically mounts Partition 2 (VTOYEFI / ESP) for existing Ventoy drives.
func MountAndResolveEFIPartition(targetDisk string) (string, error) {
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		if strings.Contains(targetDisk, "fail") {
			return "", fmt.Errorf("simulated EFI mount resolution failure for disk %s", targetDisk)
		}
		return os.TempDir(), nil
	}

	if runtime.GOOS == "darwin" {
		diskNode := disk.NormalizeDarwinDiskNode(targetDisk)
		part2 := diskNode + "s2"

		// 1. Check if part2 is already mounted
		infoCmd := execCommand("diskutil", "info", "-plist", part2)
		infoOut, infoErr := infoCmd.Output()
		if infoErr == nil {
			mount := extractPlistStringValue(string(infoOut), "MountPoint")
			if mount != "" {
				if info, err := os.Stat(mount); err == nil && info.IsDir() {
					return mount, nil
				}
			}
		}

		// 2. Force mount part2
		cmd := execCommand("diskutil", "mount", part2)
		_ = cmd.Run()

		// 3. Re-check mount point after diskutil mount
		infoCmd2 := execCommand("diskutil", "info", "-plist", part2)
		infoOut2, infoErr2 := infoCmd2.Output()
		if infoErr2 == nil {
			mount := extractPlistStringValue(string(infoOut2), "MountPoint")
			if mount != "" {
				if info, err := os.Stat(mount); err == nil && info.IsDir() {
					return mount, nil
				}
			}
		}

		vtoyEfiPath := "/Volumes/VTOYEFI"
		if info, err := os.Stat(vtoyEfiPath); err == nil && info.IsDir() {
			return vtoyEfiPath, nil
		}
	}

	return "", fmt.Errorf("could not resolve EFI boot partition for target disk %s", targetDisk)
}

func extractPlistStringValue(plistStr string, key string) string {
	keyTag := "<key>" + key + "</key>"
	idx := strings.Index(plistStr, keyTag)
	if idx == -1 {
		return ""
	}
	sub := plistStr[idx+len(keyTag):]
	startStr := strings.Index(sub, "<string>")
	if startStr == -1 {
		return ""
	}
	endStr := strings.Index(sub, "</string>")
	if endStr == -1 || endStr <= startStr+8 {
		return ""
	}
	return sub[startStr+8 : endStr]
}

// GetVolumeLabel retrieves the current volume label of the specified target disk or mount point.
func GetVolumeLabel(targetDisk string, mountPoint string) string {
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		if strings.Contains(targetDisk, "already_uniboot") || strings.Contains(mountPoint, "already_uniboot") || strings.Contains(mountPoint, "UNIBOOT") {
			return "UNIBOOT"
		}
		if strings.Contains(targetDisk, "ventoy") || strings.Contains(mountPoint, "ventoy") || strings.Contains(mountPoint, "Ventoy") {
			return "Ventoy"
		}
		return ""
	}

	if runtime.GOOS == "darwin" {
		if mountPoint != "" {
			cmd := execCommand("diskutil", "info", "-plist", mountPoint)
			if out, err := cmd.Output(); err == nil {
				val := extractPlistStringValue(string(out), "VolumeName")
				if val != "" {
					return val
				}
			}
		}

		if targetDisk != "" {
			diskNode := disk.NormalizeDarwinDiskNode(targetDisk)
			p1Node := diskNode
			if !strings.Contains(diskNode, "s") {
				p1Node = diskNode + "s1"
			}
			cmd := execCommand("diskutil", "info", "-plist", p1Node)
			if out, err := cmd.Output(); err == nil {
				val := extractPlistStringValue(string(out), "VolumeName")
				if val != "" {
					return val
				}
			}
		}

		if strings.HasPrefix(mountPoint, "/Volumes/") {
			base := filepath.Base(mountPoint)
			if base != "" && base != "Volumes" {
				return base
			}
		}
		return ""
	}

	if runtime.GOOS == "windows" {
		driveLetter := strings.TrimSuffix(mountPoint, "\\")
		driveLetter = strings.TrimSuffix(driveLetter, "/")
		if len(driveLetter) >= 2 && driveLetter[1] == ':' {
			cmd := execCommand("powershell", "-NoProfile", "-Command",
				fmt.Sprintf("(Get-Volume -DriveLetter %s).FileSystemLabel", string(driveLetter[0])))
			if out, err := cmd.Output(); err == nil {
				lbl := strings.TrimSpace(string(out))
				if lbl != "" {
					return lbl
				}
			}
		}
		return ""
	}

	if runtime.GOOS == "linux" {
		part1 := targetDisk + "1"
		if strings.Contains(targetDisk, "nvme") || strings.Contains(targetDisk, "mmcblk") {
			part1 = targetDisk + "p1"
		}
		cmd := execCommand("lsblk", "-no", "LABEL", part1)
		if out, err := cmd.Output(); err == nil {
			lbl := strings.TrimSpace(string(out))
			if lbl != "" {
				return lbl
			}
		}
		cmd2 := execCommand("blkid", "-s", "LABEL", "-o", "value", part1)
		if out, err := cmd2.Output(); err == nil {
			lbl := strings.TrimSpace(string(out))
			if lbl != "" {
				return lbl
			}
		}
		return ""
	}

	return ""
}

// UpdateVolumeLabel non-destructively renames the data partition volume label to newLabel.
// If the volume is already labeled with newLabel, the renaming operation is skipped.
// Returns the updated active mount point path if changed.
func UpdateVolumeLabel(targetDisk string, mountPoint string, newLabel string) string {
	if newLabel == "" || (mountPoint == "" && targetDisk == "") {
		return mountPoint
	}

	// 1. Inspect current volume label; skip if already labeled
	currentLabel := GetVolumeLabel(targetDisk, mountPoint)
	if currentLabel != "" && strings.EqualFold(strings.TrimSpace(currentLabel), newLabel) {
		logger.Info("Target volume is already labeled as requested, skipping rename operation",
			"target", targetDisk, "mountPoint", mountPoint, "label", currentLabel)
		return mountPoint
	}

	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		return mountPoint
	}

	logger.Info("Updating volume label...", "target", targetDisk, "mountPoint", mountPoint, "currentLabel", currentLabel, "newLabel", newLabel)

	if runtime.GOOS == "darwin" {
		target := mountPoint
		if target == "" {
			diskNode := disk.NormalizeDarwinDiskNode(targetDisk)
			if !strings.Contains(diskNode, "s") {
				target = diskNode + "s1"
			} else {
				target = diskNode
			}
		}
		cmd := execCommand("diskutil", "rename", target, newLabel)
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

