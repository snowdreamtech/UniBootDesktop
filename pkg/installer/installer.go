// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"fmt"

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
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return nil, fmt.Errorf("disk validation failed: %w", err)
	}

	if fsType == "" {
		fsType = "exFAT"
	}

	// 1. Differential Treatment: Check if target drive is ALREADY an active Ventoy drive
	isExistingVentoy := disk.IsVentoyDisk(targetDisk)
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
		mountPoint, err = FormatDiskModeA(ctx, targetDisk, fsType)
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

	// 4. Non-destructively update volume label of data partition to UNIBOOT
	mountPoint = UpdateVolumeLabel(targetDisk, mountPoint, "UNIBOOT")

	msg := fmt.Sprintf("Successfully deployed Hybrid Pro Mode A (%s/UNIBOOT) to %s (mount: %s)", fsType, targetDisk, mountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully upgraded existing Ventoy drive to UniBoot Hybrid Pro Mode A at %s (ISO data preserved)", targetDisk)
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
		res, err := DeployModeA(ctx, d, fsType)
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
	var mountPoint string
	var err error

	if isExistingVentoy {
		// Non-destructive Mode B conversion: Flash EFI Partition 2 (VTOYEFI) directly (Data partition untouched!)
		efiMountPoint, errEfi := MountAndResolveEFIPartition(targetDisk)
		if errEfi == nil {
			mountPoint = efiMountPoint
		} else {
			// Fallback: Resolve main volume mount point
			mountPoint, err = ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
			if err != nil {
				mountPoint, err = ResolveMountPointWithLabel(targetDisk, "Ventoy")
			}
			if err != nil {
				mountPoint, err = ResolveMountPointWithLabel(targetDisk, "VENTOY")
			}
			if err != nil {
				mountPoint, err = FormatDiskModeB(ctx, targetDisk)
			}
		}
	} else {
		// Fresh Mode B deployment on blank drive
		mountPoint, err = FormatDiskModeB(ctx, targetDisk)
	}
	if err != nil {
		return nil, fmt.Errorf("preparing disk for Mode B failed: %w", err)
	}

	// Extract multi-arch iPXE EFI & Legacy BIOS firmware assets to target EFI / Data volume
	if err := firmware.ExtractFirmwareModeB(mountPoint); err != nil {
		return nil, fmt.Errorf("extracting firmware assets failed: %w", err)
	}

	msg := fmt.Sprintf("Successfully deployed Cloud Pure Mode B (%s/UNIBOOT) to %s (mount: %s)", fsType, targetDisk, mountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully converted Ventoy drive to Mode B iPXE Cloud Boot by flashing EFI partition at %s (Main Data Partition untouched, ISO data preserved!)", targetDisk)
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

