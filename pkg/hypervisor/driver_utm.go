// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/snowdreamtech/unibootdesktop/internal/logger"
)

type UTMDriver struct{}

func (d *UTMDriver) Type() HypervisorType {
	return TypeUTM
}

func (d *UTMDriver) Name() string {
	return "UTM"
}

func (d *UTMDriver) Priority() int {
	return 2
}

func (d *UTMDriver) Detect() *VMStatus {
	if runtime.GOOS != "darwin" {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM",
			Installed:  false,
			Path:       "",
			Version:    "Not Supported on this OS",
			Priority:   d.Priority(),
			CanBootRaw: false,
		}
	}

	qemuInstalled := (&QEMUDriver{}).Detect().Installed

	// 1. Check utmctl command
	if path, err := exec.LookPath("utmctl"); err == nil {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM",
			Installed:  true,
			Path:       path,
			Version:    "UTM CLI (utmctl)",
			Priority:   d.Priority(),
			CanBootRaw: qemuInstalled,
		}
	}

	// 2. Check /Applications/UTM.app bundle
	appPath := "/Applications/UTM.app"
	if info, err := os.Stat(appPath); err == nil && info.IsDir() {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM",
			Installed:  true,
			Path:       appPath,
			Version:    "UTM App (/Applications/UTM.app)",
			Priority:   d.Priority(),
			CanBootRaw: qemuInstalled,
		}
	}

	return &VMStatus{
		Type:       TypeUTM,
		Name:       "UTM",
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *UTMDriver) Launch(ctx context.Context, diskPath string, bootMode string) error {
	return d.LaunchWithConfig(ctx, diskPath, VMConfig{
		BootMode:     bootMode,
		CpuCores:     GetRecommendedVCPUs(),
		MemoryMB:     GetRecommendedVMMemoryMB(),
		DisplayAccel: true,
	})
}

