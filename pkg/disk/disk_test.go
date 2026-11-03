// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"io/fs"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeDirEntry struct {
	name string
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return false }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func TestIsIgnoredVolume(t *testing.T) {
	tests := []struct {
		name     string
		volName  string
		expected bool
	}{
		{"Macintosh HD system disk", "Macintosh HD", true},
		{"Macintosh HD Data volume", "Macintosh HD - Data", true},
		{"System volume", "System", true},
		{"Recovery volume", "Recovery", true},
		{"Preboot volume", "Preboot", true},
		{"VM swap volume", "VM", true},
		{"EFI partition", "EFI", true},
		{"ESP partition", "ESP", true},
		{"EFI_BOOT partition", "EFI_BOOT", true},
		{"System Reserved", "System Reserved", true},
		{"WinRE partition", "WinRE", true},
		{"OEM partition", "OEM", true},
		{"Xcode DMG volume", "Xcode", true},
		{"Installer volume", "Installer", true},
		{"Time Machine backup", "Time Machine Backups", true},
		{"Normal USB volume", "MyUSBKey", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIgnoredVolume(tt.volName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeDarwinDiskNode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/dev/disk2", "disk2"},
		{"/dev/rdisk2", "disk2"},
		{"/dev/disk2s1", "disk2"},
		{"disk2s2", "disk2"},
		{"/dev/disk12s3", "disk12"},
		{"disk0", "disk0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeDarwinDiskNode(tt.input))
		})
	}
}

func TestValidateTargetDisk(t *testing.T) {
	assert.Error(t, ValidateTargetDisk(""))
	assert.Error(t, ValidateTargetDisk("/"))
	assert.Error(t, ValidateTargetDisk("C:"))

	var validTarget string
	switch runtime.GOOS {
	case "darwin":
		validTarget = "/Volumes/MyUSBKey"
	case "linux":
		validTarget = "/dev/sdz"
	case "windows":
		validTarget = `\\.\PhysicalDrive1`
	default:
		validTarget = "dummy_usb"
	}
	assert.NoError(t, ValidateTargetDisk(validTarget))
}

func TestParseWindowsDiskNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "physicaldrive prefix", input: `\\.\PhysicalDrive3`, want: 3},
		{name: "disk prefix", input: "disk7", want: 7},
		{name: "numeric disk id", input: "12", want: 12},
		{name: "unsafe label", input: "MyUSBKey", wantErr: true},
		{name: "system path", input: "C:", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWindowsDiskNumber(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateUserEjectTarget(t *testing.T) {
	assert.Error(t, ValidateUserEjectTarget(""))
	assert.Error(t, ValidateUserEjectTarget("/"))
	assert.Error(t, ValidateUserEjectTarget("C:"))
	assert.Error(t, ValidateUserEjectTarget("/dev/sda"))
}

func TestValidateTargetDiskSnapshot(t *testing.T) {
	var validDev, diffDev string
	switch runtime.GOOS {
	case "darwin":
		validDev = "/dev/disk4"
		diffDev = "/dev/disk5"
	case "linux":
		validDev = "/dev/sdz"
		diffDev = "/dev/sdy"
	case "windows":
		validDev = `\\.\PhysicalDrive1`
		diffDev = `\\.\PhysicalDrive2`
	default:
		validDev = "dummy_usb"
		diffDev = "dummy_usb2"
	}

	expected := DiskInfo{
		Device:       validDev,
		Size:         128000000000,
		IsRemovable:  true,
		IsSystem:     false,
		SerialNumber: "SN-123",
		Vendor:       "Example Vendor",
	}

	tests := []struct {
		name    string
		actual  DiskInfo
		wantErr bool
	}{
		{
			name:   "same device identity",
			actual: expected,
		},
		{
			name:    "device path changed",
			actual:  DiskInfo{Device: diffDev, Size: expected.Size, IsRemovable: true, SerialNumber: expected.SerialNumber, Vendor: expected.Vendor},
			wantErr: true,
		},
		{
			name:    "capacity changed",
			actual:  DiskInfo{Device: expected.Device, Size: 64000000000, IsRemovable: true, SerialNumber: expected.SerialNumber, Vendor: expected.Vendor},
			wantErr: true,
		},
		{
			name:    "serial number changed",
			actual:  DiskInfo{Device: expected.Device, Size: expected.Size, IsRemovable: true, SerialNumber: "SN-456", Vendor: expected.Vendor},
			wantErr: true,
		},
		{
			name:    "system disk detected",
			actual:  DiskInfo{Device: expected.Device, Size: expected.Size, IsRemovable: true, IsSystem: true, SerialNumber: expected.SerialNumber, Vendor: expected.Vendor},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTargetDiskSnapshot(expected, tt.actual)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTargetDiskSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckFakeUsb3(t *testing.T) {
	tests := []struct {
		name       string
		diskName   string
		usbVersion string
		usbSpeed   string
		expected   bool
	}{
		{"Fake USB 3.0 with 480 Mbps PHY", "SanDisk Ultra USB 3.0", "USB 2.0", "480 Mb/s", true},
		{"Genuine USB 3.0 with 5 Gbps PHY", "Kingston DataTraveler 3.0", "USB 3.0", "5 Gb/s", false},
		{"Fake SuperSpeed drive", "SuperSpeed USB3 Drive", "2.00", "Up to 480 Mb/s", true},
		{"Genuine USB 3.1 Gen 2", "Samsung SSD 3.1", "USB 3.1", "10 Gb/s", false},
		{"Normal USB 2.0 drive", "Old Flash Drive", "USB 2.0", "480 Mb/s", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckFakeUsb3(tt.diskName, tt.usbVersion, tt.usbSpeed)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapProtocolCode(t *testing.T) {
	assert.Equal(t, "usb4", MapProtocolCode("USB4", "40 Gb/s"))
	assert.Equal(t, "usb3_2", MapProtocolCode("USB 3.2", "10 Gb/s"))
	assert.Equal(t, "usb3_1", MapProtocolCode("USB 3.1", "10 Gb/s"))
	assert.Equal(t, "usb3_0", MapProtocolCode("USB 3.0", "5 Gb/s"))
	assert.Equal(t, "usb2", MapProtocolCode("USB 2.0", "480 Mb/s"))
}

func TestFormatBytes(t *testing.T) {
	assert.Equal(t, "500 B", FormatBytes(500))
	assert.Equal(t, "1.00 KB", FormatBytes(1024))
	assert.Equal(t, "7.50 GB", FormatBytes(8053063680))
	assert.Equal(t, "231.10 GB", FormatBytes(248145510400))
	assert.Equal(t, "1.50 TB", FormatBytes(1649267441664))
}

func TestFormatBytesDual(t *testing.T) {
	assert.Equal(t, "29.80 GB (Nominal 32 GB)", FormatBytesDual(32000000000))
}

func TestDarwinBatchDisks(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin only")
	}
	usbMap := getCachedDarwinUSBMap()
	disks, err := getDarwinDisksBatch(usbMap)
	assert.NoError(t, err)
	assert.NotNil(t, disks)
}

func TestGetRemovableDisks(t *testing.T) {
	disks, err := GetRemovableDisks()
	assert.NoError(t, err)
	assert.NotNil(t, disks)
}

func TestGetRemovableDisksCaching(t *testing.T) {
	// First call warms up cache
	_, _ = GetRemovableDisks()

	// Second call must hit cache instantaneously (< 10ms)
	start := time.Now()
	disks, err := GetRemovableDisks()
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, disks)
	assert.Less(t, elapsed, 10*time.Millisecond, "Cached GetRemovableDisks took too long: %v", elapsed)
}

func TestDiskCacheShouldReuse(t *testing.T) {
	diskCacheSnapshot = "A|B"
	assert.True(t, diskCacheShouldReuse("A|B", time.Now().Add(-2*time.Second)))
	assert.True(t, diskCacheShouldReuse("A|B", time.Now().Add(-4*time.Second)))
	assert.False(t, diskCacheShouldReuse("A|B|C", time.Now().Add(-2*time.Second)))
	assert.False(t, diskCacheShouldReuse("A|B", time.Now().Add(-10*time.Second)))
	diskCacheSnapshot = ""
}

func TestDarwinVolumeSnapshotUsesDeviceIdentity(t *testing.T) {
	infoMap := map[string]string{
		"/Volumes/UNTITLED":   "<key>DeviceIdentifier</key><string>disk2</string>",
		"/Volumes/UNTITLED 1": "<key>DeviceIdentifier</key><string>disk3</string>",
	}

	entries := []os.DirEntry{
		fakeDirEntry{name: "UNTITLED"},
		fakeDirEntry{name: "UNTITLED 1"},
	}

	snapshot := buildDarwinVolumeSnapshot(entries, infoMap)
	assert.Contains(t, snapshot, "disk2")
	assert.Contains(t, snapshot, "disk3")
	assert.NotEqual(t, "UNTITLED|UNTITLED 1", snapshot)
}
