// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !darwin && !linux && !windows

package disk

func syncPlatformBuffers() {}

func getMountFreeSpace(mountPath string) uint64 {
	return 0
}
