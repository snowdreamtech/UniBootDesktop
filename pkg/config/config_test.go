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
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIGODESKTOP_CONFIG_DIR", tmpDir)
	t.Setenv("UNIGO_DATA_DIR", tmpDir)

	cfg := GetDefaultConfig()
	cfg.GithubProxy = "https://proxy.example.com/"
	cfg.FileSystem = "NTFS"
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

	if loaded.GithubProxy != "https://proxy.example.com/" {
		t.Errorf("expected GithubProxy 'https://proxy.example.com/', got %s", loaded.GithubProxy)
	}
	if loaded.FileSystem != "NTFS" {
		t.Errorf("expected FileSystem 'NTFS', got %s", loaded.FileSystem)
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
