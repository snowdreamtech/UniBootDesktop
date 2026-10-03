// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unibootdesktop/internal/logger"
	"github.com/snowdreamtech/unibootdesktop/pkg/config"
)

//go:embed themes/*
var embeddedThemes embed.FS

//go:embed assets/ventoy_vhdboot.img
var embeddedVhdBootImg []byte

//go:embed assets/ventoy_wimboot.img
var embeddedWimBootImg []byte

// VentoyThemeConfig defines the theme configuration block in ventoy.json matching UniBoot spec.
type VentoyThemeConfig struct {
	File    string `json:"file"`
	Gfxmode string `json:"gfxmode,omitempty"`
	Display string `json:"display,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
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
// in the specified target volume mount directory, referencing UniBoot specifications and AppConfig.
func WriteVentoyConfig(mountDir string) error {
	return WriteVentoyConfigWithAppConfig(mountDir, nil)
}

// WriteVentoyConfigWithAppConfig generates ventoy.json with custom options from AppConfig.
func WriteVentoyConfigWithAppConfig(mountDir string, appCfg *config.AppConfig) error {
	if appCfg == nil {
		appCfg, _ = config.Load()
		if appCfg == nil {
			appCfg = config.GetDefaultConfig()
		}
	}

	ventoyDir := filepath.Join(mountDir, "ventoy")
	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		return fmt.Errorf("failed to create ventoy directory: %w", err)
	}

	jsonPath := filepath.Join(ventoyDir, "ventoy.json")
	var existingData []byte
	if raw, err := os.ReadFile(jsonPath); err == nil && len(raw) > 0 {
		existingData = raw
	}

	data, err := BuildVentoyConfigData(existingData, appCfg, mountDir)
	if err != nil {
		if len(existingData) > 0 {
			backupPath := filepath.Join(ventoyDir, "ventoy.json.corrupt.bak")
			_ = os.WriteFile(backupPath, existingData, 0644)
			logger.Warn("Existing ventoy.json was corrupt, backed up to ventoy.json.corrupt.bak and rebuilt clean config", "error", err)

			data, err = BuildVentoyConfigData(nil, appCfg, mountDir)
		}
		if err != nil {
			return fmt.Errorf("failed to build ventoy.json: %w", err)
		}
	}

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
    set lbl_ipxe_uefi="⚡ UniBoot Network Installation (UEFI Mode)"
    set lbl_ipxe_bios="⚡ UniBoot Network Installation (Legacy/Non-EFI Mode)"
    set lbl_return="<-- Return to Ventoy Main Menu"
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

	// 4. Ensure ventoy_vhdboot.img and ventoy_wimboot.img are deployed for VHD/WIM boot support in Hybrid Mode
	if err := DeployVentoyVhdBoot(ventoyDir); err != nil {
		return fmt.Errorf("failed to deploy ventoy_vhdboot.img: %w", err)
	}
	if err := DeployVentoyWimBoot(ventoyDir); err != nil {
		return fmt.Errorf("failed to deploy ventoy_wimboot.img: %w", err)
	}

	return nil
}

// BuildVentoyConfigData generates or non-destructively merges ventoy.json content.
// If existingData contains valid JSON, user-configured sections (e.g. persistence, auto_install,
// custom image_alias, custom controls) are strictly preserved while UniBoot settings are woven in.
func BuildVentoyConfigData(existingData []byte, appCfg *config.AppConfig, mountDir string) ([]byte, error) {
	if appCfg == nil {
		appCfg, _ = config.Load()
		if appCfg == nil {
			appCfg = config.GetDefaultConfig()
		}
	}

	controls := []map[string]interface{}{
		{"VTOY_MENU_LANGUAGE": "zh_CN"},
		{"VTOY_FILE_FLT_EFI": "1"},
		{"VTOY_FILT_DOT_UNDERSCORE_FILE": "1"},
		{"VTOY_SORT_CASE_SENSITIVE": "0"},
		{"VTOY_VHD_NO_WARNING": "1"},
	}

	defaultIso := filepath.Join(mountDir, "iso", "UniBoot.iso")
	if _, err := os.Stat(defaultIso); err == nil {
		controls = append(controls, map[string]interface{}{"VTOY_DEFAULT_IMAGE": "/iso/UniBoot.iso"})
	}

	if appCfg.VentoyWin11Bypass {
		controls = append(controls,
			map[string]interface{}{"VTOY_WIN11_BYPASS_CHECK": "1"},
			map[string]interface{}{"VTOY_WIN11_BYPASS_NRO": "1"},
		)
	}

	if appCfg.VentoySecondaryMenu {
		controls = append(controls, map[string]interface{}{"VTOY_SECONDARY_BOOT_MENU": "1"})
	} else {
		controls = append(controls, map[string]interface{}{"VTOY_SECONDARY_BOOT_MENU": "0"})
	}

	themeCfg := &VentoyThemeConfig{
		File:    "/ventoy/themes/uniboot/theme.txt",
		Gfxmode: "1280x800",
		Display: "full",
	}
	if appCfg.VentoyMenuTimeout > 0 {
		themeCfg.Timeout = appCfg.VentoyMenuTimeout
	}

	unibootAliases := []VentoyAliasConfig{
		{
			Image: "/iso/UniBoot.iso",
			Alias: "⚡ UniBoot Network & Local Installation System",
		},
	}

	// Try parsing existing JSON into a generic map to preserve user-defined sections (persistence, etc.)
	// Automatically strip any leading UTF-8 BOM bytes (\xef\xbb\xbf) generated by Windows Notepad.
	var rawMap map[string]interface{}
	cleanData := bytes.TrimPrefix(existingData, []byte("\xef\xbb\xbf"))
	cleanData = bytes.TrimSpace(cleanData)
	if len(cleanData) > 0 {
		if err := json.Unmarshal(cleanData, &rawMap); err != nil {
			return nil, fmt.Errorf("invalid json in ventoy.json: %w", err)
		}
	}
	if rawMap == nil {
		rawMap = make(map[string]interface{})
	}

	// 1. Weave Theme
	rawMap["theme"] = themeCfg

	// 2. Weave Controls non-destructively
	rawMap["control"] = mergeVentoyControls(rawMap["control"], controls)

	// 3. Weave ImageAlias non-destructively
	rawMap["image_alias"] = mergeVentoyImageAliases(rawMap["image_alias"], unibootAliases)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	if err := enc.Encode(rawMap); err != nil {
		return nil, fmt.Errorf("failed to encode ventoy.json: %w", err)
	}

	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func mergeVentoyControls(existing interface{}, unibootControls []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	handledKeys := make(map[string]bool)

	unibootMap := make(map[string]interface{})
	for _, c := range unibootControls {
		for k, v := range c {
			unibootMap[k] = v
		}
	}

	// 1. Process existing controls preserving user's order and custom keys
	if existingSlice, ok := existing.([]interface{}); ok {
		for _, item := range existingSlice {
			if m, ok := item.(map[string]interface{}); ok {
				newEntry := make(map[string]interface{})
				for k, v := range m {
					if unibootVal, exists := unibootMap[k]; exists {
						newEntry[k] = unibootVal
						handledKeys[k] = true
					} else {
						// Custom user control, preserve it
						newEntry[k] = v
					}
				}
				if len(newEntry) > 0 {
					result = append(result, newEntry)
				}
			}
		}
	}

	// 2. Append any UniBoot controls not yet handled
	for _, c := range unibootControls {
		for k, v := range c {
			if !handledKeys[k] {
				result = append(result, map[string]interface{}{k: v})
				handledKeys[k] = true
			}
		}
	}

	return result
}

func mergeVentoyImageAliases(existing interface{}, unibootAliases []VentoyAliasConfig) []VentoyAliasConfig {
	result := make([]VentoyAliasConfig, 0)
	unibootMap := make(map[string]string)
	for _, a := range unibootAliases {
		unibootMap[a.Image] = a.Alias
	}

	handledImages := make(map[string]bool)

	// Preserve existing aliases
	if existingSlice, ok := existing.([]interface{}); ok {
		for _, item := range existingSlice {
			if m, ok := item.(map[string]interface{}); ok {
				img, _ := m["image"].(string)
				alias, _ := m["alias"].(string)
				if img != "" {
					if ubAlias, exists := unibootMap[img]; exists {
						result = append(result, VentoyAliasConfig{Image: img, Alias: ubAlias})
						handledImages[img] = true
					} else {
						result = append(result, VentoyAliasConfig{Image: img, Alias: alias})
					}
				}
			}
		}
	}

	// Append any unhandled UniBoot aliases
	for _, a := range unibootAliases {
		if !handledImages[a.Image] {
			result = append(result, a)
			handledImages[a.Image] = true
		}
	}

	return result
}

// DeployVentoyVhdBoot extracts embedded ventoy_vhdboot.img to ventoyDir if not already present or incomplete.
// It is idempotent and preserves any existing non-empty file (e.g. customized versions).
func DeployVentoyVhdBoot(ventoyDir string) error {
	return deployEmbeddedImage(ventoyDir, "ventoy_vhdboot.img", embeddedVhdBootImg)
}

// DeployVentoyWimBoot extracts embedded ventoy_wimboot.img to ventoyDir if not already present or incomplete.
// It is idempotent and preserves any existing non-empty file (e.g. customized versions).
func DeployVentoyWimBoot(ventoyDir string) error {
	return deployEmbeddedImage(ventoyDir, "ventoy_wimboot.img", embeddedWimBootImg)
}

func deployEmbeddedImage(ventoyDir, filename string, embeddedData []byte) error {
	if len(embeddedData) == 0 {
		return fmt.Errorf("embedded %s is empty", filename)
	}

	targetFile := filepath.Join(ventoyDir, filename)
	info, err := os.Stat(targetFile)
	if err == nil {
		if info.Size() > 0 {
			// Already exists and non-empty, avoid redundant writing or overwriting custom user versions
			return nil
		}
		// 0-byte incomplete file, clean it up before rewrite
		_ = os.Remove(targetFile)
	}

	if err := os.MkdirAll(ventoyDir, 0755); err != nil {
		return fmt.Errorf("failed to create ventoy directory for %s: %w", filename, err)
	}

	// Write atomically via temporary file in the same directory, then rename
	tmpFile := filepath.Join(ventoyDir, fmt.Sprintf(".%s.tmp.%d", filename, os.Getpid()))
	if err := os.WriteFile(tmpFile, embeddedData, 0644); err != nil {
		// Fallback to direct write if temporary file write fails
		if directErr := os.WriteFile(targetFile, embeddedData, 0644); directErr != nil {
			return fmt.Errorf("failed to write %s directly: %w (temp err: %v)", filename, directErr, err)
		}
		return nil
	}

	if err := os.Rename(tmpFile, targetFile); err != nil {
		_ = os.Remove(tmpFile)
		// On non-standard filesystems or Windows locks, fallback to direct write
		if directErr := os.WriteFile(targetFile, embeddedData, 0644); directErr != nil {
			return fmt.Errorf("failed to commit %s: %w", filename, directErr)
		}
	}

	return nil
}
