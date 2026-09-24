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
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/privilege"
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
		if strings.HasPrefix(diskPath, "/dev/disk") || strings.HasPrefix(diskPath, "/dev/rdisk") {
			node := disk.NormalizeDarwinDiskNode(diskPath)
			return "/dev/r" + node
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

func ensureDiskPermissions(targetPath string) (func(), error) {
	noop := func() {}
	if targetPath == "" || os.Getenv("UNIBOOT_DRY_RUN") == "1" || !strings.HasPrefix(targetPath, "/dev/") {
		return noop, nil
	}
	if _, err := os.Stat(targetPath); err != nil {
		return noop, nil
	}
	f, err := os.OpenFile(targetPath, os.O_RDWR, 0)
	if err == nil {
		_ = f.Close()
		return noop, nil
	}

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		rawNode := "r" + diskNode

		logger.Info("Temporarily relaxing disk node permissions for hypervisor GUI session", "diskNode", diskNode)
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(rawNode) || !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(diskNode) {
			logger.Warn("Invalid disk node format, skipping permission elevation", "rawNode", rawNode, "diskNode", diskNode)
			return noop, nil
		}
		return privilege.RelaxRawDiskPermissionsTemporarilyWithResult("/dev/"+rawNode, "/dev/"+diskNode)
	}
	if runtime.GOOS == "linux" {
		if err := privilege.ValidateRawDevicePath(targetPath); err != nil {
			logger.Warn("Rejecting unsafe disk permission change target", "targetPath", targetPath, "error", err)
			return noop, err
		}
		return privilege.RelaxRawDiskPermissionsTemporarilyWithResult(targetPath)
	}
	return noop, nil
}

// getOrCreateVarsFile finds or generates an EFI VARS file for QEMU dual pflash drives.
func getOrCreateVarsFile() string {
	tmpVars := filepath.Join(os.TempDir(), "uniboot_vars.fd")

	varsCandidates := []string{
		"/opt/local/share/qemu/edk2-i386-vars.fd",
		"/opt/homebrew/share/qemu/edk2-i386-vars.fd",
		"/usr/share/OVMF/OVMF_VARS.fd",
		"/usr/share/ovmf/OVMF_VARS.fd",
		"/usr/share/edk2/ovmf/OVMF_VARS.fd",
		`C:\Program Files\qemu\share\edk2-i386-vars.fd`,
	}
	for _, p := range varsCandidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			data, err := os.ReadFile(p)
			if err == nil && len(data) > 0 {
				_ = os.WriteFile(tmpVars, data, 0600)
				return tmpVars
			}
		}
	}
	if _, err := os.Stat(tmpVars); err != nil {
		buf := make([]byte, 540*1024)
		_ = os.WriteFile(tmpVars, buf, 0600)
	}
	return tmpVars
}

func (d *QEMUDriver) Launch(ctx context.Context, diskPath string, bootMode string) error {
	return d.LaunchWithConfig(ctx, diskPath, VMConfig{
		CpuCores:     GetRecommendedVCPUs(),
		MemoryMB:     GetRecommendedVMMemoryMB(),
		BootMode:     bootMode,
		DisplayAccel: true,
		SecureBoot:   false,
	})
}

