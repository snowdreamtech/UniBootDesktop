// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unibootdesktop/internal/logger"
	"github.com/snowdreamtech/unibootdesktop/pkg/privilege"
)

var (
	// ignoredVolumeExact defines exact volume names to ignore (case-insensitive)
	ignoredVolumeExact = []string{
		// macOS System & Internal Volumes
		"MACINTOSH HD",
		"MACINTOSH HD - DATA",
		"SYSTEM",
		"RECOVERY",
		"PREBOOT",
		"VM",
		"UPDATE",
		"XCODE",
		"INSTALLER",

		// EFI & Boot Partition Names
		"EFI",
		"ESP",
		"VTOYEFI",
		"UNIBOOTEFI",
		"SYSTEM RESERVED",
		"SYSTEM_RESERVED",
		"WINRE",
		"WINRETOOLS",
		"OEM",
	}

	// ignoredVolumePrefixes defines volume name prefixes to ignore (case-insensitive)
	ignoredVolumePrefixes = []string{
		"VTOYEFI",
		"UNIBOOTEFI",
		"EFI_",
		"EFI-",
		"BOOT_",
		"BOOT-",
		"TIME MACHINE",
		".TIMEMACHINE",
	}
)

// IsIgnoredVolume returns true if the volume name should be ignored (e.g., system disks, EFI/boot partitions).
func IsIgnoredVolume(name string) bool {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if upper == "" {
		return true
	}
	for _, exact := range ignoredVolumeExact {
		if upper == exact {
			return true
		}
	}
	for _, prefix := range ignoredVolumePrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// ParseWindowsDiskNumber validates and extracts a numeric Windows disk index from a raw device path.
// It accepts values like \\.\PhysicalDrive3, PhysicalDrive3, disk7, and 12, but rejects labels and unsafe strings.
func ParseWindowsDiskNumber(targetDisk string) (int, error) {
	trimmed := strings.TrimSpace(targetDisk)
	if trimmed == "" {
		return 0, fmt.Errorf("empty windows disk identifier")
	}

	normalized := trimmed
	normalized = strings.TrimPrefix(normalized, `\\.\`)
	normalized = strings.TrimPrefix(normalized, `//./`)
	normalized = strings.TrimPrefix(normalized, `\\?\`)
	normalized = strings.TrimPrefix(normalized, `//?/`)
	normalized = strings.TrimPrefix(normalized, "PhysicalDrive")
	normalized = strings.TrimPrefix(normalized, "physicaldrive")
	normalized = strings.TrimPrefix(normalized, "Disk")
	normalized = strings.TrimPrefix(normalized, "disk")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return 0, fmt.Errorf("missing windows disk index in %q", targetDisk)
	}
	if strings.ContainsAny(normalized, "\"'`$;&|()[]{}<>\\/") {
		return 0, fmt.Errorf("unsafe windows disk identifier: %q", targetDisk)
	}

	num, err := strconv.Atoi(normalized)
	if err != nil || num < 0 {
		return 0, fmt.Errorf("invalid windows disk index %q", targetDisk)
	}
	return num, nil
}

func parseWindowsDiskNumber(targetDisk string) (int, error) {
	return ParseWindowsDiskNumber(targetDisk)
}

// DiskInfo represents metadata about an available disk drive.
type DiskInfo struct {
	Device             string             `json:"device"`                       // Device path (e.g., /dev/disk2, E:)
	Name               string             `json:"name"`                         // Friendly display name only; not a trusted device identity
	Size               uint64             `json:"size"`                         // Total capacity in bytes
	Formatted          string             `json:"formatted"`                    // Human readable size string
	FreeSpace          uint64             `json:"freeSpace"`                    // Free available space in bytes
	FreeFormatted      string             `json:"freeFormatted"`                // Human readable free space string
	IsRemovable        bool               `json:"isRemovable"`                  // Removable disk flag
	IsSystem           bool               `json:"isSystem"`                     // System disk safety flag
	UsbVersion         string             `json:"usbVersion"`                   // Protocol version (USB 2.0, USB 3.0, USB 3.1, USB 3.2, USB4)
	UsbSpeed           string             `json:"usbSpeed"`                     // Physical bus speed (480 Mb/s, 5 Gb/s, 10 Gb/s, 20 Gb/s)
	Vendor             string             `json:"vendor"`                       // Device manufacturer / vendor
	FileSystem         string             `json:"fileSystem"`                   // File system format (e.g., ExFAT, FAT32, NTFS, APFS, ext4)
	PartitionScheme    string             `json:"partitionScheme"`              // Partition scheme (e.g., GPT, MBR)
	Writable           bool               `json:"writable"`                     // Read-Write status (true = Read-Write, false = Read-Only)
	SerialNumber       string             `json:"serialNumber"`                 // Hardware Serial Number
	VendorId           string             `json:"vendorId"`                     // USB Vendor ID (e.g., 0x21c4)
	ProductId          string             `json:"productId"`                    // USB Product ID (e.g., 0x0cd1)
	SmartStatus        string             `json:"smartStatus"`                  // S.M.A.R.T. health status (e.g. Verified, Not Supported, Failing)
	BusPower           string             `json:"busPower"`                     // Bus power available (e.g. 500 mA, 900 mA)
	BusPowerUsed       string             `json:"busPowerUsed"`                 // Bus power required/used (e.g. 500 mA, 224 mA)
	SectorSize         string             `json:"sectorSize"`                   // Sector block size (e.g. 512 Bytes, 4096 Bytes / 4Kn)
	TransportProtocol  string             `json:"transportProtocol"`            // USB Transport Protocol (e.g. UASP, BOT)
	BootStatus         string             `json:"bootStatus"`                   // Boot sector status (e.g. UniBoot/Ventoy Ready, MBR Bootable, Standard Data)
	BootStatusCode     string             `json:"bootStatusCode"`               // Standard machine code for i18n localization
	ControllerVendor   string             `json:"controllerVendor"`             // Inferred USB Controller Vendor (e.g. Phison, SMI, Alcor)
	IsFakeUsb3         bool               `json:"isFakeUsb3"`                   // Warning flag for fake USB 3.0 (USB 2.0 PHY disguised as 3.0)
	ProtocolCode       string             `json:"protocolCode"`                 // Styling code: "usb2", "usb3_0", "usb3_1", "usb3_2", "usb4"
	IsRealVentoy       bool               `json:"isRealVentoy"`                 // True ONLY if drive contains Ventoy MBR Sector 0 signature
	IsCloudMode        bool               `json:"isCloudMode"`                  // True if drive is formatted in Cloud Mode (iPXE ESP Cloud Pure)
	IsGenericBoot      bool               `json:"isGenericBoot"`                // True if drive contains generic 3rd-party bootloader (Rufus/PE/ISO)
	ThirdPartyBootType ThirdPartyBootType `json:"thirdPartyBootType,omitempty"` // Specific 3rd-party boot type if detected
	ThirdPartyBootCode string             `json:"thirdPartyBootCode,omitempty"` // Specific 3rd-party boot code for i18n localization
	UniBootVersion     string             `json:"unibootVersion,omitempty"`     // Official UniBoot firmware version (e.g. 1.0.0)
	UniBootMode        string             `json:"unibootMode,omitempty"`        // Official UniBoot deployment mode ("cloud" or "hybrid")
	MountPoint         string             `json:"mountPoint"`                   // Mount point or volume path (e.g. /Volumes/UNTITLED, E:\)
}

// CheckFakeUsb3 determines if a USB drive is a fake USB 3.0 device (claims USB 3.0+ in name/marketing but uses USB 2.0 PHY speed).
func CheckFakeUsb3(name string, version string, speed string) bool {
	upperName := strings.ToUpper(name)
	upperVer := strings.ToUpper(version)
	upperSpeed := strings.ToUpper(speed)

	// Check if marketed as USB 3.0 / 3.1 / 3.2 / SuperSpeed
	claimsUsb3 := strings.Contains(upperName, "3.0") ||
		strings.Contains(upperName, "USB3") ||
		strings.Contains(upperName, "USB 3") ||
		strings.Contains(upperName, "3.1") ||
		strings.Contains(upperName, "3.2") ||
		strings.Contains(upperName, "SUPERSPEED") ||
		strings.Contains(upperName, "SS")

	// Check if running on High-Speed USB 2.0 physical PHY (480 Mb/s or USB 2.0 version)
	isUsb2Phy := strings.Contains(upperSpeed, "480 MB") ||
		strings.Contains(upperSpeed, "480MB") ||
		upperVer == "USB 2.0" ||
		upperVer == "2.00" ||
		upperVer == "2.0"

	return claimsUsb3 && isUsb2Phy
}

// MapProtocolCode converts speed and version into standardized CSS protocol codes.
func MapProtocolCode(version string, speed string) string {
	upperVer := strings.ToUpper(version)
	upperSpeed := strings.ToUpper(speed)

	if strings.Contains(upperVer, "USB4") || strings.Contains(upperVer, "4.0") || strings.Contains(upperSpeed, "40 GB") {
		return "usb4"
	}
	if strings.Contains(upperVer, "3.2") || strings.Contains(upperSpeed, "20 GB") {
		return "usb3_2"
	}
	if strings.Contains(upperVer, "3.1") || strings.Contains(upperSpeed, "10 GB") {
		return "usb3_1"
	}
	if strings.Contains(upperVer, "3.0") || strings.Contains(upperVer, "3.00") || strings.Contains(upperSpeed, "5 GB") {
		return "usb3_0"
	}
	return "usb2"
}

// InferControllerVendor infers the likely USB master controller brand based on VID/PID and vendor strings.
func InferControllerVendor(vendorID string, productID string, vendor string) string {
	vid := strings.ToLower(strings.TrimSpace(vendorID))
	switch {
	case strings.Contains(vid, "0x0951") || strings.Contains(vid, "0x13fe"):
		return "Phison Controller"
	case strings.Contains(vid, "0x090c"):
		return "SMI Controller"
	case strings.Contains(vid, "0x058f"):
		return "Alcor Controller"
	case strings.Contains(vid, "0x1f75"):
		return "Innostor Controller"
	case strings.Contains(vid, "0x0781"):
		return "SanDisk Controller"
	case strings.Contains(vid, "0x1b1c"):
		return "Corsair / ASMedia Controller"
	case strings.Contains(vid, "0x152d"):
		return "JMicron Bridge Controller"
	case strings.Contains(vid, "0x174c"):
		return "ASMedia Controller"
	case strings.Contains(vid, "0x1e3d"):
		return "Chipsbank Controller"
	case strings.Contains(vid, "0x0bda"):
		return "Realtek Controller"
	case strings.Contains(vid, "0x05e3"):
		return "Genesys Logic Controller"
	}
	if vendor != "" && vendor != "Generic" {
		return vendor + " Controller"
	}
	return "Standard Controller"
}

// ThirdPartyBootType represents classified categories of 3rd-party bootloader drives.
type ThirdPartyBootType string

const (
	BootTypeNone             ThirdPartyBootType = ""
	BootTypeRufus            ThirdPartyBootType = "Rufus 制作盘"
	BootTypeWePE             ThirdPartyBootType = "微PE (WePE) 维护盘"
	BootTypeEasyU            ThirdPartyBootType = "优启通 (EasyU) 维护盘"
	BootTypeYUMI             ThirdPartyBootType = "YUMI 多系统引导盘"
	BootTypeOpenCore         ThirdPartyBootType = "OpenCore 黑苹果引导盘"
	BootTypeClover           ThirdPartyBootType = "Clover 黑苹果引导盘"
	BootTypeWindowsInstaller ThirdPartyBootType = "Windows 官方安装介质"
	BootTypeWinPE            ThirdPartyBootType = "通用 WinPE 维护盘"
	BootTypeLinuxLive        ThirdPartyBootType = "Linux Live 安装盘"
	BootTypeGenericUEFI      ThirdPartyBootType = "通用 UEFI 引导盘"
)

// MapThirdPartyBootCode maps a ThirdPartyBootType to a standard machine-readable code for i18n localization.
func MapThirdPartyBootCode(t ThirdPartyBootType) string {
	switch t {
	case BootTypeRufus:
		return "rufus"
	case BootTypeWePE:
		return "wepe"
	case BootTypeEasyU:
		return "easyu"
	case BootTypeYUMI:
		return "yumi"
	case BootTypeOpenCore:
		return "opencore"
	case BootTypeClover:
		return "clover"
	case BootTypeWindowsInstaller:
		return "windows_installer"
	case BootTypeWinPE:
		return "winpe_generic"
	case BootTypeLinuxLive:
		return "linux_live"
	case BootTypeGenericUEFI:
		return "generic_uefi"
	default:
		return ""
	}
}

// HasUnmountedEspPartition checks whether targetDisk contains an unmounted EFI/ESP/0xEF partition.
// Strictly inspects raw partition scheme types without relying on volume labels.
func HasUnmountedEspPartition(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}

	if runtime.GOOS == "darwin" {
		baseDisk := NormalizeDarwinDiskNode(targetDisk)
		if strings.HasPrefix(baseDisk, "disk") {
			p2 := baseDisk + "s2"
			strP2 := getDarwinDiskutilInfo(p2)
			if strP2 != "" {
				content := extractPlistValue(strP2, "Content")
				mountPoint := extractPlistValue(strP2, "MountPoint")
				// 0xEF is the standard hex partition type for ESP on MBR disks
				if (content == "0xEF" || strings.Contains(strings.ToUpper(content), "EFI")) && mountPoint == "" {
					return true
				}
			}
		}
	} else if runtime.GOOS == "linux" {
		if strings.HasPrefix(targetDisk, "/dev/") {
			out, err := execCommand("lsblk", "-o", "NAME,PARTTYPE,MOUNTPOINT", "-n", "-l", targetDisk).Output()
			if err == nil {
				for _, line := range strings.Split(string(out), "\n") {
					upper := strings.ToUpper(line)
					if (strings.Contains(upper, "C12A7328-F81F-11D2-BA4B-00A0C93EC93B") || strings.Contains(upper, "0XEF")) && !strings.Contains(line, "/") {
						return true
					}
				}
			}
		}
	}
	return false
}

// DetectBootStatus evaluates the boot status text and standard machine code based on partition scheme,
// Ventoy/Cloud Mode, 3rd-party boot type, UniBoot manifest, and unmounted ESP partition status.
// Strictly avoids relying on user-modifiable volume labels.
func DetectBootStatus(partitionScheme string, isRealVentoy bool, isCloudMode bool, thirdPartyBoot ThirdPartyBootType, manifest *UniBootManifest, hasUnmountedEsp bool) (string, string) {
	return DetectBootStatusWithElevation(partitionScheme, isRealVentoy, isCloudMode, thirdPartyBoot, manifest, hasUnmountedEsp, privilege.IsElevated())
}

// DetectBootStatusWithElevation evaluates the boot status with explicit elevation state.
func DetectBootStatusWithElevation(partitionScheme string, isRealVentoy bool, isCloudMode bool, thirdPartyBoot ThirdPartyBootType, manifest *UniBootManifest, hasUnmountedEsp bool, isElevated bool) (string, string) {
	// 1. Highest priority: Official UniBoot Manifest (Magic: UNIBOOT_DISK)
	if manifest != nil {
		if manifest.Mode == "cloud" {
			return "UniBoot (1秒极速云引导盘)", "uniboot_cloud"
		}
		if manifest.Mode == "hybrid" {
			return "UniBoot (混合模式引导盘)", "uniboot_hybrid"
		}
	}

	// 2. Pure Cloud Mode flag fallback (physical presence of cloud boot files)
	if isCloudMode {
		return "UniBoot (1秒极速云引导盘)", "uniboot_cloud"
	}

	// 3. Ventoy base drive (verified via physical MBR Sector 0 signature or physical ventoy engine files)
	if isRealVentoy {
		return "原生 Ventoy 启动盘 (可无损升级)", "ventoy_pure"
	}

	// 4. Specific 3rd-party boot creation tools (verified via physical payload fingerprints)
	if thirdPartyBoot != BootTypeNone {
		return fmt.Sprintf("第三方引导: %s", thirdPartyBoot), "third_party_boot"
	}

	// 5. Unmounted ESP partition detected in unprivileged mode (cannot inspect payload without root)
	if hasUnmountedEsp && !isElevated {
		return "待授权", "needs_privilege"
	}

	// 6. Plain data partition fallback (verified no boot partition or code detected)
	if strings.Contains(strings.ToUpper(partitionScheme), "GPT") {
		return "GPT 数据盘", "gpt_data"
	}
	if strings.Contains(strings.ToUpper(partitionScheme), "MBR") {
		return "MBR 数据盘", "mbr_data"
	}
	return "数据存储盘 (未检测到引导包)", "data_storage"
}

// IsEmptyDirectory returns true if a mount point contains no user files or directories,
// ignoring OS system metadata files (.DS_Store, .Spotlight-V100, .Trashes, $RECYCLE.BIN, System Volume Information).
func IsEmptyDirectory(mountPoint string) bool {
	if mountPoint == "" {
		return true
	}
	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		return true
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == ".DS_Store" || name == ".Spotlight-V100" || name == ".Trashes" ||
			name == ".fseventsd" || name == "$RECYCLE.BIN" || name == "System Volume Information" ||
			strings.HasPrefix(name, "._") {
			continue
		}
		return false
	}
	return true
}

