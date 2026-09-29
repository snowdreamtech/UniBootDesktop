// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_LifecycleAndAPIs(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	ctx := context.Background()
	app.startup(ctx)

	// Greet
	assert.Equal(t, "Hello World, Welcome to UniGoDesktop!", app.Greet(""))
	assert.Equal(t, "Hello Alice, Welcome to UniGoDesktop!", app.Greet("Alice"))

	// HelloInfo
	info := app.GetHelloInfo()
	assert.NotNil(t, info)
	assert.Contains(t, info.Greeting, "Hello World")

	// SystemInfo
	sys := app.GetSystemInfo()
	assert.NotNil(t, sys)
	assert.Equal(t, "UniGoDesktop", sys.AppName)

	// Config
	cfg, err := app.GetConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	err = app.SaveConfig(cfg)
	assert.NoError(t, err)
}

func TestResolveWindowsUserDataPath(t *testing.T) {
	// Verify that resolving the Windows user data path executes safely
	path := resolveWindowsUserDataPath()
	_ = path
}

func TestApp_RestartApp(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// Mock execCommand to a command that succeeds without launching another app instance
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Use echo or true which exits immediately
		return exec.Command("true")
	}

	err := app.RestartApp()
	assert.NoError(t, err)
}

func TestApp_RestartApp_WithPendingUpdate(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	origExec := execCommand
	defer func() { execCommand = origExec }()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "apply_update.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0755))

	dataDir := env.GetDataDir()
	pending := &updater.PendingUpdate{
		Shell:      "/bin/sh",
		ScriptPath: scriptPath,
		Target:     filepath.Join(tmpDir, "target"),
		Staged:     filepath.Join(tmpDir, "staged"),
	}
	require.NoError(t, updater.SavePendingUpdate(dataDir, pending))
	defer func() { _ = updater.ClearPendingUpdate(dataDir) }()

	var executedCmd string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		executedCmd = name
		return exec.Command("true")
	}

	err := app.RestartApp()
	assert.NoError(t, err)
	assert.NotEmpty(t, executedCmd)
}

func TestApp_CheckUpdateAndURL(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// CheckUpdate should safely return a status
	status := app.CheckUpdate()
	assert.NotNil(t, status)

	// OpenURL should safely execute without panic
	app.OpenURL("")
	app.OpenURL("https://github.com/snowdreamtech/UniGoDesktop")
}
