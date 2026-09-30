// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unibootdesktop/pkg/disk"
)

// RecommendedAction represents the suggested next step after deployment failure.
type RecommendedAction string

const (
	ActionRetry    RecommendedAction = "retry"    // Retry deployment (temporary error, device intact)
	ActionRemount  RecommendedAction = "remount"  // Remount volume / reconnect disk device
	ActionReformat RecommendedAction = "reformat" // Reformat partition table / file system
)

// StepCode represents discrete step identifiers during deployment.
type StepCode string

const (
	StepValidateDisk      StepCode = "VALIDATE_DISK"
	StepFormatDisk        StepCode = "FORMAT_DISK"
	StepExtractFirmware   StepCode = "EXTRACT_FIRMWARE"
	StepWriteVentoyConfig StepCode = "WRITE_VENTOY_CONFIG"
	StepCopyIso           StepCode = "COPY_ISO"
	StepUpdateLabel       StepCode = "UPDATE_LABEL"
)

// InstallDiagnostics contains structured diagnostic data for disk installation failure recovery.
type InstallDiagnostics struct {
	TaskID            string            `json:"taskId"`            // Unique deployment task ID
	Target            string            `json:"target"`            // Target disk device path (e.g. /dev/disk2)
	DeviceSummary     string            `json:"deviceSummary"`     // Target device metadata summary
	Mode              string            `json:"mode"`              // Target deployment mode (Hybrid Mode / Cloud Mode)
	FailedStage       string            `json:"failedStage"`       // Human-readable stage title
	FailedStepCode    StepCode          `json:"failedStepCode"`    // Step identifier code
	ErrorCause        string            `json:"errorCause"`        // Detailed error cause message
	IsFormatted       bool              `json:"isFormatted"`       // Indicates if target disk was formatted
	WrittenFiles      []string          `json:"writtenFiles"`      // List of files written prior to failure
	SafeToUnplug      bool              `json:"safeToUnplug"`      // Indicates if it is safe to unplug disk drive
	RecommendedAction RecommendedAction `json:"recommendedAction"` // Recommended next action ("retry", "remount", "reformat")
	ReportSummary     string            `json:"reportSummary"`     // One-time formatted diagnostic text report
}

// FormatReport generates a human-readable, multi-line Chinese diagnostic report.
func (d *InstallDiagnostics) FormatReport() string {
	if d == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("=== UniBoot 写盘失败诊断报告 ===\n")
	sb.WriteString(fmt.Sprintf("任务 ID: %s\n", d.TaskID))
	sb.WriteString(fmt.Sprintf("目标设备: %s (%s)\n", d.Target, d.DeviceSummary))
	sb.WriteString(fmt.Sprintf("部署模式: %s\n", d.Mode))
	sb.WriteString(fmt.Sprintf("失败阶段: [%s] %s\n", d.FailedStepCode, d.FailedStage))
	sb.WriteString(fmt.Sprintf("错误原因: %s\n", d.ErrorCause))
	sb.WriteString("---------------------------------\n")

	formattedStr := "否 (未完成格式化)"
	if d.IsFormatted {
		formattedStr = "是 (已格式化)"
	}
	sb.WriteString(fmt.Sprintf("目标盘格式化状态: %s\n", formattedStr))

	if len(d.WrittenFiles) == 0 {
		sb.WriteString("已写入文件列表: (无文件写入)\n")
	} else {
		sb.WriteString(fmt.Sprintf("已写入文件列表 (%d 个文件):\n", len(d.WrittenFiles)))
		limit := 5
		for i, f := range d.WrittenFiles {
			if i >= limit {
				sb.WriteString(fmt.Sprintf("  ... 及另外 %d 个文件\n", len(d.WrittenFiles)-limit))
				break
			}
			sb.WriteString(fmt.Sprintf("  - %s\n", f))
		}
	}

	safeStr := "❌ 暂不可拔盘 (底层挂载或写句柄可能未完全释放)"
	if d.SafeToUnplug {
		safeStr = "✅ 可以安全拔盘 (写进程与句柄已释放)"
	}
	sb.WriteString(fmt.Sprintf("安全拔盘状态: %s\n", safeStr))

	var actionDesc string
	switch d.RecommendedAction {
	case ActionRetry:
		actionDesc = "🔄 建议重试 (retry) — 设备可正常访问，可重新尝试写盘"
	case ActionRemount:
		actionDesc = "🔌 建议重新挂载 (remount) — 设备卷未正常挂载，请重新插拔或手动挂载"
	case ActionReformat:
		actionDesc = "🧹 建议重新格式化 (reformat) — 分区结构或文件系统可能损坏，建议重新格式化"
	default:
		actionDesc = string(d.RecommendedAction)
	}
	sb.WriteString(fmt.Sprintf("建议下一步操作: %s\n", actionDesc))
	sb.WriteString("=================================")

	return sb.String()
}