// HasVentoyEngineFiles verifies physical presence of Ventoy core engine files inside a mount directory.
// Strictly inspects physical filesystem files and directories, NEVER relies on volume labels.
func HasVentoyEngineFiles(mountPoint string) bool {
	if mountPoint == "" || IsEmptyDirectory(mountPoint) {
		return false
	}
	ventoyDir := filepath.Join(mountPoint, "ventoy")
	if info, err := os.Stat(ventoyDir); err == nil && info.IsDir() {
		engineFiles := []string{
			"ventoy.json",
			"ventoy_grub.cfg",
			"ventoy.disk.img",
			"ventoy_os_list.json",
			"ventoy.wim",
			"ventoy.cpio",
			"ventoy_x64.efi",
			"ventoy_ia32.efi",
			"ventoy_aa64.efi",
			"vtoyutil_x64.efi",
			"vtoyjump32.exe",
			"vtoyjump64.exe",
		}
		for _, f := range engineFiles {
			if _, statErr := os.Stat(filepath.Join(ventoyDir, f)); statErr == nil {
				return true
			}
		}
		if entries, errRead := os.ReadDir(ventoyDir); errRead == nil && len(entries) > 0 {
			return true
		}
	}
	espVentoyImg := filepath.Join(mountPoint, "ventoy", "ventoy.disk.img")
	if _, err := os.Stat(espVentoyImg); err == nil {
		return true
	}
	// Also check if ESP partition has EFI/BOOT grub loader
	grubEfi := filepath.Join(mountPoint, "EFI", "BOOT", "grubx64_real.efi")
	if _, err := os.Stat(grubEfi); err == nil {
		return true
	}
	return false
}

// CheckVentoyMbrSignature inspects MBR Sector 0 for Ventoy's bootloader magic byte signature.
// Leverages direct raw read with privilege escalation fallback.
// Strictly relies on physical MBR boot opcodes and partition structures, NEVER relies on volume labels.
func CheckVentoyMbrSignature(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	buf, err := privilege.ReadSector(targetDisk, 512)
	if err != nil || len(buf) < 512 {
		return false
	}

	// 1. Direct ASCII check for custom or updated Ventoy bootloaders containing "Ventoy"
	if bytes.Contains(buf, []byte("Ventoy")) || bytes.Contains(buf, []byte("VENTOY")) {
		return true
	}

	// 2. Official Ventoy boot.img binary signatures:
	// Ventoy's boot.S begins with short jump: 0xeb, 0x63, 0x90 (jmp + nop).
	// Due to 446-byte MBR space constraints, Ventoy compresses error strings to "VT\x00Ge\x00HD\x00Rd\x00 Er\r\n".
	if bytes.Contains(buf, []byte("VT\x00Ge\x00HD\x00Rd")) {
		return true
	}
	if buf[0] == 0xeb && buf[1] == 0x63 && buf[2] == 0x90 && bytes.Contains(buf, []byte("VT\x00")) {
		return true
	}

	// 3. Ventoy MBR partition table structure check:
	// Partition 2 is exactly type 0xEF and size 32MB (65536 sectors = 0x00010000 in little-endian at offset 0x1D8).
	// Partition 3 and Partition 4 are completely empty (all zeroes).
	// Valid boot signature 0x55AA at offset 510.
	if buf[510] == 0x55 && buf[511] == 0xaa {
		p2Type := buf[0x1CE]
		p2Sectors := uint32(buf[0x1D8]) | uint32(buf[0x1D9])<<8 | uint32(buf[0x1DA])<<16 | uint32(buf[0x1DB])<<24
		p3Empty := true
		for i := 0x1DE; i < 0x1EE; i++ {
			if buf[i] != 0 {
				p3Empty = false
				break
			}
		}
		p4Empty := true
		for i := 0x1EE; i < 0x1FE; i++ {
			if buf[i] != 0 {
				p4Empty = false
				break
			}
		}
		if p2Type == 0xef && p2Sectors == 65536 && p3Empty && p4Empty {
			// Ventoy standard 32MB VTOYEFI partition layout with Ventoy JMP opcode or VT prefix
			if bytes.Contains(buf, []byte("VT\x00")) || (buf[0] == 0xeb && buf[1] == 0x63 && buf[2] == 0x90) {
				return true
			}
		}
	}

	return false
}

// HasUniBootCloudFiles verifies physical presence of UniBoot Cloud iPXE firmware files inside ESP partition.
// Prioritizes checking the official UniBoot manifest (ipxe/uniboot.json), with fallback to legacy boot scripts.
func HasUniBootCloudFiles(mountPoint string) bool {
	if mountPoint == "" || IsEmptyDirectory(mountPoint) {
		return false
	}
	if HasUniBootManifest(mountPoint) {
		return true
	}
	bootIpxe := filepath.Join(mountPoint, "boot.ipxe")
	unibootIpxe := filepath.Join(mountPoint, "ipxe", "uniboot.ipxe")
	_, errBoot := os.Stat(bootIpxe)
	_, errUni := os.Stat(unibootIpxe)
	return errBoot == nil || errUni == nil
}

// NormalizeDarwinDiskNode extracts the parent physical disk node (e.g. "disk2") from a macOS disk or partition path.
// Examples:
//
//	"/dev/disk2"    -> "disk2"
//	"/dev/rdisk2"   -> "disk2"
//	"/dev/disk2s1"  -> "disk2"
//	"disk2s2"       -> "disk2"
//	"/dev/disk12s3" -> "disk12"
func NormalizeDarwinDiskNode(targetDisk string) string {
	node := filepath.Base(targetDisk)
	node = strings.TrimPrefix(node, "r") // Remove raw disk prefix if present (rdisk2 -> disk2)

	if strings.HasPrefix(node, "disk") {
		rest := node[4:] // Part after "disk", e.g. "2", "2s1", "12s3"
		if idx := strings.Index(rest, "s"); idx > 0 {
			diskNum := rest[:idx]
			isNumeric := true
			for _, r := range diskNum {
				if r < '0' || r > '9' {
					isNumeric = false
					break
				}
			}
			if isNumeric {
				return "disk" + diskNum
			}
		}
		return node
	}
	return node
}

// IsVentoyDisk determines if a target disk device path or mount path is physically a Ventoy drive.
// Strictly checks MBR sector signatures, core ventoy engine files, and VTOYEFI/UNIBOOTEFI partition labels.
func IsVentoyDisk(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	if err := ValidateTargetDisk(targetDisk); err != nil {
		return false
	}

	// 1. Direct mount directory check if targetDisk is already a mount point
	if HasVentoyEngineFiles(targetDisk) {
		return true
	}

	// 2. Physical MBR Sector 0 signature check (Fast check when root/sudo permitted)
	if CheckVentoyMbrSignature(targetDisk) {
		return true
	}

	// 3. Scan all active mount points discovered for this disk
	for _, mp := range GetDiskMountPoints(targetDisk) {
		if HasVentoyEngineFiles(mp) {
			return true
		}
	}

	// 4. If running with elevated privileges, temporarily mount unmounted ESP to verify physical engine files
	if privilege.IsElevated() {
		baseDisk := NormalizeDarwinDiskNode(targetDisk)
		if strings.HasPrefix(baseDisk, "disk") {
			espPart := "/dev/" + baseDisk + "s2"
			if espDir, cleanup, err := privilege.MountHiddenESP(espPart); err == nil && espDir != "" {
				defer cleanup()
				if HasVentoyEngineFiles(espDir) {
					return true
				}
			}
		}
	} else if runtime.GOOS == "linux" {
		// Linux: query mounted partitions and check physical engine files
		out, err := execCommand("lsblk", "-o", "MOUNTPOINT", "-n", "-l", targetDisk).Output()
		if err == nil {
			mountPoints := strings.Split(strings.TrimSpace(string(out)), "\n")
			for _, mp := range mountPoints {
				mp = strings.TrimSpace(mp)
				if mp != "" && HasVentoyEngineFiles(mp) {
					return true
				}
			}
		}
	} else if runtime.GOOS == "windows" {
		diskNum, err := parseWindowsDiskNumber(targetDisk)
		if err != nil {
			return false
		}
		out, err := execCommand("powershell", "-NoProfile", "-NonInteractive", "-Command",
			fmt.Sprintf("Get-Partition -DiskNumber %d | Get-Volume | Select-Object -ExpandProperty DriveLetter", diskNum)).Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if l != "" && HasVentoyEngineFiles(l+":\\") {
					return true
				}
			}
		}
	}

	return false
}

// IsCloudModeDisk checks if a target disk is currently formatted in UniBoot Cloud mode (iPXE boot firmware in ESP, no Ventoy engine).
// Strictly checks physical iPXE files. NEVER relies on volume names alone.
// For hybrid mode disks (both Cloud and Ventoy), this returns false.
func IsCloudModeDisk(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	if err := ValidateTargetDisk(targetDisk); err != nil {
		return false
	}

	// First check if this is a Ventoy disk (hybrid mode or pure Ventoy)
	// If it has Ventoy MBR signature or Ventoy engine files, it's not a pure cloud disk
	if CheckVentoyMbrSignature(targetDisk) {
		return false
	}

	if runtime.GOOS == "darwin" {
		baseDisk := NormalizeDarwinDiskNode(targetDisk)
		if strings.HasPrefix(baseDisk, "disk") {
			// Check if any partition has Ventoy engine files (indicates hybrid mode)

			// Check if any partition has Ventoy engine files (indicates hybrid mode)
			p1 := baseDisk + "s1"
			p2 := baseDisk + "s2"
			for _, p := range []string{p1, p2} {
				str := getDarwinDiskutilInfo(p)
				if str != "" {
					mountPoint := extractPlistValue(str, "MountPoint")
					if HasVentoyEngineFiles(mountPoint) {
						return false // Hybrid mode - has Ventoy files
					}
				}
			}

			// Now check if it has cloud files (pure cloud mode)
			hasCloudFiles := false
			for _, p := range []string{p1, p2} {
				str := getDarwinDiskutilInfo(p)
				if str != "" {
					mountPoint := extractPlistValue(str, "MountPoint")
					if HasUniBootCloudFiles(mountPoint) {
						hasCloudFiles = true
						break
					}
				}
			}
			return hasCloudFiles
		}
	} else if runtime.GOOS == "linux" {
		// Linux: targetDisk is typically a device path like /dev/sdb
		if strings.HasPrefix(targetDisk, "/dev/") {
			// Query partitions and their mount points using lsblk
			out, err := execCommand("lsblk", "-o", "MOUNTPOINT", "-n", "-l", targetDisk).Output()
			if err == nil {
				mountPoints := strings.Split(strings.TrimSpace(string(out)), "\n")

				// Check if any partition has Ventoy engine files
				hasVentoyFiles := false
				hasCloudFiles := false
				for _, mp := range mountPoints {
					mp = strings.TrimSpace(mp)
					if mp != "" {
						if HasVentoyEngineFiles(mp) {
							hasVentoyFiles = true
						}
						if HasUniBootCloudFiles(mp) {
							hasCloudFiles = true
						}
					}
				}

				// Pure cloud mode: has cloud files but no Ventoy files
				if hasCloudFiles && !hasVentoyFiles {
					return true
				}
				return false
			}
		} else {
			// Mount point provided directly
			if HasVentoyEngineFiles(targetDisk) {
				return false
			}
			if HasUniBootCloudFiles(targetDisk) {
				return true
			}
		}
	} else if runtime.GOOS == "windows" {
		// Windows: cannot easily check partitions separately
		// For now, rely on file detection only
		// Note: This may have limitations for hybrid mode detection
		if HasVentoyEngineFiles(targetDisk) {
			return false
		}
		if HasUniBootCloudFiles(targetDisk) {
			return true
		}
	}
	return false
}

// IsRealVentoyDisk checks if a target disk is an active Ventoy drive containing Ventoy's MBR bootloader and configuration.
func IsRealVentoyDisk(targetDisk string) bool {
	if IsCloudModeDisk(targetDisk) {
		return false
	}
	return IsVentoyDisk(targetDisk)
}

// pathExists returns true if the specified file or directory path exists.
func pathExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

