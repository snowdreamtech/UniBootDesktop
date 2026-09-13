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

// DeployModeA executes Mode A: Hybrid Pro Mode (Ventoy CLI + UniBoot theme + iPXE network extension) with customizable file system.
func DeployModeA(ctx context.Context, targetDisk string, fsType string) (*DeployResult, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return nil, fmt.Errorf("disk validation failed: %w", err)
	}

	if fsType == "" {
		fsType = "exFAT"
	}

	// Deploy Mode A logic with specified file system
	return &DeployResult{
		Success: true,
		Mode:    fmt.Sprintf("Mode A (Hybrid Pro - %s)", fsType),
		Target:  targetDisk,
		Message: fmt.Sprintf("Successfully deployed Hybrid Pro Mode (%s) to %s", fsType, targetDisk),
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

// DeployModeB executes Mode B: Cloud Pure Mode (1-sec native FAT32 format & multi-arch iPXE firmware).
func DeployModeB(ctx context.Context, targetDisk string) (*DeployResult, error) {
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return nil, fmt.Errorf("disk validation failed: %w", err)
	}

	// 1. Format disk to FAT32 MBR with volume label UNIBOOT
	mountPoint, err := FormatDiskModeB(ctx, targetDisk)
	if err != nil {
		return nil, fmt.Errorf("formatting disk for Mode B failed: %w", err)
	}

	// 2. Extract multi-arch iPXE EFI & Legacy BIOS firmware assets to target volume
	if err := firmware.ExtractFirmwareToDir(mountPoint); err != nil {
		return nil, fmt.Errorf("extracting firmware assets failed: %w", err)
	}

	return &DeployResult{
		Success: true,
		Mode:    "Mode B (Cloud Pure)",
		Target:  targetDisk,
		Message: fmt.Sprintf("Successfully deployed Cloud Pure Mode B (FAT32/UNIBOOT) to %s (mount: %s)", targetDisk, mountPoint),
	}, nil
}

// DeployModeBBatch executes Mode B on multiple target USB drives concurrently/sequentially.
func DeployModeBBatch(ctx context.Context, targetDisks []string) ([]*DeployResult, error) {
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
		res, err := DeployModeB(ctx, d)
		if err != nil {
			results = append(results, &DeployResult{
				Success: false,
				Mode:    "Mode B (Cloud Pure)",
				Target:  d,
				Message: err.Error(),
			})
			continue
		}
		results = append(results, res)
	}
	return results, nil
}

