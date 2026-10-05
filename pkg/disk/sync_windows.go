// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package disk

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
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

// getSystemDriveDiskNumberWindows returns the physical disk index of the Windows SystemDrive (e.g. C:).
func getSystemDriveDiskNumberWindows() (int, error) {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	volPath := fmt.Sprintf(`\\.\%s`, strings.TrimSuffix(sysDrive, `\`))
	ptr, err := syscall.UTF16PtrFromString(volPath)
	if err != nil {
		return -1, err
	}
	handle, err := syscall.CreateFile(
		ptr,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return -1, err
	}
	defer syscall.CloseHandle(handle)

	// IOCTL_STORAGE_GET_DEVICE_NUMBER = 0x002D1080
	const ioctlStorageGetDeviceNumber = 0x002D1080
	type storageDeviceNumber struct {
		DeviceType      uint32
		DeviceNumber    uint32
		PartitionNumber uint32
	}
	var sdn storageDeviceNumber
	var bytesReturned uint32
	err = syscall.DeviceIoControl(
		handle,
		ioctlStorageGetDeviceNumber,
		nil,
		0,
		(*byte)(unsafe.Pointer(&sdn)),
		uint32(unsafe.Sizeof(sdn)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return -1, err
	}
	return int(sdn.DeviceNumber), nil
}
