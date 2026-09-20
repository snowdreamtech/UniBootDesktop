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
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/firmware"
)

// ProgressCallback is a function type for reporting deployment progress (0-100)
type ProgressCallback func(progress int)

// ProgressWithStageCallback is a function type for reporting deployment progress with stage information
type ProgressWithStageCallback func(progress int, stage string)

// BatchDeployProgress represents real-time progress for batch deployment operations
type BatchDeployProgress struct {
	TotalDisks       int     `json:"totalDisks"`       // Total number of disks to deploy
	CurrentDiskIndex int     `json:"currentDiskIndex"` // Current disk index (1-based)
	CurrentDisk      string  `json:"currentDisk"`      // Current disk device name
	CurrentStage     string  `json:"currentStage"`     // Current deployment stage description
	DiskProgress     int     `json:"diskProgress"`     // Current disk progress 0-100
	OverallProgress  int     `json:"overallProgress"`  // Overall progress 0-100
	SpeedMBps        float64 `json:"speedMBps"`        // Current I/O speed in MB/s
	ElapsedSec       int     `json:"elapsedSec"`       // Total elapsed time in seconds
	EtaSec           int     `json:"etaSec"`           // Estimated time remaining in seconds
}

// BatchProgressCallback reports batch deployment progress with disk-level details
type BatchProgressCallback func(BatchDeployProgress)

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

// DeployHybridMode executes Hybrid Mode: Hybrid Pro Mode (Ventoy + UniBoot theme + iPXE network extension) with customizable file system.
// Performs non-destructive in-place upgrade on existing Ventoy drives, or fresh partition initialization on blank drives.
func DeployHybridMode(ctx context.Context, targetDisk string, fsType string) (*DeployResult, error) {
	return DeployHybridModeWithIsoAndVentoyPath(ctx, targetDisk, fsType, "", nil, nil)
}

// DeployHybridModeWithVentoyPath executes Hybrid Mode with an optional user-configured Ventoy CLI executable path.
func DeployHybridModeWithVentoyPath(ctx context.Context, targetDisk string, fsType string, ventoyPath string) (*DeployResult, error) {
	return DeployHybridModeWithIsoAndVentoyPath(ctx, targetDisk, fsType, ventoyPath, nil, nil)
}

// DeployHybridModeWithIsoAndVentoyPath executes Hybrid Mode with customizable Ventoy CLI path, ISO file paths, and progress callback.
func DeployHybridModeWithIsoAndVentoyPath(ctx context.Context, targetDisk string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback) (*DeployResult, error) {
	return deployHybridModeWithExpectedDisk(ctx, targetDisk, fsType, ventoyPath, isoPaths, progressCb, nil)
}

// DeployHybridModeWithExpectedDisk deploys Hybrid Mode after confirming the target still matches the selected disk snapshot.
func DeployHybridModeWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected disk.DiskInfo) (*DeployResult, error) {
	return deployHybridModeWithExpectedDisk(ctx, targetDisk, fsType, ventoyPath, isoPaths, progressCb, &expected)
}

func deployHybridModeWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected *disk.DiskInfo) (*DeployResult, error) {
	if fsType == "" {
		fsType = "exFAT"
	}
	modeLabel := fmt.Sprintf("Hybrid Mode (%s)", fsType)
	tracker := NewDeployTracker(targetDisk, modeLabel, expected)

	logger.Info("Starting Hybrid Mode deployment...", "target", targetDisk, "fsType", fsType)

	// Step 1: Target Disk & Snapshot Validation
	tracker.SetStage("验证设备与底层盘状态", StepValidateDisk, ActionRetry)
	logger.Info("[Step 1/6] Validating target disk status and Ventoy CLI dependency...", "target", targetDisk)
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		errFormatted := fmt.Errorf("disk validation failed: %w", err)
		diag := tracker.BuildDiagnostics(errFormatted)
		logger.Error("Target disk validation failed", "target", targetDisk, "error", errFormatted)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errFormatted.Error(), Diagnostics: diag}, errFormatted
	}
	if err := validateLiveTargetDiskSnapshot(targetDisk, expected); err != nil {
		diag := tracker.BuildDiagnostics(err)
		logger.Error("Target disk snapshot mismatch", "target", targetDisk, "error", err)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: err.Error(), Diagnostics: diag}, err
	}

	// Step 2: Format Disk / Prepare Mount Point
	tracker.SetStage("初始化与格式化", StepFormatDisk, ActionReformat)
	isExistingVentoy := disk.IsRealVentoyDisk(targetDisk)
	var mountPoint string
	var err error

	if isExistingVentoy {
		logger.Info("[Step 2/6] Existing Ventoy partition detected, performing in-place upgrade (data preserved)...", "target", targetDisk)
		tracker.SetFormatted(true)
		mountPoint, err = ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
		if err != nil {
			mountPoint, err = ResolveMountPointWithLabel(targetDisk, "Ventoy")
		}
		if err != nil {
			mountPoint, err = ResolveMountPointWithLabel(targetDisk, "VENTOY")
		}
		if err != nil {
			mountPoint, err = FormatDiskHybridMode(ctx, targetDisk, fsType)
		}
	} else {
		logger.Info(fmt.Sprintf("[Step 2/6] Running Ventoy CLI engine to format %s disk...", fsType), "target", targetDisk)
		val := ValidateVentoyCli(ventoyPath)
		if !val.Valid {
			errVentoy := fmt.Errorf("cannot create Hybrid Mode: target disk drive is clean and no valid Ventoy directory detected. Please configure Ventoy directory in Settings first (%s)", val.Message)
			diag := tracker.BuildDiagnostics(errVentoy)
			logger.Error("Valid Ventoy CLI environment not found", "target", targetDisk, "error", errVentoy)
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
		logger.Error("Preparing Ventoy partition failed", "target", targetDisk, "error", errPrep)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errPrep.Error(), Diagnostics: diag}, errPrep
	}
	tracker.SetFormatted(true)

	// Step 3: Extract Firmware Assets
	tracker.SetStage("解压固件资源", StepExtractFirmware, ActionReformat)
	logger.Info("[Step 3/6] Writing iPXE cloud boot firmware extensions...", "mountPoint", mountPoint)
	if err := firmware.ExtractFirmwareHybridMode(mountPoint); err != nil {
		errExtract := fmt.Errorf("extracting firmware assets failed: %w", err)
		diag := tracker.BuildDiagnostics(errExtract)
		logger.Error("Writing firmware assets failed", "target", targetDisk, "error", errExtract)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errExtract.Error(), Diagnostics: diag}, errExtract
	}
	tracker.AddWrittenFiles([]string{
		filepath.Join(mountPoint, "EFI", "BOOT"),
		filepath.Join(mountPoint, "ipxe"),
	})

	// Step 4: Write Ventoy Configuration
	tracker.SetStage("写入Ventoy配置", StepWriteVentoyConfig, ActionRetry)
	logger.Info("[Step 4/6] Writing Ventoy Grub config and UniBoot visual theme pack...", "mountPoint", mountPoint)
	if err := WriteVentoyConfig(mountPoint); err != nil {
		errCfg := fmt.Errorf("writing Ventoy configuration failed: %w", err)
		diag := tracker.BuildDiagnostics(errCfg)
		logger.Error("Writing Ventoy theme config failed", "target", targetDisk, "error", errCfg)
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
		logger.Info(fmt.Sprintf("[Step 5/6] Copying %d ISO image file(s) to disk...", len(isoPaths)), "mountPoint", mountPoint)
		if err := CopyIsoFilesToDisk(mountPoint, isoPaths, progressCb); err != nil {
			errCopy := fmt.Errorf("copying selected ISO/IMG files failed: %w", err)
			diag := tracker.BuildDiagnostics(errCopy)
			logger.Error("Copying ISO files failed", "target", targetDisk, "error", errCopy)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errCopy.Error(), Diagnostics: diag}, errCopy
		}
		for _, iso := range isoPaths {
			tracker.AddWrittenFile(filepath.Join(mountPoint, "iso", filepath.Base(iso)))
		}
	}

	// Step 6: Update Volume Label
	tracker.SetStage("更新卷标", StepUpdateLabel, ActionRetry)
	logger.Info("[Step 6/6] Updating volume label to UNIBOOT...", "target", targetDisk)
	mountPoint = UpdateVolumeLabel(targetDisk, mountPoint, "UNIBOOT")

	msg := fmt.Sprintf("Successfully deployed Hybrid Mode (%s/UNIBOOT) to %s (mount: %s)", fsType, targetDisk, mountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully upgraded existing Ventoy drive to UniBoot Hybrid Mode at %s (ISO data preserved)", targetDisk)
	}
	if len(isoPaths) > 0 {
		msg += fmt.Sprintf(" (%d ISO/IMG file(s) copied)", len(isoPaths))
	}

	logger.Info("Hybrid Mode Boot Disk created successfully!", "target", targetDisk, "mountPoint", mountPoint)

	return &DeployResult{
		Success: true,
		Mode:    modeLabel,
		Target:  targetDisk,
		Message: msg,
	}, nil
}