// DeployTracker thread-safely monitors execution phases and file writes during deployment.
type DeployTracker struct {
	mu                sync.Mutex
	TaskID            string
	Target            string
	DeviceSummary     string
	Mode              string
	CurrentStage      string
	CurrentStepCode   StepCode
	IsFormatted       bool
	WrittenFiles      []string
	SafeToUnplug      bool
	RecommendedAction RecommendedAction
}

// NewDeployTracker constructs a DeployTracker instance with initial task metadata.
func NewDeployTracker(target string, mode string, expected *disk.DiskInfo) *DeployTracker {
	taskID := fmt.Sprintf("deploy-%s-%04d", time.Now().Format("20060102-150405"), time.Now().Nanosecond()/100000)
	summary := target
	if expected != nil && (expected.Name != "" || expected.Formatted != "") {
		name := expected.Name
		if name == "" {
			name = expected.Device
		}
		if expected.Formatted != "" {
			summary = fmt.Sprintf("%s (%s)", name, expected.Formatted)
		} else {
			summary = name
		}
	} else if disks, err := disk.GetRemovableDisks(); err == nil {
		for _, d := range disks {
			if d.Device == target {
				name := d.Name
				if name == "" {
					name = d.Device
				}
				if d.Formatted != "" {
					summary = fmt.Sprintf("%s (%s)", name, d.Formatted)
				} else {
					summary = name
				}
				break
			}
		}
	}

	return &DeployTracker{
		TaskID:            taskID,
		Target:            target,
		DeviceSummary:     summary,
		Mode:              mode,
		CurrentStage:      "验证设备",
		CurrentStepCode:   StepValidateDisk,
		IsFormatted:       false,
		WrittenFiles:      make([]string, 0),
		SafeToUnplug:      true,
		RecommendedAction: ActionRetry,
	}
}

// SetStage updates current execution stage and failure action recommendation.
func (t *DeployTracker) SetStage(stageName string, stepCode StepCode, actionOnFail RecommendedAction) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.CurrentStage = stageName
	t.CurrentStepCode = stepCode
	t.RecommendedAction = actionOnFail
}

// SetFormatted updates the formatted status of the target disk.
func (t *DeployTracker) SetFormatted(formatted bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.IsFormatted = formatted
}

// AddWrittenFile adds a newly written file path to tracker.
func (t *DeployTracker) AddWrittenFile(file string) {
	if file == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, f := range t.WrittenFiles {
		if f == file {
			return
		}
	}
	t.WrittenFiles = append(t.WrittenFiles, file)
}

// AddWrittenFiles appends a slice of written file paths to tracker.
func (t *DeployTracker) AddWrittenFiles(files []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, file := range files {
		if file == "" {
			continue
		}
		found := false
		for _, f := range t.WrittenFiles {
			if f == file {
				found = true
				break
			}
		}
		if !found {
			t.WrittenFiles = append(t.WrittenFiles, file)
		}
	}
}

// SetSafeToUnplug updates whether the target disk drive can safely be unplugged.
func (t *DeployTracker) SetSafeToUnplug(safe bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.SafeToUnplug = safe
}

// BuildDiagnostics constructs an InstallDiagnostics snapshot upon deployment failure.
func (t *DeployTracker) BuildDiagnostics(err error) *InstallDiagnostics {
	if err == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	diag := &InstallDiagnostics{
		TaskID:            t.TaskID,
		Target:            t.Target,
		DeviceSummary:     t.DeviceSummary,
		Mode:              t.Mode,
		FailedStage:       t.CurrentStage,
		FailedStepCode:    t.CurrentStepCode,
		ErrorCause:        err.Error(),
		IsFormatted:       t.IsFormatted,
		WrittenFiles:      append([]string(nil), t.WrittenFiles...),
		SafeToUnplug:      t.SafeToUnplug,
		RecommendedAction: t.RecommendedAction,
	}
	diag.ReportSummary = diag.FormatReport()
	return diag
}
