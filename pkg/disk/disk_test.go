// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{"VTOYEFI upper", "VTOYEFI", true},
		{"vtoyefi lower", "vtoyefi", true},
		{"VTOYEFI with prefix", "VTOYEFI_BOOT", true},
		{"UNIBOOTEFI upper", "UNIBOOTEFI", true},
		{"unibootefi lower", "unibootefi", true},
		{"UNIBOOTEFI with prefix", "UNIBOOTEFI_BOOT", true},
		{"EFI_BOOT partition", "EFI_BOOT", true},
		{"System Reserved", "System Reserved", true},
		{"WinRE partition", "WinRE", true},
		{"OEM partition", "OEM", true},
		{"Normal Ventoy volume", "Ventoy", false},
		{"Normal UniBoot volume", "UniBoot", false},
		{"Normal USB volume", "MyUSBKey", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIgnoredVolume(tt.volName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTargetDisk(t *testing.T) {
	assert.Error(t, ValidateTargetDisk(""))
	assert.Error(t, ValidateTargetDisk("/"))
	assert.Error(t, ValidateTargetDisk("C:"))
	assert.NoError(t, ValidateTargetDisk("/Volumes/MyUSBKey"))
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

func TestGetRemovableDisks(t *testing.T) {
	disks, err := GetRemovableDisks()
	assert.NoError(t, err)
	assert.NotNil(t, disks)
}

