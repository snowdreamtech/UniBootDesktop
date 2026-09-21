// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unigodesktop/internal/env"
)

var configWriteMutex sync.Mutex

func isEmptyConfig(cfg *AppConfig) bool {
	if cfg == nil {
		return true
	}
	return cfg.Mode == "" &&
		cfg.AutoCheckUpdate == false &&
		cfg.Theme == "" &&
		cfg.GithubProxy == "" &&
		cfg.FileSystem == "" &&
		cfg.ProxyProtocol == "" &&
		cfg.ProxyHost == "" &&
		cfg.ProxyPort == 0 &&
		cfg.ProxyUser == "" &&
		cfg.ProxyPassword == "" &&
		cfg.Language == "" &&
		cfg.VentoyPath == "" &&
		cfg.UniBootPath == "" &&
		cfg.VentoySecureBoot == false &&
		cfg.VentoyPartitionStyle == "" &&
		cfg.VentoyReserveSpace == 0 &&
		cfg.VentoyWin11Bypass == false &&
		cfg.VentoyMenuTimeout == 0 &&
		cfg.AutoEjectAfterDeploy == false
}

func normalizePersistedDefaults(cfg *AppConfig) bool {
	if cfg == nil {
		return false
	}
	updated := false
	if cfg.VentoyPath == "" {
		cfg.VentoyPath = env.GetVentoyDir()
		updated = true
	}
	if cfg.UniBootPath == "" {
		cfg.UniBootPath = env.GetFirmwareDir()
		updated = true
	}
	return updated
}

func validateTOMLData(data []byte) error {
	var parsed AppConfig
	if err := toml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("toml unmarshal: %w", err)
	}
	return nil
}

func restoreConfigBackup(cfgPath, backupPath string) error {
	if backupPath == "" {
		return fmt.Errorf("no config backup available")
	}
	if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove failed config: %w", err)
	}
	if err := os.Rename(backupPath, cfgPath); err != nil {
		return fmt.Errorf("restore backup: %w", err)
	}
	return nil
}

func backupExistingConfig(cfgPath string) (string, bool, error) {
	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("stat current config: %w", err)
	}
	backupPath := cfgPath + ".corrupt"
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return "", false, fmt.Errorf("remove stale backup: %w", err)
	}
	if err := os.Rename(cfgPath, backupPath); err != nil {
		return "", false, fmt.Errorf("backup config: %w", err)
	}
	return backupPath, true, nil
}

func writeConfigAtomically(cfgPath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(cfgPath), ".unibootdesktop-*.toml")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := file.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync temp config: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpPath, cfgPath); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}
	if err := os.Chmod(cfgPath, 0o600); err != nil {
		return fmt.Errorf("set config permissions: %w", err)
	}
	return nil
}

func atomicRestoreDefaultConfig(cfgPath string) error {
	defaultCfg := GetDefaultConfig()
	data, err := toml.Marshal(defaultCfg)
	if err != nil {
		return fmt.Errorf("marshal default config: %w", err)
	}
	if err := writeConfigAtomically(cfgPath, data); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}
	return nil
}

