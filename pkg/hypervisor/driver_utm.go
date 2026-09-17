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

	"github.com/snowdreamtech/unigodesktop/internal/logger"
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

	// 1. Check utmctl command
	if path, err := exec.LookPath("utmctl"); err == nil {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM",
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
			Name:       "UTM",
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

	// Generate native .utm bundle with raw disk mapping and launch via UTM app
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
	cmd := exec.Command("open", "-a", "UTM", utmBundle)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open UTM application: %w", err)
	}

	// Trigger UTM URL Scheme and utmctl to automatically boot the VM without requiring manual Play click
	time.Sleep(500 * time.Millisecond)
	utmctlPath := "/Applications/UTM.app/Contents/MacOS/utmctl"
	if _, err := os.Stat(utmctlPath); err == nil {
		_ = exec.Command(utmctlPath, "start", "UniBoot Preview").Run()
	}
	_ = exec.Command("open", "utm://run?name=UniBoot%20Preview").Run()

	go func() {
		utmctlPath := "/Applications/UTM.app/Contents/MacOS/utmctl"
		ticker := time.NewTicker(1500 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(1 * time.Hour)
		started := false

		for {
			select {
			case <-timeout:
				remountTargetDisk(targetPath)
				return
			case <-ticker.C:
				if _, err := os.Stat(utmctlPath); err == nil {
					out, err := exec.Command(utmctlPath, "status", "UniBoot Preview").Output()
					statusStr := strings.ToLower(string(out))
					if err == nil && (strings.Contains(statusStr, "started") || strings.Contains(statusStr, "running")) {
						started = true
					} else if started && (!strings.Contains(statusStr, "started") || err != nil) {
						// VM was running and has now stopped or UTM exited!
						remountTargetDisk(targetPath)
						return
					}
				}
			}
		}
	}()

	return nil
}
