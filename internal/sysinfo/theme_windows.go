// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package sysinfo

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

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
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	buildStr, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		return false
	}
	var build int
	if _, err := fmt.Sscanf(buildStr, "%d", &build); err == nil {
		return build >= 22621
	}
	return false
}
