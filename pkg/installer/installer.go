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
	"github.com/snowdreamtech/unigodesktop/pkg/firmware"
)

// DeployResult contains the output metadata of a USB deployment run.
type DeployResult struct {
	Success bool   `json:"success"`
	Mode    string `json:"mode"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

// DeployModeA executes Mode A: Hybrid Pro Mode (Ventoy + UniBoot theme + iPXE network extension) with customizable file system.
// Performs non-destructive in-place upgrade on existing Ventoy drives, or fresh partition initialization on blank drives.
func DeployModeA(ctx context.Context, targetDisk string, fsType string) (*DeployResult, error) {
	return DeployModeAWithIsoAndVentoyPath(ctx, targetDisk, fsType, "", nil, nil)
}

// DeployModeAWithVentoyPath executes Mode A with an optional user-configured Ventoy CLI executable path.
func DeployModeAWithVentoyPath(ctx context.Context, targetDisk string, fsType string, ventoyPath string) (*DeployResult, error) {
	return DeployModeAWithIsoAndVentoyPath(ctx, targetDisk, fsType, ventoyPath, nil, nil)
}

// DeployModeAWithIsoAndVentoyPath executes Mode A with customizable Ventoy CLI path, ISO file paths, and progress callback.
func DeployModeAWithIsoAndVentoyPath(ctx context.Context, targetDisk string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback) (*DeployResult, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return nil, fmt.Errorf("disk validation failed: %w", err)
	}

	if fsType == "" {
		fsType = "exFAT"
	}

	// 1. Differential Treatment: Check if target drive is ALREADY a REAL active Ventoy drive (with Ventoy MBR)
	isExistingVentoy := disk.IsRealVentoyDisk(targetDisk)
	var mountPoint string
	var err error

	if isExistingVentoy {
		// Scenario A: Existing Ventoy Drive -> Non-destructive in-place upgrade (Preserves user ISO files!)
		mountPoint, err = ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
		if err != nil {
			mountPoint, err = ResolveMountPointWithLabel(targetDisk, "Ventoy")
		}
		if err != nil {
			mountPoint, err = ResolveMountPointWithLabel(targetDisk, "VENTOY")
		}
		if err != nil {
			mountPoint, err = FormatDiskModeA(ctx, targetDisk, fsType)
		}
	} else {
		// Scenario B: Blank / Ordinary USB Drive -> Fresh initialization & formatting
		// Mode A on a blank disk STRICTLY requires a valid Ventoy directory!
		val := ValidateVentoyCli(ventoyPath)
		if !val.Valid {
			return nil, fmt.Errorf("无法制作 Mode A (混合模式)：目标 U 盘为全新纯净盘，且未检测到有效的 Ventoy 目录。请先在【设置】中配置并检测 Ventoy 目录 (%s)", val.Message)
		}
		mountPoint, err = FormatDiskWithVentoyCli(ctx, ventoyPath, targetDisk, fsType)
	}
	if err != nil {
		return nil, fmt.Errorf("preparing disk for Mode A failed: %w", err)
	}

	// 2. Extract multi-arch iPXE EFI & Legacy BIOS firmware assets to target volume
	if err := firmware.ExtractFirmwareModeA(mountPoint); err != nil {
		return nil, fmt.Errorf("extracting firmware assets failed: %w", err)
	}

	// 3. Write Ventoy configuration (ventoy.json, ventoy_grub.cfg, themes/uniboot)
	if err := WriteVentoyConfig(mountPoint); err != nil {
		return nil, fmt.Errorf("writing Ventoy configuration failed: %w", err)
	}

	// 4. Copy selected local ISO/IMG system images to target drive (/iso/ directory)
	if len(isoPaths) > 0 {
		if err := CopyIsoFilesToDisk(mountPoint, isoPaths, progressCb); err != nil {
			return nil, fmt.Errorf("copying selected ISO/IMG files failed: %w", err)
		}
	}

	// 5. Non-destructively update volume label of data partition to UNIBOOT
	mountPoint = UpdateVolumeLabel(targetDisk, mountPoint, "UNIBOOT")

	msg := fmt.Sprintf("Successfully deployed Hybrid Pro Mode A (%s/UNIBOOT) to %s (mount: %s)", fsType, targetDisk, mountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully upgraded existing Ventoy drive to UniBoot Hybrid Pro Mode A at %s (ISO data preserved)", targetDisk)
	}
	if len(isoPaths) > 0 {
		msg += fmt.Sprintf(" (%d ISO/IMG file(s) copied)", len(isoPaths))
	}

	return &DeployResult{
		Success: true,
		Mode:    fmt.Sprintf("Mode A (Hybrid Pro - %s)", fsType),
		Target:  targetDisk,
		Message: msg,
	}, nil
}


// DeployModeABatch executes Mode A on multiple target USB drives with specified file system.
func DeployModeABatch(ctx context.Context, targetDisks []string, fsType string) ([]*DeployResult, error) {
	return DeployModeABatchWithIso(ctx, targetDisks, fsType, nil, nil)
}

// DeployModeABatchWithIso executes Mode A on multiple target USB drives with optional ISO files and progress reporting.
func DeployModeABatchWithIso(ctx context.Context, targetDisks []string, fsType string, isoPaths []string, progressCb CopyIsoProgressCallback) ([]*DeployResult, error) {
	if len(targetDisks) == 0 {
		return nil, fmt.Errorf("no target disks specified for batch deployment")
	}

	for _, d := range targetDisks {
		if err := disk.ValidateTargetDisk(d); err != nil {
			return nil, fmt.Errorf("disk validation failed for %s: %w", d, err)
		}
	}

	results := make([]*DeployResult, 0, len(targetDisks))
	for _, d := range targetDisks {
		res, err := DeployModeAWithIsoAndVentoyPath(ctx, d, fsType, "", isoPaths, progressCb)
		if err != nil {
			results = append(results, &DeployResult{
				Success: false,
				Mode:    "Mode A (Hybrid Pro)",
				Target:  d,
				Message: err.Error(),
			})
			continue
		}
		results = append(results, res)
	}
	return results, nil
}

// CleanMbrBootstrapCode zero-fills bytes 0..445 of Sector 0 on targetDisk,
// preserving bytes 446-511 (Partition Table & MBR Signature) 100% intact.
// This neutralizes stale Ventoy MBR hooks when converting Mode A to Mode B, preventing Legacy BIOS boot crashes.
func CleanMbrBootstrapCode(targetDisk string) error {
	diskNode := filepath.Base(targetDisk)
	if idx := strings.Index(diskNode, "s"); idx > 0 {
		diskNode = diskNode[:idx]
	}

	var rawDev string
	if runtime.GOOS == "darwin" {
		rawDev = "/dev/r" + diskNode
		if _, err := os.Stat(rawDev); err != nil {
			rawDev = "/dev/" + diskNode
		}
	} else {
		rawDev = "/dev/" + diskNode
	}

	cmd := exec.Command("dd", "if=/dev/zero", "of="+rawDev, "bs=446", "count=1", "conv=notrunc")
	return cmd.Run()
}

// DeployModeB executes Mode B: Cloud Pure Mode (1-sec native format & multi-arch iPXE firmware) with customizable file system.
// For existing Ventoy drives, it non-destructively flashes ONLY Partition 2 (VTOYEFI / ESP), keeping Partition 1 (Data) untouched!
func DeployModeB(ctx context.Context, targetDisk string, fsType string) (*DeployResult, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return nil, fmt.Errorf("disk validation failed: %w", err)
	}

	if fsType == "" {
		fsType = "exFAT"
	}

	isExistingVentoy := disk.IsVentoyDisk(targetDisk)
	var efiMountPoint string
	var err error

	if isExistingVentoy {
		// Non-destructive Mode B conversion: Flash EFI Partition 2 (VTOYEFI) directly (Data partition untouched!)
		// Safely neutralize stale Sector 0 Ventoy MBR code to prevent Legacy BIOS boot crashes!
		_ = CleanMbrBootstrapCode(targetDisk)

		efiMountPoint, err = MountAndResolveEFIPartition(targetDisk)
		if err != nil {
			return nil, fmt.Errorf("failed to mount/resolve EFI partition (Partition 2): %w", err)
		}
	} else {
		// Fresh Mode B deployment on blank drive -> Create UNIBOOT dual partitions and resolve ESP Partition 2
		_, errFormat := FormatDiskModeB(ctx, targetDisk)
		if errFormat != nil {
			return nil, fmt.Errorf("formatting dual partitions for Mode B failed: %w", errFormat)
		}
		efiMountPoint, err = MountAndResolveEFIPartition(targetDisk)
		if err != nil {
			// Fallback: Resolve main volume mount point if EFI partition resolving fails
			efiMountPoint, err = ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("preparing EFI partition for Mode B failed: %w", err)
	}

	// Extract multi-arch iPXE EFI & Legacy BIOS firmware assets DIRECTLY into EFI partition (Partition 2)
	if err := firmware.ExtractFirmwareModeB(efiMountPoint); err != nil {
		return nil, fmt.Errorf("extracting firmware assets to EFI partition failed: %w", err)
	}

	msg := fmt.Sprintf("Successfully deployed Cloud Pure Mode B to ESP EFI Partition (%s)", efiMountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully converted Ventoy drive to Mode B iPXE Cloud Boot by flashing EFI partition at %s (Main Data Partition untouched, ISO data preserved!)", efiMountPoint)
	}

	return &DeployResult{
		Success: true,
		Mode:    fmt.Sprintf("Mode B (Cloud Pure - %s)", fsType),
		Target:  targetDisk,
		Message: msg,
	}, nil
}

// DeployModeBBatch executes Mode B on multiple target USB drives concurrently/sequentially with customizable file system.
func DeployModeBBatch(ctx context.Context, targetDisks []string, fsType string) ([]*DeployResult, error) {
	if len(targetDisks) == 0 {
		return nil, fmt.Errorf("no target disks specified for batch deployment")
	}

	for _, d := range targetDisks {
		if err := disk.ValidateTargetDisk(d); err != nil {
			return nil, fmt.Errorf("disk validation failed for %s: %w", d, err)
		}
	}

	results := make([]*DeployResult, 0, len(targetDisks))
	for _, d := range targetDisks {
		res, err := DeployModeB(ctx, d, fsType)
		if err != nil {
			results = append(results, &DeployResult{
				Success: false,
				Mode:    fmt.Sprintf("Mode B (Cloud Pure - %s)", fsType),
				Target:  d,
				Message: err.Error(),
			})
			continue
		}
		results = append(results, res)
	}
	return results, nil
}