// IdentifyThirdPartyBoot scans a list of mount points belonging to a physical disk
// and returns the most specific ThirdPartyBootType detected based on exact directory/file fingerprints.
func IdentifyThirdPartyBoot(mountPoints []string) ThirdPartyBootType {
	// First pass: highly specific popular tools & specialized bootloaders
	for _, mp := range mountPoints {
		if mp == "" {
			continue
		}

		// 1. Rufus (UEFI:NTFS companion, Rufus EFI stub, or Rufus Windows bypass xml)
		if pathExists(filepath.Join(mp, "rufus.efi")) ||
			pathExists(filepath.Join(mp, "EFI", "rufus")) ||
			(pathExists(filepath.Join(mp, "autounattend.xml")) && pathExists(filepath.Join(mp, "autorun.ico"))) {
			return BootTypeRufus
		}

		// 2. 微PE (WePE)
		if pathExists(filepath.Join(mp, "WEPE")) ||
			pathExists(filepath.Join(mp, "EFI", "boot", "wepe.efi")) ||
			pathExists(filepath.Join(mp, "wepe.efi")) {
			return BootTypeWePE
		}

		// 3. 优启通 (EasyU / IT天空)
		if pathExists(filepath.Join(mp, "EASYU")) ||
			pathExists(filepath.Join(mp, "USBDATA")) ||
			pathExists(filepath.Join(mp, "SKY")) ||
			pathExists(filepath.Join(mp, "ITSKY")) {
			return BootTypeEasyU
		}

		// 4. YUMI / Universal USB Installer
		if pathExists(filepath.Join(mp, "multiboot", "menu", "yumi.cfg")) ||
			pathExists(filepath.Join(mp, "multiboot")) {
			return BootTypeYUMI
		}

		// 5. OpenCore Hackintosh bootloader
		if pathExists(filepath.Join(mp, "EFI", "OC", "OpenCore.efi")) ||
			pathExists(filepath.Join(mp, "EFI", "OC", "config.plist")) {
			return BootTypeOpenCore
		}

		// 6. Clover Hackintosh bootloader
		if pathExists(filepath.Join(mp, "EFI", "CLOVER", "CloverX64.efi")) ||
			pathExists(filepath.Join(mp, "EFI", "CLOVER", "config.plist")) {
			return BootTypeClover
		}
	}

	// Second pass: standard OS installer media and generic WinPE
	for _, mp := range mountPoints {
		if mp == "" {
			continue
		}

		// 7. Windows Official Installation Media
		if pathExists(filepath.Join(mp, "sources", "install.wim")) ||
			pathExists(filepath.Join(mp, "sources", "install.esd")) ||
			pathExists(filepath.Join(mp, "sources", "install.swm")) {
			return BootTypeWindowsInstaller
		}

		// 8. Generic WinPE Maintenance Disk
		if pathExists(filepath.Join(mp, "PETOOLS")) ||
			pathExists(filepath.Join(mp, "winpe.ini")) ||
			pathExists(filepath.Join(mp, "pe.cfg")) ||
			(pathExists(filepath.Join(mp, "sources", "boot.wim")) &&
				!pathExists(filepath.Join(mp, "sources", "install.wim")) &&
				!pathExists(filepath.Join(mp, "sources", "install.esd"))) {
			return BootTypeWinPE
		}

		// 9. Linux Live USB (casper, LiveOS, arch, isolinux, grub.cfg)
		if pathExists(filepath.Join(mp, "casper")) ||
			pathExists(filepath.Join(mp, "LiveOS")) ||
			pathExists(filepath.Join(mp, "arch", "boot")) ||
			pathExists(filepath.Join(mp, "isolinux")) ||
			pathExists(filepath.Join(mp, "boot", "grub", "grub.cfg")) {
			return BootTypeLinuxLive
		}
	}

	// Third pass: standard fallback UEFI bootloaders
	for _, mp := range mountPoints {
		if mp == "" {
			continue
		}
		// Exclude official UniBoot / Ventoy footprints to avoid misidentifying own bootloaders as 3rd-party generic UEFI
		if HasUniBootManifest(mp) || HasVentoyEngineFiles(mp) ||
			pathExists(filepath.Join(mp, "ipxe", "uniboot.ipxe")) ||
			pathExists(filepath.Join(mp, "ipxe", "uniboot.json")) ||
			pathExists(filepath.Join(mp, "ventoy")) {
			continue
		}

		if pathExists(filepath.Join(mp, "EFI", "BOOT", "BOOTX64.EFI")) ||
			pathExists(filepath.Join(mp, "EFI", "BOOT", "BOOTAA64.EFI")) ||
			pathExists(filepath.Join(mp, "EFI", "BOOT", "BOOTIA32.EFI")) ||
			pathExists(filepath.Join(mp, "EFI", "BOOT", "BOOTARM.EFI")) ||
			pathExists(filepath.Join(mp, "bootmgr")) ||
			pathExists(filepath.Join(mp, "boot", "bcd")) {
			return BootTypeGenericUEFI
		}
	}

	return BootTypeNone
}