// AppConfig represents application-wide configuration parameters.
type AppConfig struct {
	Mode                 string `json:"mode" toml:"mode"`                                 // Hybrid Mode or Cloud Mode
	AutoCheckUpdate      bool   `json:"autoCheckUpdate" toml:"autoCheckUpdate"`           // Automatically check for updates
	Theme                string `json:"theme" toml:"theme"`                               // UI theme preference (dark/light)
	GithubProxy          string `json:"githubProxy" toml:"githubProxy"`                   // GitHub proxy server URL (e.g. https://proxy.example.com/)
	FileSystem           string `json:"fileSystem" toml:"fileSystem"`                     // Default file system for Hybrid Mode (exFAT/NTFS/FAT32/ext4)
	ProxyProtocol        string `json:"proxyProtocol" toml:"proxyProtocol"`               // Network proxy protocol: direct, http, https, socks4, socks5
	ProxyHost            string `json:"proxyHost" toml:"proxyHost"`                       // Network proxy server host
	ProxyPort            int    `json:"proxyPort" toml:"proxyPort"`                       // Network proxy server port
	ProxyUser            string `json:"proxyUser" toml:"proxyUser"`                       // Network proxy authentication username
	ProxyPassword        string `json:"proxyPassword" toml:"-"`                           // Transient input; persisted in the OS credential store
	Language             string `json:"language" toml:"language"`                         // UI Language (auto/zh-CN/en-US/zh-TW)
	VentoyPath           string `json:"ventoyPath" toml:"ventoyPath"`                     // Path to official Ventoy CLI directory / executable
	UniBootPath          string `json:"unibootPath" toml:"unibootPath"`                   // Custom local UniBoot firmware directory path (empty uses system default)
	VentoySecureBoot     bool   `json:"ventoySecureBoot" toml:"ventoySecureBoot"`         // Enable Ventoy Secure Boot support (-s)
	VentoyPartitionStyle string `json:"ventoyPartitionStyle" toml:"ventoyPartitionStyle"` // Ventoy partition style: GPT or MBR
	VentoyReserveSpace   int    `json:"ventoyReserveSpace" toml:"ventoyReserveSpace"`     // Reserved space at end of disk (MB)
	VentoyWin11Bypass    bool   `json:"ventoyWin11Bypass" toml:"ventoyWin11Bypass"`       // Auto inject Win11 TPM/CPU bypass patch
	VentoyMenuTimeout    int    `json:"ventoyMenuTimeout" toml:"ventoyMenuTimeout"`       // Auto boot timeout (seconds)
	AutoEjectAfterDeploy bool   `json:"autoEjectAfterDeploy" toml:"autoEjectAfterDeploy"` // Automatically safely eject target disks after successful deployment
}

// GetDefaultConfig returns the default application configuration.
func GetDefaultConfig() *AppConfig {
	return &AppConfig{
		Mode:                 "cloud", // Cloud Mode by default
		AutoCheckUpdate:      true,
		Theme:                "dark",
		Language:             "auto", // Auto detect OS system language by default
		GithubProxy:          "",     // Default to empty (Direct connection, no hardcoded proxy preset)
		FileSystem:           "exFAT",
		ProxyProtocol:        "direct",
		ProxyHost:            "",
		ProxyPort:            0,
		ProxyUser:            "",
		ProxyPassword:        "",
		VentoyPath:           env.GetVentoyDir(),
		UniBootPath:          env.GetFirmwareDir(),
		VentoySecureBoot:     true,  // Official Ventoy default: Enabled (Checked)
		VentoyPartitionStyle: "MBR", // Official Ventoy default: MBR
		VentoyReserveSpace:   0,     // Official Ventoy default: 0 MB
		VentoyWin11Bypass:    false, // Official Ventoy default: Disabled (False)
		VentoyMenuTimeout:    0,     // Official Ventoy default: 0 (No timeout / wait indefinitely)
		AutoEjectAfterDeploy: false, // Default: do NOT auto eject — user should verify the disk first
	}
}

// Load reads application configuration from the user config directory.
func Load() (*AppConfig, error) {
	cfgPath := env.GetGlobalConfigPath()
	if data, err := os.ReadFile(cfgPath); err == nil {
		cfg := GetDefaultConfig()
		if err := toml.Unmarshal(data, cfg); err != nil {
			backupPath := cfgPath + ".corrupt"
			if renameErr := os.Rename(cfgPath, backupPath); renameErr != nil {
				return nil, fmt.Errorf("parse config error: %w (backup failed: %v)", err, renameErr)
			}
			cfg = GetDefaultConfig()
			if saveErr := cfg.Save(); saveErr != nil {
				return nil, fmt.Errorf("parse config error: %w (default restore failed: %v)", err, saveErr)
			}
			return cfg, nil
		}
		if normalizePersistedDefaults(cfg) {
			if err := cfg.Save(); err != nil {
				return nil, fmt.Errorf("migrate default firmware directories: %w", err)
			}
		}
		return cfg, nil
	}
	defaultCfg := GetDefaultConfig()
	if err := defaultCfg.Save(); err != nil {
		return nil, fmt.Errorf("initialize default config: %w", err)
	}
	return defaultCfg, nil
}

