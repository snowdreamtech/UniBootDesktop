// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package sysinfo

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	dwmwaUseImmersiveDarkModeBefore20h1 = 19
	dwmwaUseImmersiveDarkMode           = 20
	dwmwaBorderColor                    = 34
	dwmwaCaptionColor                   = 35
	dwmwaTextColor                      = 36

	swpNosize       = 0x0001
	swpNomove       = 0x0002
	swpNozorder     = 0x0004
	swpFramechanged = 0x0020

	rdwInvalidate = 0x0001
	rdwUpdatenow  = 0x0100
	rdwFrame      = 0x0400
)

func getWindowsBuildNumber() int {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return 0
	}
	defer k.Close()

	buildStr, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		return 0
	}
	var build int
	if _, err := fmt.Sscanf(buildStr, "%d", &build); err == nil {
		return build
	}
	return 0
}

func isBuildAtLeast(minBuild int) bool {
	return getWindowsBuildNumber() >= minBuild
}

// IsSystemDarkTheme returns true if Windows is currently configured in Dark Mode.
func IsSystemDarkTheme() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false
	}
	return val == 0
}

// SupportsMicaBackdrop checks whether the current Windows build supports modern backdrop materials (Windows 11 22H2+ build >= 22621).
func SupportsMicaBackdrop() bool {
	return isBuildAtLeast(22621)
}

// GetProcessMainWindow finds the main top-level window belonging to the current process.
func GetProcessMainWindow() uintptr {
	user32 := syscall.NewLazyDLL("user32.dll")
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	procGetForegroundWindow := user32.NewProc("GetForegroundWindow")
	procFindWindowW := user32.NewProc("FindWindowW")
	procEnumWindows := user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible := user32.NewProc("IsWindowVisible")
	procGetClassNameW := user32.NewProc("GetClassNameW")
	procGetCurrentProcessId := kernel32.NewProc("GetCurrentProcessId")

	currentPid, _, _ := procGetCurrentProcessId.Call()

	// 1. Check Foreground Window (most responsive when triggered by user interaction)
	fg, _, _ := procGetForegroundWindow.Call()
	if fg != 0 {
		var pid uint32
		procGetWindowThreadProcessId.Call(fg, uintptr(unsafe.Pointer(&pid)))
		if uintptr(pid) == currentPid {
			return fg
		}
	}

	// 2. Check FindWindow by Wails class name "wailsWindow"
	clsWails, _ := syscall.UTF16PtrFromString("wailsWindow")
	hwndCls, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(clsWails)), 0)
	if hwndCls != 0 {
		var pid uint32
		procGetWindowThreadProcessId.Call(hwndCls, uintptr(unsafe.Pointer(&pid)))
		if uintptr(pid) == currentPid {
			return hwndCls
		}
	}

	// 3. Check FindWindow by application title "UniBootDesktop"
	titleWails, _ := syscall.UTF16PtrFromString("UniBootDesktop")
	hwndTitle, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titleWails)))
	if hwndTitle != 0 {
		var pid uint32
		procGetWindowThreadProcessId.Call(hwndTitle, uintptr(unsafe.Pointer(&pid)))
		if uintptr(pid) == currentPid {
			return hwndTitle
		}
	}

	// 4. Fallback: EnumWindows checking PID and class/visibility
	var mainHwnd uintptr
	cb := syscall.NewCallback(func(hwnd, lParam uintptr) uintptr {
		var pid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if uintptr(pid) == currentPid {
			buf := make([]uint16, 64)
			procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 64)
			className := syscall.UTF16ToString(buf)
			if className == "wailsWindow" {
				mainHwnd = hwnd
				return 0 // stop enumeration, found exact wails window
			}
			vis, _, _ := procIsWindowVisible.Call(hwnd)
			if vis != 0 && mainHwnd == 0 {
				mainHwnd = hwnd
			}
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return mainHwnd
}

// UpdateTitleBarTheme sets immersive dark mode and title bar caption colors, and triggers an immediate frame redraw.
func UpdateTitleBarTheme(hwnd uintptr, isDark bool) {
	if hwnd == 0 {
		return
	}
	dwmapi := syscall.NewLazyDLL("dwmapi.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	procDwmSetWindowAttribute := dwmapi.NewProc("DwmSetWindowAttribute")
	procSetWindowPos := user32.NewProc("SetWindowPos")
	procSendMessageW := user32.NewProc("SendMessageW")
	procRedrawWindow := user32.NewProc("RedrawWindow")

	var attr uintptr = dwmwaUseImmersiveDarkModeBefore20h1
	if isBuildAtLeast(18985) {
		attr = dwmwaUseImmersiveDarkMode
	}

	var winDark int32
	if isDark {
		winDark = 1
	}
	procDwmSetWindowAttribute.Call(hwnd, attr, uintptr(unsafe.Pointer(&winDark)), unsafe.Sizeof(winDark))

	if isBuildAtLeast(22000) {
		var captionColor, textColor, borderColor uint32
		if isDark {
			// #0b0f19 background, #f1f5f9 text, #1e293b border (0x00BBGGRR format)
			captionColor = 0x00190F0B
			textColor = 0x00F9F5F1
			borderColor = 0x003B291E
		} else {
			// #f8fafc background, #0f172a text, #e2e8f0 border (0x00BBGGRR format)
			captionColor = 0x00FCFAF8
			textColor = 0x002A170F
			borderColor = 0x00F0E8E2
		}
		procDwmSetWindowAttribute.Call(hwnd, dwmwaCaptionColor, uintptr(unsafe.Pointer(&captionColor)), 4)
		procDwmSetWindowAttribute.Call(hwnd, dwmwaTextColor, uintptr(unsafe.Pointer(&textColor)), 4)
		procDwmSetWindowAttribute.Call(hwnd, dwmwaBorderColor, uintptr(unsafe.Pointer(&borderColor)), 4)
	}

	// Trigger immediate non-client frame recalculation and repaint
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpNomove|swpNosize|swpNozorder|swpFramechanged)
	// Force non-client area repaint on Windows 10 via WM_NCACTIVATE de-activate/reactivate
	procSendMessageW.Call(hwnd, 0x0086, 0, 0) // WM_NCACTIVATE false
	procSendMessageW.Call(hwnd, 0x0086, 1, 0) // WM_NCACTIVATE true
	procRedrawWindow.Call(hwnd, 0, 0, rdwFrame|rdwInvalidate|rdwUpdatenow)
}

// SyncTitleBarTheme synchronizes the native window title bar on Windows with the specified theme.
func SyncTitleBarTheme(theme string) {
	isDark := false
	switch theme {
	case "dark":
		isDark = true
	case "light":
		isDark = false
	default:
		isDark = IsSystemDarkTheme()
	}

	hwnd := GetProcessMainWindow()
	if hwnd != 0 {
		UpdateTitleBarTheme(hwnd, isDark)
		// Schedule follow-up refresh to ensure repaint happens after any pending Wails message processing
		time.AfterFunc(60*time.Millisecond, func() {
			UpdateTitleBarTheme(hwnd, isDark)
		})
	}
}
