// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"testing"
)

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
	t.Setenv("UNIGO_DATA_DIR", tmpDir)

	cfg := GetDefaultConfig()
	cfg.GithubProxy = "https://mirror.ghproxy.com/"
	cfg.FileSystem = "NTFS"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if loaded.GithubProxy != "https://mirror.ghproxy.com/" {
		t.Errorf("expected GithubProxy 'https://mirror.ghproxy.com/', got %s", loaded.GithubProxy)
	}
	if loaded.FileSystem != "NTFS" {
		t.Errorf("expected FileSystem 'NTFS', got %s", loaded.FileSystem)
	}
}
