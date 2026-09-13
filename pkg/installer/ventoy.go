// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VentoyThemeConfig defines the theme configuration block in ventoy.json.
type VentoyThemeConfig struct {
	File        string `json:"file"`
	Gfxmode     string `json:"gfxmode,omitempty"`
	Display     string `json:"display,omitempty"`
	Serial      string `json:"serial,omitempty"`
	Font        string `json:"font,omitempty"`
	MenuColor   string `json:"menu_color,omitempty"`
	SelectColor string `json:"select_color,omitempty"`
}

// VentoyControlConfig defines global control options in ventoy.json.
type VentoyControlConfig struct {
	VtoyDefaultSearchRoot string `json:"VTOY_DEFAULT_SEARCH_ROOT,omitempty"`
	VtoyMenuTimeout       int    `json:"VTOY_MENU_TIMEOUT,omitempty"`
	VtoyDefaultOption     int    `json:"VTOY_DEFAULT_OPTION,omitempty"`
	VtoySecondaryBootMenu int    `json:"VTOY_SECONDARY_BOOT_MENU,omitempty"`
	VtoyHotKey            int    `json:"VTOY_HOTKEY,omitempty"`
}

// VentoyGlobalConfig represents the root JSON schema for ventoy/ventoy.json.
type VentoyGlobalConfig struct {
	Theme   *VentoyThemeConfig   `json:"theme,omitempty"`
	Control *VentoyControlConfig `json:"control,omitempty"`
}

// WriteVentoyConfig generates the ventoy/ventoy.json and ventoy/ventoy_grub.cfg files
// in the specified target volume mount directory.
func WriteVentoyConfig(mountDir string) error {
	ventoyDir := filepath.Join(mountDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		return fmt.Errorf("failed to create ventoy directory: %w", err)
	}

	// 1. Build ventoy.json
	cfg := VentoyGlobalConfig{
		Theme: &VentoyThemeConfig{
			File:      "/ventoy/theme/theme.txt",
			Gfxmode:   "1920x1080",
			MenuColor: "light-gray/black",
		},
		Control: &VentoyControlConfig{
			VtoyMenuTimeout: 10,
		},
	}

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal ventoy.json: %w", err)
	}

	jsonPath := filepath.Join(ventoyDir, "ventoy.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write ventoy.json: %w", err)
	}

	// 2. Build ventoy_grub.cfg for UniBoot iPXE Cloud Network Boot custom entry
	grubCfgContent := `
# UniBoot Custom Ventoy Menu Extensions
# Integrated Mode A: Offline Ventoy + Cloud iPXE Network Boot

menuentry "⚡ UniBoot Cloud iPXE Network Boot (云端网络引导)" --class ipxe --class net {
    echo 'Loading UniBoot iPXE Cloud Network Engine...'
    if [ "$grub_platform" = "efi" ]; then
        if [ "$grub_cpu" = "x86_64" ]; then
            chainloader /EFI/BOOT/BOOTX64.EFI
        elif [ "$grub_cpu" = "arm64" ]; then
            chainloader /EFI/BOOT/BOOTAA64.EFI
        elif [ "$grub_cpu" = "i386" ]; then
            chainloader /EFI/BOOT/BOOTIA32.EFI
        else
            chainloader /EFI/BOOT/BOOTX64.EFI
        fi
    else
        linux16 /ipxe.lkrn
    fi
}
`

	grubPath := filepath.Join(ventoyDir, "ventoy_grub.cfg")
	if err := os.WriteFile(grubPath, []byte(grubCfgContent), 0644); err != nil {
		return fmt.Errorf("failed to write ventoy_grub.cfg: %w", err)
	}

	return nil
}