func (d *QEMUDriver) LaunchWithConfig(ctx context.Context, diskPath string, cfg VMConfig) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("QEMU simulator not detected! Please install QEMU first (e.g. via brew install qemu).")
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run QEMU launch complete", "diskPath", diskPath, "bootMode", cfg.BootMode)
		return nil
	}

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	logger.Info("Executing QEMU preview simulation test", "disk", targetPath, "qemuPath", status.Path, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB, "accel", cfg.DisplayAccel)

	unmountTargetDisk(targetPath)

	ovmfFw := DetectOVMF()

	memMB := cfg.MemoryMB
	if memMB <= 0 {
		memMB = GetRecommendedVMMemoryMB()
	}
	vcpus := cfg.CpuCores
	if vcpus <= 0 {
		vcpus = GetRecommendedVCPUs()
	}

	args := []string{
		"-name", "UniBoot",
		"-snapshot",
		"-machine", "q35",
		"-smp", fmt.Sprintf("%d", vcpus),
		"-m", fmt.Sprintf("%d", memMB),
	}

	if cfg.DisplayAccel {
		switch runtime.GOOS {
		case "darwin":
			args = append(args, "-accel", "hvf")
		case "linux":
			if _, err := os.Stat("/dev/kvm"); err == nil {
				args = append(args, "-accel", "kvm")
			}
		case "windows":
			args = append(args, "-accel", "whpx")
		}
	}

	args = append(args,
		"-device", "virtio-vga,xres=1280,yres=800",
		"-netdev", "user,id=net0",
		"-device", "e1000,netdev=net0",
	)

	if runtime.GOOS == "darwin" {
		args = append(args, "-display", "cocoa,zoom-to-fit=on")
	}

	if (cfg.BootMode == BootModeUEFI || cfg.BootMode == BootModeAuto) && ovmfFw != "" {
		varsFw := getOrCreateVarsFile()
		logger.Info("Booting QEMU in UEFI mode with dual pflash firmware", "codeFw", ovmfFw, "varsFw", varsFw)
		args = append(args,
			"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", ovmfFw),
			"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", varsFw),
		)
	} else if cfg.BootMode == BootModeUEFI && ovmfFw == "" {
		return fmt.Errorf("UEFI firmware (OVMF/edk2) not found on system! Please install edk2-ovmf or switch to BIOS mode.")
	} else {
		logger.Info("Booting QEMU in Legacy BIOS mode (SeaBIOS)")
	}

	args = append(args, "-drive", fmt.Sprintf("file=%s,format=raw,file.locking=off", targetPath))

	runQEMU := func() error {
		restoreDiskPerms, err := ensureDiskPermissions(targetPath)
		if err != nil {
			remountTargetDisk(targetPath)
			return fmt.Errorf("failed to acquire target disk permissions: %w", err)
		}

		// Preflight check: verify that current process can actually open targetPath
		if strings.HasPrefix(targetPath, "/dev/") && os.Getenv("UNIBOOT_DRY_RUN") != "1" {
			f, err := os.OpenFile(targetPath, os.O_RDWR, 0)
			if err != nil {
				remountTargetDisk(targetPath)
				restoreDiskPerms()
				return fmt.Errorf("unable to access raw disk %s: %w", targetPath, err)
			}
			_ = f.Close()
		}

		cmd := exec.Command(status.Path, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if runtime.GOOS != "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		}

		if err := cmd.Start(); err != nil {
			remountTargetDisk(targetPath)
			restoreDiskPerms()
			return fmt.Errorf("failed to start QEMU process: %w", err)
		}

		waitErrCh := make(chan error, 1)
		go func() {
			err := cmd.Wait()
			// Send exit status immediately to unblock startup check if still waiting
			waitErrCh <- err

			// Post-exit teardown and notifications
			time.Sleep(300 * time.Millisecond)
			remountTargetDisk(targetPath)
			restoreDiskPerms()
			NotifyVMExited(targetPath, err)
		}()

		select {
		case err := <-waitErrCh:
			errOutput := strings.TrimSpace(stderr.String())
			if errOutput != "" {
				return fmt.Errorf("QEMU launch message: %s", errOutput)
			}
			if err != nil {
				return fmt.Errorf("QEMU exited unexpectedly: %w", err)
			}
			return fmt.Errorf("QEMU process exited immediately after startup")
		case <-time.After(1200 * time.Millisecond):
			// QEMU has started successfully and is running in the background.
			// Permissions will be safely restored when the process terminates via cmd.Wait().
			return nil
		}
	}

	err := runQEMU()
	if err != nil && (strings.Contains(err.Error(), "Resource busy") || strings.Contains(err.Error(), "busy")) {
		logger.Warn("QEMU failed with Resource busy, retrying after disk unmount", "disk", targetPath)
		time.Sleep(500 * time.Millisecond)
		unmountTargetDisk(targetPath)
		err = runQEMU()
	}
	return err
}
