// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !darwin && !windows

package sysinfo

import (
	"os/exec"
	"strings"
)

// IsSystemDarkTheme returns true if Linux/Unix desktop is currently configured in Dark Mode.
func IsSystemDarkTheme() bool {
	// 1. GNOME 42+ standard interface color-scheme
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	if out, err := cmd.Output(); err == nil {
		str := strings.ToLower(string(out))
		if strings.Contains(str, "dark") {
			return true
		}
	}

	// 2. GNOME / XFCE / MATE gtk-theme (e.g., 'Yaru-dark', 'Adwaita-dark', 'Arc-Dark')
	cmdGtk := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme")
	if outGtk, errGtk := cmdGtk.Output(); errGtk == nil {
		if strings.Contains(strings.ToLower(string(outGtk)), "dark") {
			return true
		}
	}

	// 3. FreeDesktop XDG Settings Portal (standard for modern Wayland, Flatpak, and diverse desktops)
	cmdPortal := exec.Command("dbus-send", "--session", "--dest=org.freedesktop.portal.Desktop",
		"--type=method_call", "--print-reply=literal",
		"/org/freedesktop/portal/desktop",
		"org.freedesktop.portal.Settings.Read",
		"string:org.freedesktop.appearance", "string:color-scheme")
	if outPortal, errPortal := cmdPortal.Output(); errPortal == nil {
		str := strings.TrimSpace(string(outPortal))
		if strings.Contains(str, "uint32 1") {
			return true
		}
	}

	// 4. KDE Plasma configuration
	for _, kcmd := range []string{"kreadconfig6", "kreadconfig5"} {
		if path, err := exec.LookPath(kcmd); err == nil {
			cmdKde := exec.Command(path, "--group", "General", "--key", "ColorScheme")
			if outKde, err := cmdKde.Output(); err == nil {
				if strings.Contains(strings.ToLower(string(outKde)), "dark") {
					return true
				}
			}
		}
	}

	return false
}

// SupportsMicaBackdrop always returns false on Linux/Unix.
func SupportsMicaBackdrop() bool {
	return false
}

// SyncTitleBarTheme is a no-op on Linux as window decorations are managed by the window manager/GTK.
func SyncTitleBarTheme(theme string) {
}