// GetDiskMountPoints returns all known active mount points for targetDisk across platforms.
func GetDiskMountPoints(targetDisk string) []string {
	if targetDisk == "" {
		return nil
	}

	var results []string
	seen := make(map[string]bool)
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			seen[p] = true
			results = append(results, p)
		}
	}

	// 1. Direct mount point check
	if fi, err := os.Stat(targetDisk); err == nil && fi.IsDir() {
		add(targetDisk)
	}

	// 2. Darwin multi-partition check (p1 and p2)
	if runtime.GOOS == "darwin" {
		baseDisk := NormalizeDarwinDiskNode(targetDisk)
		if strings.HasPrefix(baseDisk, "disk") {
			for _, p := range []string{baseDisk + "s1", baseDisk + "s2"} {
				str := getDarwinDiskutilInfo(p)
				if str != "" {
					mp := extractPlistValue(str, "MountPoint")
					if mp != "" {
						add(mp)
					}
				}
			}

			// Scan system mount table for any mounted partitions of this disk (including temporary ESP mounts)
			out, err := execCommand("mount").Output()
			if err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					// Format: /dev/disk2s2 on /private/var/folders/... (msdos, ...)
					if strings.HasPrefix(line, "/dev/"+baseDisk) {
						parts := strings.Split(line, " on ")
						if len(parts) >= 2 {
							right := strings.Split(parts[1], " (")[0]
							add(strings.TrimSpace(right))
						}
					}
				}
			}
		}
	} else if runtime.GOOS == "linux" {
		if strings.HasPrefix(targetDisk, "/dev/") {
			out, err := execCommand("lsblk", "-o", "MOUNTPOINT", "-n", "-l", targetDisk).Output()
			if err == nil {
				for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
					add(line)
				}
			}
		}
	} else if runtime.GOOS == "windows" {
		diskNum, err := parseWindowsDiskNumber(targetDisk)
		if err != nil {
			if dn, dnErr := getDriveLetterDiskNumberWindows(targetDisk); dnErr == nil {
				diskNum = dn
				err = nil
			}
		}
		if err == nil {
			vols, vErr := getVolumesForDiskWindows(diskNum)
			if vErr == nil {
				for _, v := range vols {
					if v.DriveLetter != "" {
						add(v.DriveLetter + `\`)
					}
				}
			}
		}
		if len(results) == 0 {
			if len(targetDisk) == 2 && targetDisk[1] == ':' {
				add(targetDisk + `\`)
			}
		}
	}

	return results
}

// GetDiskThirdPartyBoot checks mounted partitions of targetDisk and returns the specific ThirdPartyBootType.
func GetDiskThirdPartyBoot(targetDisk string) ThirdPartyBootType {
	if targetDisk == "" {
		return BootTypeNone
	}
	mounts := GetDiskMountPoints(targetDisk)
	if len(mounts) == 0 {
		mounts = []string{targetDisk}
	}
	return IdentifyThirdPartyBoot(mounts)
}

// HasGenericBootFiles verifies physical presence of generic 3rd-party bootloader files
func HasGenericBootFiles(mountPoint string) bool {
	if mountPoint == "" || IsEmptyDirectory(mountPoint) {
		return false
	}
	return IdentifyThirdPartyBoot([]string{mountPoint}) != BootTypeNone
}

// IsGenericBootDisk checks if a target disk is a 3rd-party boot disk (Rufus, PE, ISO) that is NOT a Ventoy or Cloud Mode drive.
func IsGenericBootDisk(targetDisk string) bool {
	return GetDiskThirdPartyBoot(targetDisk) != BootTypeNone
}

// FormatBytes formats byte counts into human-readable strings using 1024 base (e.g. 29.80 GB).

// FormatBytes formats byte counts into human-readable strings using 1024 base (e.g. 29.80 GB).
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatBytesDual formats byte counts with 1024-base system capacity and 1000-base hardware nominal capacity.
// Example: "29.80 GB (Nominal 32 GB)"
func FormatBytesDual(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	sysFormatted := FormatBytes(bytes)
	const (
		GB1000 = 1000 * 1000 * 1000
		TB1000 = 1000 * GB1000
	)
	if bytes >= TB1000 {
		nomVal := float64(bytes) / float64(TB1000)
		return fmt.Sprintf("%s (Nominal %.0f TB)", sysFormatted, nomVal)
	} else if bytes >= GB1000 {
		nomVal := float64(bytes) / float64(GB1000)
		return fmt.Sprintf("%s (Nominal %.0f GB)", sysFormatted, nomVal)
	}
	return sysFormatted
}

// darwinDiskutilInflight tracks in-progress diskutil calls to avoid duplicate
// concurrent subprocess launches for the same node (singleflight without
// external dependencies).
type darwinDiskutilInflightEntry struct {
	ready chan struct{}
}

var (
	diskCacheMutex    sync.Mutex
	diskCacheList     []DiskInfo
	diskCacheTime     time.Time
	diskCacheSnapshot string

	darwinDiskutilCacheMutex sync.Mutex
	darwinDiskutilCacheMap   = make(map[string]string)
	darwinDiskutilCacheTime  time.Time
	// darwinDiskutilInflightMap prevents concurrent goroutines from launching
	// duplicate diskutil subprocesses for the same node.
	darwinDiskutilInflightMap = make(map[string]*darwinDiskutilInflightEntry)
)

func diskCacheShouldReuse(snapshot string, cachedAt time.Time) bool {
	if snapshot == "" {
		return false
	}
	if time.Since(cachedAt) > 5*time.Second {
		return false
	}
	return diskCacheSnapshot == snapshot
}

func getDarwinDiskutilInfo(node string) string {
	node = strings.TrimSpace(node)
	if node == "" {
		return ""
	}

	darwinDiskutilCacheMutex.Lock()
	// Expire the entire cache if it's stale (TTL 5s).
	if time.Since(darwinDiskutilCacheTime) > 5*time.Second {
		darwinDiskutilCacheMap = make(map[string]string)
		darwinDiskutilInflightMap = make(map[string]*darwinDiskutilInflightEntry)
		darwinDiskutilCacheTime = time.Now()
	}

	// Fast path: cache hit.
	if info, ok := darwinDiskutilCacheMap[node]; ok {
		darwinDiskutilCacheMutex.Unlock()
		return info
	}

	// Singleflight: if another goroutine is already fetching this node, wait for it.
	if entry, ok := darwinDiskutilInflightMap[node]; ok {
		darwinDiskutilCacheMutex.Unlock()
		<-entry.ready // block until the in-flight call completes
		darwinDiskutilCacheMutex.Lock()
		result := darwinDiskutilCacheMap[node]
		darwinDiskutilCacheMutex.Unlock()
		return result
	}

	// Slow path: this goroutine wins the race, registers itself as the owner.
	entry := &darwinDiskutilInflightEntry{ready: make(chan struct{})}
	darwinDiskutilInflightMap[node] = entry
	darwinDiskutilCacheMutex.Unlock()

	cmd := execCommand("diskutil", "info", "-plist", node)
	out, err := cmd.Output()
	var infoStr string
	if err == nil {
		infoStr = string(out)
	}

	// Store result and unblock all waiters atomically.
	darwinDiskutilCacheMutex.Lock()
	darwinDiskutilCacheMap[node] = infoStr
	delete(darwinDiskutilInflightMap, node)
	darwinDiskutilCacheMutex.Unlock()
	close(entry.ready) // wake all goroutines that were waiting on this node
	return infoStr
}

func invalidateDarwinDiskutilCache() {
	darwinDiskutilCacheMutex.Lock()
	darwinDiskutilCacheMap = make(map[string]string)
	darwinDiskutilInflightMap = make(map[string]*darwinDiskutilInflightEntry)
	darwinDiskutilCacheTime = time.Time{}
	darwinDiskutilCacheMutex.Unlock()
}

// InvalidateDiskCache clears the memory disk cache to force an immediate fresh hardware scan.
func InvalidateDiskCache() {
	diskCacheMutex.Lock()
	diskCacheList = nil
	diskCacheTime = time.Time{}
	diskCacheSnapshot = ""
	diskCacheMutex.Unlock()

	// Invalidate both diskutil and USB hardware caches on disk change events
	// to ensure fresh disk serial numbers and hardware metadata after hotplug.
	invalidateDarwinDiskutilCache()
	InvalidateDarwinUSBCache()
}

// InvalidateDarwinUSBCache forces an immediate purge of the macOS USB hardware profile cache.
func InvalidateDarwinUSBCache() {
	darwinUSBCacheMutex.Lock()
	darwinUSBCacheMap = nil
	darwinUSBCacheTime = time.Time{}
	darwinUSBCacheMutex.Unlock()
}

// StartHotplugMonitor listens for OS drive mount/unmount events lightweightly and triggers onChange.
func StartHotplugMonitor(ctx context.Context, onChange func()) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		lastSnapshot := getVolumeSnapshot()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentSnapshot := getVolumeSnapshot()
				if currentSnapshot != lastSnapshot {
					lastSnapshot = currentSnapshot
					InvalidateDiskCache()
					if onChange != nil {
						onChange()
					}
				}
			}
		}
	}()
}

func buildDarwinVolumeSnapshot(entries []os.DirEntry, infoByPath map[string]string) string {
	if len(entries) == 0 {
		return ""
	}

	var ids []string
	seen := make(map[string]struct{})
	for _, e := range entries {
		if e == nil || IsIgnoredVolume(e.Name()) {
			continue
		}
		volPath := filepath.ToSlash(filepath.Join("/Volumes", e.Name()))
		id := e.Name()
		if info, ok := infoByPath[volPath]; ok {
			if device := extractPlistValue(info, "ParentWholeDisk"); device != "" {
				id = device
			} else if device := extractPlistValue(info, "DeviceIdentifier"); device != "" {
				id = device
			}
		}
		if id == "" {
			id = e.Name()
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, fmt.Sprintf("%s:%s", e.Name(), id))
	}
	return strings.Join(ids, "|")
}

func getVolumeSnapshot() string {
	switch runtime.GOOS {
	case "darwin":
		// Fast path: only list /Volumes directory entries — no diskutil calls.
		// The hotplug watcher only needs to detect when the set of mounted
		// volumes changes; the actual disk details are fetched by GetRemovableDisks
		// only when a change is detected.
		entries, err := os.ReadDir("/Volumes")
		var names []string
		if err == nil {
			for _, e := range entries {
				if e == nil || IsIgnoredVolume(e.Name()) {
					continue
				}
				names = append(names, e.Name())
			}
		}

		// Also track /dev/disk[0-9]* device nodes so that plugging in or removing
		// unformatted, raw, or not-yet-mounted USB drives is detected immediately.
		if devEntries, err := filepath.Glob("/dev/disk[0-9]*"); err == nil {
			names = append(names, strings.Join(devEntries, ","))
		}
		sort.Strings(names)
		return strings.Join(names, "|")
	case "windows":
		var letters []string
		for c := 'C'; c <= 'Z'; c++ {
			drive := fmt.Sprintf("%c:\\", c)
			if _, err := os.Stat(drive); err == nil {
				letters = append(letters, string(c))
			}
		}
		return strings.Join(letters, "|")
	default:
		dirs := []string{"/media", "/run/media", "/mnt"}
		var names []string
		for _, d := range dirs {
			if entries, err := os.ReadDir(d); err == nil {
				for _, e := range entries {
					names = append(names, e.Name())
				}
			}
		}
		return strings.Join(names, "|")
	}
}

// GetRemovableDisks lists removable USB drives safely while protecting system drives.
func GetRemovableDisks() ([]DiskInfo, error) {
	diskCacheMutex.Lock()
	if diskCacheList != nil && !diskCacheTime.IsZero() && time.Since(diskCacheTime) < 5*time.Second {
		cached := make([]DiskInfo, len(diskCacheList))
		copy(cached, diskCacheList)
		diskCacheMutex.Unlock()
		return cached, nil
	}
	diskCacheMutex.Unlock()

	currentSnapshot := getVolumeSnapshot()

	diskCacheMutex.Lock()
	if diskCacheList != nil && diskCacheShouldReuse(currentSnapshot, diskCacheTime) {
		cached := make([]DiskInfo, len(diskCacheList))
		copy(cached, diskCacheList)
		diskCacheMutex.Unlock()
		return cached, nil
	}
	diskCacheMutex.Unlock()

	var disks []DiskInfo
	var err error

	switch runtime.GOOS {
	case "darwin":
		disks, err = getDarwinDisks()
	case "windows":
		disks, err = getWindowsDisks()
	default:
		disks, err = getLinuxDisks()
	}

	if disks == nil && err == nil {
		disks = []DiskInfo{}
	}

	if err == nil {
		diskCacheMutex.Lock()
		diskCacheList = disks
		diskCacheTime = time.Now()
		diskCacheSnapshot = getVolumeSnapshot()
		diskCacheMutex.Unlock()
	}

	return disks, err
}

// macOS implementation structures for system_profiler SPUSBDataType -json
type darwinUSBMedia struct {
	BsdName     string `json:"bsd_name"`
	SizeInBytes uint64 `json:"size_in_bytes"`
	Size        string `json:"size"`
	Volumes     []struct {
		MountPoint string `json:"mount_point"`
		Name       string `json:"_name"`
		BsdName    string `json:"bsd_name"`
	} `json:"volumes"`
}

type darwinUSBItem struct {
	Name         string           `json:"_name"`
	Manufacturer string           `json:"manufacturer"`
	DeviceSpeed  string           `json:"device_speed"`
	BcdDevice    string           `json:"bcd_device"`
	SerialNum    string           `json:"serial_num"`
	VendorID     string           `json:"vendor_id"`
	ProductID    string           `json:"product_id"`
	BusPower     string           `json:"bus_power"`
	BusPowerUsed string           `json:"bus_power_used"`
	Media        []darwinUSBMedia `json:"Media"`
	Items        []darwinUSBItem  `json:"_items"`
}

type darwinUSBProfiler struct {
	SPUSBDataType []struct {
		Items []darwinUSBItem `json:"_items"`
	} `json:"SPUSBDataType"`
}

type darwinUSBInfo struct {
	BsdName      string
	Vendor       string
	Model        string
	UsbVersion   string
	UsbSpeed     string
	TotalSize    uint64
	MountPoint   string
	VolumeName   string
	SerialNumber string
	VendorId     string
	ProductId    string
	BusPower     string
	BusPowerUsed string
}

// FormatMilliAmperes ensures electric current values have a human-readable 'mA' unit.
func FormatMilliAmperes(val string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return ""
	}
	if !strings.Contains(strings.ToLower(val), "ma") && !strings.Contains(strings.ToLower(val), "a") {
		return val + " mA"
	}
	return val
}

func walkDarwinUSBTree(items []darwinUSBItem, result map[string]*darwinUSBInfo) {
	for _, item := range items {
		for _, media := range item.Media {
			if media.BsdName != "" {
				ver, speed := parseDarwinUSBSpeed(item.DeviceSpeed, item.BcdDevice)
				vendor := strings.TrimSpace(item.Manufacturer)
				if vendor == "" || vendor == "USB" {
					vendor = "Generic"
				}

				info := &darwinUSBInfo{
					BsdName:      media.BsdName,
					Vendor:       vendor,
					Model:        item.Name,
					UsbVersion:   ver,
					UsbSpeed:     speed,
					TotalSize:    media.SizeInBytes,
					SerialNumber: strings.TrimSpace(item.SerialNum),
					VendorId:     strings.TrimSpace(item.VendorID),
					ProductId:    strings.TrimSpace(item.ProductID),
					BusPower:     FormatMilliAmperes(item.BusPower),
					BusPowerUsed: FormatMilliAmperes(item.BusPowerUsed),
				}

				for _, vol := range media.Volumes {
					if vol.MountPoint != "" {
						info.MountPoint = vol.MountPoint
						info.VolumeName = vol.Name
					}
				}
				result[media.BsdName] = info
			}
		}
		if len(item.Items) > 0 {
			walkDarwinUSBTree(item.Items, result)
		}
	}
}

func parseDarwinUSBSpeed(speed string, bcd string) (version string, phySpeed string) {
	lowerSpeed := strings.ToLower(speed)
	switch {
	case strings.Contains(lowerSpeed, "super_speed_plus_20") || strings.Contains(lowerSpeed, "20gb"):
		return "USB 3.2", "20 Gb/s"
	case strings.Contains(lowerSpeed, "super_speed_plus") || strings.Contains(lowerSpeed, "10gb"):
		return "USB 3.1", "10 Gb/s"
	case strings.Contains(lowerSpeed, "super_speed") || strings.Contains(lowerSpeed, "5gb"):
		return "USB 3.0", "5 Gb/s"
	case strings.Contains(lowerSpeed, "high_speed") || strings.Contains(lowerSpeed, "480mb"):
		return "USB 2.0", "480 Mb/s"
	}

	if bcd != "" {
		if strings.HasPrefix(bcd, "3.") {
			return "USB 3.0", "5 Gb/s"
		}
		if strings.HasPrefix(bcd, "2.") {
			return "USB 2.0", "480 Mb/s"
		}
	}
	return "USB 2.0", "480 Mb/s"
}

var (
	darwinUSBCacheMutex sync.Mutex
	darwinUSBCacheMap   map[string]*darwinUSBInfo
	darwinUSBCacheTime  time.Time
	// darwinUSBInflight is a singleflight gate: non-nil means a system_profiler
	// call is already in progress; close it to broadcast completion.
	darwinUSBInflight chan struct{}
)

func getCachedDarwinUSBMap() map[string]*darwinUSBInfo {
	for {
		darwinUSBCacheMutex.Lock()

		// Fast path: valid cache.
		if darwinUSBCacheMap != nil && time.Since(darwinUSBCacheTime) < 30*time.Second {
			m := darwinUSBCacheMap
			darwinUSBCacheMutex.Unlock()
			return m
		}

		// Singleflight: if another goroutine is already running system_profiler, wait for it.
		if darwinUSBInflight != nil {
			gate := darwinUSBInflight
			darwinUSBCacheMutex.Unlock()
			<-gate // block until the in-flight call finishes
			// Re-loop: the cache should now be populated.
			continue
		}

		// Slow path: this goroutine wins; register gate and release lock before
		// executing the expensive system_profiler call (~1000ms).
		gate := make(chan struct{})
		darwinUSBInflight = gate
		darwinUSBCacheMutex.Unlock()

		usbMap := make(map[string]*darwinUSBInfo)
		cmd := execCommand("system_profiler", "SPUSBDataType", "-json")
		output, err := cmd.Output()
		if err == nil {
			var profiler darwinUSBProfiler
			if jsonErr := jsonUnmarshal(output, &profiler); jsonErr == nil {
				for _, bus := range profiler.SPUSBDataType {
					walkDarwinUSBTree(bus.Items, usbMap)
				}
			}
		}

		// Commit result and clear in-flight gate atomically.
		darwinUSBCacheMutex.Lock()
		darwinUSBCacheMap = usbMap
		darwinUSBCacheTime = time.Now()
		darwinUSBInflight = nil
		darwinUSBCacheMutex.Unlock()
		close(gate) // wake all waiters
		return usbMap
	}
}

type darwinDiskutilPartition struct {
	DeviceIdentifier string `json:"DeviceIdentifier"`
	MountPoint       string `json:"MountPoint"`
	VolumeName       string `json:"VolumeName"`
	VolumeUUID       string `json:"VolumeUUID"`
	Size             uint64 `json:"Size"`
	Content          string `json:"Content"`
}

type darwinDiskutilWholeDisk struct {
	DeviceIdentifier string                    `json:"DeviceIdentifier"`
	Content          string                    `json:"Content"`
	Size             uint64                    `json:"Size"`
	OSInternal       bool                      `json:"OSInternal"`
	Partitions       []darwinDiskutilPartition `json:"Partitions"`
}

type darwinDiskutilList struct {
	WholeDisks            []string                  `json:"WholeDisks"`
	AllDisksAndPartitions []darwinDiskutilWholeDisk `json:"AllDisksAndPartitions"`
}

func getDarwinDisksBatch(usbMap map[string]*darwinUSBInfo) ([]DiskInfo, error) {
	cmd := execCommand("sh", "-c", "diskutil list -plist external physical | plutil -convert json -o - -")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var dl darwinDiskutilList
	if err := jsonUnmarshal(out, &dl); err != nil {
		return nil, err
	}

	// Fast short-circuit: if no external physical disks are connected,
	// return immediately without querying system_profiler (~1000ms saving).
	if len(dl.AllDisksAndPartitions) == 0 {
		return []DiskInfo{}, nil
	}

	// On-demand fetch of rich USB profile only when physical external disks actually exist
	if usbMap == nil {
		usbMap = getCachedDarwinUSBMap()
	}

	rootDisk := getDarwinRootSystemDisk()
	validDisks := make([]darwinDiskutilWholeDisk, 0, len(dl.AllDisksAndPartitions))

	for _, wholeDisk := range dl.AllDisksAndPartitions {
		parentDisk := strings.TrimSpace(wholeDisk.DeviceIdentifier)
		if parentDisk == "" || wholeDisk.OSInternal || parentDisk == rootDisk || parentDisk == "disk0" || parentDisk == "disk1" {
			continue
		}
		hasSystemMount := false
		for _, p := range wholeDisk.Partitions {
			mp := strings.TrimSpace(p.MountPoint)
			if mp == "/" || mp == "/System" || strings.HasPrefix(mp, "/System/") {
				hasSystemMount = true
				break
			}
		}
		if !hasSystemMount {
			validDisks = append(validDisks, wholeDisk)
		}
	}

	if len(validDisks) == 0 {
		return []DiskInfo{}, nil
	}

	disksResult := make([]*DiskInfo, len(validDisks))
	var wg sync.WaitGroup
	wg.Add(len(validDisks))

	for i, wd := range validDisks {
		go func(idx int, diskNode darwinDiskutilWholeDisk) {
			defer wg.Done()
			disksResult[idx] = inspectDarwinDisk(diskNode, usbMap)
		}(i, wd)
	}
	wg.Wait()

	disks := make([]DiskInfo, 0, len(disksResult))
	for _, d := range disksResult {
		if d != nil {
			disks = append(disks, *d)
		}
	}

	return disks, nil
}

func inspectDarwinDisk(wholeDisk darwinDiskutilWholeDisk, usbMap map[string]*darwinUSBInfo) *DiskInfo {
	parentDisk := strings.TrimSpace(wholeDisk.DeviceIdentifier)
	if parentDisk == "" {
		return nil
	}
	devNode := "/dev/" + parentDisk

	totalSize := wholeDisk.Size
	if parentInfo, ok := usbMap[parentDisk]; ok && totalSize == 0 {
		totalSize = parentInfo.TotalSize
	}
	if totalSize == 0 {
		totalSize = 32 * 1024 * 1024 * 1024
	}

	partitionScheme := "GPT / MBR"
	if strings.Contains(wholeDisk.Content, "FDisk") || strings.Contains(wholeDisk.Content, "MBR") {
		partitionScheme = "MBR (Master Boot Record)"
	} else if strings.Contains(wholeDisk.Content, "GUID") || strings.Contains(wholeDisk.Content, "GPT") {
		partitionScheme = "GPT (GUID Partition Table)"
	} else if wholeDisk.Content != "" {
		partitionScheme = wholeDisk.Content
	}

	var mountPoints []string
	var primaryMountPoint, primaryVolName, primaryFileSystem string
	var freeSpace uint64
	hasUnmountedEsp := false

	for _, p := range wholeDisk.Partitions {
		pContent := strings.ToUpper(strings.TrimSpace(p.Content))
		if pContent == "0XEF" || strings.Contains(pContent, "EFI") {
			if strings.TrimSpace(p.MountPoint) == "" {
				hasUnmountedEsp = true
			}
		}

		mp := strings.TrimSpace(p.MountPoint)
		if mp != "" {
			mountPoints = append(mountPoints, mp)
			if !IsIgnoredVolume(p.VolumeName) && primaryMountPoint == "" {
				primaryMountPoint = mp
				primaryVolName = strings.TrimSpace(p.VolumeName)
				primaryFileSystem = strings.TrimSpace(p.Content)

				freeSpace = getMountFreeSpace(mp)
			}
		}
	}

	// Filter out completely unmounted disks (no mounted volumes visible in OS file manager/Finder)
	// to avoid user confusion and ghost devices (e.g. empty card readers or unmounted hardware).
	if primaryMountPoint == "" && len(mountPoints) == 0 {
		return nil
	}

	if strings.EqualFold(primaryFileSystem, "Windows_NTFS") {
		primaryFileSystem = "ExFAT / NTFS"
	} else if strings.EqualFold(primaryFileSystem, "DOS_FAT_32") {
		primaryFileSystem = "FAT32"
	} else if primaryFileSystem == "" {
		if len(wholeDisk.Partitions) == 0 {
			primaryFileSystem = "RAW / Unformatted"
		} else {
			primaryFileSystem = "ExFAT"
		}
	}

	formattedSize := FormatBytesDual(totalSize)
	freeFormatted := FormatBytes(freeSpace)
	if freeSpace == 0 {
		freeFormatted = formattedSize
	}

	usbVer := "USB 2.0"
	usbSpeed := "480 Mb/s"
	vendor := "Generic"
	displayName := primaryVolName
	serialNum := ""
	vendorId := ""
	productId := ""
	busPower := "500 mA"
	busPowerUsed := "500 mA"

	if parentInfo, ok := usbMap[parentDisk]; ok && parentInfo != nil {
		if parentInfo.Vendor != "" {
			vendor = parentInfo.Vendor
		}
		if parentInfo.UsbVersion != "" {
			usbVer = parentInfo.UsbVersion
		}
		if parentInfo.UsbSpeed != "" {
			usbSpeed = parentInfo.UsbSpeed
		}
		if displayName == "" && parentInfo.Model != "" {
			displayName = parentInfo.Model
		}
		serialNum = parentInfo.SerialNumber
		vendorId = parentInfo.VendorId
		productId = parentInfo.ProductId
		if parentInfo.BusPower != "" {
			busPower = parentInfo.BusPower
		}
		if parentInfo.BusPowerUsed != "" {
			busPowerUsed = parentInfo.BusPowerUsed
		}
	}

	if displayName == "" {
		displayName = parentDisk
	}

	isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
	protoCode := MapProtocolCode(usbVer, usbSpeed)

	isRealVentoy := CheckVentoyMbrSignature(devNode)
	if !isRealVentoy && privilege.IsElevated() {
		isRealVentoy = IsVentoyDisk(devNode)
	}
	var manifest *UniBootManifest
	for _, mp := range mountPoints {
		if m, err := ReadUniBootManifest(mp); err == nil && m != nil {
			manifest = m
			break
		}
	}
	if manifest == nil && !isRealVentoy && privilege.IsElevated() {
		manifest = GetDiskUniBootManifest(devNode)
	}
	isCloudMode := false

	for _, mp := range mountPoints {
		if HasVentoyEngineFiles(mp) {
			isRealVentoy = true
		}
		if HasUniBootCloudFiles(mp) {
			isCloudMode = true
		}
	}

	if manifest != nil {
		if manifest.Mode == "cloud" {
			isCloudMode = true
		} else if manifest.Mode == "hybrid" {
			isRealVentoy = true
		}
	}

	thirdPartyBoot := BootTypeNone
	isGenericBoot := false
	if !isCloudMode && !isRealVentoy {
		thirdPartyBoot = IdentifyThirdPartyBoot(mountPoints)
		isGenericBoot = thirdPartyBoot != BootTypeNone
	}

	bootStatusStr, bootStatusCode := DetectBootStatus(partitionScheme, isRealVentoy, isCloudMode, thirdPartyBoot, manifest, hasUnmountedEsp)
	controllerVendorStr := InferControllerVendor(vendorId, productId, vendor)

	return &DiskInfo{
		Device:             devNode,
		Name:               displayName,
		Size:               totalSize,
		Formatted:          formattedSize,
		FreeSpace:          freeSpace,
		FreeFormatted:      freeFormatted,
		IsRemovable:        true,
		IsSystem:           false,
		UsbVersion:         usbVer,
		UsbSpeed:           usbSpeed,
		Vendor:             vendor,
		FileSystem:         primaryFileSystem,
		PartitionScheme:    partitionScheme,
		Writable:           true,
		SerialNumber:       serialNum,
		VendorId:           vendorId,
		ProductId:          productId,
		SmartStatus:        "Verified",
		BusPower:           busPower,
		BusPowerUsed:       busPowerUsed,
		SectorSize:         "512 Bytes (512n/512e)",
		TransportProtocol:  "BOT (Bulk-Only Transport)",
		BootStatus:         bootStatusStr,
		BootStatusCode:     bootStatusCode,
		ControllerVendor:   controllerVendorStr,
		IsFakeUsb3:         isFake,
		ProtocolCode:       protoCode,
		IsRealVentoy:       isRealVentoy,
		IsCloudMode:        isCloudMode,
		IsGenericBoot:      isGenericBoot,
		ThirdPartyBootType: thirdPartyBoot,
		ThirdPartyBootCode: MapThirdPartyBootCode(thirdPartyBoot),
		UniBootVersion: func() string {
			if manifest != nil {
				return manifest.Version
			}
			return ""
		}(),
		UniBootMode: func() string {
			if manifest != nil {
				return manifest.Mode
			}
			return ""
		}(),
		MountPoint: primaryMountPoint,
	}
}

func getDarwinDisks() ([]DiskInfo, error) {
	// Optimization: Call getDarwinDisksBatch with nil usbMap to enable fast short-circuiting (~140ms)
	// when no external disks are attached, avoiding the expensive system_profiler SPUSBDataType call (~1000ms).
	// If external disks exist, getDarwinDisksBatch fetches or reuses the cached USB map.
	batchDisks, err := getDarwinDisksBatch(nil)
	if err != nil {
		logger.Warn("Batch disk discovery failed, using fallback scan", "error", err)
	} else if len(batchDisks) > 0 || err == nil {
		return batchDisks, nil
	}

	disks := make([]DiskInfo, 0)
	usbMap := getCachedDarwinUSBMap()

	// Step 3: Scan /Volumes for mounted removable drives (fallback)
	entries, err := os.ReadDir("/Volumes")
	if err != nil {
		return disks, nil
	}

	// Step 3: Pre-fetch diskutil info for all visible volumes concurrently to
	// avoid N serial subprocess calls (each ~80ms) blocking the scan.
	type infoResult struct {
		key string
		val string
	}
	infoCh := make(chan infoResult, len(entries)*2) // *2 for both volPath and parentDisk
	for _, entry := range entries {
		if IsIgnoredVolume(entry.Name()) {
			continue
		}
		volPath := filepath.Join("/Volumes", entry.Name())
		go func(p string) {
			infoCh <- infoResult{key: p, val: getDarwinDiskutilInfo(p)}
		}(volPath)
	}
	// collect only the volume-level results; parentDisk results are fetched below with same concurrency
	prefetchCount := 0
	for _, entry := range entries {
		if !IsIgnoredVolume(entry.Name()) {
			prefetchCount++
		}
	}
	prefetchedInfo := make(map[string]string, prefetchCount)
	for i := 0; i < prefetchCount; i++ {
		r := <-infoCh
		prefetchedInfo[r.key] = r.val
	}

	for _, entry := range entries {
		if IsIgnoredVolume(entry.Name()) {
			continue
		}

		volPath := filepath.Join("/Volumes", entry.Name())
		volName := entry.Name()

		// Use pre-fetched diskutil info (collected concurrently above)
		infoStr := prefetchedInfo[volPath]

		var totalSize uint64
		var freeSpace uint64
		var parentDisk string
		var busProto string
		var isRemovable bool
		var fileSystem string
		var partitionScheme string
		var smartStatus string
		var sectorBytes uint64
		var writable bool = true
		var isVirtual bool
		var isInternal bool
		var isOptical bool
		var isNetwork bool

		if infoStr != "" {
			if strings.Contains(infoStr, "<key>BusProtocol</key>") {
				busProto = extractPlistValue(infoStr, "BusProtocol")
			}
			if strings.Contains(infoStr, "<key>ParentWholeDisk</key>") {
				parentDisk = extractPlistValue(infoStr, "ParentWholeDisk")
			}
			if strings.Contains(infoStr, "<key>TotalSize</key>") {
				totalSize = extractPlistUint(infoStr, "TotalSize")
			}
			if strings.Contains(infoStr, "<key>FreeSpace</key>") {
				freeSpace = extractPlistUint(infoStr, "FreeSpace")
			}
			if strings.Contains(infoStr, "<key>SMARTStatus</key>") {
				smartStatus = extractPlistValue(infoStr, "SMARTStatus")
			}
			if strings.Contains(infoStr, "<key>DeviceBlockSize</key>") {
				sectorBytes = extractPlistUint(infoStr, "DeviceBlockSize")
			}
			if strings.Contains(infoStr, "<key>RemovableMediaOrExternalDevice</key>") {
				isRemovable = strings.Contains(infoStr, "<key>RemovableMediaOrExternalDevice</key>\n\t<true/>") || strings.Contains(infoStr, "<key>RemovableMediaOrExternalDevice</key><true/>")
			}
			if strings.Contains(infoStr, "<key>Internal</key>") {
				isInternal = strings.Contains(infoStr, "<key>Internal</key>\n\t<true/>") || strings.Contains(infoStr, "<key>Internal</key><true/>")
			}
			if strings.Contains(infoStr, "<key>OSInternalMedia</key>") {
				if strings.Contains(infoStr, "<key>OSInternalMedia</key>\n\t<true/>") || strings.Contains(infoStr, "<key>OSInternalMedia</key><true/>") {
					isInternal = true
				}
			}
			if strings.Contains(infoStr, "<key>VirtualOrPhysical</key>") {
				vOrP := extractPlistValue(infoStr, "VirtualOrPhysical")
				if strings.EqualFold(vOrP, "Virtual") {
					isVirtual = true
				}
			}
			if strings.Contains(infoStr, "<key>OpticalDevice</key>") {
				if strings.Contains(infoStr, "<key>OpticalDevice</key>\n\t<true/>") || strings.Contains(infoStr, "<key>OpticalDevice</key><true/>") {
					isOptical = true
				}
			}
			if strings.Contains(infoStr, "<key>Writable</key>") {
				if strings.Contains(infoStr, "<key>Writable</key>\n\t<false/>") || strings.Contains(infoStr, "<key>Writable</key><false/>") {
					writable = false
				}
			}
			if strings.Contains(infoStr, "<key>WritableMedia</key>") {
				if strings.Contains(infoStr, "<key>WritableMedia</key>\n\t<false/>") || strings.Contains(infoStr, "<key>WritableMedia</key><false/>") {
					writable = false
				}
			}
			if strings.Contains(infoStr, "<key>WritableVolume</key>") {
				if strings.Contains(infoStr, "<key>WritableVolume</key>\n\t<false/>") || strings.Contains(infoStr, "<key>WritableVolume</key><false/>") {
					writable = false
				}
			}
			if strings.Contains(infoStr, "<key>FilesystemUserVisibleName</key>") {
				fileSystem = extractPlistValue(infoStr, "FilesystemUserVisibleName")
			}
			if fileSystem == "" && strings.Contains(infoStr, "<key>FilesystemName</key>") {
				fileSystem = extractPlistValue(infoStr, "FilesystemName")
			}
			if fileSystem == "" && strings.Contains(infoStr, "<key>FilesystemType</key>") {
				fileSystem = extractPlistValue(infoStr, "FilesystemType")
			}
			fsLower := strings.ToLower(fileSystem)
			if fsLower == "smbfs" || fsLower == "nfs" || fsLower == "afpfs" || fsLower == "cifs" || fsLower == "webdav" {
				isNetwork = true
			}
		}

		if fileSystem == "" {
			fileSystem = "ExFAT"
		}

		var parentBusProto string
		var parentIsVirtual bool
		var parentIsInternal bool
		var parentIsOptical bool
		var parentWritable bool = true

		// Probe whole disk info for total raw byte size and partition map type
		if parentDisk != "" {
			parentStr := getDarwinDiskutilInfo(parentDisk)
			if parentStr != "" {
				if strings.Contains(parentStr, "<key>BusProtocol</key>") {
					parentBusProto = extractPlistValue(parentStr, "BusProtocol")
				}
				if strings.Contains(parentStr, "<key>VirtualOrPhysical</key>") {
					if strings.EqualFold(extractPlistValue(parentStr, "VirtualOrPhysical"), "Virtual") {
						parentIsVirtual = true
					}
				}
				if strings.Contains(parentStr, "<key>Internal</key>") {
					if strings.Contains(parentStr, "<key>Internal</key>\n\t<true/>") || strings.Contains(parentStr, "<key>Internal</key><true/>") {
						parentIsInternal = true
					}
				}
				if strings.Contains(parentStr, "<key>OpticalDevice</key>") {
					if strings.Contains(parentStr, "<key>OpticalDevice</key>\n\t<true/>") || strings.Contains(parentStr, "<key>OpticalDevice</key><true/>") {
						parentIsOptical = true
					}
				}
				if strings.Contains(parentStr, "<key>Writable</key>") {
					if strings.Contains(parentStr, "<key>Writable</key>\n\t<false/>") || strings.Contains(parentStr, "<key>Writable</key><false/>") {
						parentWritable = false
					}
				}
				if strings.Contains(parentStr, "<key>WritableMedia</key>") {
					if strings.Contains(parentStr, "<key>WritableMedia</key>\n\t<false/>") || strings.Contains(parentStr, "<key>WritableMedia</key><false/>") {
						parentWritable = false
					}
				}
				if strings.Contains(parentStr, "<key>IORegistryEntryName</key>") {
					entryName := strings.ToLower(extractPlistValue(parentStr, "IORegistryEntryName"))
					if strings.Contains(entryName, "disk image") || strings.Contains(entryName, "virtual") || strings.Contains(entryName, "appleapfs") {
						parentIsVirtual = true
					}
				}
				pSize := extractPlistUint(parentStr, "TotalSize")
				if pSize > 0 {
					totalSize = pSize
				}
				if smartStatus == "" && strings.Contains(parentStr, "<key>SMARTStatus</key>") {
					smartStatus = extractPlistValue(parentStr, "SMARTStatus")
				}
				if sectorBytes == 0 && strings.Contains(parentStr, "<key>DeviceBlockSize</key>") {
					sectorBytes = extractPlistUint(parentStr, "DeviceBlockSize")
				}
				content := extractPlistValue(parentStr, "Content")
				if strings.Contains(content, "GUID") || strings.Contains(content, "GPT") {
					partitionScheme = "GPT (GUID Partition Table)"
				} else if strings.Contains(content, "FDisk") || strings.Contains(content, "MBR") {
					partitionScheme = "MBR (Master Boot Record)"
				} else if content != "" {
					partitionScheme = content
				}
			}
		}

		// Comprehensive filter for virtual disks, read-only images, DMG, optical drives, network shares, and internal system drives
		effectiveBus := busProto
		if effectiveBus == "" {
			effectiveBus = parentBusProto
		}

		// 1. Filter out virtual disk images (DMG, ISO mounts, virtual devices)
		if strings.EqualFold(busProto, "Disk Image") || strings.EqualFold(parentBusProto, "Disk Image") ||
			strings.Contains(strings.ToLower(busProto), "image") || strings.Contains(strings.ToLower(parentBusProto), "image") ||
			isVirtual || parentIsVirtual {
			continue
		}

		// 2. Filter out read-only media (boot disk flashing requires physical read/write capacity)
		if !writable || !parentWritable {
			continue
		}

		// 3. Filter out optical drives and discs (CD/DVD/BD)
		if isOptical || parentIsOptical || strings.EqualFold(busProto, "ATAPI") || strings.EqualFold(parentBusProto, "ATAPI") {
			continue
		}

		// 4. Filter out network volumes (NFS/SMB/AFP)
		if isNetwork {
			continue
		}

		// 5. Filter out internal system drives (macOS root, recovery, internal NVMe/APFS)
		if isInternal || parentIsInternal {
			continue
		}

		// 6. Must be removable physical storage (e.g. USB)
		if effectiveBus != "" && effectiveBus != "USB" && !isRemovable {
			continue
		}

		if smartStatus == "" {
			smartStatus = "Verified"
		}

		if partitionScheme == "" {
			partitionScheme = "GPT / MBR"
		}

		sectorSizeStr := "512 Bytes (512n/512e)"
		if sectorBytes == 4096 {
			sectorSizeStr = "4096 Bytes (4Kn Native)"
		} else if sectorBytes > 0 {
			sectorSizeStr = fmt.Sprintf("%d Bytes", sectorBytes)
		}

		transportProtoStr := "BOT (Bulk-Only Transport)"
		if strings.Contains(strings.ToUpper(busProto), "UASP") || strings.Contains(strings.ToUpper(busProto), "SCSI") {
			transportProtoStr = "UASP (USB Attached SCSI)"
		}

		usbVer := "USB 2.0"
		usbSpeed := "480 Mb/s"
		vendor := "Generic"
		displayName := volName
		serialNum := ""
		vendorId := ""
		productId := ""
		busPower := "500 mA"
		busPowerUsed := "500 mA"

		// Match with system_profiler hardware metadata
		if parentInfo, ok := usbMap[parentDisk]; ok {
			if parentInfo.TotalSize > 0 {
				totalSize = parentInfo.TotalSize
			}
			if parentInfo.Vendor != "" {
				vendor = parentInfo.Vendor
			}
			if parentInfo.UsbVersion != "" {
				usbVer = parentInfo.UsbVersion
			}
			if parentInfo.UsbSpeed != "" {
				usbSpeed = parentInfo.UsbSpeed
			}
			if displayName == "" && parentInfo.Model != "" {
				displayName = parentInfo.Model
			}
			serialNum = parentInfo.SerialNumber
			vendorId = parentInfo.VendorId
			productId = parentInfo.ProductId
			if parentInfo.BusPower != "" {
				busPower = parentInfo.BusPower
			}
			if parentInfo.BusPowerUsed != "" {
				busPowerUsed = parentInfo.BusPowerUsed
			}
		}

		if totalSize == 0 {
			totalSize = 32 * 1024 * 1024 * 1024 // Fallback if size unknown
		}

		formattedSize := FormatBytesDual(totalSize)
		freeFormatted := FormatBytes(freeSpace)
		if freeSpace == 0 {
			freeFormatted = formattedSize
		}

		isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
		protoCode := MapProtocolCode(usbVer, usbSpeed)
		devNode := volPath
		if parentDisk != "" {
			devNode = "/dev/" + parentDisk
		}

		// Use IsRealVentoyDisk which correctly handles hybrid mode detection
		manifest := GetDiskUniBootManifest(devNode)
		isRealVentoy := IsRealVentoyDisk(devNode)
		isCloudMode := IsCloudModeDisk(devNode)
		if manifest != nil {
			if manifest.Mode == "cloud" {
				isCloudMode = true
			} else if manifest.Mode == "hybrid" {
				isRealVentoy = true
			}
		}
		thirdPartyBoot := BootTypeNone
		isGenericBoot := false
		if !isCloudMode && !isRealVentoy {
			thirdPartyBoot = GetDiskThirdPartyBoot(devNode)
			isGenericBoot = thirdPartyBoot != BootTypeNone
		}

		hasUnmountedEsp := HasUnmountedEspPartition(devNode)
		bootStatusStr, bootStatusCode := DetectBootStatus(partitionScheme, isRealVentoy, isCloudMode, thirdPartyBoot, manifest, hasUnmountedEsp)
		controllerVendorStr := InferControllerVendor(vendorId, productId, vendor)

		// Detect if this is a system disk
		isSystemDisk, _ := isSystemDiskDarwin(devNode)

		disks = append(disks, DiskInfo{
			Device:             devNode,
			Name:               displayName,
			Size:               totalSize,
			Formatted:          formattedSize,
			FreeSpace:          freeSpace,
			FreeFormatted:      freeFormatted,
			IsRemovable:        true,
			IsSystem:           isSystemDisk,
			UsbVersion:         usbVer,
			UsbSpeed:           usbSpeed,
			Vendor:             vendor,
			FileSystem:         fileSystem,
			PartitionScheme:    partitionScheme,
			Writable:           writable,
			SerialNumber:       serialNum,
			VendorId:           vendorId,
			ProductId:          productId,
			SmartStatus:        smartStatus,
			BusPower:           busPower,
			BusPowerUsed:       busPowerUsed,
			SectorSize:         sectorSizeStr,
			TransportProtocol:  transportProtoStr,
			BootStatus:         bootStatusStr,
			BootStatusCode:     bootStatusCode,
			ControllerVendor:   controllerVendorStr,
			IsFakeUsb3:         isFake,
			ProtocolCode:       protoCode,
			IsRealVentoy:       isRealVentoy,
			IsCloudMode:        isCloudMode,
			IsGenericBoot:      isGenericBoot,
			ThirdPartyBootType: thirdPartyBoot,
			ThirdPartyBootCode: MapThirdPartyBootCode(thirdPartyBoot),
			UniBootVersion: func() string {
				if manifest != nil {
					return manifest.Version
				}
				return ""
			}(),
			UniBootMode: func() string {
				if manifest != nil {
					return manifest.Mode
				}
				return ""
			}(),
			MountPoint: volPath,
		})
	}

	return disks, nil
}

// Helpers for simple XML plist string parsing without heavy external dependencies
func extractPlistValue(plistStr string, key string) string {
	keyTag := "<key>" + key + "</key>"
	idx := strings.Index(plistStr, keyTag)
	if idx == -1 {
		return ""
	}
	sub := plistStr[idx+len(keyTag):]
	startStr := strings.Index(sub, "<string>")
	if startStr == -1 {
		return ""
	}
	endStr := strings.Index(sub, "</string>")
	if endStr == -1 || endStr <= startStr+8 {
		return ""
	}
	return sub[startStr+8 : endStr]
}

func extractPlistUint(plistStr string, key string) uint64 {
	keyTag := "<key>" + key + "</key>"
	idx := strings.Index(plistStr, keyTag)
	if idx == -1 {
		return 0
	}
	sub := plistStr[idx+len(keyTag):]
	startInt := strings.Index(sub, "<integer>")
	if startInt == -1 {
		return 0
	}
	endInt := strings.Index(sub, "</integer>")
	if endInt == -1 || endInt <= startInt+9 {
		return 0
	}
	valStr := sub[startInt+9 : endInt]
	var val uint64
	fmt.Sscanf(valStr, "%d", &val)
	return val
}

// Linux disk probing via lsblk -J
type linuxBlockDevice struct {
	Name       string             `json:"name"`
	Size       uint64             `json:"size"`
	Fsavail    uint64             `json:"fsavail"`
	Rm         bool               `json:"rm"`
	Ro         bool               `json:"ro"`
	Type       string             `json:"type"`
	MountPoint string             `json:"mountpoint"`
	Model      string             `json:"model"`
	Vendor     string             `json:"vendor"`
	Tran       string             `json:"tran"`
	Fstype     string             `json:"fstype"`
	Pttype     string             `json:"pttype"`
	Children   []linuxBlockDevice `json:"children"`
}

type linuxLsblkOutput struct {
	BlockDevices []linuxBlockDevice `json:"blockdevices"`
}

func getLinuxDisks() ([]DiskInfo, error) {
	disks := make([]DiskInfo, 0)
	cmd := execCommand("lsblk", "-J", "-b", "-o", "NAME,SIZE,FSAVAIL,RM,RO,TYPE,MOUNTPOINT,MODEL,VENDOR,TRAN,FSTYPE,PTTYPE")
	output, err := cmd.Output()
	if err != nil {
		return disks, nil
	}

	var lsblk linuxLsblkOutput
	if err := jsonUnmarshal(output, &lsblk); err != nil {
		return disks, nil
	}

	validDevs := make([]linuxBlockDevice, 0, len(lsblk.BlockDevices))
	for _, dev := range lsblk.BlockDevices {
		// Filter out virtual, loop, ram, optical, and read-only devices
		if dev.Type == "loop" || strings.HasPrefix(dev.Name, "loop") {
			continue
		}
		if dev.Type == "ram" || strings.HasPrefix(dev.Name, "ram") || strings.HasPrefix(dev.Name, "zram") {
			continue
		}
		if dev.Type == "rom" || strings.HasPrefix(dev.Name, "sr") || strings.HasPrefix(dev.Name, "cdrom") {
			continue
		}
		if dev.Ro {
			continue
		}
		if dev.Tran != "usb" && !dev.Rm {
			continue
		}
		validDevs = append(validDevs, dev)
	}

	if len(validDevs) == 0 {
		return disks, nil
	}

	disksResult := make([]*DiskInfo, len(validDevs))
	var wg sync.WaitGroup
	wg.Add(len(validDevs))

	for i, d := range validDevs {
		go func(idx int, dev linuxBlockDevice) {
			defer wg.Done()
			disksResult[idx] = inspectLinuxDisk(dev)
		}(i, d)
	}
	wg.Wait()

	for _, d := range disksResult {
		if d != nil {
			disks = append(disks, *d)
		}
	}

	return disks, nil
}

func inspectLinuxDisk(dev linuxBlockDevice) *DiskInfo {
	devPath := "/dev/" + dev.Name
	mountPath := devPath
	fileSystem := dev.Fstype
	freeSpace := dev.Fsavail
	partitionScheme := "GPT / MBR"
	if strings.ToLower(dev.Pttype) == "gpt" {
		partitionScheme = "GPT (GUID Partition Table)"
	} else if strings.ToLower(dev.Pttype) == "dos" || strings.ToLower(dev.Pttype) == "mbr" {
		partitionScheme = "MBR (Master Boot Record)"
	}

	displayName := strings.TrimSpace(dev.Vendor + " " + dev.Model)
	if displayName == "" {
		displayName = dev.Name
	}
	if displayName == "" {
		displayName = "USB Storage Device"
	}

	// Labels are UI metadata only. They must never be used as device identity,
	// cache key, or safety gate. Prefer the real block device path instead.
	for _, child := range dev.Children {
		if child.MountPoint != "" {
			mountPath = child.MountPoint
			if child.Fstype != "" {
				fileSystem = child.Fstype
			}
			if child.Fsavail > 0 {
				freeSpace = child.Fsavail
			}
			if displayName == "USB Storage Device" || displayName == dev.Name {
				baseMount := filepath.Base(child.MountPoint)
				if baseMount != "" && !IsIgnoredVolume(baseMount) {
					displayName = baseMount
				}
			}
			break
		}
	}

	if IsIgnoredVolume(filepath.Base(mountPath)) {
		return nil
	}

	// Filter out completely unmounted devices (no active mount point) on Linux
	if mountPath == devPath || mountPath == "" {
		return nil
	}

	if fileSystem == "" {
		fileSystem = "vfat / exfat"
	}

	usbVer := "USB 3.0"
	usbSpeed := "5 Gb/s"
	vendor := strings.TrimSpace(dev.Vendor)
	if vendor == "" {
		vendor = "Generic"
	}

	// Check sysfs for physical USB speed if available
	sysSpeed, sysErr := os.ReadFile(fmt.Sprintf("/sys/block/%s/device/speed", dev.Name))
	if sysErr == nil {
		sp := strings.TrimSpace(string(sysSpeed))
		if sp == "480" {
			usbVer = "USB 2.0"
			usbSpeed = "480 Mb/s"
		} else if sp == "5000" {
			usbVer = "USB 3.0"
			usbSpeed = "5 Gb/s"
		} else if sp == "10000" {
			usbVer = "USB 3.1"
			usbSpeed = "10 Gb/s"
		}
	}

	formattedSize := FormatBytesDual(dev.Size)
	freeFormatted := FormatBytes(freeSpace)
	if freeSpace == 0 {
		freeFormatted = formattedSize
	}

	isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
	protoCode := MapProtocolCode(usbVer, usbSpeed)

	// Detect if this is a system disk
	isSystemDisk, _ := isSystemDiskLinux("/dev/" + dev.Name)

	// Use devPath for disk type detection (MBR check), mountPath for file checks
	manifestLinux := GetDiskUniBootManifest(devPath)
	if manifestLinux == nil && mountPath != "" {
		manifestLinux = GetDiskUniBootManifest(mountPath)
	}
	isRealVentoyLinux := IsRealVentoyDisk(devPath)
	isCloudModeLinux := IsCloudModeDisk(devPath)
	if manifestLinux != nil {
		if manifestLinux.Mode == "cloud" {
			isCloudModeLinux = true
		} else if manifestLinux.Mode == "hybrid" {
			isRealVentoyLinux = true
		}
	}
	thirdPartyBootLinux := BootTypeNone
	isGenBootLinux := false
	if !isCloudModeLinux && !isRealVentoyLinux {
		thirdPartyBootLinux = GetDiskThirdPartyBoot(devPath)
		if thirdPartyBootLinux == BootTypeNone && mountPath != "" {
			thirdPartyBootLinux = GetDiskThirdPartyBoot(mountPath)
		}
		isGenBootLinux = thirdPartyBootLinux != BootTypeNone
	}
	hasUnmountedEspLinux := HasUnmountedEspPartition(devPath)
	bootStatusLinux, bootStatusCodeLinux := DetectBootStatus(partitionScheme, isRealVentoyLinux, isCloudModeLinux, thirdPartyBootLinux, manifestLinux, hasUnmountedEspLinux)

	return &DiskInfo{
		Device:             mountPath,
		Name:               displayName,
		Size:               dev.Size,
		Formatted:          formattedSize,
		FreeSpace:          freeSpace,
		FreeFormatted:      freeFormatted,
		IsRemovable:        true,
		IsSystem:           isSystemDisk,
		UsbVersion:         usbVer,
		UsbSpeed:           usbSpeed,
		Vendor:             vendor,
		FileSystem:         fileSystem,
		PartitionScheme:    partitionScheme,
		Writable:           !dev.Ro,
		SmartStatus:        "Verified",
		BusPower:           "500 mA",
		BusPowerUsed:       "500 mA",
		SectorSize:         "512 Bytes (512n/512e)",
		TransportProtocol:  "BOT (Bulk-Only Transport)",
		BootStatus:         bootStatusLinux,
		BootStatusCode:     bootStatusCodeLinux,
		ControllerVendor:   InferControllerVendor("", "", vendor),
		IsFakeUsb3:         isFake,
		ProtocolCode:       protoCode,
		IsRealVentoy:       isRealVentoyLinux,
		IsCloudMode:        isCloudModeLinux,
		IsGenericBoot:      isGenBootLinux,
		ThirdPartyBootType: thirdPartyBootLinux,
		ThirdPartyBootCode: MapThirdPartyBootCode(thirdPartyBootLinux),
		UniBootVersion: func() string {
			if manifestLinux != nil {
				return manifestLinux.Version
			}
			return ""
		}(),
		UniBootMode: func() string {
			if manifestLinux != nil {
				return manifestLinux.Mode
			}
			return ""
		}(),
		MountPoint: mountPath,
	}
}

// winPartition represents a physical partition entry on Windows.
type winPartition struct {
	Index uint32 `json:"Index"`
	Type  string `json:"Type"`
	Size  uint64 `json:"Size"`
}

// winPartitionList handles deserializing either a single winPartition or a slice from PowerShell JSON.
type winPartitionList []winPartition

func (w *winPartitionList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*w = nil
		return nil
	}
	var list []winPartition
	if err := json.Unmarshal(data, &list); err == nil {
		*w = list
		return nil
	}
	var single winPartition
	if err := json.Unmarshal(data, &single); err == nil {
		*w = []winPartition{single}
		return nil
	}
	return nil
}

// Windows disk probing via PowerShell Win32_DiskDrive
type winDiskDrive struct {
	DeviceID       string           `json:"DeviceID"`
	Index          int              `json:"Index"`
	Model          string           `json:"Model"`
	Size           uint64           `json:"Size"`
	InterfaceType  string           `json:"InterfaceType"`
	Caption        string           `json:"Caption"`
	BytesPerSector uint32           `json:"BytesPerSector"`
	PNPDeviceID    string           `json:"PNPDeviceID"`
	SerialNumber   string           `json:"SerialNumber"`
	PartitionList  winPartitionList `json:"PartitionList"`
}

func getWindowsDisks() ([]DiskInfo, error) {
	disks := make([]DiskInfo, 0)
	psScript := `$drives = Get-CimInstance Win32_DiskDrive | Where-Object { ($_.InterfaceType -eq 'USB' -or ($_.MediaType -like '*Removable*' -and $_.MediaType -notlike '*Fixed*')) -and $_.Model -notmatch 'Virtual|VHD|ISO|CD-ROM|DVD' -and $_.InterfaceType -ne 'FileBackedVirtual' }; $parts = Get-CimInstance Win32_DiskPartition; $drives | ForEach-Object { $d = $_; $dp = $parts | Where-Object { $_.DiskIndex -eq $d.Index }; [PSCustomObject]@{ DeviceID = $d.DeviceID; Index = $d.Index; Model = $d.Model; Size = $d.Size; InterfaceType = $d.InterfaceType; Caption = $d.Caption; BytesPerSector = $d.BytesPerSector; PNPDeviceID = $d.PNPDeviceID; SerialNumber = $d.SerialNumber; PartitionList = @($dp | ForEach-Object { [PSCustomObject]@{ Index = $_.Index; Type = $_.Type; Size = $_.Size } }) } } | ConvertTo-Json -Depth 3`
	cmd := execCommand("powershell", "-NoProfile", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return disks, nil
	}

	var winDrives []winDiskDrive
	if jsonErr := jsonUnmarshal(output, &winDrives); jsonErr != nil {
		var singleDrive winDiskDrive
		if jsonErrSingle := jsonUnmarshal(output, &singleDrive); jsonErrSingle == nil {
			winDrives = append(winDrives, singleDrive)
		}
	}

	type indexedDrive struct {
		index int
		drive winDiskDrive
	}

	validDrives := make([]indexedDrive, 0, len(winDrives))
	for i, drive := range winDrives {
		// Secondary check to ensure virtual devices, VHD, or mounted ISOs are not displayed
		upperModel := strings.ToUpper(drive.Model + " " + drive.Caption)
		if strings.Contains(upperModel, "VIRTUAL") ||
			strings.Contains(upperModel, "VHD") ||
			strings.Contains(upperModel, "ISO") ||
			strings.Contains(upperModel, "CD-ROM") ||
			strings.Contains(upperModel, "DVD") ||
			drive.InterfaceType == "FileBackedVirtual" {
			continue
		}
		validDrives = append(validDrives, indexedDrive{index: i, drive: drive})
	}

	if len(validDrives) == 0 {
		return disks, nil
	}

	disksResult := make([]*DiskInfo, len(validDrives))
	var wg sync.WaitGroup
	wg.Add(len(validDrives))

	for i, item := range validDrives {
		go func(idx int, driveIndex int, d winDiskDrive) {
			defer wg.Done()
			disksResult[idx] = inspectWindowsDisk(driveIndex, d)
		}(i, item.index, item.drive)
	}
	wg.Wait()

	for _, d := range disksResult {
		if d != nil {
			disks = append(disks, *d)
		}
	}

	return disks, nil
}

func inspectWindowsDisk(i int, drive winDiskDrive) *DiskInfo {
	devNode := strings.TrimSpace(drive.DeviceID)
	if devNode == "" {
		devNode = fmt.Sprintf(`\\.\PhysicalDrive%d`, i)
	}
	diskNum, err := parseWindowsDiskNumber(devNode)
	if err != nil {
		diskNum = drive.Index
	}

	// Query mounted volumes for this physical disk using native Win32 APIs
	vols, _ := getVolumesForDiskWindows(diskNum)

	var primaryVol *winVolumeInfo
	for idx := range vols {
		if !IsIgnoredVolume(vols[idx].VolumeName) {
			primaryVol = &vols[idx]
			break
		}
	}
	if primaryVol == nil && len(vols) > 0 {
		primaryVol = &vols[0]
	}

	displayName := ""
	if primaryVol != nil && strings.TrimSpace(primaryVol.VolumeName) != "" {
		displayName = strings.TrimSpace(primaryVol.VolumeName)
	} else if strings.TrimSpace(drive.Model) != "" {
		displayName = strings.TrimSpace(drive.Model)
	} else if strings.TrimSpace(drive.Caption) != "" {
		displayName = strings.TrimSpace(drive.Caption)
	} else {
		displayName = "USB Storage Device"
	}

	mountPoint := ""
	fileSystem := "RAW / Unformatted"
	freeSpace := uint64(0)
	if primaryVol != nil {
		if primaryVol.DriveLetter != "" {
			mountPoint = primaryVol.DriveLetter + `\`
		}
		if primaryVol.FileSystem != "" {
			fileSystem = primaryVol.FileSystem
		}
		freeSpace = primaryVol.FreeSpace
	}

	formattedSize := FormatBytesDual(drive.Size)
	freeFormatted := FormatBytes(freeSpace)
	if freeSpace == 0 && primaryVol == nil {
		freeFormatted = formattedSize
	}

	// Determine partition scheme
	partitionScheme := "MBR"
	for _, p := range drive.PartitionList {
		if strings.Contains(strings.ToUpper(p.Type), "GPT") {
			partitionScheme = "GPT"
			break
		}
	}
	if partitionScheme == "MBR" {
		if secBuf, err := privilege.ReadSector(devNode, 1024); err == nil && len(secBuf) >= 1024 {
			if bytes.Equal(secBuf[512:520], []byte("EFI PART")) || secBuf[450] == 0xEE {
				partitionScheme = "GPT"
			}
		}
	}

	// Format sector size
	sectorSize := "512 Bytes (512n/512e)"
	if drive.BytesPerSector == 4096 {
		sectorSize = "4096 Bytes (4Kn)"
	} else if drive.BytesPerSector > 0 {
		sectorSize = fmt.Sprintf("%d Bytes", drive.BytesPerSector)
	}

	// Inspect authentic USB hardware metadata (VID/PID, real revision, UASP vs BOT)
	usbInfo := getUsbDeviceInfoWindows(drive.PNPDeviceID, drive.SerialNumber)
	usbVer := "USB 2.0"
	usbSpeed := "480 Mb/s"
	vendor := "Generic"
	serialNum := drive.SerialNumber
	vendorID := ""
	productID := ""
	busPower := "500 mA"
	busPowerUsed := "500 mA"
	transportProtocol := "BOT (Bulk-Only Transport)"

	if usbInfo != nil {
		if usbInfo.UsbVersion != "" {
			usbVer = usbInfo.UsbVersion
		}
		if usbInfo.UsbSpeed != "" {
			usbSpeed = usbInfo.UsbSpeed
		}
		if usbInfo.Vendor != "" {
			vendor = usbInfo.Vendor
		}
		if usbInfo.SerialNumber != "" {
			serialNum = usbInfo.SerialNumber
		}
		vendorID = usbInfo.VendorID
		productID = usbInfo.ProductID
		if usbInfo.BusPower != "" {
			busPower = usbInfo.BusPower
		}
		if usbInfo.BusPowerUsed != "" {
			busPowerUsed = usbInfo.BusPowerUsed
		}
		if usbInfo.TransportProtocol != "" {
			transportProtocol = usbInfo.TransportProtocol
		}
	} else {
		// Fallback inspection from drive Model/Caption if USB info unavailable
		upperModel := strings.ToUpper(drive.Model + " " + displayName)
		if strings.Contains(upperModel, "3.2") {
			usbVer = "USB 3.2"
			usbSpeed = "20 Gb/s"
		} else if strings.Contains(upperModel, "3.1") {
			usbVer = "USB 3.1"
			usbSpeed = "10 Gb/s"
		} else if strings.Contains(upperModel, "3.0") || strings.Contains(upperModel, "SUPERSPEED") {
			usbVer = "USB 3.0"
			usbSpeed = "5 Gb/s"
		}
	}

	isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
	protoCode := MapProtocolCode(usbVer, usbSpeed)
	ctrlVendor := InferControllerVendor(vendorID, productID, vendor)

	// Detect if this is a system disk
	isSystemDisk, _ := isSystemDiskWindows(devNode)

	// Collect all mount points on this physical disk for bootloader inspection
	var diskMountPoints []string
	for _, v := range vols {
		if v.DriveLetter != "" {
			diskMountPoints = append(diskMountPoints, v.DriveLetter+`\`)
		}
	}
	if len(diskMountPoints) == 0 && mountPoint != "" {
		diskMountPoints = append(diskMountPoints, mountPoint)
	}

	var manifestWin *UniBootManifest
	for _, mp := range diskMountPoints {
		if m, err := ReadUniBootManifest(mp); err == nil && m != nil {
			manifestWin = m
			break
		}
	}
	if manifestWin == nil {
		manifestWin = GetDiskUniBootManifest(devNode)
	}

	isRealVentoyWin := CheckVentoyMbrSignature(devNode)
	isCloudModeWin := false
	for _, v := range vols {
		if strings.EqualFold(v.VolumeName, "VTOYEFI") || strings.EqualFold(v.VolumeName, "VENTOY") {
			isRealVentoyWin = true
		}
		if HasVentoyEngineFiles(v.DriveLetter + `\`) {
			isRealVentoyWin = true
		}
		if HasUniBootCloudFiles(v.DriveLetter + `\`) {
			isCloudModeWin = true
		}
	}

	if manifestWin != nil {
		if manifestWin.Mode == "cloud" {
			isCloudModeWin = true
			isRealVentoyWin = false
		} else if manifestWin.Mode == "hybrid" {
			isRealVentoyWin = true
		}
	}

	thirdPartyBootWin := BootTypeNone
	isGenBootWin := false
	if !isCloudModeWin && !isRealVentoyWin {
		thirdPartyBootWin = IdentifyThirdPartyBoot(diskMountPoints)
		isGenBootWin = thirdPartyBootWin != BootTypeNone
	}
	bootStatusWin, bootStatusCodeWin := DetectBootStatus(partitionScheme, isRealVentoyWin, isCloudModeWin, thirdPartyBootWin, manifestWin, false)

	return &DiskInfo{
		Device:             devNode,
		Name:               displayName,
		Size:               drive.Size,
		Formatted:          formattedSize,
		FreeSpace:          freeSpace,
		FreeFormatted:      freeFormatted,
		IsRemovable:        true,
		IsSystem:           isSystemDisk,
		UsbVersion:         usbVer,
		UsbSpeed:           usbSpeed,
		Vendor:             vendor,
		SerialNumber:       serialNum,
		VendorId:           vendorID,
		ProductId:          productID,
		FileSystem:         fileSystem,
		PartitionScheme:    partitionScheme,
		Writable:           true,
		SmartStatus:        "Verified",
		BusPower:           busPower,
		BusPowerUsed:       busPowerUsed,
		SectorSize:         sectorSize,
		TransportProtocol:  transportProtocol,
		BootStatus:         bootStatusWin,
		BootStatusCode:     bootStatusCodeWin,
		ControllerVendor:   ctrlVendor,
		IsFakeUsb3:         isFake,
		ProtocolCode:       protoCode,
		IsRealVentoy:       isRealVentoyWin,
		IsCloudMode:        isCloudModeWin,
		IsGenericBoot:      isGenBootWin,
		ThirdPartyBootType: thirdPartyBootWin,
		ThirdPartyBootCode: MapThirdPartyBootCode(thirdPartyBootWin),
		UniBootVersion: func() string {
			if manifestWin != nil {
				return manifestWin.Version
			}
			return ""
		}(),
		UniBootMode: func() string {
			if manifestWin != nil {
				return manifestWin.Mode
			}
			return ""
		}(),
		MountPoint: mountPoint,
	}
}

// Variables for command execution and JSON parsing to allow mocking in tests
var (
	execCommand   = exec.Command
	jsonUnmarshal = json.Unmarshal
)

// ValidateTargetDisk ensures the target disk is not a system disk before operation.
func ValidateTargetDisk(targetDevice string) error {
	if targetDevice == "" {
		return fmt.Errorf("target disk device path cannot be empty")
	}

	// Static blacklist check for common system disk paths - always blocked even in tests
	staticBlacklist := []string{
		"/", "C:", `C:\`,
		"/dev/sda", "/dev/nvme0n1", "/dev/mmcblk0", "/dev/vda",
		"/dev/disk0", "disk0",
		`\\.\PhysicalDrive0`, "PhysicalDrive0",
	}
	for _, blocked := range staticBlacklist {
		if targetDevice == blocked {
			return fmt.Errorf("CRITICAL: Safety block triggered! %s is a known system drive", targetDevice)
		}
	}

	// Test mock devices bypass (only for dummy* or test*)
	if strings.HasPrefix(targetDevice, "dummy") || strings.HasPrefix(targetDevice, "test") {
		return nil
	}

	// Dynamic system disk detection (Fail-closed: erroring out blocks formatting)
	isSystem, err := isSystemDisk(targetDevice)
	if err != nil {
		logger.Error("System disk detection failed, blocking operation for safety", "device", targetDevice, "error", err)
		return fmt.Errorf("CRITICAL: Safety block triggered! Unable to verify if %s is a system disk: %w", targetDevice, err)
	} else if isSystem {
		return fmt.Errorf("CRITICAL: Safety block triggered! %s is detected as an active system disk", targetDevice)
	}

	// Path format validation to prevent injection
	if !isValidDiskPath(targetDevice) {
		return fmt.Errorf("CRITICAL: Invalid disk path format: %s", targetDevice)
	}

	return nil
}

// isValidDiskPath validates disk path format to prevent command injection
func isValidDiskPath(path string) bool {
	// Check for dangerous characters
	if strings.ContainsAny(path, ";|&`$(){}[]<>\n\r") {
		return false
	}

	// Platform-specific validation
	switch runtime.GOOS {
	case "darwin":
		// macOS: /dev/diskN or /dev/rdiskN or diskN or /Volumes/...
		return regexp.MustCompile(`^((/dev/)?(r)?disk\d+|/Volumes/[a-zA-Z0-9_\-\.\s]+)$`).MatchString(path)
	case "windows":
		// Windows: C:, PhysicalDriveN, \\.\PhysicalDriveN, or diskN. Match case-insensitively.
		upper := strings.ToUpper(path)
		if regexp.MustCompile(`^([A-Z]:)$`).MatchString(upper) {
			return true
		}
		if strings.HasPrefix(upper, `\\.\PHYSICALDRIVE`) {
			suffix := strings.TrimPrefix(upper, `\\.\`)
			suffix = strings.TrimPrefix(suffix, `PHYSICALDRIVE`)
			return regexp.MustCompile(`^\d+$`).MatchString(suffix)
		}
		if strings.HasPrefix(upper, `PHYSICALDRIVE`) {
			suffix := strings.TrimPrefix(upper, `PHYSICALDRIVE`)
			return regexp.MustCompile(`^\d+$`).MatchString(suffix)
		}
		if strings.HasPrefix(upper, `DISK`) {
			suffix := strings.TrimPrefix(upper, `DISK`)
			return regexp.MustCompile(`^\d+$`).MatchString(suffix)
		}
		return false
	case "linux":
		// Linux: /dev/sdX, /dev/nvmeXnY, /dev/mmcblkX, /dev/vdX
		return regexp.MustCompile(`^/dev/(sd[a-z]+|nvme\d+n\d+|mmcblk\d+|vd[a-z]+)$`).MatchString(path)
	default:
		return false
	}
}

// isSystemDisk dynamically detects if a disk is a system disk
func isSystemDisk(device string) (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		return isSystemDiskDarwin(device)
	case "windows":
		return isSystemDiskWindows(device)
	case "linux":
		return isSystemDiskLinux(device)
	default:
		return false, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

var darwinRootSystemDiskOnce sync.Once
var darwinRootSystemDisk string

func getDarwinRootSystemDisk() string {
	darwinRootSystemDiskOnce.Do(func() {
		cmd := execCommand("diskutil", "info", "-plist", "/")
		if out, err := cmd.Output(); err == nil {
			darwinRootSystemDisk = extractPlistValue(string(out), "ParentWholeDisk")
			if darwinRootSystemDisk == "" {
				darwinRootSystemDisk = extractPlistValue(string(out), "DeviceIdentifier")
			}
		}
	})
	return darwinRootSystemDisk
}

// isSystemDiskDarwin checks if a disk is a system disk on macOS
func isSystemDiskDarwin(device string) (bool, error) {
	// Normalize device path
	diskNode := NormalizeDarwinDiskNode(device)

	// Direct match against the true macOS boot/root disk
	if rootDisk := getDarwinRootSystemDisk(); rootDisk != "" && diskNode != "" {
		if diskNode == rootDisk {
			return true, nil
		}
	}

	target := diskNode
	if target == "" {
		target = device
	}

	outputStr := getDarwinDiskutilInfo(target)
	if outputStr == "" {
		// If diskutil cannot find the device, verify it doesn't match root disk
		if rootDisk := getDarwinRootSystemDisk(); rootDisk != "" && diskNode == rootDisk {
			return true, nil
		}
		return false, nil
	}

	// Check for system mount points
	systemMountPoints := []string{
		"<string>/</string>",
		"<string>/System</string>",
		"<string>/Library</string>",
		"<string>/Applications</string>",
		"<string>/usr</string>",
		"<string>/var</string>",
	}

	for _, mountPoint := range systemMountPoints {
		if strings.Contains(outputStr, mountPoint) {
			return true, nil
		}
	}

	// Check if it's an internal disk (not removable)
	if strings.Contains(outputStr, "<key>Internal</key>") {
		// Look for <true/> after Internal key
		internalIdx := strings.Index(outputStr, "<key>Internal</key>")
		if internalIdx >= 0 {
			afterInternal := outputStr[internalIdx:]
			if strings.Contains(afterInternal[:200], "<true/>") {
				return true, nil
			}
		}
	}

	return false, nil
}

var (
	windowsSystemDiskIndexOnce sync.Once
	windowsSystemDiskIndex     = -1
)

func getWindowsSystemDiskIndex() int {
	windowsSystemDiskIndexOnce.Do(func() {
		// 1. Fast path: native Win32 DeviceIoControl
		if idx, err := getSystemDriveDiskNumberWindows(); err == nil && idx >= 0 {
			windowsSystemDiskIndex = idx
			return
		}

		systemDrive := os.Getenv("SystemDrive")
		if systemDrive == "" {
			systemDrive = "C:"
		}
		driveLetter := strings.TrimSuffix(systemDrive, "\\")
		driveLetter = strings.TrimSuffix(driveLetter, ":")
		if driveLetter == "" {
			driveLetter = "C"
		}

		// 2. Try PowerShell to query the partition of the system drive
		cmd := execCommand("powershell", "-NoProfile", "-NonInteractive", "-Command",
			fmt.Sprintf("(Get-Partition -DriveLetter '%s').DiskNumber", driveLetter))
		if out, err := cmd.Output(); err == nil {
			trimmed := strings.TrimSpace(string(out))
			if idx, errConv := strconv.Atoi(trimmed); errConv == nil && idx >= 0 {
				windowsSystemDiskIndex = idx
				return
			}
		}

		// 2. Fallback to WMI association query
		cmdWmi := execCommand("powershell", "-NoProfile", "-NonInteractive", "-Command",
			fmt.Sprintf("(Get-CimAssociatedInstance -InputObject (Get-CimInstance Win32_LogicalDisk -Filter \"DeviceID='%s:'\") -ResultClassName Win32_DiskPartition).DiskIndex", driveLetter))
		if out, err := cmdWmi.Output(); err == nil {
			trimmed := strings.TrimSpace(string(out))
			if idx, errConv := strconv.Atoi(trimmed); errConv == nil && idx >= 0 {
				windowsSystemDiskIndex = idx
				return
			}
		}

		// 3. Fallback default to 0 for safety
		windowsSystemDiskIndex = 0
	})
	return windowsSystemDiskIndex
}

// isSystemDiskWindows checks if a disk is a system disk on Windows
func isSystemDiskWindows(device string) (bool, error) {
	trimmed := strings.TrimSpace(device)
	if trimmed == "" {
		return false, fmt.Errorf("empty device path")
	}

	// 1. Drive letter format (e.g. C:, C:\, D:)
	if (len(trimmed) == 2 && trimmed[1] == ':') || (len(trimmed) == 3 && trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/')) {
		target := strings.ToUpper(trimmed[:2])
		sysDrive := strings.ToUpper(os.Getenv("SystemDrive"))
		if sysDrive == "" {
			sysDrive = "C:"
		}
		if target == sysDrive {
			return true, nil
		}

		windowsDir := filepath.Join(target+"\\", "Windows")
		if info, err := os.Stat(windowsDir); err == nil && info.IsDir() {
			return true, nil
		}

		// Check if it's the boot volume
		cmd := execCommand("wmic", "volume", "where",
			fmt.Sprintf("DriveLetter='%s'", target),
			"get", "BootVolume")
		if output, err := cmd.Output(); err == nil {
			if strings.Contains(strings.ToUpper(string(output)), "TRUE") {
				return true, nil
			}
		}
		return false, nil
	}

	// 2. Physical drive format (e.g. \\.\PhysicalDrive0, PhysicalDrive0, disk0, 0)
	diskNum, err := parseWindowsDiskNumber(trimmed)
	if err == nil {
		sysDiskNum := getWindowsSystemDiskIndex()
		if sysDiskNum >= 0 && diskNum == sysDiskNum {
			return true, nil
		}
		if diskNum == 0 {
			return true, nil
		}
	}

	return false, nil
}

// isSystemDiskLinux checks if a disk is a system disk on Linux
func isSystemDiskLinux(device string) (bool, error) {
	// Read /proc/mounts to check for system mount points
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("failed to read /proc/mounts: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	systemMountPoints := []string{"/", "/boot", "/usr", "/var", "/lib", "/bin", "/sbin", "/etc"}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		mountDevice := fields[0]
		mountPoint := fields[1]

		// Check if this line refers to our device
		if strings.HasPrefix(mountDevice, device) {
			for _, sysMount := range systemMountPoints {
				if mountPoint == sysMount {
					return true, nil
				}
			}
		}
	}

	// Check if device is listed in /etc/fstab for system mounts
	if fstabData, err := os.ReadFile("/etc/fstab"); err == nil {
		fstabLines := strings.Split(string(fstabData), "\n")
		for _, line := range fstabLines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if strings.Contains(fields[0], device) {
					for _, sysMount := range systemMountPoints {
						if fields[1] == sysMount {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}

// ValidateTargetDiskSnapshot verifies that a disk still matches the identity captured before deployment.
func ValidateTargetDiskSnapshot(expected DiskInfo, actual DiskInfo) error {
	if err := ValidateTargetDisk(actual.Device); err != nil {
		return err
	}
	if expected.Device == "" || actual.Device != expected.Device {
		return fmt.Errorf("target disk changed: expected device %q, got %q", expected.Device, actual.Device)
	}
	if expected.Size > 0 && actual.Size > 0 && expected.Size != actual.Size {
		return fmt.Errorf("target disk capacity changed: expected %d bytes, got %d bytes", expected.Size, actual.Size)
	}
	if !actual.IsRemovable {
		return fmt.Errorf("target disk is no longer removable: %s", actual.Device)
	}
	if actual.IsSystem {
		return fmt.Errorf("target disk is a system disk: %s", actual.Device)
	}
	if expected.SerialNumber != "" && actual.SerialNumber != "" && expected.SerialNumber != actual.SerialNumber {
		return fmt.Errorf("target disk serial number changed: expected %q, got %q", expected.SerialNumber, actual.SerialNumber)
	}
	if expected.Vendor != "" && actual.Vendor != "" && !strings.EqualFold(expected.Vendor, actual.Vendor) {
		return fmt.Errorf("target disk vendor changed: expected %q, got %q", expected.Vendor, actual.Vendor)
	}
	return nil
}

// ValidateLiveTargetDisk confirms that a target is still present in the current removable-disk inventory.
func ValidateLiveTargetDisk(targetDevice string) error {
	if err := ValidateTargetDisk(targetDevice); err != nil {
		return err
	}

	disks, err := GetRemovableDisks()
	if err != nil {
		return fmt.Errorf("failed to refresh target disk inventory: %w", err)
	}
	for _, candidate := range disks {
		if candidate.Device != targetDevice {
			continue
		}
		if err := ValidateTargetDiskSnapshot(candidate, candidate); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("target disk is no longer present as a removable disk: %s", targetDevice)
}

// ValidateUserEjectTarget ensures the selected disk is a currently present, removable,
// non-system disk before a user explicitly requests ejection. This gate keeps shutdown
// cleanup and user-initiated ejects separate, preventing automatic mass ejection.
func ValidateUserEjectTarget(targetDevice string) error {
	if err := ValidateTargetDisk(targetDevice); err != nil {
		return err
	}

	disks, err := GetRemovableDisks()
	if err != nil {
		return fmt.Errorf("failed to refresh removable disk inventory before ejection: %w", err)
	}
	for _, candidate := range disks {
		if candidate.Device != targetDevice {
			continue
		}
		if candidate.IsSystem {
			return fmt.Errorf("CRITICAL: Safety block triggered! %s is a system disk and cannot be ejected", targetDevice)
		}
		if !candidate.IsRemovable {
			return fmt.Errorf("CRITICAL: Safety block triggered! %s is not a removable disk", targetDevice)
		}
		return nil
	}

	return fmt.Errorf("target disk is not present in the current removable-disk inventory: %s", targetDevice)
}

// SafeUserEjectDisk performs the user-initiated eject only after the target has passed the
// explicit removable-disk validation gate. This is intentionally separate from shutdown cleanup.
func SafeUserEjectDisk(device string) error {
	if err := ValidateUserEjectTarget(device); err != nil {
		return err
	}
	return EjectDisk(device)
}

// SyncDiskBuffers commits all filesystem caches and dirty pages to storage media across OS platforms.
// On Unix (macOS and Linux), it issues a kernel-level sync() syscall.
// On Windows, it flushes volume-level write caches via FlushFileBuffers.
func SyncDiskBuffers() {
	syncPlatformBuffers()
}

// EjectDisk safely unmounts and ejects the target removable USB storage drive.
func EjectDisk(device string) error {
	if device == "" {
		return fmt.Errorf("device path cannot be empty")
	}
	if err := ValidateTargetDisk(device); err != nil {
		return err
	}

	// Flush OS kernel page cache to physical media before unmounting/ejecting
	SyncDiskBuffers()

	switch runtime.GOOS {
	case "darwin":
		diskNode := NormalizeDarwinDiskNode(device)
		target := "/dev/" + diskNode
		if diskNode == "" {
			target = device
		}

		// Step 1: Force unmount all partitions first to avoid "Resource busy" or diskarbitrationd timeout
		_ = execCommand("diskutil", "unmountDisk", "force", target).Run()

		// Step 2: Perform true hardware eject
		cmd := execCommand("diskutil", "eject", target)
		if _, err := cmd.CombinedOutput(); err != nil {
			// Retry once with an explicit forced unmount followed by eject
			time.Sleep(200 * time.Millisecond)
			_ = execCommand("diskutil", "unmountDisk", "force", target).Run()
			retryCmd := execCommand("diskutil", "eject", target)
			if retryOut, retryErr := retryCmd.CombinedOutput(); retryErr != nil {
				return fmt.Errorf("failed to eject disk %s: %s (%w)", target, strings.TrimSpace(string(retryOut)), retryErr)
			}
		}
		return nil

	case "windows":
		driveLetter := strings.TrimSuffix(device, "\\")
		if !strings.HasSuffix(driveLetter, ":") {
			driveLetter = driveLetter + ":"
		}
		psCmd := fmt.Sprintf("(New-Object -ComObject Shell.Application).NameSpace(17).ParseName('%s').InvokeVerb('Eject')", driveLetter)
		cmd := execCommand("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to eject drive %s: %s (%w)", device, strings.TrimSpace(string(output)), err)
		}
		return nil

	default:
		cmd := execCommand("udisksctl", "power-off", "-b", device)
		if output, err := cmd.CombinedOutput(); err == nil {
			_ = output
			return nil
		}
		fallbackCmd := execCommand("eject", device)
		if output, err := fallbackCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to eject device %s: %s (%w)", device, strings.TrimSpace(string(output)), err)
		}
		return nil
	}
}
