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

// AppConfig represents application-wide configuration parameters.
type AppConfig struct {
	Mode            string `json:"mode" toml:"mode"`                         // Mode A (hybrid) or Mode B (cloud)
	AutoCheckUpdate bool   `json:"autoCheckUpdate" toml:"autoCheckUpdate"`   // Automatically check for updates
	Theme           string `json:"theme" toml:"theme"`                       // UI theme preference (dark/light)
	GithubProxy     string `json:"githubProxy" toml:"githubProxy"`           // GitHub proxy server URL (e.g. https://ghproxy.net/)
	FileSystem      string `json:"fileSystem" toml:"fileSystem"`             // Default file system for Mode A (exFAT/NTFS/FAT32/ext4)
}

// GetDefaultConfig returns the default application configuration.
func GetDefaultConfig() *AppConfig {
	return &AppConfig{
		Mode:            "cloud", // Mode B Cloud Pure Mode by default
		AutoCheckUpdate: true,
		Theme:           "dark",
		GithubProxy:     "https://ghproxy.net/",
		FileSystem:      "exFAT",
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
	return os.WriteFile(cfgPath, data, 0644)
}
