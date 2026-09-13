// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package qemu

import (
	"context"
	"testing"
)

func TestQEMUDetect_DryRun(t *testing.T) {
	t.Setenv("UNIBOOT_DRY_RUN", "true")

	status := Detect()
	if status == nil {
		t.Fatalf("expected non-nil QEMUStatus")
	}
	if !status.Installed {
		t.Errorf("expected Installed to be true under dry-run")
	}
}

func TestLaunchTest_EmptyDisk(t *testing.T) {
	ctx := context.Background()
	err := LaunchTest(ctx, "")
	if err == nil {
		t.Errorf("expected error for empty diskPath")
	}
}

func TestLaunchTest_DryRun(t *testing.T) {
	t.Setenv("UNIBOOT_DRY_RUN", "true")
	ctx := context.Background()

	err := LaunchTest(ctx, "testdisk1")
	if err != nil {
		t.Errorf("expected nil error under dry-run, got: %v", err)
	}
}

func TestResolveRawDiskDevice(t *testing.T) {
	node := ResolveRawDiskDevice("/dev/disk2s1")
	if node != "/dev/rdisk2" {
		t.Errorf("expected /dev/rdisk2 for /dev/disk2s1, got %s", node)
	}
}
