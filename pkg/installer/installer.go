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

// DeployResult contains the output metadata of a disk deployment run.
type DeployResult struct {
	Success     bool                `json:"success"`
	Mode        string              `json:"mode"`
	Target      string              `json:"target"`
	Message     string              `json:"message"`
	Diagnostics *InstallDiagnostics `json:"diagnostics,omitempty"`
}

func validateLiveTargetDisk(targetDisk string) error {
	return validateLiveTargetDiskSnapshot(targetDisk, nil)
}

func validateLiveTargetDiskSnapshot(targetDisk string, expected *disk.DiskInfo) error {
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(targetDisk, "dummy") || strings.HasPrefix(targetDisk, "test") {
		return nil
	}
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		return fmt.Errorf("final target disk validation failed: %w", err)
	}

	disks, err := disk.GetRemovableDisks()
	if err != nil {
		return fmt.Errorf("final target disk validation failed: refresh disk inventory: %w", err)
	}
	for _, actual := range disks {
		if actual.Device != targetDisk {
			continue
		}
		if expected == nil {
			expected = &actual
		}
		if err := disk.ValidateTargetDiskSnapshot(*expected, actual); err != nil {
			return fmt.Errorf("final target disk validation failed: %w", err)
		}
		return nil
	}
	return fmt.Errorf("final target disk validation failed: target disk is no longer present as a removable disk: %s", targetDisk)
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
	return deployModeAWithExpectedDisk(ctx, targetDisk, fsType, ventoyPath, isoPaths, progressCb, nil)
}

// DeployModeAWithExpectedDisk deploys Mode A after confirming the target still matches the selected disk snapshot.
func DeployModeAWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected disk.DiskInfo) (*DeployResult, error) {
	return deployModeAWithExpectedDisk(ctx, targetDisk, fsType, ventoyPath, isoPaths, progressCb, &expected)
}

func deployModeAWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected *disk.DiskInfo) (*DeployResult, error) {
	if fsType == "" {
		fsType = "exFAT"
	}
	modeLabel := fmt.Sprintf("Hybrid Mode (%s)", fsType)
	tracker := NewDeployTracker(targetDisk, modeLabel, expected)

	// Step 1: Target Disk & Snapshot Validation
	tracker.SetStage("验证设备与底层盘状态", StepValidateDisk, ActionRetry)
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		errFormatted := fmt.Errorf("disk validation failed: %w", err)
		diag := tracker.BuildDiagnostics(errFormatted)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errFormatted.Error(), Diagnostics: diag}, errFormatted
	}
	if err := validateLiveTargetDiskSnapshot(targetDisk, expected); err != nil {
		diag := tracker.BuildDiagnostics(err)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: err.Error(), Diagnostics: diag}, err
	}

	// Step 2: Format Disk / Prepare Mount Point
	tracker.SetStage("初始化与格式化", StepFormatDisk, ActionReformat)
	isExistingVentoy := disk.IsRealVentoyDisk(targetDisk)
	var mountPoint string
	var err error

	if isExistingVentoy {
		tracker.SetFormatted(true)
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
		val := ValidateVentoyCli(ventoyPath)
		if !val.Valid {
			errVentoy := fmt.Errorf("cannot create Hybrid Mode: target disk drive is clean and no valid Ventoy directory detected. Please configure Ventoy directory in Settings first (%s)", val.Message)
			diag := tracker.BuildDiagnostics(errVentoy)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errVentoy.Error(), Diagnostics: diag}, errVentoy
		}
		mountPoint, err = FormatDiskWithVentoyCli(ctx, ventoyPath, targetDisk, fsType)
		if err == nil {
			tracker.SetFormatted(true)
		}
	}
	if err != nil {
		errPrep := fmt.Errorf("preparing disk for Hybrid Mode failed: %w", err)
		diag := tracker.BuildDiagnostics(errPrep)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errPrep.Error(), Diagnostics: diag}, errPrep
	}
	tracker.SetFormatted(true)

	// Step 3: Extract Firmware Assets
	tracker.SetStage("解压固件资源", StepExtractFirmware, ActionReformat)
	if err := firmware.ExtractFirmwareModeA(mountPoint); err != nil {
		errExtract := fmt.Errorf("extracting firmware assets failed: %w", err)
		diag := tracker.BuildDiagnostics(errExtract)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errExtract.Error(), Diagnostics: diag}, errExtract
	}
	tracker.AddWrittenFiles([]string{
		filepath.Join(mountPoint, "EFI", "BOOT"),
		filepath.Join(mountPoint, "ipxe"),
	})

	// Step 4: Write Ventoy Configuration
	tracker.SetStage("写入Ventoy配置", StepWriteVentoyConfig, ActionRetry)
	if err := WriteVentoyConfig(mountPoint); err != nil {
		errCfg := fmt.Errorf("writing Ventoy configuration failed: %w", err)
		diag := tracker.BuildDiagnostics(errCfg)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errCfg.Error(), Diagnostics: diag}, errCfg
	}
	tracker.AddWrittenFiles([]string{
		filepath.Join(mountPoint, "ventoy", "ventoy.json"),
		filepath.Join(mountPoint, "ventoy", "ventoy_grub.cfg"),
		filepath.Join(mountPoint, "ventoy", "themes", "uniboot"),
	})

	// Step 5: Copy Selected ISO / IMG Files
	if len(isoPaths) > 0 {
		tracker.SetStage("复制系统镜像", StepCopyIso, ActionRemount)
		if err := CopyIsoFilesToDisk(mountPoint, isoPaths, progressCb); err != nil {
			errCopy := fmt.Errorf("copying selected ISO/IMG files failed: %w", err)
			diag := tracker.BuildDiagnostics(errCopy)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errCopy.Error(), Diagnostics: diag}, errCopy
		}
		for _, iso := range isoPaths {
			tracker.AddWrittenFile(filepath.Join(mountPoint, "iso", filepath.Base(iso)))
		}
	}

	// Step 6: Update Volume Label
	tracker.SetStage("更新卷标", StepUpdateLabel, ActionRetry)
	mountPoint = UpdateVolumeLabel(targetDisk, mountPoint, "UNIBOOT")

	msg := fmt.Sprintf("Successfully deployed Hybrid Mode (%s/UNIBOOT) to %s (mount: %s)", fsType, targetDisk, mountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully upgraded existing Ventoy drive to UniBoot Hybrid Mode at %s (ISO data preserved)", targetDisk)
	}
	if len(isoPaths) > 0 {
		msg += fmt.Sprintf(" (%d ISO/IMG file(s) copied)", len(isoPaths))
	}

	return &DeployResult{
		Success: true,
		Mode:    modeLabel,
		Target:  targetDisk,
		Message: msg,
	}, nil
}

// DeployHybridModeBatch executes Hybrid Mode on multiple target disk drives with specified file system.
func DeployModeABatch(ctx context.Context, targetDisks []string, fsType string) ([]*DeployResult, error) {
	return DeployModeABatchWithIso(ctx, targetDisks, fsType, nil, nil)
}

// DeployHybridModeBatchWithIso executes Hybrid Mode on multiple target disk drives with optional ISO files and progress reporting.
func DeployModeABatchWithIso(ctx context.Context, targetDisks []string, fsType string, isoPaths []string, progressCb CopyIsoProgressCallback) ([]*DeployResult, error) {
	return DeployModeABatchWithVentoyAndIso(ctx, targetDisks, fsType, "", isoPaths, progressCb)
}

// DeployHybridModeBatchWithVentoyAndIso executes Hybrid Mode on multiple target disk drives with customizable Ventoy CLI path, ISO files, and progress reporting.
func DeployModeABatchWithVentoyAndIso(ctx context.Context, targetDisks []string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback) ([]*DeployResult, error) {
	return deployModeABatchWithExpectedDisks(ctx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, nil)
}

// DeployModeABatchWithExpectedDisks deploys Mode A only after all selected disk snapshots pass final validation.
func DeployModeABatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected []disk.DiskInfo) ([]*DeployResult, error) {
	return deployModeABatchWithExpectedDisks(ctx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, expected)
}

func deployModeABatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected []disk.DiskInfo) ([]*DeployResult, error) {
	if len(targetDisks) == 0 {
		return nil, fmt.Errorf("no target disks specified for batch deployment")
	}
	if len(expected) > 0 && len(expected) != len(targetDisks) {
		return nil, fmt.Errorf("target disk snapshot count does not match target disk count")
	}

	for index, d := range targetDisks {
		if err := disk.ValidateTargetDisk(d); err != nil {
			return nil, fmt.Errorf("disk validation failed for %s: %w", d, err)
		}
		var snapshot *disk.DiskInfo
		if len(expected) > 0 {
			snapshot = &expected[index]
		}
		if err := validateLiveTargetDiskSnapshot(d, snapshot); err != nil {
			return nil, err
		}
	}

	results := make([]*DeployResult, 0, len(targetDisks))
	for index, d := range targetDisks {
		var snapshot disk.DiskInfo
		if len(expected) > 0 {
			snapshot = expected[index]
		}
		snapshotPtr := snapshotPointer(snapshot, len(expected) > 0)
		res, err := deployModeAWithExpectedDisk(ctx, d, fsType, ventoyPath, isoPaths, progressCb, snapshotPtr)
		if err != nil {
			if res != nil && res.Diagnostics != nil {
				results = append(results, res)
			} else {
				tracker := NewDeployTracker(d, fmt.Sprintf("Hybrid Mode (%s)", fsType), snapshotPtr)
				results = append(results, &DeployResult{
					Success:     false,
					Mode:        fmt.Sprintf("Hybrid Mode (%s)", fsType),
					Target:      d,
					Message:     err.Error(),
					Diagnostics: tracker.BuildDiagnostics(err),
				})
			}
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
	return deployModeBWithExpectedDisk(ctx, targetDisk, fsType, nil)
}

// DeployModeBWithExpectedDisk deploys Mode B after confirming the target still matches the selected disk snapshot.
func DeployModeBWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, expected disk.DiskInfo) (*DeployResult, error) {
	return deployModeBWithExpectedDisk(ctx, targetDisk, fsType, &expected)
}

func deployModeBWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, expected *disk.DiskInfo) (*DeployResult, error) {
	if fsType == "" {
		fsType = "exFAT"
	}
	modeLabel := fmt.Sprintf("Cloud Mode (%s)", fsType)
	tracker := NewDeployTracker(targetDisk, modeLabel, expected)

	// Step 1: Target Disk & Snapshot Validation
	tracker.SetStage("验证设备与底层盘状态", StepValidateDisk, ActionRetry)
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		errFormatted := fmt.Errorf("disk validation failed: %w", err)
		diag := tracker.BuildDiagnostics(errFormatted)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errFormatted.Error(), Diagnostics: diag}, errFormatted
	}
	if err := validateLiveTargetDiskSnapshot(targetDisk, expected); err != nil {
		diag := tracker.BuildDiagnostics(err)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: err.Error(), Diagnostics: diag}, err
	}

	// Step 2: Format Disk / Prepare EFI Partition
	tracker.SetStage("初始化与格式化分区", StepFormatDisk, ActionReformat)
	isExistingVentoy := disk.IsVentoyDisk(targetDisk)
	var efiMountPoint string
	var err error

	if isExistingVentoy {
		_ = CleanMbrBootstrapCode(targetDisk)
		tracker.SetFormatted(true)
		efiMountPoint, err = MountAndResolveEFIPartition(targetDisk)
		if err != nil {
			errMount := fmt.Errorf("failed to mount/resolve EFI partition (Partition 2): %w", err)
			diag := tracker.BuildDiagnostics(errMount)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errMount.Error(), Diagnostics: diag}, errMount
		}
	} else {
		_, errFormat := FormatDiskModeB(ctx, targetDisk)
		if errFormat != nil {
			errFmt := fmt.Errorf("formatting dual partitions for Cloud Mode failed: %w", errFormat)
			diag := tracker.BuildDiagnostics(errFmt)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errFmt.Error(), Diagnostics: diag}, errFmt
		}
		tracker.SetFormatted(true)
		efiMountPoint, err = MountAndResolveEFIPartition(targetDisk)
		if err != nil {
			efiMountPoint, err = ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
		}
	}
	if err != nil {
		errPrep := fmt.Errorf("preparing EFI partition for Cloud Mode failed: %w", err)
		diag := tracker.BuildDiagnostics(errPrep)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errPrep.Error(), Diagnostics: diag}, errPrep
	}
	tracker.SetFormatted(true)

	// Step 3: Extract Firmware Assets to ESP Partition
	tracker.SetStage("解压ESP固件资源", StepExtractFirmware, ActionReformat)
	if err := firmware.ExtractFirmwareModeB(efiMountPoint); err != nil {
		errExtract := fmt.Errorf("extracting firmware assets to EFI partition failed: %w", err)
		diag := tracker.BuildDiagnostics(errExtract)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errExtract.Error(), Diagnostics: diag}, errExtract
	}
	tracker.AddWrittenFiles([]string{
		filepath.Join(efiMountPoint, "EFI", "BOOT"),
		filepath.Join(efiMountPoint, "ipxe"),
	})

	msg := fmt.Sprintf("Successfully deployed Cloud Mode to ESP EFI Partition (%s)", efiMountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully converted Ventoy drive to Cloud Mode iPXE Cloud Boot by flashing EFI partition at %s (Main Data Partition untouched, ISO data preserved!)", efiMountPoint)
	}

	return &DeployResult{
		Success: true,
		Mode:    modeLabel,
		Target:  targetDisk,
		Message: msg,
	}, nil
}

