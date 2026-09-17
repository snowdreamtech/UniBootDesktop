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
	"runtime"
	"strings"
	"sync"
	"time"
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

// DiskInfo represents metadata about an available disk/USB drive.
type DiskInfo struct {
	Device            string `json:"device"`            // Device path (e.g., /dev/disk2, E:)
	Name              string `json:"name"`              // Friendly label / vendor model
	Size              uint64 `json:"size"`              // Total capacity in bytes
	Formatted         string `json:"formatted"`         // Human readable size string
	FreeSpace         uint64 `json:"freeSpace"`         // Free available space in bytes
	FreeFormatted     string `json:"freeFormatted"`     // Human readable free space string
	IsRemovable       bool   `json:"isRemovable"`       // Removable USB flag
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
	IsModeB           bool   `json:"isModeB"`           // True if drive is formatted in Mode B (iPXE ESP Cloud Pure)
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

// DetectBootStatus evaluates the boot status text based on partition scheme, volume label, Ventoy/Mode B, and generic boot flags.
func DetectBootStatus(volName string, partitionScheme string, isRealVentoy bool, isModeB bool, isGenericBoot bool) string {
	if isRealVentoy {
		return "Ventoy / UniBoot (混合模式)"
	}
	if isModeB {
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
		diskNode := filepath.Base(targetDisk)
		if strings.HasPrefix(diskNode, "disk") {
			baseDisk := diskNode
			if idx := strings.Index(diskNode, "s"); idx != -1 {
				baseDisk = diskNode[:idx]
			}

			p1 := baseDisk + "s1"
			p2 := baseDisk + "s2"

			// Inspect Partition 2 (VTOYEFI / UNIBOOTEFI ESP Partition)
			outP2, errP2 := exec.Command("diskutil", "info", "-plist", p2).Output()
			if errP2 == nil {
				strP2 := string(outP2)
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
			outP1, errP1 := exec.Command("diskutil", "info", "-plist", p1).Output()
			if errP1 == nil {
				strP1 := string(outP1)
				volNameP1 := strings.ToUpper(extractPlistValue(strP1, "VolumeName"))
				mountP1 := extractPlistValue(strP1, "MountPoint")

				if mountP1 != "" && HasVentoyEngineFiles(mountP1) {
					return true
				}

				if (strings.Contains(volNameP1, "VENTOY") || strings.Contains(volNameP1, "UNIBOOT")) && errP2 == nil {
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
		out, err := exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("Get-Partition -DiskNumber (Get-Disk | Where-Object {$_.Path -like '*%s*'}).DiskNumber | Get-Volume | Select-Object -ExpandProperty FileSystemLabel", filepath.Base(targetDisk))).Output()
		if err == nil {
			upperOut := strings.ToUpper(string(out))
			if strings.Contains(upperOut, "VTOYEFI") || strings.Contains(upperOut, "UNIBOOTEFI") {
				return true
			}
		}
	}

	return false
}

// IsModeBDisk checks if a target disk is currently formatted in UniBoot Cloud mode (iPXE boot firmware in ESP, no Ventoy engine).
// Strictly checks physical iPXE files. NEVER relies on volume names alone.
func IsModeBDisk(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	if runtime.GOOS == "darwin" {
		diskNode := filepath.Base(targetDisk)
		if strings.HasPrefix(diskNode, "disk") {
			// Single partition short-circuit: Mode B requires dual partitions (Partition 1 Data + Partition 2 ESP).
			if !strings.Contains(diskNode, "s") {
				p2 := diskNode + "s2"
				if err := exec.Command("diskutil", "info", p2).Run(); err != nil {
					return false
				}
			}
			p1 := diskNode
			p2 := diskNode
			if !strings.Contains(diskNode, "s") {
				p1 = diskNode + "s1"
				p2 = diskNode + "s2"
			}
			for _, p := range []string{p1, p2} {
				out, err := exec.Command("diskutil", "info", "-plist", p).Output()
				if err == nil {
					mountPoint := extractPlistValue(string(out), "MountPoint")
					if HasUniBootCloudFiles(mountPoint) && !HasVentoyEngineFiles(mountPoint) {
						return true
					}
				}
			}
		}
	} else {
		if HasUniBootCloudFiles(targetDisk) && !HasVentoyEngineFiles(targetDisk) {
			return true
		}
	}
	return false
}

// IsRealVentoyDisk checks if a target disk is an active Ventoy drive containing Ventoy's MBR bootloader and configuration.
func IsRealVentoyDisk(targetDisk string) bool {
	if IsModeBDisk(targetDisk) {
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

// IsGenericBootDisk checks if a target disk is a 3rd-party boot disk (Rufus, PE, ISO) that is NOT a Ventoy or Mode B drive.
func IsGenericBootDisk(targetDisk string) bool {
	if targetDisk == "" {
		return false
	}
	if IsRealVentoyDisk(targetDisk) || IsModeBDisk(targetDisk) {
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
				out, err := exec.Command("diskutil", "info", "-plist", p).Output()
				if err == nil {
					mountPoint := extractPlistValue(string(out), "MountPoint")
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
)

// InvalidateDiskCache clears the memory disk cache to force an immediate fresh hardware scan.
func InvalidateDiskCache() {
	diskCacheMutex.Lock()
	diskCacheList = nil
	diskCacheMutex.Unlock()
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
	if diskCacheList != nil && time.Since(diskCacheTime) < 3*time.Second {
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

	// Cache hardware profile details for 6 seconds to prevent system_profiler CPU spikes
	if darwinUSBCacheMap != nil && time.Since(darwinUSBCacheTime) < 6*time.Second {
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

	// Step 1: Probe system_profiler for rich hardware details (cached for 6s to eliminate CPU spikes)
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
		infoCmd := execCommand("diskutil", "info", "-plist", volPath)
		infoOut, infoErr := infoCmd.Output()

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

		if infoErr == nil {
			infoStr := string(infoOut)
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
				isRemovable = strings.Contains(infoStr, "<true/>")
			}
			if strings.Contains(infoStr, "<key>Writable</key>") {
				writable = strings.Contains(infoStr, "<key>Writable</key>\n\t<true/>") || strings.Contains(infoStr, "<key>Writable</key><true/>") || !strings.Contains(infoStr, "<key>Writable</key>\n\t<false/>")
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
		}

		if fileSystem == "" {
			fileSystem = "ExFAT"
		}

		// Skip non-USB / internal disks if bus protocol is available
		if busProto != "" && busProto != "USB" && !isRemovable {
			continue
		}

		// Probe whole disk info for total raw byte size and partition map type
		if parentDisk != "" {
			parentCmd := execCommand("diskutil", "info", "-plist", parentDisk)
			parentOut, parentErr := parentCmd.Output()
			if parentErr == nil {
				parentStr := string(parentOut)
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

		isRealVentoy := IsRealVentoyDisk(devNode)
		isModeB := IsModeBDisk(devNode)
		isGenericBoot := IsGenericBootDisk(devNode)
		bootStatusStr := DetectBootStatus(volName, partitionScheme, isRealVentoy, isModeB, isGenericBoot)
		controllerVendorStr := InferControllerVendor(vendorId, productId, vendor)

		disks = append(disks, DiskInfo{
			Device:            devNode,
			Name:              displayName,
			Size:              totalSize,
			Formatted:         formattedSize,
			FreeSpace:         freeSpace,
			FreeFormatted:     freeFormatted,
			IsRemovable:       true,
			IsSystem:          false,
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
			IsModeB:           isModeB,
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

		disks = append(disks, DiskInfo{
			Device:            mountPath,
			Name:              label,
			Size:              dev.Size,
			Formatted:         formattedSize,
			FreeSpace:         freeSpace,
			FreeFormatted:     freeFormatted,
			IsRemovable:       true,
			IsSystem:          false,
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
			BootStatus:        DetectBootStatus(label, partitionScheme, IsRealVentoyDisk(mountPath), IsModeBDisk(mountPath), IsGenericBootDisk(mountPath)),
			ControllerVendor:  InferControllerVendor("", "", vendor),
			IsFakeUsb3:        isFake,
			ProtocolCode:      protoCode,
			IsRealVentoy:      IsRealVentoyDisk(mountPath),
			IsModeB:           IsModeBDisk(mountPath),
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
		"Get-CimInstance Win32_DiskDrive | Where-Object { $_.InterfaceType -eq 'USB' -or $_.MediaType -like '*Removable*' } | Select-Object DeviceID, Model, Size, InterfaceType, Caption | ConvertTo-Json")
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

		disks = append(disks, DiskInfo{
			Device:            driveLetter,
			Name:              displayName,
			Size:              drive.Size,
			Formatted:         formattedSize,
			FreeSpace:         freeSpace,
			FreeFormatted:     freeFormatted,
			IsRemovable:       true,
			IsSystem:          false,
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
			BootStatus:        DetectBootStatus(displayName, "GPT / MBR", IsRealVentoyDisk(driveLetter), IsModeBDisk(driveLetter), IsGenericBootDisk(driveLetter)),
			ControllerVendor:  InferControllerVendor("", "", "Generic"),
			IsFakeUsb3:        isFake,
			ProtocolCode:      protoCode,
			IsRealVentoy:      IsRealVentoyDisk(driveLetter),
			IsModeB:           IsModeBDisk(driveLetter),
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
	if targetDevice == "/" || targetDevice == "C:" || targetDevice == "/dev/sda" || targetDevice == "/dev/nvme0n1" {
		return fmt.Errorf("CRITICAL: Safety block triggered! %s is a system drive", targetDevice)
	}
	return nil
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
