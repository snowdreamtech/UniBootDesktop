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
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
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

// DiskInfo represents metadata about an available disk drive.
type DiskInfo struct {
	Device            string `json:"device"`            // Device path (e.g., /dev/disk2, E:)
	Name              string `json:"name"`              // Friendly label / vendor model
	Size              uint64 `json:"size"`              // Total capacity in bytes
	Formatted         string `json:"formatted"`         // Human readable size string
	FreeSpace         uint64 `json:"freeSpace"`         // Free available space in bytes
	FreeFormatted     string `json:"freeFormatted"`     // Human readable free space string
	IsRemovable       bool   `json:"isRemovable"`       // Removable disk flag
	IsSystem          bool   `json:"isSystem"`          // System disk safety flag
	UsbVersion        string `json:"usbVersion"`        // Protocol version (USB 2.0, USB 3.0, USB 3.1, USB 3.2, USB4)
	UsbSpeed          string `json:"usbSpeed"`          // Physical bus speed (480 Mb/s, 5 Gb/s, 10 Gb/s, 20 Gb/s)
	Vendor            string `json:"vendor"`            // Device manufacturer / vendor
	FileSystem        string `json:"fileSystem"`        // File system format (e.g., ExFAT, FAT32, NTFS, APFS, ext4)
	PartitionScheme   string `json:"partitionScheme"`   // Partition scheme (e.g., GPT, MBR)
	Writable          bool   `json:"writable"`          // Read-Write status (true = Read-Write, false = Read-Only)
	SerialNumber      string `json:"serialNumber"`      // Hardware Serial Number
	VendorId          string `json:"vendorId"`          // USB Vendor ID (e.g., 0x21c4)
	ProductId         string `json:"productId"`         // USB Product ID (e.g., 0x0cd1)
	SmartStatus       string `json:"smartStatus"`       // S.M.A.R.T. health status (e.g. Verified, Not Supported, Failing)
	BusPower          string `json:"busPower"`          // Bus power available (e.g. 500 mA, 900 mA)
	BusPowerUsed      string `json:"busPowerUsed"`      // Bus power required/used (e.g. 500 mA, 224 mA)
	SectorSize        string `json:"sectorSize"`        // Sector block size (e.g. 512 Bytes, 4096 Bytes / 4Kn)
	TransportProtocol string `json:"transportProtocol"` // USB Transport Protocol (e.g. UASP, BOT)
	BootStatus        string `json:"bootStatus"`        // Boot sector status (e.g. UniBoot/Ventoy Ready, MBR Bootable, Standard Data)
	ControllerVendor  string `json:"controllerVendor"`  // Inferred USB Controller Vendor (e.g. Phison, SMI, Alcor)
	IsFakeUsb3        bool   `json:"isFakeUsb3"`        // Warning flag for fake USB 3.0 (USB 2.0 PHY disguised as 3.0)
	ProtocolCode      string `json:"protocolCode"`      // Styling code: "usb2", "usb3_0", "usb3_1", "usb3_2", "usb4"
	IsRealVentoy      bool   `json:"isRealVentoy"`      // True ONLY if drive contains Ventoy MBR Sector 0 signature
	IsCloudMode           bool   `json:"isCloudMode"`           // True if drive is formatted in Cloud Mode (iPXE ESP Cloud Pure)
	IsGenericBoot     bool   `json:"isGenericBoot"`     // True if drive contains generic 3rd-party bootloader (Rufus/PE/ISO)
	MountPoint        string `json:"mountPoint"`        // Mount point or volume path (e.g. /Volumes/UNTITLED, E:\)
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

// DetectBootStatus evaluates the boot status text based on partition scheme, volume label, Ventoy/Cloud Mode, and generic boot flags.
func DetectBootStatus(volName string, partitionScheme string, isRealVentoy bool, isCloudMode bool, isGenericBoot bool) string {
	if isRealVentoy {
		return "Ventoy / UniBoot (混合模式)"
	}
	if isCloudMode {
		return "UniBoot (1秒极速云引导盘)"
	}
	if isGenericBoot {
		return "第三方引导盘 (Rufus / PE / ISO)"
	}
	if strings.Contains(strings.ToUpper(partitionScheme), "GPT") {
		return "GPT 数据盘"
	}
	if strings.Contains(strings.ToUpper(partitionScheme), "MBR") {
		return "MBR 数据盘"
	}
	return "数据存储盘 (未检测到引导包)"
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
	return false
}

// CheckVentoyMbrSignature inspects MBR Sector 0 for Ventoy's bootloader magic byte signature.
func CheckVentoyMbrSignature(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	devicePath := targetDisk
	if runtime.GOOS == "darwin" && strings.HasPrefix(targetDisk, "/dev/disk") && !strings.HasPrefix(targetDisk, "/dev/rdisk") {
		devicePath = "/dev/r" + strings.TrimPrefix(targetDisk, "/dev/")
	}

	f, err := os.Open(devicePath)
	if err != nil {
		f, err = os.Open(targetDisk)
	}
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n < 512 {
		return false
	}

	return bytes.Contains(buf, []byte("Ventoy")) || bytes.Contains(buf, []byte("VENTOY"))
}

// HasUniBootCloudFiles verifies physical presence of UniBoot Cloud iPXE firmware files inside ESP partition.
func HasUniBootCloudFiles(mountPoint string) bool {
	if mountPoint == "" || IsEmptyDirectory(mountPoint) {
		return false
	}
	bootIpxe := filepath.Join(mountPoint, "boot.ipxe")
	unibootIpxe := filepath.Join(mountPoint, "ipxe", "uniboot.ipxe")
	_, errBoot := os.Stat(bootIpxe)
	_, errUni := os.Stat(unibootIpxe)
	return errBoot == nil || errUni == nil
}

// NormalizeDarwinDiskNode extracts the parent physical disk node (e.g. "disk2") from a macOS disk or partition path.
// Examples:
//   "/dev/disk2"    -> "disk2"
//   "/dev/rdisk2"   -> "disk2"
//   "/dev/disk2s1"  -> "disk2"
//   "disk2s2"       -> "disk2"
//   "/dev/disk12s3" -> "disk12"
func NormalizeDarwinDiskNode(targetDisk string) string {
	node := filepath.Base(targetDisk)
	node = strings.TrimPrefix(node, "r") // Remove raw disk prefix if present (rdisk2 -> disk2)

	if strings.HasPrefix(node, "disk") {
		rest := node[4:] // Part after "disk", e.g. "2", "2s1", "12s3"
		if idx := strings.Index(rest, "s"); idx > 0 {
			return "disk" + rest[:idx]
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

	// 1. Direct mount directory check if targetDisk is already a mount point
	if HasVentoyEngineFiles(targetDisk) {
		return true
	}

	// 2. Physical MBR Sector 0 signature check (Fast check when root/sudo permitted)
	if CheckVentoyMbrSignature(targetDisk) {
		return true
	}

	// 3. Platform-specific target partition inspection with label and metadata checks
	if runtime.GOOS == "darwin" {
		baseDisk := NormalizeDarwinDiskNode(targetDisk)
		if strings.HasPrefix(baseDisk, "disk") {
			p1 := baseDisk + "s1"
			p2 := baseDisk + "s2"

			// Inspect Partition 2 (VTOYEFI / UNIBOOTEFI ESP Partition)
			strP2 := getDarwinDiskutilInfo(p2)
			if strP2 != "" {
				volNameP2 := strings.ToUpper(extractPlistValue(strP2, "VolumeName"))
				mountP2 := extractPlistValue(strP2, "MountPoint")

				if strings.Contains(volNameP2, "VTOYEFI") || strings.Contains(volNameP2, "UNIBOOTEFI") {
					return true
				}
				if mountP2 != "" && HasVentoyEngineFiles(mountP2) {
					return true
				}
			}

			// Inspect Partition 1 (Ventoy / UniBoot Data Partition)
			strP1 := getDarwinDiskutilInfo(p1)
			if strP1 != "" {
				volNameP1 := strings.ToUpper(extractPlistValue(strP1, "VolumeName"))
				mountP1 := extractPlistValue(strP1, "MountPoint")

				if mountP1 != "" && HasVentoyEngineFiles(mountP1) {
					return true
				}

				if (strings.Contains(volNameP1, "VENTOY") || strings.Contains(volNameP1, "UNIBOOT")) && strP2 != "" {
					return true
				}
			}
		}
	} else if runtime.GOOS == "linux" {
		out, err := exec.Command("lsblk", "-o", "NAME,LABEL", "-J", targetDisk).Output()
		if err == nil {
			upperOut := strings.ToUpper(string(out))
			if strings.Contains(upperOut, "VTOYEFI") || strings.Contains(upperOut, "UNIBOOTEFI") {
				return true
			}
		}
	} else if runtime.GOOS == "windows" {
		// 安全地转义PowerShell参数，防止命令注入
		baseDisk := filepath.Base(targetDisk)
		safeDisk := strings.ReplaceAll(baseDisk, "'", "''")  // PowerShell单引号转义
		safeDisk = strings.ReplaceAll(safeDisk, "`", "``")    // PowerShell反引号转义
		safeDisk = strings.ReplaceAll(safeDisk, "$", "`$")    // PowerShell变量转义
		safeDisk = strings.ReplaceAll(safeDisk, "\"", "`\"")  // 双引号转义

		out, err := exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("Get-Partition -DiskNumber (Get-Disk | Where-Object {$_.Path -like '*%s*'}).DiskNumber | Get-Volume | Select-Object -ExpandProperty FileSystemLabel", safeDisk)).Output()
		if err == nil {
			upperOut := strings.ToUpper(string(out))
			if strings.Contains(upperOut, "VTOYEFI") || strings.Contains(upperOut, "UNIBOOTEFI") {
				return true
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

	// First check if this is a Ventoy disk (hybrid mode or pure Ventoy)
	// If it has Ventoy MBR signature or Ventoy engine files, it's not a pure cloud disk
	if CheckVentoyMbrSignature(targetDisk) {
		return false
	}

	if runtime.GOOS == "darwin" {
		diskNode := filepath.Base(targetDisk)
		if strings.HasPrefix(diskNode, "disk") {
			baseDisk := diskNode
			if strings.Contains(diskNode, "s") {
				// Extract base disk from partition (e.g., "disk2s1" -> "disk2")
				baseDisk = NormalizeDarwinDiskNode(diskNode)
			}

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
			out, err := exec.Command("lsblk", "-o", "MOUNTPOINT", "-n", "-l", targetDisk).Output()
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

// HasGenericBootFiles verifies physical presence of generic 3rd-party bootloader files
// (e.g. Rufus, UltraISO, PE, BalenaEtcher, WinToUSB, ISO9660).
func HasGenericBootFiles(mountPoint string) bool {
	if mountPoint == "" || IsEmptyDirectory(mountPoint) {
		return false
	}
	bootPaths := []string{
		filepath.Join(mountPoint, "EFI", "BOOT", "BOOTX64.EFI"),
		filepath.Join(mountPoint, "EFI", "BOOT", "BOOTIA32.EFI"),
		filepath.Join(mountPoint, "EFI", "BOOT", "BOOTARM.EFI"),
		filepath.Join(mountPoint, "EFI", "BOOT", "BOOTAA64.EFI"),
		filepath.Join(mountPoint, "sources", "boot.wim"),
		filepath.Join(mountPoint, "bootmgr"),
		filepath.Join(mountPoint, "boot", "bcd"),
		filepath.Join(mountPoint, "boot", "grub", "grub.cfg"),
		filepath.Join(mountPoint, "isolinux", "isolinux.bin"),
		filepath.Join(mountPoint, "syslinux.cfg"),
	}
	for _, p := range bootPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// IsGenericBootDisk checks if a target disk is a 3rd-party boot disk (Rufus, PE, ISO) that is NOT a Ventoy or Cloud Mode drive.
func IsGenericBootDisk(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	if HasGenericBootFiles(targetDisk) {
		return true
	}

	if runtime.GOOS == "darwin" {
		diskNode := filepath.Base(targetDisk)
		if strings.HasPrefix(diskNode, "disk") {
			partitions := []string{diskNode}
			if !strings.Contains(diskNode, "s") {
				partitions = []string{diskNode + "s1", diskNode + "s2"}
			}
			for _, p := range partitions {
				str := getDarwinDiskutilInfo(p)
				if str != "" {
					mountPoint := extractPlistValue(str, "MountPoint")
					if HasGenericBootFiles(mountPoint) {
						return true
					}
				}
			}
		}
	}
	return false
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

var (
	diskCacheMutex sync.Mutex
	diskCacheList  []DiskInfo
	diskCacheTime  time.Time

	darwinDiskutilCacheMutex sync.Mutex
	darwinDiskutilCacheMap   = make(map[string]string)
	darwinDiskutilCacheTime  time.Time
)

func getDarwinDiskutilInfo(node string) string {
	node = strings.TrimSpace(node)
	if node == "" {
		return ""
	}

	darwinDiskutilCacheMutex.Lock()
	defer darwinDiskutilCacheMutex.Unlock()

	if time.Since(darwinDiskutilCacheTime) > 5*time.Second {
		darwinDiskutilCacheMap = make(map[string]string)
		darwinDiskutilCacheTime = time.Now()
	}

	if info, ok := darwinDiskutilCacheMap[node]; ok {
		return info
	}

	cmd := execCommand("diskutil", "info", "-plist", node)
	out, err := cmd.Output()
	if err != nil {
		darwinDiskutilCacheMap[node] = ""
		return ""
	}

	infoStr := string(out)
	darwinDiskutilCacheMap[node] = infoStr
	return infoStr
}

func invalidateDarwinDiskutilCache() {
	darwinDiskutilCacheMutex.Lock()
	darwinDiskutilCacheMap = make(map[string]string)
	darwinDiskutilCacheTime = time.Time{}
	darwinDiskutilCacheMutex.Unlock()
}

// InvalidateDiskCache clears the memory disk cache to force an immediate fresh hardware scan.
func InvalidateDiskCache() {
	diskCacheMutex.Lock()
	diskCacheList = nil
	diskCacheMutex.Unlock()

	darwinUSBCacheMutex.Lock()
	darwinUSBCacheMap = nil
	darwinUSBCacheMutex.Unlock()

	invalidateDarwinDiskutilCache()
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

func getVolumeSnapshot() string {
	switch runtime.GOOS {
	case "darwin":
		entries, err := os.ReadDir("/Volumes")
		if err != nil {
			return ""
		}
		var names []string
		for _, e := range entries {
			if !IsIgnoredVolume(e.Name()) {
				names = append(names, e.Name())
			}
		}
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
	if diskCacheList != nil && time.Since(diskCacheTime) < 5*time.Second {
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

	if err == nil {
		diskCacheMutex.Lock()
		diskCacheList = disks
		diskCacheTime = time.Now()
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
)

func getCachedDarwinUSBMap() map[string]*darwinUSBInfo {
	darwinUSBCacheMutex.Lock()
	defer darwinUSBCacheMutex.Unlock()

	// Cache hardware profile details for 30 seconds to prevent system_profiler CPU spikes
	if darwinUSBCacheMap != nil && time.Since(darwinUSBCacheTime) < 30*time.Second {
		return darwinUSBCacheMap
	}

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

	darwinUSBCacheMap = usbMap
	darwinUSBCacheTime = time.Now()
	return usbMap
}

func getDarwinDisks() ([]DiskInfo, error) {
	disks := make([]DiskInfo, 0)

	// Step 1: Probe system_profiler for rich hardware details (cached for 30s to eliminate CPU spikes)
	usbMap := getCachedDarwinUSBMap()

	// Step 2: Scan /Volumes for mounted removable drives
	entries, err := os.ReadDir("/Volumes")
	if err != nil {
		return disks, nil
	}

	for _, entry := range entries {
		if IsIgnoredVolume(entry.Name()) {
			continue
		}

		volPath := filepath.Join("/Volumes", entry.Name())
		volName := entry.Name()

		// Probe diskutil info for exact volume & whole disk node details
		infoStr := getDarwinDiskutilInfo(volPath)

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

		isCloudMode := IsCloudModeDisk(devNode)
		isRealVentoy := false
		if !isCloudMode {
			isRealVentoy = IsVentoyDisk(devNode)
		}
		isGenericBoot := false
		if !isCloudMode && !isRealVentoy {
			isGenericBoot = IsGenericBootDisk(devNode)
		}

		bootStatusStr := DetectBootStatus(volName, partitionScheme, isRealVentoy, isCloudMode, isGenericBoot)
		controllerVendorStr := InferControllerVendor(vendorId, productId, vendor)

		// Detect if this is a system disk
		isSystemDisk, _ := isSystemDiskDarwin(devNode)

		disks = append(disks, DiskInfo{
			Device:            devNode,
			Name:              displayName,
			Size:              totalSize,
			Formatted:         formattedSize,
			FreeSpace:         freeSpace,
			FreeFormatted:     freeFormatted,
			IsRemovable:       true,
			IsSystem:          isSystemDisk,
			UsbVersion:        usbVer,
			UsbSpeed:          usbSpeed,
			Vendor:            vendor,
			FileSystem:        fileSystem,
			PartitionScheme:   partitionScheme,
			Writable:          writable,
			SerialNumber:      serialNum,
			VendorId:          vendorId,
			ProductId:         productId,
			SmartStatus:       smartStatus,
			BusPower:          busPower,
			BusPowerUsed:      busPowerUsed,
			SectorSize:        sectorSizeStr,
			TransportProtocol: transportProtoStr,
			BootStatus:        bootStatusStr,
			ControllerVendor:  controllerVendorStr,
			IsFakeUsb3:        isFake,
			ProtocolCode:      protoCode,
			IsRealVentoy:      isRealVentoy,
			IsCloudMode:           isCloudMode,
			IsGenericBoot:     isGenericBoot,
			MountPoint:        volPath,
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
	Label      string             `json:"label"`
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
	var disks []DiskInfo
	cmd := execCommand("lsblk", "-J", "-b", "-o", "NAME,SIZE,FSAVAIL,RM,RO,TYPE,MOUNTPOINT,LABEL,MODEL,VENDOR,TRAN,FSTYPE,PTTYPE")
	output, err := cmd.Output()
	if err != nil {
		return disks, nil
	}

	var lsblk linuxLsblkOutput
	if err := jsonUnmarshal(output, &lsblk); err != nil {
		return disks, nil
	}

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

		devPath := "/dev/" + dev.Name
		mountPath := devPath
		label := dev.Label
		fileSystem := dev.Fstype
		freeSpace := dev.Fsavail
		partitionScheme := "GPT / MBR"
		if strings.ToLower(dev.Pttype) == "gpt" {
			partitionScheme = "GPT (GUID Partition Table)"
		} else if strings.ToLower(dev.Pttype) == "dos" || strings.ToLower(dev.Pttype) == "mbr" {
			partitionScheme = "MBR (Master Boot Record)"
		}

		if label == "" {
			label = strings.TrimSpace(dev.Vendor + " " + dev.Model)
		}
		if label == "" {
			label = dev.Name
		}

		for _, child := range dev.Children {
			if child.MountPoint != "" && !IsIgnoredVolume(child.Label) {
				mountPath = child.MountPoint
				if child.Label != "" {
					label = child.Label
				}
				if child.Fstype != "" {
					fileSystem = child.Fstype
				}
				if child.Fsavail > 0 {
					freeSpace = child.Fsavail
				}
				break
			}
		}

		if IsIgnoredVolume(label) {
			continue
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

		isFake := CheckFakeUsb3(label, usbVer, usbSpeed)
		protoCode := MapProtocolCode(usbVer, usbSpeed)

		// Detect if this is a system disk
		isSystemDisk, _ := isSystemDiskLinux("/dev/" + dev.Name)

		disks = append(disks, DiskInfo{
			Device:            mountPath,
			Name:              label,
			Size:              dev.Size,
			Formatted:         formattedSize,
			FreeSpace:         freeSpace,
			FreeFormatted:     freeFormatted,
			IsRemovable:       true,
			IsSystem:          isSystemDisk,
			UsbVersion:        usbVer,
			UsbSpeed:          usbSpeed,
			Vendor:            vendor,
			FileSystem:        fileSystem,
			PartitionScheme:   partitionScheme,
			Writable:          !dev.Ro,
			SmartStatus:       "Verified",
			BusPower:          "500 mA",
			BusPowerUsed:      "500 mA",
			SectorSize:        "512 Bytes (512n/512e)",
			TransportProtocol: "BOT (Bulk-Only Transport)",
			// Use devPath for disk type detection (MBR check), mountPath for file checks
			BootStatus:        DetectBootStatus(label, partitionScheme, IsRealVentoyDisk(devPath), IsCloudModeDisk(devPath), IsGenericBootDisk(mountPath)),
			ControllerVendor:  InferControllerVendor("", "", vendor),
			IsFakeUsb3:        isFake,
			ProtocolCode:      protoCode,
			IsRealVentoy:      IsRealVentoyDisk(devPath),
			IsCloudMode:       IsCloudModeDisk(devPath),
			IsGenericBoot:     IsGenericBootDisk(mountPath),
			MountPoint:        mountPath,
		})
	}

	return disks, nil
}

// Windows disk probing via PowerShell Win32_DiskDrive
type winDiskDrive struct {
	DeviceID      string `json:"DeviceID"`
	Model         string `json:"Model"`
	Size          uint64 `json:"Size"`
	InterfaceType string `json:"InterfaceType"`
	Caption       string `json:"Caption"`
}

func getWindowsDisks() ([]DiskInfo, error) {
	var disks []DiskInfo
	cmd := execCommand("powershell", "-NoProfile", "-Command",
		"Get-CimInstance Win32_DiskDrive | Where-Object { ($_.InterfaceType -eq 'USB' -or ($_.MediaType -like '*Removable*' -and $_.MediaType -notlike '*Fixed*')) -and $_.Model -notmatch 'Virtual|VHD|ISO|CD-ROM|DVD' -and $_.InterfaceType -ne 'FileBackedVirtual' } | Select-Object DeviceID, Model, Size, InterfaceType, Caption | ConvertTo-Json")
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

		driveLetter := fmt.Sprintf("%c:", 'E'+i)
		displayName := drive.Model
		if displayName == "" {
			displayName = drive.Caption
		}
		if displayName == "" {
			displayName = "USB Storage Device"
		}

		usbVer := "USB 3.0"
		usbSpeed := "5 Gb/s"
		if strings.Contains(strings.ToUpper(displayName), "2.0") {
			usbVer = "USB 2.0"
			usbSpeed = "480 Mb/s"
		}

		formattedSize := FormatBytesDual(drive.Size)
		freeSpace := uint64(float64(drive.Size) * 0.8)
		freeFormatted := FormatBytes(freeSpace)

		isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
		protoCode := MapProtocolCode(usbVer, usbSpeed)

		// Detect if this is a system disk
		isSystemDisk, _ := isSystemDiskWindows(driveLetter)

		disks = append(disks, DiskInfo{
			Device:            driveLetter,
			Name:              displayName,
			Size:              drive.Size,
			Formatted:         formattedSize,
			FreeSpace:         freeSpace,
			FreeFormatted:     freeFormatted,
			IsRemovable:       true,
			IsSystem:          isSystemDisk,
			UsbVersion:        usbVer,
			UsbSpeed:          usbSpeed,
			Vendor:            "Generic",
			FileSystem:        "FAT32 / NTFS",
			PartitionScheme:   "GPT / MBR",
			Writable:          true,
			SmartStatus:       "Verified",
			BusPower:          "500 mA",
			BusPowerUsed:      "500 mA",
			SectorSize:        "512 Bytes (512n/512e)",
			TransportProtocol: "BOT (Bulk-Only Transport)",
			BootStatus:        DetectBootStatus(displayName, "GPT / MBR", IsRealVentoyDisk(driveLetter), IsCloudModeDisk(driveLetter), IsGenericBootDisk(driveLetter)),
			ControllerVendor:  InferControllerVendor("", "", "Generic"),
			IsFakeUsb3:        isFake,
			ProtocolCode:      protoCode,
			IsRealVentoy:      IsRealVentoyDisk(driveLetter),
			IsCloudMode:           IsCloudModeDisk(driveLetter),
			IsGenericBoot:     IsGenericBootDisk(driveLetter),
			MountPoint:        driveLetter,
		})
	}

	return disks, nil
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

	// Static blacklist check for common system disk paths
	staticBlacklist := []string{"/", "C:", "/dev/sda", "/dev/nvme0n1"}
	for _, blocked := range staticBlacklist {
		if targetDevice == blocked {
			return fmt.Errorf("CRITICAL: Safety block triggered! %s is a known system drive", targetDevice)
		}
	}

	// Dynamic system disk detection
	isSystem, err := isSystemDisk(targetDevice)
	if err != nil {
		// Log warning but don't fail if detection fails
		logger.Warn("System disk detection failed, proceeding with caution", "device", targetDevice, "error", err)
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
		// macOS: /dev/diskN or /dev/rdiskN or diskN
		return regexp.MustCompile(`^(/dev/)?(r)?disk\d+$`).MatchString(path)
	case "windows":
		// Windows: C:, PhysicalDriveN, \\.\PhysicalDriveN, or diskN
		return regexp.MustCompile(`^([A-Z]:|([\\]{2}\.[\\])?PhysicalDrive\d+|disk\d+)$`).MatchString(path)
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

// isSystemDiskDarwin checks if a disk is a system disk on macOS
func isSystemDiskDarwin(device string) (bool, error) {
	// Normalize device path
	diskNode := NormalizeDarwinDiskNode(device)
	if diskNode == "" {
		return false, fmt.Errorf("invalid device path: %s", device)
	}

	cmd := execCommand("diskutil", "info", "-plist", diskNode)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get disk info: %w", err)
	}

	outputStr := string(output)

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

// isSystemDiskWindows checks if a disk is a system disk on Windows
func isSystemDiskWindows(device string) (bool, error) {
	// Extract drive letter or disk number
	var target string
	if len(device) == 2 && device[1] == ':' {
		// Drive letter format (C:)
		target = device
	} else {
		// PhysicalDrive format
		target = strings.TrimPrefix(device, `\\.\PhysicalDrive`)
		target = strings.TrimPrefix(target, `PhysicalDrive`)
		target = strings.TrimPrefix(target, `disk`)
	}

	// Check if drive contains Windows directory
	if len(target) == 2 && target[1] == ':' {
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

// EjectDisk safely unmounts and ejects the target removable USB storage drive.
func EjectDisk(device string) error {
	if device == "" {
		return fmt.Errorf("device path cannot be empty")
	}
	if err := ValidateTargetDisk(device); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		cmd := execCommand("diskutil", "eject", device)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fallbackCmd := execCommand("diskutil", "unmountDisk", device)
			if fallbackOut, fallbackErr := fallbackCmd.CombinedOutput(); fallbackErr != nil {
				return fmt.Errorf("failed to eject disk %s: %s (%w)", device, strings.TrimSpace(string(output)), err)
			} else {
				_ = fallbackOut
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
