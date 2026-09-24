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
	return d.LaunchWithConfig(ctx, diskPath, VMConfig{
		BootMode:     bootMode,
		CpuCores:     GetRecommendedVCPUs(),
		MemoryMB:     GetRecommendedVMMemoryMB(),
		DisplayAccel: true,
	})
}

// LaunchWithConfig launches VMware with custom VMConfig and monitors guest OS shutdown.
func (d *VMwareDriver) LaunchWithConfig(ctx context.Context, diskPath string, cfg VMConfig) error {
	status := d.Detect()
	if !status.Installed && os.Getenv("UNIBOOT_DRY_RUN") == "" {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run VMware launch complete", "diskPath", diskPath, "bootMode", cfg.BootMode)
		return nil
	}

	logger.Info("Executing VMware preview test instance", "disk", diskPath, "vmwarePath", status.Path, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	restoreDiskPerms := ensureDiskPermissions(targetPath)
	unmountTargetDisk(targetPath)

	if err := launchVMwareVM(ctx, status, targetPath, cfg, restoreDiskPerms); err != nil {
		remountTargetDisk(targetPath)
		restoreDiskPerms()
		return err
	}

	return nil
}

func launchVMwareVM(ctx context.Context, status *VMStatus, targetPath string, cfg VMConfig, restoreDiskPerms func()) error {
	tmpDir := filepath.Join(os.TempDir(), "uniboot_vmware")
	_ = os.RemoveAll(tmpDir)
	_ = os.MkdirAll(tmpDir, 0755)

	vmdkBase := filepath.Join(tmpDir, "uniboot_raw")
	vmdkPath := vmdkBase + ".vmdk"
	vmxPath := filepath.Join(tmpDir, "UniBoot.vmx")

	fwSetting := "efi"
	if cfg.BootMode == BootModeBIOS {
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

	memMB := cfg.MemoryMB
	if memMB <= 0 {
		memMB = GetRecommendedVMMemoryMB()
	}
	vcpus := cfg.CpuCores
	if vcpus <= 0 {
		vcpus = GetRecommendedVCPUs()
	}

	// 3. Generate clean VMX configuration with automatic power-on
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
gui.powerOnAtStartup = "TRUE"
`, vcpus, memMB, fwSetting)
	_ = os.WriteFile(vmxPath, []byte(vmxContent), 0600)

	// 4. Launch VMware GUI without blocking on the whole host application (-W removed)
	vmrunPath := findVmrunPath()
	var startCmd *exec.Cmd

	if vmrunPath != "" {
		if runtime.GOOS == "darwin" {
			startCmd = exec.Command(vmrunPath, "-T", "fusion", "start", vmxPath, "gui")
		} else {
			startCmd = exec.Command(vmrunPath, "-T", "ws", "start", vmxPath, "gui")
		}
	} else if runtime.GOOS == "darwin" {
		startCmd = exec.Command("open", "-a", "VMware Fusion", vmxPath)
	} else {
		startCmd = exec.Command(status.Path, vmxPath)
	}

	if err := startCmd.Start(); err != nil {
		if runtime.GOOS == "darwin" && vmrunPath != "" {
			fallbackCmd := exec.Command("open", "-a", "VMware Fusion", vmxPath)
			if fErr := fallbackCmd.Start(); fErr != nil {
				return fmt.Errorf("failed to start VMware: %w (fallback: %v)", err, fErr)
			}
		} else {
			return fmt.Errorf("failed to start VMware: %w", err)
		}
	}

	// 5. Watch for the guest system power-off (instead of waiting for VMware application to quit)
	go monitorVMwareVM(ctx, vmrunPath, tmpDir, vmxPath, targetPath, restoreDiskPerms)

	return nil
}

func monitorVMwareVM(ctx context.Context, vmrunPath, tmpDir, vmxPath, targetPath string, restoreDiskPerms func()) {
	// Wait up to 15s for the guest VM to enter the running state
	for i := 0; i < 30; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}
		time.Sleep(500 * time.Millisecond)
		if isVMwareVMRunning(vmrunPath, tmpDir, vmxPath) {
			logger.Info("VMware guest system confirmed running", "vmx", vmxPath)
			break
		}
	}

	// Poll until the guest VM powers off or user cancels
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	consecutiveInactive := 0
	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled during VMware session, terminating", "vmx", vmxPath)
			goto cleanup
		case <-ticker.C:
			running := isVMwareVMRunning(vmrunPath, tmpDir, vmxPath)
			if !running {
				consecutiveInactive++
				// Require 2 consecutive checks (2 seconds) to avoid transient state during reboot
				if consecutiveInactive >= 2 {
					logger.Info("VMware guest system shutdown detected, cleaning up session", "vmx", vmxPath)
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

func findVmrunPath() string {
	if path, err := exec.LookPath("vmrun"); err == nil {
		return path
	}
	if path, err := exec.LookPath("vmrun.exe"); err == nil {
		return path
	}
	candidates := []string{
		"/Applications/VMware Fusion.app/Contents/Library/vmrun",
		"/Applications/VMware Fusion Tech Preview.app/Contents/Library/vmrun",
		`C:\Program Files (x86)\VMware\VMware Workstation\vmrun.exe`,
		`C:\Program Files\VMware\VMware Workstation\vmrun.exe`,
		`C:\Program Files (x86)\VMware\VMware Player\vmrun.exe`,
		`C:\Program Files\VMware\VMware Player\vmrun.exe`,
		"/usr/bin/vmrun",
		"/usr/local/bin/vmrun",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func isVMwareVMRunning(vmrunPath, tmpDir, vmxPath string) bool {
	// 1. Authoritative check via vmrun CLI
	if vmrunPath != "" {
		if running, ok := isVMRunningViaVmrun(vmrunPath, vmxPath); ok {
			return running
		}
	}

	// 2. Check VMware disk locks in tmpDir (.lck)
	// VMware creates *.lck directories when the VM is powered ON, and removes them when powered OFF.
	if hasVMwareLocks(tmpDir) {
		return true
	}

	// 3. Check vmware-vmx process
	if isVMwareProcessRunning(vmxPath) {
		return true
	}

	return false
}

func isVMRunningViaVmrun(vmrunPath, vmxPath string) (bool, bool) {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command(vmrunPath, "-T", "fusion", "list")
	} else {
		cmd = exec.Command(vmrunPath, "-T", "ws", "list")
	}
	out, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command(vmrunPath, "list")
		out, err = cmd2.Output()
		if err != nil {
			return false, false
		}
	}
	outStr := string(out)
	vmxBase := filepath.Base(vmxPath)
	return strings.Contains(outStr, vmxBase) || strings.Contains(outStr, vmxPath), true
}

func hasVMwareLocks(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".lck") {
			return true
		}
	}
	return false
}

func isVMwareProcessRunning(vmxPath string) bool {
	vmxName := filepath.Base(vmxPath)
	if runtime.GOOS != "windows" {
		cmd := exec.Command("pgrep", "-f", vmxName)
		if err := cmd.Run(); err == nil {
			return true
		}
		return false
	}
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq vmware-vmx.exe")
	out, err := cmd.Output()
	return err == nil && strings.Contains(string(out), "vmware-vmx.exe")
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
					valStr := rest[startInt+len("<integer>") : endInt]
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
