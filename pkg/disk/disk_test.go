// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
		{"Xcode DMG volume", "Xcode", true},
		{"Installer volume", "Installer", true},
		{"Time Machine backup", "Time Machine Backups", true},
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

func TestIsSystemDiskWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}

	isSys, err := isSystemDiskWindows("C:")
	assert.NoError(t, err)
	assert.True(t, isSys, "C: must be detected as system disk")

	isSys, err = isSystemDiskWindows(`\\.\PhysicalDrive0`)
	assert.NoError(t, err)
	assert.True(t, isSys, `\\.\PhysicalDrive0 must be detected as system disk`)

	// Non-existent or removable drive should not be detected as system disk
	isSys, err = isSystemDiskWindows(`\\.\PhysicalDrive99`)
	assert.NoError(t, err)
	assert.False(t, isSys, `PhysicalDrive99 should not be detected as system disk`)
}

func TestGetVolumesForDiskWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}

	sysDiskNum, err := getSystemDriveDiskNumberWindows()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, sysDiskNum, 0)

	vols, err := getVolumesForDiskWindows(sysDiskNum)
	assert.NoError(t, err)
	assert.NotEmpty(t, vols, "System disk should have at least one mounted volume")

	foundC := false
	for _, v := range vols {
		if strings.EqualFold(v.DriveLetter, "C:") {
			foundC = true
			assert.NotEmpty(t, v.FileSystem, "C: volume should report a valid file system (e.g. NTFS)")
			assert.Greater(t, v.TotalSize, uint64(0), "C: total size should be > 0")
			assert.Greater(t, v.FreeSpace, uint64(0), "C: free space should be > 0")
		}
	}
	assert.True(t, foundC, "C: must be present in volumes discovered for system disk")
	assert.Greater(t, getMountFreeSpace(`C:\`), uint64(0), "C: mount point free space should be > 0")
}

func TestParsePnpStorageID_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}

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
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}

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
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}

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
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}
	_ = hasConnectedUSBStorageWindows()
}

func TestGetVolumeSnapshotWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}
	snap := getVolumeSnapshotWindows()
	assert.NotEmpty(t, snap)
	assert.Contains(t, snap, "mask:")
	assert.Contains(t, snap, "usb:")
	assert.Contains(t, snap, "phys:")
}

func TestGetWindowsDisksNative(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}
	disks, err := getWindowsDisksNative()
	assert.NoError(t, err)
	_ = disks
}

func TestInvalidateWindowsUSBCache(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows only")
	}
	InvalidateWindowsUSBCache()
	assert.Empty(t, winUSBCacheMap)
}

func TestValidateUserEjectTarget(t *testing.T) {
	assert.Error(t, ValidateUserEjectTarget(""))
	assert.Error(t, ValidateUserEjectTarget("/"))
	assert.Error(t, ValidateUserEjectTarget("C:"))
	assert.Error(t, ValidateUserEjectTarget("/dev/sda"))
}

func TestEjectDisk_Validation(t *testing.T) {
	assert.Error(t, EjectDisk(""))
	assert.Error(t, EjectDisk("C:"))
	assert.Error(t, EjectDisk(`\\.\PhysicalDrive0`))
	assert.Error(t, EjectDisk("/dev/sda"))
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

func TestIdentifyThirdPartyBoot_Fingerprints(t *testing.T) {
	tests := []struct {
		name          string
		relativeDirs  []string
		relativeFiles []string
		expected      ThirdPartyBootType
	}{
		{
			name:          "Rufus Disk with rufus.efi",
			relativeDirs:  []string{"EFI/BOOT"},
			relativeFiles: []string{"rufus.efi", "EFI/BOOT/BOOTX64.EFI"},
			expected:      BootTypeRufus,
		},
		{
			name:          "Rufus Disk with autounattend.xml bypass",
			relativeDirs:  []string{"sources"},
			relativeFiles: []string{"autounattend.xml", "autorun.ico", "sources/boot.wim"},
			expected:      BootTypeRufus,
		},
		{
			name:          "WePE Maintenance Disk with WEPE directory",
			relativeDirs:  []string{"WEPE"},
			relativeFiles: []string{"WEPE/WEPE64.WIM"},
			expected:      BootTypeWePE,
		},
		{
			name:          "EasyU Maintenance Disk with EASYU directory",
			relativeDirs:  []string{"EASYU"},
			relativeFiles: []string{"EASYU/EasyU64.wim"},
			expected:      BootTypeEasyU,
		},
		{
			name:          "EasyU Maintenance Disk with USBDATA directory",
			relativeDirs:  []string{"USBDATA"},
			relativeFiles: []string{"USBDATA/SKY.wim"},
			expected:      BootTypeEasyU,
		},
		{
			name:          "YUMI Multiboot USB",
			relativeDirs:  []string{"multiboot/menu"},
			relativeFiles: []string{"multiboot/menu/yumi.cfg"},
			expected:      BootTypeYUMI,
		},
		{
			name:          "OpenCore Hackintosh Bootloader",
			relativeDirs:  []string{"EFI/OC"},
			relativeFiles: []string{"EFI/OC/OpenCore.efi", "EFI/OC/config.plist"},
			expected:      BootTypeOpenCore,
		},
		{
			name:          "Clover Hackintosh Bootloader",
			relativeDirs:  []string{"EFI/CLOVER"},
			relativeFiles: []string{"EFI/CLOVER/CloverX64.efi"},
			expected:      BootTypeClover,
		},
		{
			name:          "Generic WinPE Maintenance Disk with winpe.ini",
			relativeDirs:  []string{"sources"},
			relativeFiles: []string{"winpe.ini", "sources/boot.wim"},
			expected:      BootTypeWinPE,
		},
		{
			name:          "Windows Official Installer with install.wim",
			relativeDirs:  []string{"sources"},
			relativeFiles: []string{"sources/boot.wim", "sources/install.wim"},
			expected:      BootTypeWindowsInstaller,
		},
		{
			name:          "Windows Official Installer with install.esd",
			relativeDirs:  []string{"sources"},
			relativeFiles: []string{"sources/install.esd"},
			expected:      BootTypeWindowsInstaller,
		},
		{
			name:          "Linux Live USB (Ubuntu casper)",
			relativeDirs:  []string{"casper", "boot/grub"},
			relativeFiles: []string{"boot/grub/grub.cfg"},
			expected:      BootTypeLinuxLive,
		},
		{
			name:          "Linux Live USB (Fedora LiveOS)",
			relativeDirs:  []string{"LiveOS"},
			relativeFiles: []string{"LiveOS/squashfs.img"},
			expected:      BootTypeLinuxLive,
		},
		{
			name:          "Generic UEFI Fallback USB",
			relativeDirs:  []string{"EFI/BOOT"},
			relativeFiles: []string{"EFI/BOOT/BOOTX64.EFI"},
			expected:      BootTypeGenericUEFI,
		},
		{
			name:          "Plain Data USB (no boot files)",
			relativeDirs:  []string{"Documents", "Photos"},
			relativeFiles: []string{"Documents/resume.pdf", "Photos/trip.jpg"},
			expected:      BootTypeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, d := range tt.relativeDirs {
				err := os.MkdirAll(filepath.Join(tmpDir, d), 0755)
				assert.NoError(t, err)
			}
			for _, f := range tt.relativeFiles {
				filePath := filepath.Join(tmpDir, f)
				err := os.MkdirAll(filepath.Dir(filePath), 0755)
				assert.NoError(t, err)
				err = os.WriteFile(filePath, []byte("dummy binary"), 0644)
				assert.NoError(t, err)
			}

			detected := IdentifyThirdPartyBoot([]string{tmpDir})
			assert.Equal(t, tt.expected, detected)
		})
	}
}

func TestDetectBootStatus_Classification(t *testing.T) {
	// 1. UniBoot Cloud Mode
	status, code := DetectBootStatus("GPT", false, true, BootTypeNone, nil, false)
	assert.Equal(t, "UniBoot (1秒极速云引导盘)", status)
	assert.Equal(t, "uniboot_cloud", code)

	// 2. UniBoot Hybrid Mode
	hybridManifest := &UniBootManifest{Magic: MagicUniBootDisk, Mode: "hybrid", Version: "1.0.0"}
	status, code = DetectBootStatus("GPT", true, false, BootTypeNone, hybridManifest, false)
	assert.Equal(t, "UniBoot (混合模式引导盘)", status)
	assert.Equal(t, "uniboot_hybrid", code)

	// 2b. UniBoot Hybrid Mode even if isRealVentoy is temporarily false (e.g. unprivileged raw MBR read)
	status, code = DetectBootStatus("GPT", false, false, BootTypeNone, hybridManifest, false)
	assert.Equal(t, "UniBoot (混合模式引导盘)", status)
	assert.Equal(t, "uniboot_hybrid", code)

	// 3. Genuine Ventoy Disk (no UniBoot manifest)
	status, code = DetectBootStatus("GPT", true, false, BootTypeNone, nil, false)
	assert.Equal(t, "原生 Ventoy 启动盘 (可无损升级)", status)
	assert.Equal(t, "ventoy_pure", code)

	// 4. Third-party popular tools & boot disks
	status, code = DetectBootStatus("GPT", false, false, BootTypeRufus, nil, false)
	assert.Equal(t, "第三方引导: Rufus 制作盘", status)
	assert.Equal(t, "third_party_boot", code)
	assert.Equal(t, "rufus", MapThirdPartyBootCode(BootTypeRufus))
	assert.Equal(t, "wepe", MapThirdPartyBootCode(BootTypeWePE))
	assert.Equal(t, "easyu", MapThirdPartyBootCode(BootTypeEasyU))
	assert.Equal(t, "yumi", MapThirdPartyBootCode(BootTypeYUMI))
	assert.Equal(t, "opencore", MapThirdPartyBootCode(BootTypeOpenCore))
	assert.Equal(t, "clover", MapThirdPartyBootCode(BootTypeClover))
	assert.Equal(t, "windows_installer", MapThirdPartyBootCode(BootTypeWindowsInstaller))
	assert.Equal(t, "winpe_generic", MapThirdPartyBootCode(BootTypeWinPE))
	assert.Equal(t, "linux_live", MapThirdPartyBootCode(BootTypeLinuxLive))
	assert.Equal(t, "generic_uefi", MapThirdPartyBootCode(BootTypeGenericUEFI))

	// 5. Unmounted ESP partition detected in unprivileged mode
	status, code = DetectBootStatusWithElevation("MBR", false, false, BootTypeNone, nil, true, false)
	assert.Equal(t, "待授权", status)
	assert.Equal(t, "needs_privilege", code)

	// 6. Normal data disks
	status, code = DetectBootStatus("GPT", false, false, BootTypeNone, nil, false)
	assert.Equal(t, "GPT 数据盘", status)
	assert.Equal(t, "gpt_data", code)

	status, code = DetectBootStatus("MBR", false, false, BootTypeNone, nil, false)
	assert.Equal(t, "MBR 数据盘", status)
	assert.Equal(t, "mbr_data", code)

	status, code = DetectBootStatus("", false, false, BootTypeNone, nil, false)
	assert.Equal(t, "数据存储盘 (未检测到引导包)", status)
	assert.Equal(t, "data_storage", code)
}
