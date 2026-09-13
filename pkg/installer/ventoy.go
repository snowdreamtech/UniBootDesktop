// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed themes/*
var embeddedThemes embed.FS

// VentoyThemeConfig defines the theme configuration block in ventoy.json matching UniBoot spec.
type VentoyThemeConfig struct {
	File    string `json:"file"`
	Gfxmode string `json:"gfxmode,omitempty"`
	Display string `json:"display,omitempty"`
}

// VentoyAliasConfig defines image_alias items in ventoy.json matching Ventoy plugin spec.
type VentoyAliasConfig struct {
	Image string `json:"image"`
	Alias string `json:"alias"`
}

// VentoyGlobalConfig represents the root JSON schema for ventoy/ventoy.json matching UniBoot spec.
type VentoyGlobalConfig struct {
	Theme      *VentoyThemeConfig       `json:"theme,omitempty"`
	ImageAlias []VentoyAliasConfig      `json:"image_alias,omitempty"`
	Control    []map[string]interface{} `json:"control,omitempty"`
}

// WriteVentoyConfig generates the ventoy/ventoy.json, ventoy/ventoy_grub.cfg, and extracts theme assets
// in the specified target volume mount directory, referencing UniBoot specifications.
func WriteVentoyConfig(mountDir string) error {
	ventoyDir := filepath.Join(mountDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		return fmt.Errorf("failed to create ventoy directory: %w", err)
	}

	// 1. Build ventoy.json matching official UniBoot specification
	cfg := VentoyGlobalConfig{
		Theme: &VentoyThemeConfig{
			File:    "/ventoy/themes/uniboot/theme.txt",
			Gfxmode: "1280x800",
			Display: "full",
		},
		ImageAlias: []VentoyAliasConfig{
			{
				Image: "/iso/UniBoot.iso",
				Alias: "⚡ UniBoot 统一网络与本地安装系统",
			},
		},
		Control: []map[string]interface{}{
			{"VTOY_DEFAULT_IMAGE": "/iso/UniBoot.iso"},
			{"VTOY_MENU_LANGUAGE": "zh_CN"},
			{"VTOY_FILE_FLT_EFI": "1"},
			{"VTOY_FILT_DOT_UNDERSCORE_FILE": "1"},
			{"VTOY_SORT_CASE_SENSITIVE": "0"},
			{"VTOY_WIN11_BYPASS_CHECK": "1"},
			{"VTOY_WIN11_BYPASS_NRO": "1"},
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

	// 2. Extract embedded UniBoot theme files to ventoy/themes/uniboot/
	themeTargetDir := filepath.Join(ventoyDir, "themes", "uniboot")
	if err := os.MkdirAll(themeTargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}

	err = fs.WalkDir(embeddedThemes, "themes/uniboot", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		relPath, err := filepath.Rel("themes/uniboot", path)
		if err != nil {
			return err
		}
		fileData, err := embeddedThemes.ReadFile(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(themeTargetDir, relPath)
		_ = os.MkdirAll(filepath.Dir(dest), 0755)
		return os.WriteFile(dest, fileData, 0644)
	})
	if err != nil {
		return fmt.Errorf("failed to extract embedded theme files: %w", err)
	}

	// 2. Build ventoy_grub.cfg matching official UniBoot specification (with i18n & multi-arch iPXE support)
	grubCfgContent := `# UniBoot Ventoy Custom GRUB Menu Configuration
# Press F6 in Ventoy main menu to access custom menu entries

# --- i18n Localization Engine ---
if [ -z "${lang}" ]; then
    set lang=zh_CN
fi

if [ "${lang}" = "zh_CN" -o "${lang}" = "zh_TW" -o "${lang}" = "zh_HK" ]; then
    set lbl_ipxe_uefi="⚡ UniBoot Network Installation (统一网络安装 - UEFI)"
    set lbl_ipxe_bios="⚡ UniBoot Network Installation (统一网络安装 - BIOS/非EFI)"
    set lbl_return="<-- 返回 Ventoy 主菜单"
else
    set lbl_ipxe_uefi="⚡ UniBoot Network Installation (UEFI Mode)"
    set lbl_ipxe_bios="⚡ UniBoot Network Installation (Legacy/Non-EFI Mode)"
    set lbl_return="<-- Return to Main Menu"
fi

if [ "$grub_platform" = "efi" ]; then
    menuentry "$lbl_ipxe_uefi" --class netboot {
        set ipxe_file="/ipxe/ipxe-${grub_cpu}.efi"
        search --no-floppy --set=root --file $ipxe_file
        if [ $? -ne 0 ]; then
            set ipxe_file="/EFI/BOOT/BOOTX64.EFI"
            if [ "$grub_cpu" = "arm64" ]; then
                set ipxe_file="/EFI/BOOT/BOOTAA64.EFI"
            elif [ "$grub_cpu" = "i386" ]; then
                set ipxe_file="/EFI/BOOT/BOOTIA32.EFI"
            fi
            search --no-floppy --set=root --file $ipxe_file
        fi
        chainloader $ipxe_file
    }
else
    menuentry "$lbl_ipxe_bios" --class netboot {
        set lkrn_file="/ipxe/ipxe.lkrn"
        if [ "$grub_cpu" = "riscv64" ]; then
            set lkrn_file="/ipxe/ipxe-riscv64.lkrn"
        elif [ "$grub_cpu" = "riscv32" ]; then
            set lkrn_file="/ipxe/ipxe-riscv32.lkrn"
        fi
        
        search --no-floppy --set=root --file $lkrn_file
        if [ $? -ne 0 ]; then
            set lkrn_file="/ipxe.lkrn"
            search --no-floppy --set=root --file $lkrn_file
        fi
        
        # x86 legacy bios uses linux16, others use linux
        if [ "$grub_cpu" = "i386" -o "$grub_cpu" = "x86_64" -o -z "$grub_cpu" ]; then
            linux16 $lkrn_file
            initrd /ipxe/uniboot.ipxe
        else
            linux $lkrn_file
            initrd /ipxe/uniboot.ipxe
        fi
    }
fi

menuentry "$lbl_return" --class=vtoyret VTOY_RET {
    true
}
`

	grubPath := filepath.Join(ventoyDir, "ventoy_grub.cfg")
	if err := os.WriteFile(grubPath, []byte(grubCfgContent), 0644); err != nil {
		return fmt.Errorf("failed to write ventoy_grub.cfg: %w", err)
	}

	return nil
}
