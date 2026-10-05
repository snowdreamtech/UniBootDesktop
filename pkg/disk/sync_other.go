// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !darwin && !linux && !windows

package disk

import "fmt"

func syncPlatformBuffers() {}

func getMountFreeSpace(mountPath string) uint64 {
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

// getDriveLetterDiskNumberWindows is a stub on other platforms.
func getDriveLetterDiskNumberWindows(driveLetter string) (int, error) {
	return -1, fmt.Errorf("not supported on this platform")
}

// getSystemDriveDiskNumberWindows is a stub on other platforms.
func getSystemDriveDiskNumberWindows() (int, error) {
	return -1, fmt.Errorf("not supported on this platform")
}

// getVolumesForDiskWindows is a stub on other platforms.
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

// getUsbDeviceInfoWindows is a stub on other platforms.
func getUsbDeviceInfoWindows(pnpDeviceID string, serialNumber string) *winUsbDeviceInfo {
	return nil
}

// winPhysicalDiskInfo represents hardware attributes returned directly from native Win32 physical disk IOCTLs.
type winPhysicalDiskInfo struct {
	DiskNumber     int
	DevicePath     string
	Size           uint64
	BytesPerSector uint32
	BusType        uint32
	IsRemovable    bool
	Vendor         string
	Product        string
	Revision       string
	SerialNumber   string
	UsbVersion     string
	UsbSpeed       string
}

// queryPhysicalDiskWindows is a stub on other platforms.
func queryPhysicalDiskWindows(diskNum int) (*winPhysicalDiskInfo, error) {
	return nil, fmt.Errorf("not supported on this platform")
}

// InvalidateWindowsUSBCache is a stub on other platforms.
func InvalidateWindowsUSBCache() {}

// hasConnectedUSBStorageWindows is a stub on other platforms.
func hasConnectedUSBStorageWindows() bool {
	return false
}

// getVolumeSnapshotWindows is a stub on other platforms.
func getVolumeSnapshotWindows() string {
	return ""
}

// getWindowsDisksNative is a stub on other platforms.
func getWindowsDisksNative() ([]DiskInfo, error) {
	return nil, fmt.Errorf("not supported on this platform")
}
