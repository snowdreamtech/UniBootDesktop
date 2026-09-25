// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	if !cfg.AutoCheckUpdate {
		t.Errorf("expected default AutoCheckUpdate true, got %v", cfg.AutoCheckUpdate)
	}
	if cfg.Theme != "system" {
		t.Errorf("expected default Theme 'system', got %s", cfg.Theme)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default Language 'auto', got %s", cfg.Language)
	}
	if cfg.GithubProxy != "" {
		t.Errorf("expected default GithubProxy '', got %s", cfg.GithubProxy)
	}
	if cfg.ProxyProtocol != "direct" {
		t.Errorf("expected default ProxyProtocol 'direct', got %s", cfg.ProxyProtocol)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIGODESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIGO_DATA_DIR", tmpDir)

	cfg := GetDefaultConfig()
	cfg.Theme = "dark"
	cfg.Language = "zh-CN"
	cfg.GithubProxy = "https://proxy.example.com/"
	cfg.ProxyProtocol = "socks5"
	cfg.ProxyHost = "127.0.0.1"
	cfg.ProxyPort = 1080
	cfg.ProxyUser = "dummy_user"
	cfg.ProxyPassword = "dummy_password"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if loaded.Theme != "dark" {
		t.Errorf("expected Theme 'dark', got %s", loaded.Theme)
	}
	if loaded.Language != "zh-CN" {
		t.Errorf("expected Language 'zh-CN', got %s", loaded.Language)
	}
	if loaded.GithubProxy != "https://proxy.example.com/" {
		t.Errorf("expected GithubProxy 'https://proxy.example.com/', got %s", loaded.GithubProxy)
	}
	if loaded.ProxyProtocol != "socks5" || loaded.ProxyHost != "127.0.0.1" || loaded.ProxyPort != 1080 || loaded.ProxyUser != "dummy_user" || loaded.ProxyPassword != "dummy_password" {
		t.Errorf("proxy config mismatch: %+v", loaded)
	}
}
