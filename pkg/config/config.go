// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/snowdreamtech/unigodesktop/internal/env"
)

var configWriteMutex sync.Mutex

var (
	configCacheMu      sync.RWMutex
	configCachePath    string
	configCacheModTime time.Time
	configCache        *AppConfig
)

// InvalidateConfigCache discards the in-memory config cache, forcing the next
// Load() call to re-read the config file from disk. Useful in tests or
// whenever the config file may have been modified externally.
func InvalidateConfigCache() {
	configCacheMu.Lock()
	configCache = nil
	configCachePath = ""
	configCacheModTime = time.Time{}
	configCacheMu.Unlock()
}

func isEmptyConfig(cfg *AppConfig) bool {
	if cfg == nil {
		return true
	}
	return !cfg.Debug &&
		!cfg.AutoCheckUpdate &&
		cfg.Theme == "" &&
		cfg.GithubProxy == "" &&
		cfg.ProxyProtocol == "" &&
		cfg.ProxyHost == "" &&
		cfg.ProxyPort == 0 &&
		cfg.ProxyUser == "" &&
		cfg.ProxyPassword == "" &&
		cfg.Language == "" &&
		!cfg.EnableTray &&
		cfg.CloseAction == ""
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
	file, err := os.CreateTemp(filepath.Dir(cfgPath), ".unigodesktop-*.toml")
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

// AppConfig represents universal desktop application settings.
type AppConfig struct {
	Debug           bool   `json:"debug" toml:"debug"`                     // Debug mode flag
	AutoCheckUpdate bool   `json:"autoCheckUpdate" toml:"autoCheckUpdate"` // Automatically check for updates
	Theme           string `json:"theme" toml:"theme"`                     // UI theme preference (dark, light, system)
	Language        string `json:"language" toml:"language"`               // UI Language (auto, zh-CN, en-US, etc.)
	EnableTray      bool   `json:"enableTray" toml:"enableTray"`           // Whether to enable system tray (default: false)
	CloseAction     string `json:"closeAction" toml:"closeAction"`         // Window close action: "quit" or "minimize_to_tray" (default: "quit")
	GithubProxy     string `json:"githubProxy" toml:"githubProxy"`         // GitHub proxy acceleration mirror
	ProxyProtocol   string `json:"proxyProtocol" toml:"proxyProtocol"`     // Network proxy protocol: direct, http, https, socks5
	ProxyHost       string `json:"proxyHost" toml:"proxyHost"`             // Network proxy server host
	ProxyPort       int    `json:"proxyPort" toml:"proxyPort"`             // Network proxy server port
	ProxyUser       string `json:"proxyUser" toml:"proxyUser"`             // Network proxy authentication username
	ProxyPassword   string `json:"proxyPassword" toml:"-"`                 // Stored securely in OS credential store, excluded from TOML file
}

// GetDefaultConfig returns the default application configuration.
func GetDefaultConfig() *AppConfig {
	return &AppConfig{
		Debug:           false,
		AutoCheckUpdate: true,
		Theme:           "system",
		Language:        "auto", // Auto detect OS system language by default
		EnableTray:      false,  // Disabled by default
		CloseAction:     "quit", // Quit application on close by default
		GithubProxy:     "",     // Direct connection by default
		ProxyProtocol:   "direct",
		ProxyHost:       "",
		ProxyPort:       0,
		ProxyUser:       "",
		ProxyPassword:   "",
	}
}

// Load reads application configuration from the user config directory.
// Repeated calls return a cached copy without touching disk as long as the file
// on disk has not been modified.
func Load() (*AppConfig, error) {
	cfgPath := env.GetGlobalConfigPath()

	// Fast path: return cached config without parsing TOML if file is unchanged.
	if fi, statErr := os.Stat(cfgPath); statErr == nil {
		configCacheMu.RLock()
		if configCache != nil && configCachePath == cfgPath && fi.ModTime().Equal(configCacheModTime) {
			copy := *configCache
			configCacheMu.RUnlock()
			return &copy, nil
		}
		configCacheMu.RUnlock()
	}

	if data, err := os.ReadFile(cfgPath); err == nil {
		cfg := GetDefaultConfig()
		if err := toml.Unmarshal(data, cfg); err != nil {
			backupPath := cfgPath + ".corrupt"
			if renameErr := os.Rename(cfgPath, backupPath); renameErr != nil {
				return nil, fmt.Errorf("parse config error: %w (backup failed: %w)", err, renameErr)
			}
			cfg = GetDefaultConfig()
			if saveErr := cfg.Save(); saveErr != nil {
				return nil, fmt.Errorf("parse config error: %w (default restore failed: %w)", err, saveErr)
			}
			return cfg, nil
		}

		// Sanitize legacy or default public proxy mirrors to maintain legal compliance
		if strings.Contains(cfg.GithubProxy, "ghproxy") || strings.Contains(cfg.GithubProxy, "ghfast") {
			cfg.GithubProxy = ""
			_ = cfg.Save()
		}

		// Populate cache
		configCacheMu.Lock()
		cacheCopy := *cfg
		configCache = &cacheCopy
		configCachePath = cfgPath
		if fi, statErr := os.Stat(cfgPath); statErr == nil {
			configCacheModTime = fi.ModTime()
		}
		configCacheMu.Unlock()
		return cfg, nil
	}

	defaultCfg := GetDefaultConfig()
	if err := defaultCfg.Save(); err != nil {
		return nil, fmt.Errorf("initialize default config: %w", err)
	}
	return defaultCfg, nil
}

// Save writes application configuration to disk atomically with file locking and backup rollback.
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
				return fmt.Errorf("save failed and backup restore failed: %w (restoreErr=%w)", writeErr, restoreErr)
			}
		}
		return fmt.Errorf("write config file: %w", writeErr)
	}

	if err := validateTOMLDataFromFile(cfgPath); err != nil {
		if backupFile != "" {
			if restoreErr := restoreConfigBackup(cfgPath, backupFile); restoreErr != nil {
				return fmt.Errorf("post-write validation failed and backup restore failed: %w (restoreErr=%w)", err, restoreErr)
			}
		} else if restoreErr := atomicRestoreDefaultConfig(cfgPath); restoreErr != nil {
			return fmt.Errorf("post-write validation failed and default restore failed: %w (restoreErr=%w)", err, restoreErr)
		}
		return fmt.Errorf("post-write validation failed: %w", err)
	}

	if backupFile != "" {
		_ = os.Remove(backupFile)
	}

	// Update in-memory cache
	configCacheMu.Lock()
	cacheCopy := *c
	cacheCopy.ProxyPassword = ""
	configCache = &cacheCopy
	configCachePath = cfgPath
	if fi, statErr := os.Stat(cfgPath); statErr == nil {
		configCacheModTime = fi.ModTime()
	}
	configCacheMu.Unlock()

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
	ConfigPath string
	Exists     bool
	Valid      bool
	HasBackup  bool
	Issue      string
}

// HealthCheck inspects the persisted config file and reports whether it is readable,
// whether a backup exists, and whether configuration format is valid.
func HealthCheck() (*ConfigHealth, error) {
	cfgPath := env.GetGlobalConfigPath()
	status := &ConfigHealth{
		ConfigPath: cfgPath,
		Exists:     false,
		Valid:      true,
		HasBackup:  false,
		Issue:      "",
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

	return status, nil
}
