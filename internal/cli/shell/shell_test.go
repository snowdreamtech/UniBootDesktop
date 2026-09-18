// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unigodesktop/internal/cli/output"
)

func TestDetectShell(t *testing.T) {
	os.Setenv("UNIBOOTDESKTOP_SHELL", "/bin/zsh")
	s, err := DetectShell()
	if err != nil {
		t.Fatalf("unexpected error detecting zsh: %v", err)
	}
	if s != ShellZsh {
		t.Errorf("expected ShellZsh, got: %s", s)
	}

	os.Setenv("UNIBOOTDESKTOP_SHELL", "/usr/bin/bash")
	s, err = DetectShell()
	if err != nil {
		t.Fatalf("unexpected error detecting bash: %v", err)
	}
	if s != ShellBash {
		t.Errorf("expected ShellBash, got: %s", s)
	}

	os.Setenv("UNIBOOTDESKTOP_SHELL", "/usr/local/bin/fish")
	s, err = DetectShell()
	if err != nil {
		t.Fatalf("unexpected error detecting fish: %v", err)
	}
	if s != ShellFish {
		t.Errorf("expected ShellFish, got: %s", s)
	}
	os.Unsetenv("UNIBOOTDESKTOP_SHELL")
}

func TestShellConfigManagerGetConfigPath(t *testing.T) {
	fmtter := output.DefaultFormatter()
	mgr := NewShellConfigManager(fmtter, true)

	zshPath, err := mgr.GetConfigPath(ShellZsh)
	if err != nil {
		t.Fatalf("failed to get zsh path: %v", err)
	}
	if filepath.Base(zshPath) != ".zshrc" {
		t.Errorf("expected .zshrc in path, got %s", zshPath)
	}

	bashPath, err := mgr.GetConfigPath(ShellBash)
	if err != nil {
		t.Fatalf("failed to get bash path: %v", err)
	}
	if filepath.Base(bashPath) != ".bashrc" {
		t.Errorf("expected .bashrc in path, got %s", bashPath)
	}

	fishPath, err := mgr.GetConfigPath(ShellFish)
	if err != nil {
		t.Fatalf("failed to get fish path: %v", err)
	}
	if filepath.Base(fishPath) != "config.fish" {
		t.Errorf("expected config.fish in path, got %s", fishPath)
	}
}
