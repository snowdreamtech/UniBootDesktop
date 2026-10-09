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
	"strings"
	"testing"

	"github.com/snowdreamtech/unibootdesktop/internal/env"
	internalUpdater "github.com/snowdreamtech/unibootdesktop/internal/updater"
	"github.com/snowdreamtech/unibootdesktop/pkg/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_LifecycleAndAPIs(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	ctx := context.Background()
	app.startup(ctx)

	// Greet
	assert.Equal(t, "Hello World, Welcome to UniBootDesktop!", app.Greet(""))
	assert.Equal(t, "Hello Alice, Welcome to UniBootDesktop!", app.Greet("Alice"))

	// HelloInfo
	info := app.GetHelloInfo()
	assert.NotNil(t, info)
	assert.Contains(t, info.Greeting, "Hello World")

	// SystemInfo
	sys := app.GetSystemInfo()
	assert.NotNil(t, sys)
	assert.Equal(t, "UniBootDesktop", sys.AppName)

	// Config
	cfg, err := app.GetConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	err = app.SaveConfig(cfg)
	assert.NoError(t, err)

	// Theme synchronization
	app.SetTheme("dark")
	app.SetTheme("light")
	app.SetTheme("system")
}

func TestResolveWindowsUserDataPath(t *testing.T) {
	// Verify that resolving the Windows user data path executes safely
	path := resolveWindowsUserDataPath()
	_ = path
}

func mockNoopCommand() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd.exe", "/c", "exit 0")
	}
	return exec.Command("true")
}

func TestApp_RestartApp(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// Mock execCommand to a command that succeeds without launching another app instance
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
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
	app.OpenURL("https://github.com/snowdreamtech/UniBootDesktop")
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

func TestApp_UniBootOperations(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// Status
	status := app.GetUniBootStatus()
	assert.NotNil(t, status)
	assert.True(t, status.Ready)

	// Firmware mappings
	mappings := app.GetFirmwareList()
	assert.NotEmpty(t, mappings)

	// Log buffer
	app.LogAction("INFO", "test message", "test details")
	logs := app.GetRecentLogs()
	assert.NotEmpty(t, logs)
	app.ClearLogs()
	assert.Empty(t, app.GetRecentLogs())

	// Hypervisor checks
	qemuStatus := app.CheckQEMU()
	assert.NotNil(t, qemuStatus)
	assert.NotNil(t, app.DetectHypervisors())

	// Deployment cancellation when no deployment is active
	assert.False(t, app.CancelDeployment())

	// Validate Ventoy CLI with non-existent, empty, and excessively long paths
	res := app.ValidateVentoyCli("non-existent-ventoy-path")
	assert.NotNil(t, res)

	resEmpty := app.ValidateVentoyCli("")
	assert.NotNil(t, resEmpty)

	resLong := app.ValidateVentoyCli(strings.Repeat("a", 5000))
	assert.NotNil(t, resLong)
	assert.False(t, resLong.Valid)
	assert.Equal(t, "path_too_long", resLong.Code)

	// App info
	appInfo := app.GetAppInfo()
	assert.Equal(t, env.ProjectName, appInfo.ProjectName)
}

func TestApp_SaveConfigAndThemeLanguageValidation(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	cfg, err := app.GetConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	allLocales := []string{
		"auto", "zh-CN", "en-US", "zh-TW", "ja-JP", "ko-KR", "de-DE", "fr-FR",
		"es-ES", "es-LA", "ru-RU", "pt-BR", "pt-PT", "it-IT", "tr-TR", "pl-PL",
		"vi-VN", "ar-SA", "ur-PK", "az-AZ", "da-DK", "ka-GE", "fa-IR", "sl-SI",
		"oc-FR", "cs-CZ", "sk-SK", "bn-BD", "hi-IN", "nl-NL", "ro-RO", "hr-HR",
		"hu-HU", "sr-Latn", "sr-Cyrl", "th-TH", "lt-LT", "mk-MK", "he-IL", "id-ID",
		"nb-NO", "no-NO", "uk-UA", "el-GR", "sv-SE", "bg-BG", "hy-AM", "fi-FI",
		"gl-ES", "ca-ES", "ta-IN", "be-BY", "ml-IN", "et-EE",
	}

	for _, loc := range allLocales {
		cfg.Language = loc
		cfg.Theme = "light"
		err := app.SaveConfig(cfg)
		assert.NoError(t, err, "locale %s must be accepted by SaveConfig", loc)
	}

	for _, th := range []string{"light", "dark", "system"} {
		cfg.Theme = th
		cfg.Language = "zh-CN"
		err := app.SaveConfig(cfg)
		assert.NoError(t, err, "theme %s must be accepted by SaveConfig", th)
	}

	// Invalid language should be rejected
	cfg.Language = "invalid-lang-code"
	err = app.SaveConfig(cfg)
	assert.Error(t, err)

	// Invalid theme should normalize to system
	cfg.Language = "zh-CN"
	cfg.Theme = "invalid-theme"
	err = app.SaveConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "system", cfg.Theme)
}
