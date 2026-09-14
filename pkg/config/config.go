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
	GithubProxy     string `json:"githubProxy" toml:"githubProxy"`           // GitHub proxy server URL (e.g. https://proxy.example.com/)
	FileSystem      string `json:"fileSystem" toml:"fileSystem"`             // Default file system for Mode A (exFAT/NTFS/FAT32/ext4)
	ProxyProtocol   string `json:"proxyProtocol" toml:"proxyProtocol"`       // Network proxy protocol: direct, http, https, socks4, socks5
	ProxyHost       string `json:"proxyHost" toml:"proxyHost"`               // Network proxy server host
	ProxyPort       int    `json:"proxyPort" toml:"proxyPort"`               // Network proxy server port
	ProxyUser       string `json:"proxyUser" toml:"proxyUser"`               // Network proxy authentication username
	ProxyPassword   string `json:"proxyPassword" toml:"proxyPassword"`       // Network proxy authentication password
	VentoyPath           string `json:"ventoyPath" toml:"ventoyPath"`                     // Path to official Ventoy CLI directory / executable
	VentoySecureBoot     bool   `json:"ventoySecureBoot" toml:"ventoySecureBoot"`         // Enable Ventoy Secure Boot support (-s)
	VentoyPartitionStyle string `json:"ventoyPartitionStyle" toml:"ventoyPartitionStyle"` // Ventoy partition style: GPT or MBR
	VentoyReserveSpace   int    `json:"ventoyReserveSpace" toml:"ventoyReserveSpace"`     // Reserved space at end of disk (MB)
	VentoyWin11Bypass    bool   `json:"ventoyWin11Bypass" toml:"ventoyWin11Bypass"`       // Auto inject Win11 TPM/CPU bypass patch
	VentoyMenuTimeout    int    `json:"ventoyMenuTimeout" toml:"ventoyMenuTimeout"`       // Auto boot timeout (seconds)
}

// GetDefaultConfig returns the default application configuration.
func GetDefaultConfig() *AppConfig {
	return &AppConfig{
		Mode:                 "cloud", // Mode B Cloud Pure Mode by default
		AutoCheckUpdate:      true,
		Theme:                "dark",
		GithubProxy:          "", // Default to empty (Direct connection, no hardcoded proxy preset)
		FileSystem:           "exFAT",
		ProxyProtocol:        "direct",
		ProxyHost:            "",
		ProxyPort:            0,
		ProxyUser:            "",
		ProxyPassword:        "",
		VentoyPath:           "",
		VentoySecureBoot:     true,
		VentoyPartitionStyle: "GPT",
		VentoyReserveSpace:   0,
		VentoyWin11Bypass:    true,
		VentoyMenuTimeout:    10,
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
