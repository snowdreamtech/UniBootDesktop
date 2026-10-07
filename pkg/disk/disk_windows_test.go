//go:build windows

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePnpStorageID_Windows(t *testing.T) {
	ven, prod, rev, inst := parsePnpStorageID(`USBSTOR\DISK&VEN_SANDISK&PROD_ULTRA&REV_1.00\0123456789ABCDEF&0`)
	assert.Equal(t, "SANDISK", ven)
	assert.Equal(t, "ULTRA", prod)
	assert.Equal(t, "1.00", rev)
	assert.Equal(t, "0123456789ABCDEF", inst)

	ven, prod, rev, inst = parsePnpStorageID(`USBSTOR\Disk&Ven_VendorCo&Prod_ProductCode&Rev_2.00\6275981091237119587&0`)
	assert.Equal(t, "VendorCo", ven)
	assert.Equal(t, "ProductCode", prod)
	assert.Equal(t, "2.00", rev)
	assert.Equal(t, "6275981091237119587", inst)
}

func TestParseUsbRevision_Windows(t *testing.T) {
	v, s := parseUsbRevision("REV_0200")
	assert.Equal(t, "USB 2.0", v)
	assert.Equal(t, "480 Mb/s", s)

	v, s = parseUsbRevision("REV_0300")
	assert.Equal(t, "USB 3.0", v)
	assert.Equal(t, "5 Gb/s", s)

	v, s = parseUsbRevision("REV_0310")
	assert.Equal(t, "USB 3.1", v)
	assert.Equal(t, "10 Gb/s", s)

	v, s = parseUsbRevision("REV_0320")
	assert.Equal(t, "USB 3.2", v)
	assert.Equal(t, "20 Gb/s", s)

	v, s = parseUsbRevision("REV_0400")
	assert.Equal(t, "USB4", v)
	assert.Equal(t, "40 Gb/s", s)
}

func TestGetUsbDeviceInfoWindows(t *testing.T) {
	info := getUsbDeviceInfoWindows(`USBSTOR\Disk&Ven_VendorCo&Prod_ProductCode&Rev_2.00\6275981091237119587&0`, "")
	assert.NotNil(t, info)
	assert.Equal(t, "VendorCo", info.Vendor)
	assert.Equal(t, "ProductCode", info.Product)
	assert.Equal(t, "6275981091237119587", info.SerialNumber)
	assert.Equal(t, "0x346d", info.VendorID)
	assert.Equal(t, "0x5678", info.ProductID)
	assert.Equal(t, "USB 2.0", info.UsbVersion)
	assert.Equal(t, "480 Mb/s", info.UsbSpeed)
	assert.Equal(t, "BOT (Bulk-Only Transport)", info.TransportProtocol)
}

func TestHasConnectedUSBStorageWindows(t *testing.T) {
	_ = hasConnectedUSBStorageWindows()
}

func TestGetVolumeSnapshotWindows(t *testing.T) {
	snap := getVolumeSnapshotWindows()
	assert.NotEmpty(t, snap)
	assert.Contains(t, snap, "mask:")
	assert.Contains(t, snap, "usb:")
	assert.Contains(t, snap, "phys:")
}

func TestGetWindowsDisksNative(t *testing.T) {
	disks, err := getWindowsDisksNative()
	assert.NoError(t, err)
	_ = disks
}

func TestInvalidateWindowsUSBCache(t *testing.T) {
	InvalidateWindowsUSBCache()
	assert.Empty(t, winUSBCacheMap)
}

func TestInspectWindowsDisk_UsbVersionPrecedence(t *testing.T) {
	drive := winDiskDrive{
		DeviceID:       `\\.\PhysicalDrive9`,
		Index:          9,
		Model:          "SanDisk Ultra 3.0 Flash Drive",
		Caption:        "SanDisk Ultra 3.0 Flash Drive",
		Size:           16000000000,
		BytesPerSector: 512,
		UsbVersion:     "USB 2.0",
		UsbSpeed:       "480 Mb/s",
	}
	info := inspectWindowsDisk(9, drive)
	assert.NotNil(t, info)
	assert.Equal(t, "USB 2.0", info.UsbVersion)
	assert.Equal(t, "480 Mb/s", info.UsbSpeed)
}
