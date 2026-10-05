// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build darwin || linux

package disk

import (
	"fmt"
	"syscall"
)

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

// winVolumeInfo contains volume metadata discovered for a mounted drive letter on Windows.
type winVolumeInfo struct {
	DriveLetter string
	VolumeName  string
	FileSystem  string
	FreeSpace   uint64
	TotalSize   uint64
	DriveType   uint32
}

// getDriveLetterDiskNumberWindows is a stub on Unix.
func getDriveLetterDiskNumberWindows(driveLetter string) (int, error) {
	return -1, fmt.Errorf("not supported on unix")
}

// getSystemDriveDiskNumberWindows is a stub on Unix.
func getSystemDriveDiskNumberWindows() (int, error) {
	return -1, fmt.Errorf("not supported on unix")
}

// getVolumesForDiskWindows is a stub on Unix.
func getVolumesForDiskWindows(diskNumber int) ([]winVolumeInfo, error) {
	return nil, nil
}

// winUsbDeviceInfo contains detailed hardware attributes for a USB storage device on Windows.
type winUsbDeviceInfo struct {
	VendorID          string
	ProductID         string
	SerialNumber      string
	UsbVersion        string
	UsbSpeed          string
	TransportProtocol string
	Vendor            string
	Product           string
	BusPower          string
	BusPowerUsed      string
}

// getUsbDeviceInfoWindows is a stub on Unix.
func getUsbDeviceInfoWindows(pnpDeviceID string, serialNumber string) *winUsbDeviceInfo {
	return nil
}
