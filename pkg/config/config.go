// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unigodesktop/internal/env"
)

// AppConfig represents universal desktop application settings.
type AppConfig struct {
	AutoCheckUpdate bool   `json:"autoCheckUpdate" toml:"autoCheckUpdate"` // Automatically check for updates
	Theme           string `json:"theme" toml:"theme"`                     // UI theme preference (dark, light, system)
	Language        string `json:"language" toml:"language"`               // UI Language (auto, zh-CN, en-US, etc.)
	GithubProxy     string `json:"githubProxy" toml:"githubProxy"`         // GitHub proxy acceleration mirror
	ProxyProtocol   string `json:"proxyProtocol" toml:"proxyProtocol"`     // Network proxy protocol: direct, http, https, socks5
	ProxyHost       string `json:"proxyHost" toml:"proxyHost"`             // Network proxy server host
	ProxyPort       int    `json:"proxyPort" toml:"proxyPort"`             // Network proxy server port
	ProxyUser       string `json:"proxyUser" toml:"proxyUser"`             // Network proxy authentication username
	ProxyPassword   string `json:"proxyPassword" toml:"proxyPassword"`     // Network proxy authentication password
}

// GetDefaultConfig returns the default application configuration.
func GetDefaultConfig() *AppConfig {
	return &AppConfig{
		AutoCheckUpdate: true,
		Theme:           "system",
		Language:        "auto", // Auto detect OS system language by default
		GithubProxy:     "",     // Direct connection by default
		ProxyProtocol:   "direct",
		ProxyHost:       "",
		ProxyPort:       0,
		ProxyUser:       "",
		ProxyPassword:   "",
	}
}

// Load reads application configuration from the user config directory.
func Load() (*AppConfig, error) {
	cfgPath := env.GetGlobalConfigPath()
	if data, err := os.ReadFile(cfgPath); err == nil {
		cfg := GetDefaultConfig()
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config error: %w", err)
		}
		return cfg, nil
	}
	return GetDefaultConfig(), nil
}

// Save writes application configuration to disk.
func (c *AppConfig) Save() error {
	cfgPath := env.GetGlobalConfigPath()
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return fmt.Errorf("create config dir error: %w", err)
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config error: %w", err)
	}
	return os.WriteFile(cfgPath, data, 0600)
}
