// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package disk

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// syncPlatformBuffers flushes file buffers across all accessible Windows volumes,
// forcing pending filesystem metadata and write-back caches to physical storage.
func syncPlatformBuffers() {
	for c := 'A'; c <= 'Z'; c++ {
		volPath := fmt.Sprintf(`\\.\%c:`, c)
		ptr, err := syscall.UTF16PtrFromString(volPath)
		if err != nil {
			continue
		}
		handle, err := syscall.CreateFile(
			ptr,
			syscall.GENERIC_READ|syscall.GENERIC_WRITE,
			syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
			nil,
			syscall.OPEN_EXISTING,
			0,
			0,
		)
		if err == nil {
			_ = syscall.FlushFileBuffers(handle)
			_ = syscall.CloseHandle(handle)
		}
	}
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

// getMountFreeSpace returns available free bytes for a mounted Windows path or drive root.
func getMountFreeSpace(mountPath string) uint64 {
	trimmed := strings.TrimSpace(mountPath)
	if trimmed == "" {
		return 0
	}
	if !strings.HasSuffix(trimmed, `\`) && !strings.HasSuffix(trimmed, `/`) {
		trimmed += `\`
	}
	ptr, err := syscall.UTF16PtrFromString(trimmed)
	if err != nil {
		return 0
	}
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(
		ptr,
		&freeBytesAvailable,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return 0
	}
	return totalFreeBytes
}

// getDriveLetterDiskNumberWindows returns the physical disk index of a Windows drive letter (e.g. C: or C:\).
func getDriveLetterDiskNumberWindows(driveLetter string) (int, error) {
	driveLetter = strings.TrimSpace(driveLetter)
	driveLetter = strings.TrimSuffix(driveLetter, `\`)
	driveLetter = strings.TrimSuffix(driveLetter, `/`)
	if len(driveLetter) == 1 {
		driveLetter += ":"
	}
	if len(driveLetter) != 2 || driveLetter[1] != ':' {
		return -1, fmt.Errorf("invalid windows drive letter %q", driveLetter)
	}

	volPath := fmt.Sprintf(`\\.\%s`, driveLetter)
	ptr, err := syscall.UTF16PtrFromString(volPath)
	if err != nil {
		return -1, err
	}
	handle, err := syscall.CreateFile(
		ptr,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return -1, err
	}
	defer syscall.CloseHandle(handle)

	// IOCTL_STORAGE_GET_DEVICE_NUMBER = 0x002D1080
	const ioctlStorageGetDeviceNumber = 0x002D1080
	type storageDeviceNumber struct {
		DeviceType      uint32
		DeviceNumber    uint32
		PartitionNumber uint32
	}
	var sdn storageDeviceNumber
	var bytesReturned uint32
	err = syscall.DeviceIoControl(
		handle,
		ioctlStorageGetDeviceNumber,
		nil,
		0,
		(*byte)(unsafe.Pointer(&sdn)),
		uint32(unsafe.Sizeof(sdn)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return -1, err
	}
	return int(sdn.DeviceNumber), nil
}

// getSystemDriveDiskNumberWindows returns the physical disk index of the Windows SystemDrive (e.g. C:).
func getSystemDriveDiskNumberWindows() (int, error) {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	return getDriveLetterDiskNumberWindows(sysDrive)
}

// getVolumesForDiskWindows scans all mounted Windows drive letters and returns those residing on diskNumber.
func getVolumesForDiskWindows(diskNumber int) ([]winVolumeInfo, error) {
	drivesMask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, err
	}

	var results []winVolumeInfo
	for c := 'A'; c <= 'Z'; c++ {
		if (drivesMask & (1 << (c - 'A'))) == 0 {
			continue
		}
		driveLetter := fmt.Sprintf("%c:", c)
		dn, err := getDriveLetterDiskNumberWindows(driveLetter)
		if err != nil || dn != diskNumber {
			continue
		}

		rootPath := fmt.Sprintf("%c:\\", c)
		rootPtr, err := syscall.UTF16PtrFromString(rootPath)
		if err != nil {
			continue
		}

		var volNameBuf [260]uint16
		var fsNameBuf [260]uint16
		var serialNum, maxLen, flags uint32
		_ = windows.GetVolumeInformation(
			rootPtr,
			&volNameBuf[0],
			uint32(len(volNameBuf)),
			&serialNum,
			&maxLen,
			&flags,
			&fsNameBuf[0],
			uint32(len(fsNameBuf)),
		)
		volName := syscall.UTF16ToString(volNameBuf[:])
		fsName := syscall.UTF16ToString(fsNameBuf[:])

		var freeBytesAvailable, totalBytes, totalFreeBytes uint64
		_ = windows.GetDiskFreeSpaceEx(
			rootPtr,
			&freeBytesAvailable,
			&totalBytes,
			&totalFreeBytes,
		)
		dt := windows.GetDriveType(rootPtr)

		results = append(results, winVolumeInfo{
			DriveLetter: driveLetter,
			VolumeName:  strings.TrimSpace(volName),
			FileSystem:  strings.TrimSpace(fsName),
			FreeSpace:   totalFreeBytes,
			TotalSize:   totalBytes,
			DriveType:   dt,
		})
	}
	return results, nil
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

// parsePnpStorageID extracts vendor, product, revision, and serial/instance from a Windows PNPDeviceID.
// Example: USBSTOR\DISK&VEN_SANDISK&PROD_ULTRA&REV_1.00\0123456789ABCDEF&0
func parsePnpStorageID(pnpID string) (vendor, product, rev, instance string) {
	parts := strings.Split(strings.TrimSpace(pnpID), `\`)
	if len(parts) >= 3 {
		instance = parts[len(parts)-1]
		instance = strings.TrimSuffix(instance, "&0")
		instance = strings.TrimSuffix(instance, "&1")
		desc := parts[1]
		for _, token := range strings.Split(desc, "&") {
			upper := strings.ToUpper(token)
			if strings.HasPrefix(upper, "VEN_") {
				vendor = strings.TrimPrefix(token, token[:4])
				vendor = strings.ReplaceAll(vendor, "_", " ")
				vendor = strings.TrimSpace(vendor)
			} else if strings.HasPrefix(upper, "PROD_") {
				product = strings.TrimPrefix(token, token[:5])
				product = strings.ReplaceAll(product, "_", " ")
				product = strings.TrimSpace(product)
			} else if strings.HasPrefix(upper, "REV_") {
				rev = strings.TrimPrefix(token, token[:4])
			}
		}
	} else if len(parts) == 2 {
		instance = parts[1]
	}
	return
}

// parseUsbRevision maps a USB BCD revision string to human-readable USB version and physical speed.
func parseUsbRevision(rev string) (version, speed string) {
	upper := strings.ToUpper(strings.TrimSpace(rev))
	upper = strings.TrimPrefix(upper, "REV_")
	switch {
	case strings.HasPrefix(upper, "04") || strings.HasPrefix(upper, "4"):
		return "USB4", "40 Gb/s"
	case strings.HasPrefix(upper, "032") || strings.HasPrefix(upper, "3.2"):
		return "USB 3.2", "20 Gb/s"
	case strings.HasPrefix(upper, "031") || strings.HasPrefix(upper, "3.1"):
		return "USB 3.1", "10 Gb/s"
	case strings.HasPrefix(upper, "03") || strings.HasPrefix(upper, "3.0") || strings.HasPrefix(upper, "3"):
		return "USB 3.0", "5 Gb/s"
	case strings.HasPrefix(upper, "02") || strings.HasPrefix(upper, "2.0") || strings.HasPrefix(upper, "2"):
		return "USB 2.0", "480 Mb/s"
	case strings.HasPrefix(upper, "01") || strings.HasPrefix(upper, "1.1") || strings.HasPrefix(upper, "1.0"):
		return "USB 1.1", "12 Mb/s"
	default:
		return "USB 2.0", "480 Mb/s"
	}
}

var (
	winUSBCacheMutex sync.RWMutex
	winUSBCacheMap   = make(map[string]*winUsbDeviceInfo)
	winUSBCacheTime  time.Time
)

// InvalidateWindowsUSBCache purges the Windows USB hardware profile cache.
func InvalidateWindowsUSBCache() {
	winUSBCacheMutex.Lock()
	winUSBCacheMap = make(map[string]*winUsbDeviceInfo)
	winUSBCacheTime = time.Time{}
	winUSBCacheMutex.Unlock()
}

// getUsbDeviceInfoWindows queries hardware metadata (VID/PID, real USB revision, UASP vs BOT) from the Windows PnP Registry.
func getUsbDeviceInfoWindows(pnpDeviceID string, serialNumber string) *winUsbDeviceInfo {
	cacheKey := strings.ToUpper(strings.TrimSpace(pnpDeviceID)) + "|" + strings.ToUpper(strings.TrimSpace(serialNumber))
	winUSBCacheMutex.RLock()
	if winUSBCacheMap != nil && time.Since(winUSBCacheTime) < 30*time.Second {
		if cached, ok := winUSBCacheMap[cacheKey]; ok {
			winUSBCacheMutex.RUnlock()
			return cached
		}
	}
	winUSBCacheMutex.RUnlock()

	parsedVen, parsedProd, parsedRev, parsedInst := parsePnpStorageID(pnpDeviceID)
	effectiveSerial := strings.TrimSpace(serialNumber)
	if effectiveSerial == "" {
		effectiveSerial = parsedInst
	}

	transportProto := "BOT (Bulk-Only Transport)"
	if strings.Contains(strings.ToUpper(pnpDeviceID), "UASP") || strings.HasPrefix(strings.ToUpper(pnpDeviceID), "SCSI\\") {
		transportProto = "UASP (USB Attached SCSI)"
	}

	info := &winUsbDeviceInfo{
		Vendor:            parsedVen,
		Product:           parsedProd,
		SerialNumber:      effectiveSerial,
		UsbVersion:        "USB 2.0",
		UsbSpeed:          "480 Mb/s",
		TransportProtocol: transportProto,
		BusPower:          "500 mA",
		BusPowerUsed:      "500 mA",
	}

	if parsedRev != "" {
		v, s := parseUsbRevision(parsedRev)
		info.UsbVersion = v
		info.UsbSpeed = s
	}

	// Open HKLM\SYSTEM\CurrentControlSet\Enum\USB to correlate with the parent USB entity
	usbKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Enum\USB`, registry.READ)
	if err != nil {
		return info
	}
	defer usbKey.Close()

	subkeys, err := usbKey.ReadSubKeyNames(-1)
	if err != nil {
		return info
	}

	for _, sub := range subkeys {
		upperSub := strings.ToUpper(sub)
		if !strings.HasPrefix(upperSub, "VID_") {
			continue
		}
		devKey, err := registry.OpenKey(usbKey, sub, registry.READ)
		if err != nil {
			continue
		}
		instNames, err := devKey.ReadSubKeyNames(-1)
		if err != nil {
			devKey.Close()
			continue
		}

		matched := false
		for _, inst := range instNames {
			upperInst := strings.ToUpper(inst)
			cleanInst := strings.TrimSuffix(upperInst, "&0")
			cleanInst = strings.TrimSuffix(cleanInst, "&1")

			// Match against serial or instance id
			isSerialMatch := effectiveSerial != "" && (cleanInst == strings.ToUpper(effectiveSerial) ||
				strings.Contains(strings.ToUpper(effectiveSerial), cleanInst) ||
				strings.Contains(cleanInst, strings.ToUpper(effectiveSerial)))
			isInstMatch := parsedInst != "" && (cleanInst == strings.ToUpper(parsedInst) ||
				strings.Contains(strings.ToUpper(pnpDeviceID), cleanInst))

			if isSerialMatch || isInstMatch {
				matched = true

				// Extract VID and PID from "VID_xxxx&PID_yyyy"
				tokens := strings.Split(upperSub, "&")
				for _, tok := range tokens {
					if strings.HasPrefix(tok, "VID_") {
						info.VendorID = "0x" + strings.ToLower(strings.TrimPrefix(tok, "VID_"))
					} else if strings.HasPrefix(tok, "PID_") {
						info.ProductID = "0x" + strings.ToLower(strings.TrimPrefix(tok, "PID_"))
					}
				}

				instKey, err := registry.OpenKey(devKey, inst, registry.READ)
				if err == nil {
					// Detect UASP vs BOT service
					if svc, _, errSvc := instKey.GetStringValue("Service"); errSvc == nil {
						if strings.EqualFold(svc, "UASPSTOR") {
							info.TransportProtocol = "UASP (USB Attached SCSI)"
						}
					}
					// Parse HardwareID for REV_xxxx
					if hwIDs, _, errHw := instKey.GetStringsValue("HardwareID"); errHw == nil {
						for _, hw := range hwIDs {
							if strings.Contains(hw, "REV_") {
								revParts := strings.Split(hw, "REV_")
								if len(revParts) >= 2 {
									v, s := parseUsbRevision(revParts[1])
									info.UsbVersion = v
									info.UsbSpeed = s
									break
								}
							}
						}
					}
					// Check manufacturer if not present in disk PNP string
					if mfg, _, errMfg := instKey.GetStringValue("Mfg"); errMfg == nil && mfg != "" && !strings.HasPrefix(mfg, "@") {
						if info.Vendor == "" || strings.EqualFold(info.Vendor, "Generic") {
							info.Vendor = strings.TrimSpace(mfg)
						}
					}
					instKey.Close()
				}
				break
			}
		}
		devKey.Close()
		if matched {
			break
		}
	}

	if strings.HasPrefix(info.UsbVersion, "USB 3") || strings.HasPrefix(info.UsbVersion, "USB4") {
		info.BusPower = "900 mA"
		info.BusPowerUsed = "900 mA"
	}

	winUSBCacheMutex.Lock()
	if winUSBCacheMap == nil {
		winUSBCacheMap = make(map[string]*winUsbDeviceInfo)
	}
	winUSBCacheMap[cacheKey] = info
	winUSBCacheTime = time.Now()
	winUSBCacheMutex.Unlock()

	return info
}

// winPhysicalDiskInfo represents hardware attributes returned directly from native Win32 physical disk IOCTLs.
type winPhysicalDiskInfo struct {
	DiskNumber     int
	DevicePath     string
	Size           uint64
	BytesPerSector uint32
	BusType        uint32 // 0x07 = BusTypeUsb
	IsRemovable    bool
	Vendor         string
	Product        string
	Revision       string
	SerialNumber   string
}

// queryPhysicalDiskWindows queries physical drive geometry and hardware properties directly via native Win32 IOCTLs (0ms, no PowerShell).
func queryPhysicalDiskWindows(diskNum int) (*winPhysicalDiskInfo, error) {
	devPath := fmt.Sprintf(`\\.\PhysicalDrive%d`, diskNum)
	ptr, err := syscall.UTF16PtrFromString(devPath)
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(
		ptr,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(handle)

	info := &winPhysicalDiskInfo{
		DiskNumber: diskNum,
		DevicePath: devPath,
	}

	// 1. IOCTL_DISK_GET_DRIVE_GEOMETRY_EX = 0x000700A0
	const ioctlDiskGetDriveGeometryEx = 0x000700A0
	type diskGeometryEx struct {
		Cylinders         int64
		MediaType         uint32
		TracksPerCylinder uint32
		SectorsPerTrack   uint32
		BytesPerSector    uint32
		DiskSize          int64
	}
	var geom diskGeometryEx
	var bytesRet uint32
	err = syscall.DeviceIoControl(
		handle,
		ioctlDiskGetDriveGeometryEx,
		nil,
		0,
		(*byte)(unsafe.Pointer(&geom)),
		uint32(unsafe.Sizeof(geom)),
		&bytesRet,
		nil,
	)
	if err == nil {
		if geom.DiskSize > 0 {
			info.Size = uint64(geom.DiskSize)
		}
		info.BytesPerSector = geom.BytesPerSector
	}

	// 2. IOCTL_STORAGE_QUERY_PROPERTY = 0x002D1400
	const ioctlStorageQueryProperty = 0x002D1400
	type storagePropertyQuery struct {
		PropertyId uint32
		QueryType  uint32
		Params     [1]byte
	}
	query := storagePropertyQuery{
		PropertyId: 0, // StorageDeviceProperty
		QueryType:  0, // PropertyStandardQuery
	}
	var buf [1024]byte
	err = syscall.DeviceIoControl(
		handle,
		ioctlStorageQueryProperty,
		(*byte)(unsafe.Pointer(&query)),
		uint32(unsafe.Sizeof(query)),
		&buf[0],
		uint32(len(buf)),
		&bytesRet,
		nil,
	)
	if err == nil && bytesRet >= 24 {
		if buf[6] == 1 {
			info.IsRemovable = true
		}
		vendorOffset := *(*uint32)(unsafe.Pointer(&buf[8]))
		productOffset := *(*uint32)(unsafe.Pointer(&buf[12]))
		revisionOffset := *(*uint32)(unsafe.Pointer(&buf[16]))
		serialOffset := *(*uint32)(unsafe.Pointer(&buf[20]))
		if bytesRet >= 28 {
			info.BusType = *(*uint32)(unsafe.Pointer(&buf[24]))
		}

		readASCII := func(offset uint32) string {
			if offset == 0 || offset >= bytesRet {
				return ""
			}
			end := offset
			for end < bytesRet && buf[end] != 0 {
				end++
			}
			return strings.TrimSpace(string(buf[offset:end]))
		}

		info.Vendor = readASCII(vendorOffset)
		info.Product = readASCII(productOffset)
		info.Revision = readASCII(revisionOffset)
		info.SerialNumber = readASCII(serialOffset)
	}

	return info, nil
}

// hasConnectedUSBStorageWindows checks quickly (~0.1ms) whether any USB mass storage device is registered or attached.
func hasConnectedUSBStorageWindows() bool {
	readEnumCount := func(svc string) int {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\`+svc+`\Enum`, registry.QUERY_VALUE)
		if err != nil {
			return 0
		}
		defer k.Close()
		count, _, err := k.GetIntegerValue("Count")
		if err != nil {
			return 0
		}
		return int(count)
	}

	if readEnumCount("USBSTOR") > 0 || readEnumCount("UASPSTOR") > 0 {
		return true
	}

	// Also check if any mounted drive letter is DRIVE_REMOVABLE
	mask, err := windows.GetLogicalDrives()
	if err == nil {
		for c := 'A'; c <= 'Z'; c++ {
			shift := c - 'A'
			if (mask & (1 << shift)) != 0 {
				root := fmt.Sprintf("%c:\\", c)
				rootPtr, _ := syscall.UTF16PtrFromString(root)
				if windows.GetDriveType(rootPtr) == windows.DRIVE_REMOVABLE {
					return true
				}
			}
		}
	}

	return false
}

// getVolumeSnapshotWindows computes a fast volume and hardware snapshot on Windows without blocking filesystem calls.
func getVolumeSnapshotWindows() string {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return ""
	}

	readEnumCount := func(svc string) int {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\`+svc+`\Enum`, registry.QUERY_VALUE)
		if err != nil {
			return 0
		}
		defer k.Close()
		count, _, err := k.GetIntegerValue("Count")
		if err != nil {
			return 0
		}
		return int(count)
	}

	usbCount := readEnumCount("USBSTOR")
	uaspCount := readEnumCount("UASPSTOR")

	physMask := uint32(0)
	for i := 0; i < 16; i++ {
		devPath := fmt.Sprintf(`\\.\PhysicalDrive%d`, i)
		ptr, err := syscall.UTF16PtrFromString(devPath)
		if err != nil {
			continue
		}
		handle, err := syscall.CreateFile(
			ptr,
			0,
			syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
			nil,
			syscall.OPEN_EXISTING,
			0,
			0,
		)
		if err == nil {
			physMask |= (1 << i)
			syscall.CloseHandle(handle)
		}
	}

	return fmt.Sprintf("mask:%x|usb:%d|uasp:%d|phys:%x", mask, usbCount, uaspCount, physMask)
}

// getWindowsDisksNative discovers USB and removable storage devices using direct Win32 IOCTLs (0.5ms vs 900ms PowerShell).
func getWindowsDisksNative() ([]DiskInfo, error) {
	if !hasConnectedUSBStorageWindows() {
		// Fast short-circuit: if no USB storage service entries and no removable drives exist,
		// probe PhysicalDrive1. If PhysicalDrive1 does not exist, return immediately (~0.1ms).
		p1, err := syscall.UTF16PtrFromString(`\\.\PhysicalDrive1`)
		if err == nil {
			h, err := syscall.CreateFile(p1, 0, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, 0, 0)
			if err != nil {
				return []DiskInfo{}, nil
			}
			syscall.CloseHandle(h)
		}
	}

	type indexedDrive struct {
		index int
		drive winDiskDrive
	}
	var validDrives []indexedDrive

	consecutiveFails := 0
	for i := 0; i < 16; i++ {
		pInfo, err := queryPhysicalDiskWindows(i)
		if err != nil {
			consecutiveFails++
			if i > 0 && consecutiveFails >= 3 {
				break
			}
			continue
		}
		consecutiveFails = 0

		// Check if this physical disk is USB or removable
		// BusTypeUsb = 0x07, BusType1394 = 0x04, BusTypeSd = 0x0C, BusTypeMmc = 0x0D
		isUSB := pInfo.BusType == 0x07 || pInfo.BusType == 0x04 || pInfo.BusType == 0x0C || pInfo.BusType == 0x0D || pInfo.IsRemovable
		if !isUSB {
			continue
		}

		devPath := fmt.Sprintf(`\\.\PhysicalDrive%d`, i)
		if isSys, _ := isSystemDiskWindows(devPath); isSys {
			continue
		}

		model := strings.TrimSpace(pInfo.Vendor + " " + pInfo.Product)
		if model == "" {
			model = "USB Storage Device"
		}

		upperModel := strings.ToUpper(model)
		if strings.Contains(upperModel, "VIRTUAL") ||
			strings.Contains(upperModel, "VHD") ||
			strings.Contains(upperModel, "ISO") ||
			strings.Contains(upperModel, "CD-ROM") ||
			strings.Contains(upperModel, "DVD") {
			continue
		}

		drive := winDiskDrive{
			DeviceID:       pInfo.DevicePath,
			Index:          pInfo.DiskNumber,
			Model:          model,
			Size:           pInfo.Size,
			InterfaceType:  "USB",
			Caption:        model,
			BytesPerSector: pInfo.BytesPerSector,
			SerialNumber:   pInfo.SerialNumber,
		}

		validDrives = append(validDrives, indexedDrive{index: i, drive: drive})
	}

	if len(validDrives) == 0 {
		return []DiskInfo{}, nil
	}

	disksResult := make([]*DiskInfo, len(validDrives))
	var wg sync.WaitGroup
	wg.Add(len(validDrives))

	for idx, item := range validDrives {
		go func(resultIdx int, driveIndex int, d winDiskDrive) {
			defer wg.Done()
			disksResult[resultIdx] = inspectWindowsDisk(driveIndex, d)
		}(idx, item.index, item.drive)
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