// DeployHybridModeBatch executes Hybrid Mode on multiple target disk drives with specified file system.
func DeployHybridModeBatch(ctx context.Context, targetDisks []string, fsType string) ([]*DeployResult, error) {
	return DeployHybridModeBatchWithIso(ctx, targetDisks, fsType, nil, nil)
}

// DeployHybridModeBatchWithIso executes Hybrid Mode on multiple target disk drives with optional ISO files and progress reporting.
func DeployHybridModeBatchWithIso(ctx context.Context, targetDisks []string, fsType string, isoPaths []string, progressCb CopyIsoProgressCallback) ([]*DeployResult, error) {
	return DeployHybridModeBatchWithVentoyAndIso(ctx, targetDisks, fsType, "", isoPaths, progressCb)
}

// DeployHybridModeBatchWithVentoyAndIso executes Hybrid Mode on multiple target disk drives with customizable Ventoy CLI path, ISO files, and progress reporting.
func DeployHybridModeBatchWithVentoyAndIso(ctx context.Context, targetDisks []string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback) ([]*DeployResult, error) {
	return deployHybridModeBatchWithExpectedDisks(ctx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, nil)
}

// DeployHybridModeBatchWithExpectedDisks deploys Hybrid Mode only after all selected disk snapshots pass final validation.
func DeployHybridModeBatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected []disk.DiskInfo) ([]*DeployResult, error) {
	return deployHybridModeBatchWithExpectedDisks(ctx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, expected)
}

func deployHybridModeBatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, ventoyPath string, isoPaths []string, progressCb CopyIsoProgressCallback, expected []disk.DiskInfo) ([]*DeployResult, error) {
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
		res, err := deployHybridModeWithExpectedDisk(ctx, d, fsType, ventoyPath, isoPaths, progressCb, snapshotPtr)
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
// This neutralizes stale Ventoy MBR hooks when converting Hybrid Mode to Cloud Mode, preventing Legacy BIOS boot crashes.
func CleanMbrBootstrapCode(targetDisk string) error {
	diskNode := disk.NormalizeDarwinDiskNode(targetDisk)

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

// DeployCloudMode executes Cloud Mode: Cloud Pure Mode (1-sec native format & multi-arch iPXE firmware) with customizable file system.
// For existing Ventoy drives, it non-destructively flashes ONLY Partition 2 (VTOYEFI / ESP), keeping Partition 1 (Data) untouched!
func DeployCloudMode(ctx context.Context, targetDisk string, fsType string) (*DeployResult, error) {
	return deployCloudModeWithExpectedDisk(ctx, targetDisk, fsType, nil, nil)
}

// DeployCloudModeWithExpectedDisk deploys Cloud Mode after confirming the target still matches the selected disk snapshot.
func DeployCloudModeWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, expected disk.DiskInfo, progressCallback ProgressCallback) (*DeployResult, error) {
	return deployCloudModeWithExpectedDisk(ctx, targetDisk, fsType, &expected, progressCallback)
}

func deployCloudModeWithExpectedDisk(ctx context.Context, targetDisk string, fsType string, expected *disk.DiskInfo, progressCallback ProgressCallback) (*DeployResult, error) {
	// Convert simple progress callback to stage-aware callback
	var stageCallback ProgressWithStageCallback
	if progressCallback != nil {
		stageCallback = func(progress int, stage string) {
			progressCallback(progress)
		}
	}
	return deployCloudModeWithStage(ctx, targetDisk, fsType, expected, stageCallback)
}

func deployCloudModeWithStage(ctx context.Context, targetDisk string, fsType string, expected *disk.DiskInfo, progressCallback ProgressWithStageCallback) (*DeployResult, error) {
	if fsType == "" {
		fsType = "exFAT"
	}
	modeLabel := fmt.Sprintf("Cloud Mode (%s)", fsType)
	tracker := NewDeployTracker(targetDisk, modeLabel, expected)

	logger.Info("Starting Cloud Mode deployment...", "target", targetDisk, "fsType", fsType)

	// Report initial progress
	if progressCallback != nil {
		progressCallback(10, "Validating disk")
	}

	// Step 1: Target Disk & Snapshot Validation
	tracker.SetStage("验证设备与底层盘状态", StepValidateDisk, ActionRetry)
	logger.Info("[Step 1/3] Validating target disk drive and read-only protection status...", "target", targetDisk)
	if err := disk.ValidateTargetDisk(targetDisk); err != nil {
		errFormatted := fmt.Errorf("disk validation failed: %w", err)
		diag := tracker.BuildDiagnostics(errFormatted)
		logger.Error("Target disk validation failed", "target", targetDisk, "error", errFormatted)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errFormatted.Error(), Diagnostics: diag}, errFormatted
	}
	if err := validateLiveTargetDiskSnapshot(targetDisk, expected); err != nil {
		diag := tracker.BuildDiagnostics(err)
		logger.Error("Target disk snapshot mismatch", "target", targetDisk, "error", err)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: err.Error(), Diagnostics: diag}, err
	}

	// Step 1 completed
	if progressCallback != nil {
		progressCallback(30, "Formatting partitions")
	}

	// Step 2: Format Disk / Prepare EFI Partition
	tracker.SetStage("初始化与格式化分区", StepFormatDisk, ActionReformat)
	isExistingVentoy := disk.IsVentoyDisk(targetDisk)
	var efiMountPoint string
	var err error

	if isExistingVentoy {
		logger.Info("[Step 2/3] Existing Ventoy/UniBoot partition detected, upgrading ESP partition...", "target", targetDisk)
		_ = CleanMbrBootstrapCode(targetDisk)
		tracker.SetFormatted(true)
		efiMountPoint, err = MountAndResolveEFIPartition(targetDisk)
		if err != nil {
			errMount := fmt.Errorf("failed to mount/resolve EFI partition (Partition 2): %w", err)
			diag := tracker.BuildDiagnostics(errMount)
			logger.Error("Mounting ESP boot partition failed", "target", targetDisk, "error", errMount)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errMount.Error(), Diagnostics: diag}, errMount
		}
	} else {
		logger.Info(fmt.Sprintf("[Step 2/3] Formatting dual partitions (Main Data Partition %s + 64MB FAT32 ESP)...", fsType), "target", targetDisk)
		_, errFormat := FormatDiskCloudMode(ctx, targetDisk)
		if errFormat != nil {
			errFmt := fmt.Errorf("formatting dual partitions for Cloud Mode failed: %w", errFormat)
			diag := tracker.BuildDiagnostics(errFmt)
			logger.Error("Formatting dual partitions failed", "target", targetDisk, "error", errFmt)
			return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errFmt.Error(), Diagnostics: diag}, errFmt
		}
		tracker.SetFormatted(true)
		logger.Info("Mounting and resolving newly created ESP boot partition...", "target", targetDisk)
		efiMountPoint, err = MountAndResolveEFIPartition(targetDisk)
		if err != nil {
			efiMountPoint, err = ResolveMountPointWithLabel(targetDisk, "UNIBOOT")
		}
	}
	if err != nil {
		errPrep := fmt.Errorf("preparing EFI partition for Cloud Mode failed: %w", err)
		diag := tracker.BuildDiagnostics(errPrep)
		logger.Error("Resolving ESP boot partition failed", "target", targetDisk, "error", errPrep)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errPrep.Error(), Diagnostics: diag}, errPrep
	}
	tracker.SetFormatted(true)

	// Step 2 completed
	if progressCallback != nil {
		progressCallback(70, "Writing firmware")
	}

	// Step 3: Extract Firmware Assets to ESP Partition
	tracker.SetStage("解压ESP固件资源", StepExtractFirmware, ActionReformat)
	logger.Info("[Step 3/3] Extracting iPXE multi-arch cloud boot firmware to ESP partition...", "efiMountPoint", efiMountPoint)
	if err := firmware.ExtractFirmwareCloudMode(efiMountPoint); err != nil {
		errExtract := fmt.Errorf("extracting firmware assets to EFI partition failed: %w", err)
		diag := tracker.BuildDiagnostics(errExtract)
		logger.Error("Extracting iPXE cloud firmware assets failed", "target", targetDisk, "error", errExtract)
		return &DeployResult{Success: false, Mode: modeLabel, Target: targetDisk, Message: errExtract.Error(), Diagnostics: diag}, errExtract
	}
	tracker.AddWrittenFiles([]string{
		filepath.Join(efiMountPoint, "EFI", "BOOT"),
		filepath.Join(efiMountPoint, "ipxe"),
	})

	// Step 3 completed
	if progressCallback != nil {
		progressCallback(100, "Completed")
	}

	msg := fmt.Sprintf("Successfully deployed Cloud Mode to ESP EFI Partition (%s)", efiMountPoint)
	if isExistingVentoy {
		msg = fmt.Sprintf("Successfully converted Ventoy drive to Cloud Mode iPXE Cloud Boot by flashing EFI partition at %s (Main Data Partition untouched, ISO data preserved!)", efiMountPoint)
	}

	logger.Info("Cloud Boot Disk created successfully!", "target", targetDisk, "efiMountPoint", efiMountPoint)

	return &DeployResult{
		Success: true,
		Mode:    modeLabel,
		Target:  targetDisk,
		Message: msg,
	}, nil
}

