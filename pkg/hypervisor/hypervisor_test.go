// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"os"
	"testing"
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

	var foundQemu, foundHyperV, foundParallels, foundKVM bool
	for _, st := range statuses {
		if st.Type == TypeQEMU {
			foundQemu = true
		}
		if st.Type == TypeHyperV {
			foundHyperV = true
		}
		if st.Type == TypeParallels {
			foundParallels = true
		}
		if st.Type == TypeKVM {
			foundKVM = true
		}
	}
	if !foundQemu || !foundHyperV || !foundParallels || !foundKVM {
		t.Errorf("expected QEMU, Hyper-V, Parallels, and KVM driver status in DetectAll results")
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
	err := mgr.LaunchBest(nil, "dummy_disk", BootModeAuto)
	if err != nil {
		t.Fatalf("unexpected error during dry-run launch: %v", err)
	}
}
