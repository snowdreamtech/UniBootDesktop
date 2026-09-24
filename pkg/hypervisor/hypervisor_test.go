// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestHypervisorManager_DetectAll(t *testing.T) {
	mgr := GetManager()
	if mgr == nil {
		t.Fatalf("expected non-nil Hypervisor Manager singleton")
	}

	statuses := mgr.DetectAll()
	if len(statuses) < 7 {
		t.Fatalf("expected at least 7 registered drivers, got %d", len(statuses))
	}

	var foundQemu, foundHyperV, foundParallels, foundKVM, foundVMware, foundVBox, foundUtm bool
	for _, st := range statuses {
		switch st.Type {
		case TypeQEMU:
			foundQemu = true
		case TypeHyperV:
			foundHyperV = true
		case TypeParallels:
			foundParallels = true
		case TypeKVM:
			foundKVM = true
		case TypeVMware:
			foundVMware = true
		case TypeVirtualBox:
			foundVBox = true
		case TypeUTM:
			foundUtm = true
		}
	}
	if !foundQemu || !foundHyperV || !foundParallels || !foundKVM || !foundVMware || !foundVBox || !foundUtm {
		t.Errorf("expected all 7 hypervisor drivers registered in DetectAll results")
	}
}

func TestHypervisorManager_DetectBest(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "1")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mgr := GetManager()
	best := mgr.DetectBest()
	if best == nil {
		t.Fatalf("expected non-nil VMStatus for DetectBest")
	}

	if !best.Installed {
		t.Errorf("expected Installed to be true under dry-run mode")
	}
}

func TestHypervisorManager_LaunchBest_DryRun(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "1")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mgr := GetManager()
	err := mgr.LaunchBest(context.Background(), "dummy_disk", BootModeAuto)
	if err != nil {
		t.Fatalf("unexpected error during dry-run launch: %v", err)
	}
}

func TestRunCommandWithTimeout(t *testing.T) {
	name := "sh"
	args := []string{"-c", "sleep 30"}
	if runtime.GOOS == "windows" {
		name = "powershell"
		args = []string{"-NoProfile", "-NonInteractive", "-Command", "Start-Sleep -Seconds 30"}
	}

	start := time.Now()
	err := runCommandWithTimeout(200*time.Millisecond, name, args...)
	if err == nil {
		t.Fatalf("expected command timeout error but got nil")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("command exceeded timeout budget: %v", elapsed)
	}
}

func TestRecommendedResources(t *testing.T) {
	vcpus := GetRecommendedVCPUs()
	if vcpus < 1 || vcpus > 4 {
		t.Errorf("recommended vCPUs should be between 1 and 4, got %d", vcpus)
	}

	ram := GetRecommendedVMMemoryMB()
	if ram < 2048 || ram > 8192 {
		t.Errorf("recommended RAM should be between 2048MB and 8192MB, got %d", ram)
	}
}

func TestUnmountedDisksTrackingAndCleanup(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "1")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	dummyDisk := "/dev/disk999"
	TrackDiskUnmounted(dummyDisk)

	var found bool
	unmountedDisksTracker.Range(func(key, value any) bool {
		if path, ok := key.(string); ok && path == dummyDisk {
			found = true
		}
		return true
	})
	if !found {
		t.Fatalf("expected dummyDisk to be tracked in unmountedDisksTracker")
	}

	GetManager().CleanupAllUnmountedDisks()

	foundAfter := false
	unmountedDisksTracker.Range(func(key, value any) bool {
		if path, ok := key.(string); ok && path == dummyDisk {
			foundAfter = true
		}
		return true
	})
	if foundAfter {
		t.Errorf("expected dummyDisk to be cleared after CleanupAllUnmountedDisks")
	}
}

func TestLaunchSpecified_DryRun(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "1")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	mgr := GetManager()
	drivers := []HypervisorType{TypeQEMU, TypeVMware, TypeVirtualBox}
	for _, drv := range drivers {
		err := mgr.LaunchSpecified(context.Background(), "/dev/disk_test", drv, BootModeUEFI)
		if err != nil {
			t.Errorf("unexpected error launching specified driver %s in dry-run: %v", drv, err)
		}
	}
}

func TestResolveRawDiskDevice(t *testing.T) {
	node := ResolveRawDiskDevice("/dev/disk2s1")
	if node != "/dev/rdisk2" {
		t.Errorf("expected /dev/rdisk2 for /dev/disk2s1, got %s", node)
	}
}

func TestRegisterVMExitHandlerAndNotify(t *testing.T) {
	called := false
	var receivedDisk string
	var receivedErr error

	RegisterVMExitHandler(func(targetDisk string, err error) {
		called = true
		receivedDisk = targetDisk
		receivedErr = err
	})

	testErr := fmt.Errorf("qemu exit test")
	NotifyVMExited("/dev/disk42", testErr)

	if !called {
		t.Fatalf("expected VMExitHandler to be called")
	}
	if receivedDisk != "/dev/disk42" {
		t.Errorf("expected disk /dev/disk42, got %s", receivedDisk)
	}
	if receivedErr != testErr {
		t.Errorf("expected error %v, got %v", testErr, receivedErr)
	}
}