// DeployCloudModeBatch executes Cloud Mode on multiple target disk drives concurrently/sequentially with customizable file system.
func DeployCloudModeBatch(ctx context.Context, targetDisks []string, fsType string) ([]*DeployResult, error) {
	return deployCloudModeBatchWithExpectedDisks(ctx, targetDisks, fsType, nil, nil)
}

// DeployCloudModeBatchWithExpectedDisks deploys Cloud Mode only after all selected disk snapshots pass final validation.
func DeployCloudModeBatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, expected []disk.DiskInfo, progressCallback BatchProgressCallback) ([]*DeployResult, error) {
	return deployCloudModeBatchWithExpectedDisks(ctx, targetDisks, fsType, expected, progressCallback)
}

func deployCloudModeBatchWithExpectedDisks(ctx context.Context, targetDisks []string, fsType string, expected []disk.DiskInfo, progressCallback BatchProgressCallback) ([]*DeployResult, error) {
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
	totalDisks := len(targetDisks)
	batchStartTime := time.Now()

	for index, d := range targetDisks {
		var snapshot *disk.DiskInfo
		if len(expected) > 0 {
			snapshot = &expected[index]
		}

		diskStartTime := time.Now()

		// Get disk size for speed calculation
		var diskSize int64
		if snapshot != nil {
			diskSize = int64(snapshot.Size)
		}

		// Calculate progress for batch operations: each disk contributes equally
		diskProgressCallback := func(diskProgress int, stage string) {
			if progressCallback != nil {
				// Calculate elapsed time
				elapsed := int(time.Since(batchStartTime).Seconds())

				// Calculate overall progress
				overallProgress := (index*100 + diskProgress) / totalDisks

				// Calculate speed (MB/s) based on disk progress and size
				var speedMBps float64
				if diskSize > 0 && diskProgress > 0 {
					diskElapsed := time.Since(diskStartTime).Seconds()
					if diskElapsed > 0 {
						bytesProcessed := float64(diskSize) * float64(diskProgress) / 100.0
						speedMBps = bytesProcessed / diskElapsed / (1024 * 1024)
					}
				}

				// Calculate ETA (estimated time remaining)
				var etaSec int
				if overallProgress > 0 && elapsed > 0 {
					totalEstimated := (elapsed * 100) / overallProgress
					etaSec = totalEstimated - elapsed
					if etaSec < 0 {
						etaSec = 0
					}
				}

				progressCallback(BatchDeployProgress{
					TotalDisks:       totalDisks,
					CurrentDiskIndex: index + 1, // 1-based index for display
					CurrentDisk:      d,
					CurrentStage:     stage,
					DiskProgress:     diskProgress,
					OverallProgress:  overallProgress,
					SpeedMBps:        speedMBps,
					ElapsedSec:       elapsed,
					EtaSec:           etaSec,
				})
			}
		}

		res, err := deployCloudModeWithStage(ctx, d, fsType, snapshot, diskProgressCallback)
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
