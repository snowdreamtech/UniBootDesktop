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

	"golang.org/x/sys/windows"
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

// winVolumeInfo contains volume metadata discovered for a mounted drive letter on Windows.
type winVolumeInfo struct {
	DriveLetter string
	VolumeName  string
	FileSystem  string
	FreeSpace   uint64
	TotalSize   uint64
	DriveType   uint32
}

// getMountFreeSpace returns available free bytes for a mounted Windows path or drive root.
func getMountFreeSpace(mountPath string) uint64 {
	trimmed := strings.TrimSpace(mountPath)
	if trimmed == "" {
		return 0
	}
	if !strings.HasSuffix(trimmed, `\`) && !strings.HasSuffix(trimmed, `/`) {
		trimmed += `\`
	}
	ptr, err := syscall.UTF16PtrFromString(trimmed)
	if err != nil {
		return 0
	}
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(
		ptr,
		&freeBytesAvailable,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return 0
	}
	return totalFreeBytes
}

// getDriveLetterDiskNumberWindows returns the physical disk index of a Windows drive letter (e.g. C: or C:\).
func getDriveLetterDiskNumberWindows(driveLetter string) (int, error) {
	driveLetter = strings.TrimSpace(driveLetter)
	driveLetter = strings.TrimSuffix(driveLetter, `\`)
	driveLetter = strings.TrimSuffix(driveLetter, `/`)
	if len(driveLetter) == 1 {
		driveLetter += ":"
	}
	if len(driveLetter) != 2 || driveLetter[1] != ':' {
		return -1, fmt.Errorf("invalid windows drive letter %q", driveLetter)
	}

	volPath := fmt.Sprintf(`\\.\%s`, driveLetter)
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

// getSystemDriveDiskNumberWindows returns the physical disk index of the Windows SystemDrive (e.g. C:).
func getSystemDriveDiskNumberWindows() (int, error) {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	return getDriveLetterDiskNumberWindows(sysDrive)
}

// getVolumesForDiskWindows scans all mounted Windows drive letters and returns those residing on diskNumber.
func getVolumesForDiskWindows(diskNumber int) ([]winVolumeInfo, error) {
	drivesMask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, err
	}

	var results []winVolumeInfo
	for c := 'A'; c <= 'Z'; c++ {
		if (drivesMask & (1 << (c - 'A'))) == 0 {
			continue
		}
		driveLetter := fmt.Sprintf("%c:", c)
		dn, err := getDriveLetterDiskNumberWindows(driveLetter)
		if err != nil || dn != diskNumber {
			continue
		}

		rootPath := fmt.Sprintf("%c:\\", c)
		rootPtr, err := syscall.UTF16PtrFromString(rootPath)
		if err != nil {
			continue
		}

		var volNameBuf [260]uint16
		var fsNameBuf [260]uint16
		var serialNum, maxLen, flags uint32
		_ = windows.GetVolumeInformation(
			rootPtr,
			&volNameBuf[0],
			uint32(len(volNameBuf)),
			&serialNum,
			&maxLen,
			&flags,
			&fsNameBuf[0],
			uint32(len(fsNameBuf)),
		)
		volName := syscall.UTF16ToString(volNameBuf[:])
		fsName := syscall.UTF16ToString(fsNameBuf[:])

		var freeBytesAvailable, totalBytes, totalFreeBytes uint64
		_ = windows.GetDiskFreeSpaceEx(
			rootPtr,
			&freeBytesAvailable,
			&totalBytes,
			&totalFreeBytes,
		)
		dt := windows.GetDriveType(rootPtr)

		results = append(results, winVolumeInfo{
			DriveLetter: driveLetter,
			VolumeName:  strings.TrimSpace(volName),
			FileSystem:  strings.TrimSpace(fsName),
			FreeSpace:   totalFreeBytes,
			TotalSize:   totalBytes,
			DriveType:   dt,
		})
	}
	return results, nil
}
