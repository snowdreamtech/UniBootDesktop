// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unigodesktop/cmd"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

// resolveWindowsUserDataPath returns the user data path for WebView2 on Windows.
// It detects portable mode by checking for a portable marker file (portable.dat or .portable)
// next to the executable, isolating user data in the local "data" directory.
func resolveWindowsUserDataPath() string {
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if _, err := os.Stat(filepath.Join(exeDir, "portable.dat")); err == nil {
			return filepath.Join(exeDir, "data", "webview2")
		}
		if _, err := os.Stat(filepath.Join(exeDir, ".portable")); err == nil {
			return filepath.Join(exeDir, "data", "webview2")
		}
	}

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return filepath.Join(localAppData, "UniGoDesktop", "webview2")
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "UniGoDesktop", "webview2")
	}
	return ""
}

// RunWails initializes and launches the Wails v2 desktop GUI application.
func RunWails() error {
	fmt.Println(">>> Starting Wails GUI Runtime...")
	app := NewApp()

	return wails.Run(&options.App{
		Title:     "UniGoDesktop",
		Width:     1180,
		Height:    820,
		MinWidth:  1024,
		MinHeight: 728,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewUserDataPath:  resolveWindowsUserDataPath(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			Theme:                windows.SystemDefault,
			BackdropType:         windows.Auto,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
			},
			Appearance:           mac.DefaultAppearance,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "UniGoDesktop",
				Message: "Universal Go Desktop Suite",
			},
		},
		Linux: &linux.Options{
			Icon:                appIcon,
			WindowIsTranslucent: false,
			ProgramName:         "unigodesktop",
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
		},
	})
}

func init() {
	cmd.WailsRunner = RunWails
}
