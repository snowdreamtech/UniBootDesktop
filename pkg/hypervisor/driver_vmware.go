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

type VMwareDriver struct{}

func (d *VMwareDriver) Type() HypervisorType {
	return TypeVMware
}

func (d *VMwareDriver) Name() string {
	if runtime.GOOS == "darwin" {
		return "VMware Fusion"
	}
	return "VMware Workstation"
}

func (d *VMwareDriver) Priority() int {
	return 3
}

func (d *VMwareDriver) Detect() *VMStatus {
	// 1. Look for vmrun command
	if path, err := exec.LookPath("vmrun"); err == nil {
		return &VMStatus{
			Type:       TypeVMware,
			Name:       d.Name(),
			Installed:  true,
			Path:       path,
			Version:    fmt.Sprintf("%s (vmrun CLI)", d.Name()),
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Platform specific paths
	commonPaths := []string{}
	if runtime.GOOS == "darwin" {
		commonPaths = append(commonPaths,
			"/Applications/VMware Fusion.app/Contents/Library/vmrun",
			"/Applications/VMware Fusion.app",
		)
	} else if runtime.GOOS == "windows" {
		commonPaths = append(commonPaths,
			`C:\Program Files (x86)\VMware\VMware Workstation\vmrun.exe`,
			`C:\Program Files\VMware\VMware Workstation\vmrun.exe`,
			`C:\Program Files (x86)\VMware\VMware Workstation\vmware.exe`,
		)
	} else if runtime.GOOS == "linux" {
		commonPaths = append(commonPaths,
			"/usr/bin/vmrun",
			"/usr/bin/vmware",
		)
	}

	for _, p := range commonPaths {
		if info, err := os.Stat(p); err == nil {
			if !info.IsDir() || strings.HasSuffix(p, ".app") {
				return &VMStatus{
					Type:       TypeVMware,
					Name:       d.Name(),
					Installed:  true,
					Path:       p,
					Version:    fmt.Sprintf("%s (%s)", d.Name(), filepath.Base(p)),
					Priority:   d.Priority(),
					CanBootRaw: true,
				}
			}
		}
	}

	return &VMStatus{
		Type:       TypeVMware,
		Name:       d.Name(),
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *VMwareDriver) Launch(ctx context.Context, diskPath string, bootMode string) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run VMware launch complete", "diskPath", diskPath, "bootMode", bootMode)
		return nil
	}

	logger.Info("Executing VMware preview test instance", "disk", diskPath, "vmwarePath", status.Path, "bootMode", bootMode)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	unmountTargetDisk(targetPath)
	ensureDiskPermissions(targetPath)

	if err := launchVMwareVM(status, targetPath, bootMode); err == nil {
		return nil
	}

	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", "VMware Fusion")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to launch VMware Fusion: %w", err)
		}
		return nil
	}

	cmd := exec.Command(status.Path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start VMware: %w", err)
	}

	return nil
}

func launchVMwareVM(status *VMStatus, targetPath string, bootMode string) error {
	tmpDir := filepath.Join(os.TempDir(), "uniboot_vmware")
	_ = os.MkdirAll(tmpDir, 0755)
	vmdkPath := filepath.Join(tmpDir, "uniboot_raw.vmdk")
	vmxPath := filepath.Join(tmpDir, "UniBootPreview.vmx")

	fwSetting := "efi"
	if bootMode == BootModeBIOS {
		fwSetting = "bios"
	}

	rawDiskContent := fmt.Sprintf(`# Disk DescriptorFile
version=1
encoding="UTF-8"
CID=fffffffe
parentCID=ffffffff
isNativeSnapshot="no"
createType="fullDevice"

# Extent description
RW 20000000 FLAT "%s" 0

# The Disk Data Base 
DDB
ddb.adapterType = "lsilogic"
ddb.geometry.cylinders = "1024"
ddb.geometry.heads = "255"
ddb.geometry.sectors = "63"
ddb.longContentID = "1234567890"
ddb.virtualHWVersion = "18"
`, targetPath)
	_ = os.WriteFile(vmdkPath, []byte(rawDiskContent), 0644)

	vmxContent := fmt.Sprintf(`.encoding = "UTF-8"
config.version = "8"
virtualHW.version = "18"
pciBridge0.present = "TRUE"
mks.enable3d = "TRUE"
memory = "2048"
firmware = "%s"
sata0.present = "TRUE"
sata0:0.present = "TRUE"
sata0:0.fileName = "%s"
sata0:0.deviceType = "rawDisk"
displayName = "UniBoot Boot Preview"
guestOS = "other-64"
`, fwSetting, vmdkPath)
	_ = os.WriteFile(vmxPath, []byte(vmxContent), 0644)

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", "-W", "-a", "VMware Fusion", vmxPath)
	} else if strings.HasSuffix(status.Path, "vmrun") || strings.HasSuffix(status.Path, "vmrun.exe") {
		cmd = exec.Command(status.Path, "-T", "ws", "start", vmxPath, "gui")
	} else {
		cmd = exec.Command(status.Path, vmxPath)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		_ = cmd.Wait()
		remountTargetDisk(targetPath)
	}()

	return nil
}
