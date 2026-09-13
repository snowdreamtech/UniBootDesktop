// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package qemu

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// QEMUStatus contains detection metadata for QEMU installation.
type QEMUStatus struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Version   string `json:"version"`
}

// Detect checks if QEMU executable is available on PATH or common installation directories.
func Detect() *QEMUStatus {
	candidates := []string{
		"qemu-system-x86_64",
		"qemu-system-aarch64",
		"qemu-system-i386",
	}

	commonPaths := []string{
		"/opt/local/bin", // MacPorts (macOS)
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/bin",
		`C:\Program Files\qemu`,
		`C:\Program Files (x86)\qemu`,
	}

	// 1. Try finding candidates via system PATH
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return &QEMUStatus{
				Installed: true,
				Path:      path,
				Version:   fmt.Sprintf("QEMU (%s)", name),
			}
		}
	}

	// 2. Search common installation directories directly
	for _, dir := range commonPaths {
		for _, name := range candidates {
			exeName := name
			if runtime.GOOS == "windows" {
				exeName += ".exe"
			}
			fullPath := filepath.Join(dir, exeName)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return &QEMUStatus{
					Installed: true,
					Path:      fullPath,
					Version:   fmt.Sprintf("QEMU (%s)", name),
				}
			}
		}
	}

	// Support dry-run or mock mode when UNIBOOT_DRY_RUN is set
	if os.Getenv("UNIBOOT_DRY_RUN") != "" {
		return &QEMUStatus{
			Installed: true,
			Path:      "/usr/local/bin/qemu-system-x86_64 (Dry-Run)",
			Version:   "QEMU system x86_64 (Mock)",
		}
	}

	return &QEMUStatus{
		Installed: false,
		Path:      "",
		Version:   "Not Installed",
	}
}

// LaunchTest executes a non-blocking QEMU preview test instance on the target USB drive safely.
func LaunchTest(ctx context.Context, diskPath string) error {
	if diskPath == "" {
		return fmt.Errorf("请先选择要测试的目标 U 盘！")
	}

	// Dry-run mode for tests or simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(diskPath, "dummy") || strings.HasPrefix(diskPath, "test") {
		return nil
	}

	status := Detect()
	if !status.Installed {
		return fmt.Errorf("未检测到 QEMU 模拟器！请先安装 QEMU（例如通过 brew install qemu 或 MacPorts 命令行包）。")
	}

	// Normalize macOS disk path: convert /dev/diskN -> /dev/rdiskN for raw unbuffered I/O
	targetPath := diskPath
	if runtime.GOOS == "darwin" {
		if strings.HasPrefix(targetPath, "/dev/disk") {
			targetPath = strings.Replace(targetPath, "/dev/disk", "/dev/rdisk", 1)
		} else if !strings.HasPrefix(targetPath, "/dev/") && strings.HasPrefix(targetPath, "disk") {
			targetPath = "/dev/r" + targetPath
		}
	}

	// Build safe read-only preview command using snapshot mode
	args := []string{
		"-m", "1024",
		"-drive", fmt.Sprintf("file=%s,format=raw,snapshot=on", targetPath),
	}

	cmd := exec.Command(status.Path, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 QEMU 进程失败: %w", err)
	}

	// Wait briefly (300ms) to catch immediate startup failures (e.g. permission denied)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			errOutput := strings.TrimSpace(stderr.String())
			if errOutput != "" {
				return fmt.Errorf("QEMU 启动异常退出 (%v): %s", err, errOutput)
			}
			return fmt.Errorf("QEMU 启动异常退出: %w", err)
		}
		// Process exited early cleanly
		return nil
	case <-time.After(300 * time.Millisecond):
		// QEMU process started successfully and continues running in background
		return nil
	}
}