// DeployCloudModeBatch executes Cloud Mode on multiple target disk drives concurrently/sequentially with customizable file system.
func DeployModeBBatch(ctx context.Context, targetDisks []string, fsType string) ([]*DeployResult, error) {
	return deployModeBBatchWithExpectedDisks(ctx, targetDisks, fsType, nil)
}

// DeployModeBBatchWithExpectedDisks deploys Mode B only after all selected disk snapshots pass final validation.
func DeployModeBBatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, expected []disk.DiskInfo) ([]*DeployResult, error) {
	return deployModeBBatchWithExpectedDisks(ctx, targetDisks, fsType, expected)
}

func deployModeBBatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, expected []disk.DiskInfo) ([]*DeployResult, error) {
	if len(targetDisks) == 0 {
		return nil, fmt.Errorf("no target disks specified for batch deployment")
	}
	if len(expected) > 0 && len(expected) != len(targetDisks) {
		return nil, fmt.Errorf("target disk snapshot count does not match target disk count")
	}

	for index, d := range targetDisks {
		if err := disk.ValidateTargetDisk(d); err != nil {
			return nil, fmt.Errorf("disk validation failed for %s: %w", d, err)
		}
		var snapshot *disk.DiskInfo
		if len(expected) > 0 {
			snapshot = &expected[index]
		}
		if err := validateLiveTargetDiskSnapshot(d, snapshot); err != nil {
			return nil, err
		}
	}

	results := make([]*DeployResult, 0, len(targetDisks))
	for index, d := range targetDisks {
		var snapshot *disk.DiskInfo
		if len(expected) > 0 {
			snapshot = &expected[index]
		}
		res, err := deployModeBWithExpectedDisk(ctx, d, fsType, snapshot)
		if err != nil {
			if res != nil && res.Diagnostics != nil {
				results = append(results, res)
			} else {
				tracker := NewDeployTracker(d, fmt.Sprintf("Cloud Mode (%s)", fsType), snapshot)
				results = append(results, &DeployResult{
					Success:     false,
					Mode:        fmt.Sprintf("Cloud Mode (%s)", fsType),
					Target:      d,
					Message:     err.Error(),
					Diagnostics: tracker.BuildDiagnostics(err),
				})
			}
			continue
		}
		results = append(results, res)
	}
	return results, nil
}

func snapshotPointer(snapshot disk.DiskInfo, enabled bool) *disk.DiskInfo {
	if !enabled {
		return nil
	}
	return &snapshot
}
