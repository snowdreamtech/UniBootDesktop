// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build darwin || linux

package disk

import "syscall"

// syncPlatformBuffers issues a kernel-level sync() on Unix systems (macOS and Linux),
// flushing all unwritten filesystem dirty pages and metadata to underlying block devices.
func syncPlatformBuffers() {
	syscall.Sync()
}

// getMountFreeSpace returns available free bytes on a mounted filesystem using statfs.
func getMountFreeSpace(mountPath string) uint64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(mountPath, &stat); err == nil {
		return stat.Bavail * uint64(stat.Bsize)
	}
	return 0
}
