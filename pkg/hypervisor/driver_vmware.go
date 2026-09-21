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
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
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
	if !status.Installed && os.Getenv("UNIBOOT_DRY_RUN") == "" {
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
	ensureDiskPermissions(targetPath)
	unmountTargetDisk(targetPath)

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
	_ = os.RemoveAll(tmpDir)
	_ = os.MkdirAll(tmpDir, 0755)

	vmdkBase := filepath.Join(tmpDir, "uniboot_raw")
	vmdkPath := vmdkBase + ".vmdk"
	vmxPath := filepath.Join(tmpDir, "UniBoot.vmx")

	fwSetting := "efi"
	if bootMode == BootModeBIOS {
		fwSetting = "bios"
	}

	diskDev := targetPath
	if runtime.GOOS == "darwin" {
		diskDev = "/dev/" + disk.NormalizeDarwinDiskNode(targetPath)
	}

	// 1. On macOS, use official vmware-rawdiskCreator if available
	rawCreator := "/Applications/VMware Fusion.app/Contents/Library/vmware-rawdiskCreator"
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(rawCreator); err == nil {
			logger.Info("Creating native VMware raw disk VMDK via vmware-rawdiskCreator", "diskDev", diskDev, "vmdkPath", vmdkPath)
			_ = os.Remove(vmdkPath)
			cmd := exec.Command(rawCreator, "create", diskDev, "fullDevice", vmdkBase, "ide")
			if err := cmd.Run(); err != nil {
				logger.Warn("vmware-rawdiskCreator returned error, using fallback descriptor", "error", err)
			}
		}
	}

	// 2. Fallback to manual descriptor if creator didn't generate valid file
	if info, err := os.Stat(vmdkPath); err != nil || info.Size() == 0 {
		sectors := getDiskSectorCount(diskDev)
		cylinders := sectors / (255 * 63)
		if cylinders == 0 {
			cylinders = 1024
		}
		rawDiskContent := fmt.Sprintf(`# Disk DescriptorFile
version=1
encoding="UTF-8"
CID=fffffffe
parentCID=ffffffff
createType="fullDevice"

# Extent description
RW %d FLAT "%s" 0

# The Disk Data Base
#DDB
ddb.adapterType = "ide"
ddb.geometry.cylinders = "%d"
ddb.geometry.heads = "255"
ddb.geometry.sectors = "63"
ddb.longContentID = "1234567890"
ddb.virtualHWVersion = "14"
`, sectors, diskDev, cylinders)
		_ = os.WriteFile(vmdkPath, []byte(rawDiskContent), 0600)
	}

	memMB := GetRecommendedVMMemoryMB()
	vcpus := GetRecommendedVCPUs()

	// 2. Generate clean VMX configuration
	vmxContent := fmt.Sprintf(`.encoding = "UTF-8"
config.version = "8"
virtualHW.version = "18"
pciBridge0.present = "TRUE"
mks.enable3d = "TRUE"
numvcpus = "%d"
memsize = "%d"
firmware = "%s"
nvram = "UniBoot.nvram"
floppy0.present = "FALSE"
sata0.present = "TRUE"
sata0:0.present = "TRUE"
sata0:0.fileName = "uniboot_raw.vmdk"
sata0:0.mode = "independent-nonpersistent"
ethernet0.present = "TRUE"
ethernet0.virtualDev = "e1000"
ethernet0.connectionType = "nat"
ethernet0.addressType = "generated"
displayName = "UniBoot"
guestOS = "other-64"
`, vcpus, memMB, fwSetting)
	_ = os.WriteFile(vmxPath, []byte(vmxContent), 0600)

	// 3. Launch VMware Fusion
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

func getDiskSectorCount(diskDev string) int64 {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("diskutil", "info", "-plist", diskDev)
		output, err := cmd.Output()
		if err == nil {
			plistStr := string(output)
			keyPattern := "<key>TotalSize</key>"
			if idx := strings.Index(plistStr, keyPattern); idx != -1 {
				rest := plistStr[idx+len(keyPattern):]
				startInt := strings.Index(rest, "<integer>")
				endInt := strings.Index(rest, "</integer>")
				if startInt != -1 && endInt != -1 && startInt < endInt {
					valStr := rest[startInt+len("<integer>"):endInt]
					var size int64
					if _, fmtErr := fmt.Sscanf(strings.TrimSpace(valStr), "%d", &size); fmtErr == nil && size > 0 {
						return size / 512
					}
				}
			}
		}
	}
	if info, err := os.Stat(diskDev); err == nil && info.Size() > 0 {
		return info.Size() / 512
	}
	return 15728640 // Default fallback ~8GB
}
