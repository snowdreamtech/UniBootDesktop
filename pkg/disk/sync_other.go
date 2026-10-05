// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !darwin && !linux && !windows

package disk

import "fmt"

func syncPlatformBuffers() {}

func getMountFreeSpace(mountPath string) uint64 {
	return 0
}

// getSystemDriveDiskNumberWindows is a stub on other platforms.
func getSystemDriveDiskNumberWindows() (int, error) {
	return -1, fmt.Errorf("not supported on this platform")
}