// Save writes application configuration to disk.
func (c *AppConfig) Save() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	configWriteMutex.Lock()
	defer configWriteMutex.Unlock()

	cfgPath := env.GetGlobalConfigPath()
	lockFile, unlockLock, err := acquireConfigLock(cfgPath)
	if err != nil {
		return fmt.Errorf("acquire config lock: %w", err)
	}
	defer unlockLock()
	defer func() {
		if lockFile != nil {
			_ = lockFile.Close()
		}
	}()

	if isEmptyConfig(c) {
		if _, statErr := os.Stat(cfgPath); statErr != nil && os.IsNotExist(statErr) {
			return nil
		}
		return nil
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config error: %w", err)
	}
	if err := validateTOMLData(data); err != nil {
		return fmt.Errorf("validate config payload: %w", err)
	}

	backupFile, _, backupErr := backupExistingConfig(cfgPath)
	if backupErr != nil {
		return fmt.Errorf("prepare backup for config write: %w", backupErr)
	}
	if writeErr := writeConfigAtomically(cfgPath, data); writeErr != nil {
		if backupFile != "" {
			if restoreErr := restoreConfigBackup(cfgPath, backupFile); restoreErr != nil {
				return fmt.Errorf("save failed and backup restore failed: %w (restoreErr=%v)", writeErr, restoreErr)
			}
		}
		return fmt.Errorf("write config file: %w", writeErr)
	}

	if err := validateTOMLDataFromFile(cfgPath); err != nil {
		if backupFile != "" {
			if restoreErr := restoreConfigBackup(cfgPath, backupFile); restoreErr != nil {
				return fmt.Errorf("post-write validation failed and backup restore failed: %w (restoreErr=%v)", err, restoreErr)
			}
		} else if restoreErr := atomicRestoreDefaultConfig(cfgPath); restoreErr != nil {
			return fmt.Errorf("post-write validation failed and default restore failed: %w (restoreErr=%v)", err, restoreErr)
		}
		return fmt.Errorf("post-write validation failed: %w", err)
	}

	if backupFile != "" {
		_ = os.Remove(backupFile)
	}
	return nil
}

func validateTOMLDataFromFile(cfgPath string) error {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	return validateTOMLData(data)
}

// ConfigHealth describes the current health of the persisted config state.
type ConfigHealth struct {
	ConfigPath      string
	Exists          bool
	Valid           bool
	HasBackup       bool
	DefaultPathsSet bool
	Issue           string
}

// HealthCheck inspects the persisted config file and reports whether it is readable,
// whether a backup exists, and whether default UniBoot/Ventoy paths are available.
func HealthCheck() (*ConfigHealth, error) {
	cfgPath := env.GetGlobalConfigPath()
	status := &ConfigHealth{
		ConfigPath:      cfgPath,
		Exists:          false,
		Valid:           true,
		HasBackup:       false,
		DefaultPathsSet: true,
		Issue:           "",
	}

	if _, err := os.Stat(cfgPath); err == nil {
		status.Exists = true
		data, readErr := os.ReadFile(cfgPath)
		if readErr != nil {
			status.Valid = false
			status.Issue = fmt.Sprintf("config file unreadable: %v", readErr)
			return status, nil
		}
		if unmarshalErr := validateTOMLData(data); unmarshalErr != nil {
			status.Valid = false
			status.Issue = fmt.Sprintf("invalid TOML: %v", unmarshalErr)
		}
	} else if !os.IsNotExist(err) {
		status.Valid = false
		status.Issue = fmt.Sprintf("config path check failed: %v", err)
		return status, nil
	}

	if _, err := os.Stat(cfgPath + ".corrupt"); err == nil {
		status.HasBackup = true
	}

	defaults := GetDefaultConfig()
	status.DefaultPathsSet = defaults.VentoyPath != "" && defaults.UniBootPath != ""
	if !status.Valid {
		status.DefaultPathsSet = status.DefaultPathsSet && defaults.VentoyPath != "" && defaults.UniBootPath != ""
	}
	return status, nil
}
