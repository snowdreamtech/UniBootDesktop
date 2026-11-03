// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	internalUpdater "github.com/snowdreamtech/unigodesktop/internal/updater"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
)

func mockNoopCommand() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd.exe", "/c", "exit 0")
	}
	return exec.Command("true")
}

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

	// ClearProxyPassword executes safely
	_ = app.ClearProxyPassword()
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
		// Use cross-platform command that exits immediately
		return mockNoopCommand()
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
	scriptName := "apply_update.sh"
	scriptContent := "#!/bin/sh\nexit 0\n"
	shellCmd := "/bin/sh"
	if runtime.GOOS == "windows" {
		scriptName = "apply_update.bat"
		scriptContent = "@echo off\r\nexit 0\r\n"
		shellCmd = "cmd.exe"
	}
	scriptPath := filepath.Join(tmpDir, scriptName)
	require.NoError(t, os.WriteFile(scriptPath, []byte(scriptContent), 0o755))

	dataDir := env.GetDataDir()
	pending := &updater.PendingUpdate{
		Shell:      shellCmd,
		ScriptPath: scriptPath,
		Target:     filepath.Join(tmpDir, "target"),
		Staged:     filepath.Join(tmpDir, "staged"),
	}
	require.NoError(t, updater.SavePendingUpdate(dataDir, pending))
	defer func() { _ = updater.ClearPendingUpdate(dataDir) }()

	var executedCmd string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		executedCmd = name
		return mockNoopCommand()
	}

	err := app.RestartApp()
	assert.NoError(t, err)
	assert.NotEmpty(t, executedCmd)
}

type testMockTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (m *testMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = m.target.Scheme
	req.URL.Host = m.target.Host
	return m.base.RoundTrip(req)
}

func TestApp_CheckUpdateAndURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(internalUpdater.ReleaseInfo{
			TagName: "v99.0.0",
		})
	}))
	defer ts.Close()

	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()

	targetURL, _ := url.Parse(ts.URL)
	http.DefaultTransport = &testMockTransport{
		target: targetURL,
		base:   origTransport,
	}

	app := NewApp()
	require.NotNil(t, app)

	// CheckUpdate should safely return a status without hitting external network
	status := app.CheckUpdate()
	assert.NotNil(t, status)

	// OpenURL should safely execute without panic
	app.OpenURL("")
	app.OpenURL("https://github.com/snowdreamtech/UniGoDesktop")
}

func TestApp_PerformGuiUpdate_Concurrency(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// Lock mutex to simulate an ongoing update
	app.mu.Lock()
	res, err := app.PerformGuiUpdate()
	app.mu.Unlock()

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "already in progress")
}

func TestApp_ContextLifecycle(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	ctx := context.Background()
	app.startup(ctx)
	require.NotNil(t, app.ctx)
	require.NotNil(t, app.cancel)

	// Context should not be canceled yet
	assert.NoError(t, app.ctx.Err())

	// Shutdown should trigger cancel
	app.shutdown(ctx)
	assert.Error(t, app.ctx.Err())
	assert.Equal(t, context.Canceled, app.ctx.Err())
}

func TestApp_HardwareAndHypervisorAPIs(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	ctx := context.Background()
	app.startup(ctx)
	defer app.shutdown(ctx)

	// GetDiskList returns slice without panic
	disks, err := app.GetDiskList()
	assert.NoError(t, err)
	assert.NotNil(t, disks)

	// EjectDisk validation errors
	assert.Error(t, app.EjectDisk(""))
	assert.Error(t, app.EjectDisk(string(make([]byte, 513))))

	// Hypervisor detection
	hvs := app.DetectHypervisors()
	assert.NotNil(t, hvs)

	best := app.DetectBestHypervisor()
	assert.NotNil(t, best)

	cfg := app.GetDefaultVMConfig()
	assert.NotNil(t, cfg)
	assert.True(t, cfg.CpuCores > 0)
	assert.True(t, cfg.MemoryMB > 0)

	// VM validation checks
	assert.Error(t, app.LaunchVM("", "auto", "auto"))
	assert.Error(t, app.LaunchVM("/dev/invalid_test_disk", "invalid_hypervisor", "auto"))
	assert.Error(t, app.LaunchVM("/dev/invalid_test_disk", "qemu", "invalid_boot_mode"))

	// StopVM executes cleanly
	err = app.StopVM()
	assert.NoError(t, err)

	// SetTheme executes safely
	app.SetTheme("dark")
	app.SetTheme("light")
	app.SetTheme("system")

	// IsPrivileged returns boolean
	_ = app.IsPrivileged()

	// CheckConfigHealth returns diagnostic report
	report, err := app.CheckConfigHealth()
	assert.NoError(t, err)
	assert.NotNil(t, report)

	// CalculateFileChecksum
	tmpFile := filepath.Join(t.TempDir(), "test_hash.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("Hello UniGoDesktop"), 0o644))
	res, err := app.CalculateFileChecksum(tmpFile, "sha256")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Hash)
}
