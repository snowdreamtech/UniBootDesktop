// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

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

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
)

type QEMUDriver struct{}

func (d *QEMUDriver) Type() HypervisorType {
	return TypeQEMU
}

func (d *QEMUDriver) Name() string {
	return "QEMU"
}

func (d *QEMUDriver) Priority() int {
	return 1
}

func (d *QEMUDriver) Detect() *VMStatus {
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
			return &VMStatus{
				Type:       TypeQEMU,
				Name:       "QEMU",
				Installed:  true,
				Path:       path,
				Version:    fmt.Sprintf("QEMU (%s)", name),
				Priority:   d.Priority(),
				CanBootRaw: true,
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
				return &VMStatus{
					Type:       TypeQEMU,
					Name:       "QEMU",
					Installed:  true,
					Path:       fullPath,
					Version:    fmt.Sprintf("QEMU (%s)", name),
					Priority:   d.Priority(),
					CanBootRaw: true,
				}
			}
		}
	}

	return &VMStatus{
		Type:       TypeQEMU,
		Name:       "QEMU",
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

// DetectOVMF searches common system paths for edk2 / OVMF UEFI firmware image across macOS, Linux & Windows.
func DetectOVMF() string {
	searchPaths := []string{
		"/opt/local/share/qemu/edk2-x86_64-code.fd",
		"/opt/homebrew/share/qemu/edk2-x86_64-code.fd",
		"/usr/share/OVMF/OVMF_CODE.fd",
		"/usr/share/ovmf/OVMF.fd",
		"/usr/share/qemu/ovmf-x86_64-code.bin",
		"/usr/share/edk2/ovmf/OVMF_CODE.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_CODE.fd",
		`C:\Program Files\qemu\share\edk2-x86_64-code.fd`,
	}

	for _, path := range searchPaths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// ResolveRawDiskDevice resolves volume mount paths to raw block device paths suitable for QEMU.
func ResolveRawDiskDevice(diskPath string) string {
	diskPath = strings.TrimSpace(diskPath)
	if diskPath == "" {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		if strings.HasPrefix(diskPath, "/dev/rdisk") {
			return diskPath
		}
		if strings.HasPrefix(diskPath, "/dev/disk") {
			rawNode := strings.Replace(diskPath, "/dev/disk", "/dev/rdisk", 1)
			base := filepath.Base(rawNode)
			if strings.HasPrefix(base, "rdisk") {
				diskNumPart := base[len("rdisk"):]
				if idx := strings.Index(diskNumPart, "s"); idx != -1 {
					diskNumPart = diskNumPart[:idx]
				}
				base = "rdisk" + diskNumPart
			}
			return filepath.Join(filepath.Dir(rawNode), base)
		}
		cmd := exec.Command("diskutil", "info", "-plist", diskPath)
		output, err := cmd.Output()
		if err == nil {
			plistStr := string(output)
			if parentDisk := extractPlistString(plistStr, "ParentWholeDisk"); parentDisk != "" {
				return "/dev/r" + parentDisk
			}
		}
		if strings.HasPrefix(diskPath, "disk") || strings.HasPrefix(filepath.Base(diskPath), "disk") {
			node := disk.NormalizeDarwinDiskNode(diskPath)
			return "/dev/r" + node
		}
	case "linux":
		if strings.HasPrefix(diskPath, "/dev/") {
			base := diskPath
			if strings.Contains(base, "nvme") || strings.Contains(base, "mmcblk") {
				if idx := strings.LastIndex(base, "p"); idx != -1 && idx > len("/dev/nvme") {
					base = base[:idx]
				}
			} else {
				base = strings.TrimRight(base, "0123456789")
			}
			return base
		}
	case "windows":
		cleanDrive := strings.TrimRight(diskPath, `\`)
		if len(cleanDrive) == 2 && cleanDrive[1] == ':' {
			return fmt.Sprintf(`\\.\%s`, cleanDrive)
		}
		if strings.HasPrefix(diskPath, `disk`) || strings.HasPrefix(diskPath, `Disk`) {
			diskIdx := strings.TrimPrefix(strings.TrimPrefix(diskPath, "disk"), "Disk")
			return fmt.Sprintf(`\\.\PhysicalDrive%s`, diskIdx)
		}
	}

	return diskPath
}

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

func ensureDiskPermissions(targetPath string) {
	if targetPath == "" || os.Getenv("UNIBOOT_DRY_RUN") == "1" || !strings.HasPrefix(targetPath, "/dev/") {
		return
	}
	if _, err := os.Stat(targetPath); err != nil {
		return
	}
	f, err := os.OpenFile(targetPath, os.O_RDWR, 0)
	if err == nil {
		_ = f.Close()
		return
	}

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		rawNode := "r" + diskNode

		logger.Info("Elevating disk node permissions for QEMU GUI session via osascript", "diskNode", diskNode)
		script := fmt.Sprintf(`do shell script "chmod 666 /dev/%s /dev/%s" with administrator privileges`, rawNode, diskNode)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			logger.Warn("Failed to elevate disk node permissions via osascript", "error", err)
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("pkexec", "chmod", "666", targetPath)
		_ = cmd.Run()
	}
}

func (d *QEMUDriver) Launch(ctx context.Context, diskPath string) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("QEMU simulator not detected! Please install QEMU first (e.g. via brew install qemu).")
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run QEMU launch complete", "diskPath", diskPath)
		return nil
	}

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	logger.Info("Executing QEMU preview simulation test", "disk", targetPath, "qemuPath", status.Path)

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		unmountCmd := exec.Command("diskutil", "unmountDisk", "force", fmt.Sprintf("/dev/%s", diskNode))
		_ = unmountCmd.Run()
		time.Sleep(300 * time.Millisecond)
	} else if runtime.GOOS == "linux" {
		unmountCmd := exec.Command("udisksctl", "unmount", "-b", diskPath)
		_ = unmountCmd.Run()
	}

	ensureDiskPermissions(targetPath)

	ovmfFw := DetectOVMF()

	args := []string{
		"-machine", "q35",
		"-m", "2048",
		"-device", "virtio-vga,xres=1280,yres=800",
		"-netdev", "user,id=net0",
		"-device", "e1000,netdev=net0",
	}

	if runtime.GOOS == "darwin" {
		args = append(args, "-display", "cocoa,zoom-to-fit=on")
	} else if runtime.GOOS == "linux" {
		if _, err := os.Stat("/dev/kvm"); err == nil {
			args = append(args, "-enable-kvm")
		}
	}

	if ovmfFw != "" {
		args = append(args, "-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", ovmfFw))
	}

	args = append(args, "-drive", fmt.Sprintf("file=%s,format=raw", targetPath))

	cmd := exec.Command(status.Path, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start QEMU process: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			errOutput := strings.TrimSpace(stderr.String())
			if errOutput != "" {
				return fmt.Errorf("QEMU launch message: %s", errOutput)
			}
			return fmt.Errorf("QEMU exited unexpectedly: %w", err)
		}
		return nil
	case <-time.After(800 * time.Millisecond):
		return nil
	}
}
