// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

type UTMDriver struct{}

func (d *UTMDriver) Type() HypervisorType {
	return TypeUTM
}

func (d *UTMDriver) Name() string {
	return "UTM Virtual Machine"
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

	// 1. Check utmctl command
	if path, err := exec.LookPath("utmctl"); err == nil {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM Virtual Machine",
			Installed:  true,
			Path:       path,
			Version:    "UTM CLI (utmctl)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Check /Applications/UTM.app bundle
	appPath := "/Applications/UTM.app"
	if info, err := os.Stat(appPath); err == nil && info.IsDir() {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM Virtual Machine",
			Installed:  true,
			Path:       appPath,
			Version:    "UTM App (/Applications/UTM.app)",
			Priority:   d.Priority(),
			CanBootRaw: true,
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
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run UTM launch complete", "diskPath", diskPath, "bootMode", bootMode)
		return nil
	}

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	ensureDiskPermissions(targetPath)
	unmountTargetDisk(targetPath)

	// If system has qemu-system-x86_64 / qemu-system-aarch64 installed, leverage QEMU backend directly
	qemuDrv := &QEMUDriver{}
	if qemuStatus := qemuDrv.Detect(); qemuStatus.Installed {
		logger.Info("UTM selected, delegating execution to embedded/system QEMU engine", "disk", targetPath, "bootMode", bootMode)
		return qemuDrv.Launch(ctx, diskPath, bootMode)
	}

	// Fallback: Generate native .utm bundle with raw disk mapping and launch via UTM app
	tmpDir := "/tmp/uniboot_utm"
	_ = os.RemoveAll(tmpDir)
	_ = os.MkdirAll(tmpDir, 0755)

	utmBundle := "/tmp/uniboot_utm/UniBootPreview.utm"
	_ = os.MkdirAll(utmBundle, 0755)

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
		<string>UniBoot Preview</string>
	</dict>
	<key>System</key>
	<dict>
		<key>Architecture</key>
		<string>x86_64</string>
		<key>CPUCount</key>
		<integer>2</integer>
		<key>MemorySize</key>
		<integer>2048</integer>
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
`, targetPath)

	_ = os.WriteFile(utmBundle+"/config.plist", []byte(plistContent), 0644)

	logger.Info("Opening UTM application with native raw disk bundle", "bundle", utmBundle, "targetPath", targetPath)
	cmd := exec.Command("open", "-W", "-a", "UTM", utmBundle)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open UTM application: %w", err)
	}

	go func() {
		_ = cmd.Wait()
		remountTargetDisk(targetPath)
	}()

	return nil
}
