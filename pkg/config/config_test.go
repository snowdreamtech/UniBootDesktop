// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type memorySecretStore struct {
	password string
}

func newMemorySecretStore() *memorySecretStore {
	return &memorySecretStore{}
}

func (s *memorySecretStore) Get(_ string, _ string) (string, error) {
	if s.password == "" {
		return "", errors.New("credential not found")
	}
	return s.password, nil
}

func (s *memorySecretStore) Set(_ string, _ string, password string) error {
	s.password = password
	return nil
}

func (s *memorySecretStore) Delete(_ string, _ string) error {
	if s.password == "" {
		return errors.New("credential not found")
	}
	s.password = ""
	return nil
}

func TestDefaultConfig(t *testing.T) {
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", t.TempDir())
	cfg := GetDefaultConfig()
	if cfg.Mode != "cloud" {
		t.Errorf("expected default Mode 'cloud', got %s", cfg.Mode)
	}
	if cfg.GithubProxy != "" {
		t.Errorf("expected default GithubProxy '', got %s", cfg.GithubProxy)
	}
	if cfg.FileSystem != "exFAT" {
		t.Errorf("expected default FileSystem 'exFAT', got %s", cfg.FileSystem)
	}
	if cfg.VentoyPath == "" || cfg.UniBootPath == "" {
		t.Fatalf("expected default Ventoy and UniBoot paths, got Ventoy=%q UniBoot=%q", cfg.VentoyPath, cfg.UniBootPath)
	}
	if cfg.EnableTray != false {
		t.Errorf("expected default EnableTray false, got %v", cfg.EnableTray)
	}
	if cfg.CloseAction != "quit" {
		t.Errorf("expected default CloseAction 'quit', got %s", cfg.CloseAction)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", tmpDir)

	cfg := GetDefaultConfig()
	cfg.GithubProxy = "https://proxy.example.com/"
	cfg.FileSystem = "NTFS"
	cfg.ProxyProtocol = "socks5"
	cfg.ProxyHost = "127.0.0.1"
	cfg.ProxyPort = 1080
	cfg.ProxyUser = "dummy_user"
	cfg.ProxyPassword = "dummy_password"
	cfg.EnableTray = true
	cfg.CloseAction = "minimize_to_tray"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if loaded.GithubProxy != "https://proxy.example.com/" {
		t.Errorf("expected GithubProxy 'https://proxy.example.com/', got %s", loaded.GithubProxy)
	}
	if loaded.FileSystem != "NTFS" {
		t.Errorf("expected FileSystem 'NTFS', got %s", loaded.FileSystem)
	}
	if loaded.EnableTray != true || loaded.CloseAction != "minimize_to_tray" {
		t.Errorf("tray config mismatch: %+v", loaded)
	}
	if loaded.ProxyProtocol != "socks5" || loaded.ProxyHost != "127.0.0.1" || loaded.ProxyPort != 1080 || loaded.ProxyUser != "dummy_user" || loaded.ProxyPassword != "" {
		t.Errorf("proxy config mismatch: %+v", loaded)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "unibootdesktop.toml"))
	if err != nil {
		t.Fatalf("read saved config failed: %v", err)
	}
	if strings.Contains(string(data), "dummy_password") || strings.Contains(string(data), "proxyPassword") {
		t.Fatalf("proxy password must not be written to config file: %s", data)
	}
	info, err := os.Stat(filepath.Join(tmpDir, "unibootdesktop.toml"))
	if err != nil {
		t.Fatalf("stat saved config failed: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Fatalf("expected config mode 0600, got %04o", mode)
	}
}

func TestSaveSkipsEmptyConfigDuringInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", filepath.Join(tmpDir, "data"))

	cfg := &AppConfig{}
	if err := cfg.Save(); err != nil {
		t.Fatalf("empty config save should be a no-op during initialization: %v", err)
	}

	configPath := filepath.Join(tmpDir, "unibootdesktop.toml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("expected empty config to skip writing defaults, but %s exists", configPath)
	}
	if cfg.VentoyPath != "" || cfg.UniBootPath != "" {
		t.Fatalf("expected empty config to remain unset, got Ventoy=%q UniBoot=%q", cfg.VentoyPath, cfg.UniBootPath)
	}
}

func TestLoadMigratesEmptyFirmwareDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", filepath.Join(tmpDir, "data"))

	legacyConfig := "mode = 'cloud'\nventoyPath = ''\nunibootPath = ''\n"
	configPath := filepath.Join(tmpDir, "unibootdesktop.toml")
	if err := os.WriteFile(configPath, []byte(legacyConfig), 0600); err != nil {
		t.Fatalf("write legacy config failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load legacy config failed: %v", err)
	}
	if loaded.VentoyPath == "" || loaded.UniBootPath == "" {
		t.Fatalf("expected migrated firmware directories, got Ventoy=%q UniBoot=%q", loaded.VentoyPath, loaded.UniBootPath)
	}

	migrated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read migrated config failed: %v", err)
	}
	if !strings.Contains(string(migrated), "ventoyPath = '") || !strings.Contains(string(migrated), "unibootPath = '") {
		t.Fatalf("expected migrated paths in config: %s", migrated)
	}
}

func TestLoadRestoresCorruptConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", filepath.Join(tmpDir, "data"))

	configPath := filepath.Join(tmpDir, "unibootdesktop.toml")
	if err := os.WriteFile(configPath, []byte("not valid toml =\n"), 0600); err != nil {
		t.Fatalf("write corrupt config failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("restore corrupt config failed: %v", err)
	}
	if loaded.VentoyPath == "" || loaded.UniBootPath == "" {
		t.Fatalf("expected default paths after restore, got Ventoy=%q UniBoot=%q", loaded.VentoyPath, loaded.UniBootPath)
	}
	if _, err := os.Stat(configPath + ".corrupt"); err != nil {
		t.Fatalf("expected corrupt config backup: %v", err)
	}
}

func TestHealthCheckReportsCorruptConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIBOOTDESKTOP_DATA_DIR", filepath.Join(tmpDir, "data"))

	configPath := filepath.Join(tmpDir, "unibootdesktop.toml")
	if err := os.WriteFile(configPath, []byte("bad =\n"), 0o600); err != nil {
		t.Fatalf("write corrupt config: %v", err)
	}
	if err := os.WriteFile(configPath+".corrupt", []byte("mode = 'cloud'\n"), 0o600); err != nil {
		t.Fatalf("write stale backup file: %v", err)
	}

	health, err := HealthCheck()
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if health.Valid {
		t.Fatal("expected invalid config to be reported unhealthy")
	}
	if !strings.Contains(health.Issue, "invalid TOML") {
		t.Fatalf("expected invalid TOML issue, got %q", health.Issue)
	}
	if !health.HasBackup {
		t.Fatal("expected backup flag to be true when a corrupt backup exists")
	}
}

func TestProxyPasswordStore(t *testing.T) {
	previousStore := proxyPasswordStore
	store := newMemorySecretStore()
	proxyPasswordStore = store
	t.Cleanup(func() { proxyPasswordStore = previousStore })

	if err := SaveProxyPassword("secret"); err != nil {
		t.Fatalf("save proxy password failed: %v", err)
	}
	password, err := LoadProxyPassword()
	if err != nil {
		t.Fatalf("load proxy password failed: %v", err)
	}
	if password != "secret" {
		t.Fatalf("expected stored password, got %q", password)
	}

	if err := DeleteProxyPassword(); err != nil {
		t.Fatalf("delete proxy password failed: %v", err)
	}
	if _, err := LoadProxyPassword(); err == nil {
		t.Fatal("expected deleted proxy password to be unavailable")
	}
}
