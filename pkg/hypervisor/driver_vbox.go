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
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run VirtualBox launch complete", "diskPath", diskPath, "bootMode", bootMode)
		return nil
	}

	logger.Info("Executing VirtualBox preview test instance", "disk", diskPath, "vboxPath", status.Path, "bootMode", bootMode)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	unmountTargetDisk(targetPath)
	ensureDiskPermissions(targetPath)

	vboxManage, _ := exec.LookPath("VBoxManage")
	if vboxManage == "" && runtime.GOOS == "darwin" {
		if info, err := os.Stat("/Applications/VirtualBox.app/Contents/MacOS/VBoxManage"); err == nil && !info.IsDir() {
			vboxManage = "/Applications/VirtualBox.app/Contents/MacOS/VBoxManage"
		} else if info, err := os.Stat("/usr/local/bin/VBoxManage"); err == nil && !info.IsDir() {
			vboxManage = "/usr/local/bin/VBoxManage"
		}
	}

	if vboxManage != "" {
		if err := launchVirtualBoxVM(vboxManage, targetPath, bootMode); err == nil {
			return nil
		}
	}

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

func launchVirtualBoxVM(vboxManage string, targetPath string, bootMode string) error {
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
	if bootMode == BootModeBIOS {
		fwSetting = "bios"
	}

	_ = exec.Command(vboxManage, "storagectl", vmName, "--name", "SATA", "--add", "sata", "--controller", "IntelAhci").Run()
	_ = exec.Command(vboxManage, "storageattach", vmName, "--storagectl", "SATA", "--port", "0", "--device", "0", "--type", "hdd", "--medium", vmdkPath).Run()
	_ = exec.Command(vboxManage, "modifyvm", vmName, "--firmware", fwSetting, "--cpus", "2", "--memory", "4096").Run()

	startCmd := exec.Command(vboxManage, "startvm", vmName)
	if err := startCmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = startCmd.Wait()
		remountTargetDisk(targetPath)
	}()
	return nil
}
