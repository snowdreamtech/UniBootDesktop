// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

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
)

type VirtualBoxDriver struct{}

func (d *VirtualBoxDriver) Type() HypervisorType {
	return TypeVirtualBox
}

func (d *VirtualBoxDriver) Name() string {
	return "Oracle VM VirtualBox"
}

func (d *VirtualBoxDriver) Priority() int {
	return 4
}

func (d *VirtualBoxDriver) Detect() *VMStatus {
	// 1. Check system PATH for VBoxManage
	if path, err := exec.LookPath("VBoxManage"); err == nil {
		return &VMStatus{
			Type:       TypeVirtualBox,
			Name:       d.Name(),
			Installed:  true,
			Path:       path,
			Version:    "VirtualBox (VBoxManage CLI)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Platform specific paths
	commonPaths := []string{}
	if runtime.GOOS == "darwin" {
		commonPaths = append(commonPaths,
			"/usr/local/bin/VBoxManage",
			"/Applications/VirtualBox.app/Contents/MacOS/VBoxManage",
			"/Applications/VirtualBox.app",
		)
	} else if runtime.GOOS == "windows" {
		commonPaths = append(commonPaths,
			`C:\Program Files\Oracle\VirtualBox\VBoxManage.exe`,
			`C:\Program Files (x86)\Oracle\VirtualBox\VBoxManage.exe`,
			`C:\Program Files\Oracle\VirtualBox\VirtualBox.exe`,
		)
	} else if runtime.GOOS == "linux" {
		commonPaths = append(commonPaths,
			"/usr/bin/VBoxManage",
			"/usr/bin/virtualbox",
		)
	}

	for _, p := range commonPaths {
		if info, err := os.Stat(p); err == nil {
			if !info.IsDir() || strings.HasSuffix(p, ".app") {
				return &VMStatus{
					Type:       TypeVirtualBox,
					Name:       d.Name(),
					Installed:  true,
					Path:       p,
					Version:    fmt.Sprintf("VirtualBox (%s)", filepath.Base(p)),
					Priority:   d.Priority(),
					CanBootRaw: true,
				}
			}
		}
	}

	return &VMStatus{
		Type:       TypeVirtualBox,
		Name:       d.Name(),
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *VirtualBoxDriver) Launch(ctx context.Context, diskPath string, bootMode string) error {
	return d.LaunchWithConfig(ctx, diskPath, VMConfig{
		BootMode:     bootMode,
		CpuCores:     GetRecommendedVCPUs(),
		MemoryMB:     GetRecommendedVMMemoryMB(),
		DisplayAccel: true,
	})
}

// LaunchWithConfig launches VirtualBox with custom VMConfig and monitors guest OS shutdown.
func (d *VirtualBoxDriver) LaunchWithConfig(ctx context.Context, diskPath string, cfg VMConfig) error {
	status := d.Detect()
	if !status.Installed && os.Getenv("UNIBOOT_DRY_RUN") == "" {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run VirtualBox launch complete", "diskPath", diskPath, "bootMode", cfg.BootMode)
		return nil
	}

	logger.Info("Executing VirtualBox preview test instance", "disk", diskPath, "vboxPath", status.Path, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	restoreDiskPerms := ensureDiskPermissions(targetPath)
	unmountTargetDisk(targetPath)

	vboxManage, _ := exec.LookPath("VBoxManage")
	if vboxManage == "" && runtime.GOOS == "darwin" {
		if info, err := os.Stat("/Applications/VirtualBox.app/Contents/MacOS/VBoxManage"); err == nil && !info.IsDir() {
			vboxManage = "/Applications/VirtualBox.app/Contents/MacOS/VBoxManage"
		} else if info, err := os.Stat("/usr/local/bin/VBoxManage"); err == nil && !info.IsDir() {
			vboxManage = "/usr/local/bin/VBoxManage"
		}
	}

	if vboxManage != "" {
		if err := launchVirtualBoxVM(ctx, vboxManage, targetPath, cfg, restoreDiskPerms); err == nil {
			return nil
		}
	}

	remountTargetDisk(targetPath)
	restoreDiskPerms()

	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", "VirtualBox")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open VirtualBox application: %w", err)
		}
		return nil
	}

	cmd := exec.Command(status.Path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start VirtualBox: %w", err)
	}

	return nil
}

func launchVirtualBoxVM(ctx context.Context, vboxManage string, targetPath string, cfg VMConfig, restoreDiskPerms func()) error {
	tmpDir := filepath.Join(os.TempDir(), "uniboot_vbox")
	_ = os.MkdirAll(tmpDir, 0755)
	vmdkPath := filepath.Join(tmpDir, "uniboot_raw.vmdk")
	_ = os.Remove(vmdkPath)

	createCmd := exec.Command(vboxManage, "internalcommands", "createrawvmdk", "-filename", vmdkPath, "-rawdisk", targetPath)
	if err := createCmd.Run(); err != nil {
		logger.Warn("VBoxManage createrawvmdk failed, falling back to GUI app launch", "error", err)
		return err
	}

	vmName := "UniBoot"
	_ = exec.Command(vboxManage, "unregistervm", vmName, "--delete").Run()

	if err := exec.Command(vboxManage, "createvm", "--name", vmName, "--ostype", "Other_64", "--register").Run(); err != nil {
		logger.Warn("VBoxManage createvm failed", "error", err)
		return err
	}

	fwSetting := "efi"
	if cfg.BootMode == BootModeBIOS {
		fwSetting = "bios"
	}

	memMB := cfg.MemoryMB
	if memMB <= 0 {
		memMB = GetRecommendedVMMemoryMB()
	}
	vcpus := cfg.CpuCores
	if vcpus <= 0 {
		vcpus = GetRecommendedVCPUs()
	}

	_ = exec.Command(vboxManage, "storagectl", vmName, "--name", "SATA", "--add", "sata", "--controller", "IntelAhci").Run()
	_ = exec.Command(vboxManage, "storageattach", vmName, "--storagectl", "SATA", "--port", "0", "--device", "0", "--type", "hdd", "--medium", vmdkPath, "--mtype", "immutable").Run()
	_ = exec.Command(vboxManage, "modifyvm", vmName, "--firmware", fwSetting, "--cpus", fmt.Sprintf("%d", vcpus), "--memory", fmt.Sprintf("%d", memMB)).Run()

	startCmd := exec.Command(vboxManage, "startvm", vmName)
	if err := startCmd.Start(); err != nil {
		return err
	}

	go monitorVirtualBoxVM(ctx, vboxManage, vmName, targetPath, restoreDiskPerms)
	return nil
}

func monitorVirtualBoxVM(ctx context.Context, vboxManage, vmName, targetPath string, restoreDiskPerms func()) {
	for i := 0; i < 20; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}
		time.Sleep(500 * time.Millisecond)
		if isVirtualBoxVMRunning(vboxManage, vmName) {
			logger.Info("VirtualBox guest system confirmed running", "vm", vmName)
			break
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	consecutiveInactive := 0
	for {
		select {
		case <-ctx.Done():
			goto cleanup
		case <-ticker.C:
			if !isVirtualBoxVMRunning(vboxManage, vmName) {
				consecutiveInactive++
				if consecutiveInactive >= 2 {
					logger.Info("VirtualBox guest system power-off detected, resetting simulation state", "vm", vmName)
					goto cleanup
				}
			} else {
				consecutiveInactive = 0
			}
		}
	}

cleanup:
	time.Sleep(300 * time.Millisecond)
	remountTargetDisk(targetPath)
	if restoreDiskPerms != nil {
		restoreDiskPerms()
	}
	NotifyVMExited(targetPath, nil)
}

func isVirtualBoxVMRunning(vboxManage, vmName string) bool {
	cmd := exec.Command(vboxManage, "list", "runningvms")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf("%q", vmName)) || strings.Contains(string(out), vmName)
}
