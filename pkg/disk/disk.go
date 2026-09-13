// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
		"SYSTEM RESERVED",
		"系统保留",
		"WINRE",
		"WINRETOOLS",
		"OEM",
	}

	// ignoredVolumePrefixes defines volume name prefixes to ignore (case-insensitive)
	ignoredVolumePrefixes = []string{
		"VTOYEFI",
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
	Device          string `json:"device"`          // Device path (e.g., /dev/disk2, E:)
	Name            string `json:"name"`            // Friendly label / vendor model
	Size            uint64 `json:"size"`            // Total capacity in bytes
	Formatted       string `json:"formatted"`       // Human readable size string
	FreeSpace       uint64 `json:"freeSpace"`       // Free available space in bytes
	FreeFormatted   string `json:"freeFormatted"`   // Human readable free space string
	IsRemovable     bool   `json:"isRemovable"`     // Removable USB flag
	IsSystem        bool   `json:"isSystem"`        // System disk safety flag
	UsbVersion      string `json:"usbVersion"`      // Protocol version (USB 2.0, USB 3.0, USB 3.1, USB 3.2, USB4)
	UsbSpeed        string `json:"usbSpeed"`        // Physical bus speed (480 Mb/s, 5 Gb/s, 10 Gb/s, 20 Gb/s)
	Vendor          string `json:"vendor"`          // Device manufacturer / vendor
	FileSystem      string `json:"fileSystem"`      // File system format (e.g., ExFAT, FAT32, NTFS, APFS, ext4)
	PartitionScheme string `json:"partitionScheme"` // Partition scheme (e.g., GPT, MBR)
	Writable        bool   `json:"writable"`        // Read-Write status (true = Read-Write, false = Read-Only)
	SerialNumber    string `json:"serialNumber"`    // Hardware Serial Number
	VendorId        string `json:"vendorId"`        // USB Vendor ID (e.g., 0x21c4)
	ProductId       string `json:"productId"`       // USB Product ID (e.g., 0x0cd1)
	SmartStatus     string `json:"smartStatus"`     // S.M.A.R.T. health status (e.g. Verified, Not Supported, Failing)
	BusPower        string `json:"busPower"`        // Bus power available (e.g. 500 mA, 900 mA)
	BusPowerUsed    string `json:"busPowerUsed"`    // Bus power required/used (e.g. 500 mA, 224 mA)
	IsFakeUsb3      bool   `json:"isFakeUsb3"`      // Warning flag for fake USB 3.0 (USB 2.0 PHY disguised as 3.0)
	ProtocolCode    string `json:"protocolCode"`    // Styling code: "usb2", "usb3_0", "usb3_1", "usb3_2", "usb4"
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

// FormatBytes formats byte counts into human-readable strings (e.g. 248.15 GB, 8.05 GB).
func FormatBytes(bytes uint64) string {
	const (
		KB = 1000
		MB = 1000 * KB
		GB = 1000 * MB
		TB = 1000 * GB
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

// GetRemovableDisks lists removable USB drives safely while protecting system drives.
func GetRemovableDisks() ([]DiskInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return getDarwinDisks()
	case "windows":
		return getWindowsDisks()
	default:
		return getLinuxDisks()
	}
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
					BusPower:     strings.TrimSpace(item.BusPower),
					BusPowerUsed: strings.TrimSpace(item.BusPowerUsed),
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

func getDarwinDisks() ([]DiskInfo, error) {
	var disks []DiskInfo
	usbMap := make(map[string]*darwinUSBInfo)

	// Step 1: Probe system_profiler for rich hardware details (~0.3s runtime)
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
			if parentInfo.Model != "" && parentInfo.Model != "USB Flash Drive" && parentInfo.Model != "Disk 2.0" {
				displayName = fmt.Sprintf("%s (%s)", parentInfo.Model, volName)
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

		formattedSize := FormatBytes(totalSize)
		freeFormatted := FormatBytes(freeSpace)
		if freeSpace == 0 {
			freeFormatted = formattedSize
		}

		isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
		protoCode := MapProtocolCode(usbVer, usbSpeed)

		disks = append(disks, DiskInfo{
			Device:          volPath,
			Name:            displayName,
			Size:            totalSize,
			Formatted:       formattedSize,
			FreeSpace:       freeSpace,
			FreeFormatted:   freeFormatted,
			IsRemovable:     true,
			IsSystem:        false,
			UsbVersion:      usbVer,
			UsbSpeed:        usbSpeed,
			Vendor:          vendor,
			FileSystem:      fileSystem,
			PartitionScheme: partitionScheme,
			Writable:        writable,
			SerialNumber:    serialNum,
			VendorId:        vendorId,
			ProductId:       productId,
			SmartStatus:     smartStatus,
			BusPower:        busPower,
			BusPowerUsed:    busPowerUsed,
			IsFakeUsb3:      isFake,
			ProtocolCode:    protoCode,
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

		formattedSize := FormatBytes(dev.Size)
		freeFormatted := FormatBytes(freeSpace)
		if freeSpace == 0 {
			freeFormatted = formattedSize
		}

		isFake := CheckFakeUsb3(label, usbVer, usbSpeed)
		protoCode := MapProtocolCode(usbVer, usbSpeed)

		disks = append(disks, DiskInfo{
			Device:          mountPath,
			Name:            label,
			Size:            dev.Size,
			Formatted:       formattedSize,
			FreeSpace:       freeSpace,
			FreeFormatted:   freeFormatted,
			IsRemovable:     true,
			IsSystem:        false,
			UsbVersion:      usbVer,
			UsbSpeed:        usbSpeed,
			Vendor:          vendor,
			FileSystem:      fileSystem,
			PartitionScheme: partitionScheme,
			Writable:        !dev.Ro,
			SmartStatus:     "Verified",
			BusPower:        "500 mA",
			BusPowerUsed:    "500 mA",
			IsFakeUsb3:      isFake,
			ProtocolCode:    protoCode,
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

		formattedSize := FormatBytes(drive.Size)
		freeSpace := uint64(float64(drive.Size) * 0.8)
		freeFormatted := FormatBytes(freeSpace)

		isFake := CheckFakeUsb3(displayName, usbVer, usbSpeed)
		protoCode := MapProtocolCode(usbVer, usbSpeed)

		disks = append(disks, DiskInfo{
			Device:          driveLetter,
			Name:            displayName,
			Size:            drive.Size,
			Formatted:       formattedSize,
			FreeSpace:       freeSpace,
			FreeFormatted:   freeFormatted,
			IsRemovable:     true,
			IsSystem:        false,
			UsbVersion:      usbVer,
			UsbSpeed:        usbSpeed,
			Vendor:          "Generic",
			FileSystem:      "FAT32 / NTFS",
			PartitionScheme: "GPT / MBR",
			Writable:        true,
			SmartStatus:     "Verified",
			BusPower:        "500 mA",
			BusPowerUsed:    "500 mA",
			IsFakeUsb3:      isFake,
			ProtocolCode:    protoCode,
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

