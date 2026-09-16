// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"errors"
	"testing"

	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallDiagnosticsFormatReport(t *testing.T) {
	diag := &InstallDiagnostics{
		TaskID:            "deploy-20260917-010500-1234",
		Target:            "/dev/disk2",
		DeviceSummary:     "SanDisk Ultra 16GB",
		Mode:              "Mode A (Hybrid Pro - exFAT)",
		FailedStage:       "解压固件资源",
		FailedStepCode:    StepExtractFirmware,
		ErrorCause:        "writing /Volumes/UNIBOOT/EFI/BOOT/BOOTX64.EFI: i/o error",
		IsFormatted:       true,
		WrittenFiles:      []string{"/Volumes/UNIBOOT/EFI/BOOT/grubx64.efi", "/Volumes/UNIBOOT/ipxe/ipxe.efi"},
		SafeToUnplug:      true,
		RecommendedAction: ActionReformat,
	}

	report := diag.FormatReport()
	assert.Contains(t, report, "=== UniBoot 写盘失败诊断报告 ===")
	assert.Contains(t, report, "deploy-20260917-010500-1234")
	assert.Contains(t, report, "/dev/disk2 (SanDisk Ultra 16GB)")
	assert.Contains(t, report, "[EXTRACT_FIRMWARE] 解压固件资源")
	assert.Contains(t, report, "writing /Volumes/UNIBOOT/EFI/BOOT/BOOTX64.EFI: i/o error")
	assert.Contains(t, report, "是 (已格式化)")
	assert.Contains(t, report, "已写入文件列表 (2 个文件)")
	assert.Contains(t, report, "✅ 可以安全拔盘")
	assert.Contains(t, report, "🧹 建议重新格式化")
}

func TestDeployTrackerPhases(t *testing.T) {
	expected := &disk.DiskInfo{
		Device:    "/dev/disk3",
		Name:      "Kingston DataTraveler",
		Formatted: "32 GB",
	}

	tracker := NewDeployTracker("/dev/disk3", "Mode A (Hybrid Pro - exFAT)", expected)
	assert.Equal(t, "/dev/disk3", tracker.Target)
	assert.Contains(t, tracker.DeviceSummary, "Kingston DataTraveler")
	assert.True(t, tracker.SafeToUnplug)
	assert.False(t, tracker.IsFormatted)

	// Step 1: Format phase
	tracker.SetStage("初始化与格式化", StepFormatDisk, ActionReformat)
	tracker.SetFormatted(true)
	assert.True(t, tracker.IsFormatted)

	// Step 2: Write files
	tracker.AddWrittenFile("/Volumes/UNIBOOT/ventoy/ventoy.json")
	tracker.AddWrittenFiles([]string{"/Volumes/UNIBOOT/ventoy/ventoy_grub.cfg", "/Volumes/UNIBOOT/iso/ubuntu.iso"})
	assert.Len(t, tracker.WrittenFiles, 3)

	// Step 3: Trigger error snapshot
	testErr := errors.New("copying ISO failed: disk full")
	diag := tracker.BuildDiagnostics(testErr)
	require.NotNil(t, diag)

	assert.Equal(t, tracker.TaskID, diag.TaskID)
	assert.Equal(t, "初始化与格式化", diag.FailedStage)
	assert.Equal(t, StepFormatDisk, diag.FailedStepCode)
	assert.Equal(t, "copying ISO failed: disk full", diag.ErrorCause)
	assert.True(t, diag.IsFormatted)
	assert.Len(t, diag.WrittenFiles, 3)
	assert.Equal(t, ActionReformat, diag.RecommendedAction)
	assert.NotEmpty(t, diag.ReportSummary)
}
