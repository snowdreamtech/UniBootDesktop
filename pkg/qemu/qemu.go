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
	"syscall"
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

// extractPlistString parses a string value for a given key from a plist string.
func extractPlistString(plistStr string, key string) string {
	keyPattern := fmt.Sprintf("<key>%s</key>", key)
	idx := strings.Index(plistStr, keyPattern)
	if idx == -1 {
		return ""
	}
	rest := plistStr[idx+len(keyPattern):]
	startStr := strings.Index(rest, "<string>")
	if startStr == -1 {
		return ""
	}
	rest = rest[startStr+len("<string>"):]
	endStr := strings.Index(rest, "</string>")
	if endStr == -1 {
		return ""
	}
	return strings.TrimSpace(rest[:endStr])
}

// ResolveRawDiskDevice resolves volume mount paths (e.g. /Volumes/Ventoy), partition paths (e.g. /dev/disk2s1),
// or device names (e.g. disk2) to raw unbuffered block device paths suitable for QEMU (e.g. /dev/rdisk2).
func ResolveRawDiskDevice(diskPath string) string {
	diskPath = strings.TrimSpace(diskPath)
	if diskPath == "" {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		// Case 1: Already a raw disk node (e.g. /dev/rdisk2)
		if strings.HasPrefix(diskPath, "/dev/rdisk") {
			return diskPath
		}

		// Case 2: Standard block disk node (e.g. /dev/disk2 or /dev/disk2s1)
		if strings.HasPrefix(diskPath, "/dev/disk") {
			rawNode := strings.Replace(diskPath, "/dev/disk", "/dev/rdisk", 1)
			if idx := strings.Index(filepath.Base(rawNode), "s"); idx != -1 {
				base := filepath.Base(rawNode)[:idx]
				rawNode = filepath.Join(filepath.Dir(rawNode), base)
			}
			return rawNode
		}

		// Case 3: Volume mount path (e.g. /Volumes/Ventoy, /Volumes/UNIBOOT)
		cmd := exec.Command("diskutil", "info", "-plist", diskPath)
		output, err := cmd.Output()
		if err == nil {
			plistStr := string(output)
			if parentDisk := extractPlistString(plistStr, "ParentWholeDisk"); parentDisk != "" {
				return "/dev/r" + parentDisk
			}
			if devNode := extractPlistString(plistStr, "DeviceNode"); devNode != "" {
				rawNode := strings.Replace(devNode, "/dev/disk", "/dev/rdisk", 1)
				if idx := strings.Index(filepath.Base(rawNode), "s"); idx != -1 {
					base := filepath.Base(rawNode)[:idx]
					rawNode = filepath.Join(filepath.Dir(rawNode), base)
				}
				return rawNode
			}
		}

		if strings.HasPrefix(diskPath, "disk") {
			node := diskPath
			if idx := strings.Index(node, "s"); idx != -1 {
				node = node[:idx]
			}
			return "/dev/r" + node
		}
	}

	return diskPath
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

	// Resolve volume/mount path (e.g. /Volumes/Ventoy) to underlying raw block device (/dev/rdisk2)
	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}

	if runtime.GOOS == "darwin" {
		// Extract whole disk identifier like disk2
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}

		// 1. Unmount target disk volumes to release macOS disk arbitration lock
		unmountCmd := exec.Command("diskutil", "unmountDisk", diskNode)
		_ = unmountCmd.Run()
	}

	// Build safe read-only preview command using snapshot mode
	args := []string{
		"-m", "1024",
		"-drive", fmt.Sprintf("file=%s,format=raw,snapshot=on", targetPath),
	}

	cmd := exec.Command(status.Path, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Detach QEMU process group from parent Wails app so the window stays alive independently
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}

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
				// If permission is denied on macOS, try executing via osascript administrator privileges
				if runtime.GOOS == "darwin" && strings.Contains(errOutput, "Permission denied") {
					script := fmt.Sprintf(`do shell script "'%s' -m 1024 -drive 'file=%s,format=raw,snapshot=on' >/dev/null 2>&1 &" with administrator privileges`, status.Path, targetPath)
					adminCmd := exec.Command("osascript", "-e", script)
					if adminErr := adminCmd.Run(); adminErr == nil {
						return nil
					}
				}
				return fmt.Errorf("QEMU 启动提示: %s", errOutput)
			}
			return fmt.Errorf("QEMU 启动异常退出: %w", err)
		}
		return nil
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}
