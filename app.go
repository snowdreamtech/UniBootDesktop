// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/config"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// SystemInfo holds runtime and OS metadata.
type SystemInfo struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"goVersion"`
	AppName   string `json:"appName"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	DataDir   string `json:"dataDir"`
	ConfigDir string `json:"configDir"`
}

// NetworkTestResult holds connectivity test metrics.
type NetworkTestResult struct {
	Connected bool   `json:"connected"`
	LatencyMs int64  `json:"latencyMs"`
	TargetURL string `json:"targetUrl"`
	Error     string `json:"error,omitempty"`
}

// HelloInfo provides greeting and runtime demonstration.
type HelloInfo struct {
	Greeting  string `json:"greeting"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Timestamp string `json:"timestamp"`
}

// App struct manages Wails GUI lifecycle and frontend bound APIs.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails application starts up.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Info("UniGoDesktop Wails GUI runtime started successfully")
}

// Greet returns a friendly greeting for demonstration.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello %s, Welcome to UniGoDesktop!", name)
}

// GetHelloInfo returns structured hello greeting and runtime environment details.
func (a *App) GetHelloInfo() *HelloInfo {
	return &HelloInfo{
		Greeting:  "Hello World From UniGoDesktop!",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// GetSystemInfo returns system environment and runtime diagnostic metadata.
func (a *App) GetSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		AppName:   "UniGoDesktop",
		Version:   env.GitTag,
		Commit:    env.CommitHash,
		BuildTime: env.BuildTime,
		DataDir:   env.GetDataDir(),
		ConfigDir: env.GetConfigDir(),
	}
}

// TestNetwork checks network connectivity and latency against target endpoint.
func (a *App) TestNetwork(targetURL string) *NetworkTestResult {
	if targetURL == "" {
		targetURL = "https://api.github.com"
	}
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &NetworkTestResult{
			Connected: false,
			LatencyMs: latency,
			TargetURL: targetURL,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	return &NetworkTestResult{
		Connected: resp.StatusCode >= 200 && resp.StatusCode < 400,
		LatencyMs: latency,
		TargetURL: targetURL,
	}
}

// CheckUpdate returns GitHub release update metadata.
func (a *App) CheckUpdate() *updater.UpdateStatus {
	return updater.CheckUpdate(a.ctx)
}

// GetConfig loads the application settings.
func (a *App) GetConfig() (*config.AppConfig, error) {
	return config.Load()
}

// SaveConfig updates and saves application settings.
func (a *App) SaveConfig(cfg *config.AppConfig) error {
	if cfg == nil {
		return config.GetDefaultConfig().Save()
	}
	return cfg.Save()
}

// OpenURL opens the specified URL in the native desktop browser.
func (a *App) OpenURL(url string) {
	if url != "" {
		wailsRuntime.BrowserOpenURL(a.ctx, url)
	}
}