func (d *UTMDriver) LaunchWithConfig(ctx context.Context, diskPath string, cfg VMConfig) error {
	status := d.Detect()
	if !status.Installed && os.Getenv("UNIBOOT_DRY_RUN") == "" {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run UTM launch complete", "diskPath", diskPath, "bootMode", cfg.BootMode)
		return nil
	}

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}

	// macOS App Sandbox (com.utmapp.UTM) strictly forbids UTM.app from reading raw host block devices (/dev/rdiskN).
	// If system QEMU is installed, delegate physical disk preview testing to host QEMU engine IMMEDIATELY
	// before any disk permissions are modified or any 2-second restore timers are scheduled.
	if strings.HasPrefix(targetPath, "/dev/") {
		qemuDrv := &QEMUDriver{}
		if qemuStatus := qemuDrv.Detect(); qemuStatus.Installed {
			logger.Info("UTM.app is sandboxed on macOS and cannot access raw block devices directly; delegating physical disk preview test to host QEMU engine", "disk", targetPath, "bootMode", cfg.BootMode)
			return qemuDrv.LaunchWithConfig(ctx, diskPath, cfg)
		}
		return fmt.Errorf("UTM on macOS is sandboxed and cannot access raw physical disks (%s). Please install QEMU via 'brew install qemu' to enable raw USB emulation", targetPath)
	}

	restoreDiskPerms, err := ensureDiskPermissions(targetPath)
	if err != nil {
		remountTargetDisk(targetPath)
		return fmt.Errorf("failed to acquire target disk permissions: %w", err)
	}
	unmountTargetDisk(targetPath)

	// Generate native .utm bundle with raw disk mapping and launch via UTM app
	tmpDir := "/tmp/uniboot_utm"
	_ = os.RemoveAll(tmpDir)
	_ = os.MkdirAll(tmpDir, 0755)

	utmBundle := "/tmp/uniboot_utm/UniBoot.utm"
	_ = os.MkdirAll(utmBundle, 0755)

	memMB := cfg.MemoryMB
	if memMB <= 0 {
		memMB = GetRecommendedVMMemoryMB()
	}
	vcpus := cfg.CpuCores
	if vcpus <= 0 {
		vcpus = GetRecommendedVCPUs()
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>ConfigurationVersion</key>
	<integer>4</integer>
	<key>Information</key>
	<dict>
		<key>Icon</key>
		<string>disk</string>
		<key>Name</key>
		<string>UniBoot</string>
	</dict>
	<key>System</key>
	<dict>
		<key>Architecture</key>
		<string>x86_64</string>
		<key>CPUCount</key>
		<integer>%d</integer>
		<key>MemorySize</key>
		<integer>%d</integer>
		<key>Target</key>
		<string>q35</string>
	</dict>
	<key>Drives</key>
	<array>
		<dict>
			<key>DriveType</key>
			<string>Disk</string>
			<key>Interface</key>
			<string>USB</string>
			<key>ImagePath</key>
			<string>%s</string>
		</dict>
	</array>
</dict>
</plist>
`, vcpus, memMB, targetPath)

	_ = os.WriteFile(utmBundle+"/config.plist", []byte(plistContent), 0600)

	logger.Info("Opening UTM application with native raw disk bundle", "bundle", utmBundle, "targetPath", targetPath)
	cmd := exec.Command("open", "-a", "UTM", utmBundle)
	if err := cmd.Run(); err != nil {
		remountTargetDisk(targetPath)
		restoreDiskPerms()
		return fmt.Errorf("failed to open UTM application: %w", err)
	}

	time.Sleep(1 * time.Second)
	utmctlPath := "/Applications/UTM.app/Contents/MacOS/utmctl"
	vmUUID := getLatestUTMUUID()

	if vmUUID != "" {
		if _, err := os.Stat(utmctlPath); err == nil {
			logger.Info("Triggering auto-start for UTM VM", "uuid", vmUUID)
			_ = exec.Command(utmctlPath, "start", vmUUID).Run()
		}
	}

	go func() {
		ticker := time.NewTicker(1500 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(1 * time.Hour)
		started := false

		for {
			select {
			case <-timeout:
				time.Sleep(300 * time.Millisecond)
				remountTargetDisk(targetPath)
				restoreDiskPerms()
				NotifyVMExited(targetPath, nil)
				return
			case <-ticker.C:
				// Check if UTM process itself was closed
				utmCheck := exec.Command("pgrep", "-x", "UTM")
				if err := utmCheck.Run(); err != nil {
					// UTM app process has exited!
					logger.Info("UTM process terminated, auto-remounting target disk", "targetPath", targetPath)
					time.Sleep(300 * time.Millisecond)
					remountTargetDisk(targetPath)
					restoreDiskPerms()
					NotifyVMExited(targetPath, nil)
					return
				}

				if vmUUID != "" {
					if _, err := os.Stat(utmctlPath); err == nil {
						out, err := exec.Command(utmctlPath, "status", vmUUID).Output()
						statusStr := strings.ToLower(string(out))
						if err == nil && (strings.Contains(statusStr, "started") || strings.Contains(statusStr, "running")) {
							started = true
						} else if started && (!strings.Contains(statusStr, "started") || err != nil) {
							// VM was running and has now stopped
							logger.Info("UTM VM stopped, auto-remounting target disk", "targetPath", targetPath)
							time.Sleep(300 * time.Millisecond)
							remountTargetDisk(targetPath)
							restoreDiskPerms()
							NotifyVMExited(targetPath, nil)
							return
						}
					}
				}
			}
		}
	}()

	return nil
}

func getLatestUTMUUID() string {
	utmctlPath := "/Applications/UTM.app/Contents/MacOS/utmctl"
	out, err := exec.Command(utmctlPath, "list").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	var lastUUID string
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && len(fields[0]) == 36 && strings.Count(fields[0], "-") == 4 {
			lastUUID = fields[0]
		}
	}
	return lastUUID
}
