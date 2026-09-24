// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package disk

import (
	"fmt"
	"syscall"
)

// syncPlatformBuffers flushes file buffers across all accessible Windows volumes,
// forcing pending filesystem metadata and write-back caches to physical storage.
func syncPlatformBuffers() {
	for c := 'A'; c <= 'Z'; c++ {
		volPath := fmt.Sprintf(`\\.\%c:`, c)
		ptr, err := syscall.UTF16PtrFromString(volPath)
		if err != nil {
			continue
		}
		handle, err := syscall.CreateFile(
			ptr,
			syscall.GENERIC_READ|syscall.GENERIC_WRITE,
			syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
			nil,
			syscall.OPEN_EXISTING,
			0,
			0,
		)
		if err == nil {
			_ = syscall.FlushFileBuffers(handle)
			_ = syscall.CloseHandle(handle)
		}
	}
}

// getMountFreeSpace on Windows build is a stub for Darwin inspection routines.
func getMountFreeSpace(mountPath string) uint64 {
	return 0
}
