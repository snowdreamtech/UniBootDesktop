// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"time"

	"github.com/snowdreamtech/unibootdesktop/cmd"
	"github.com/snowdreamtech/unibootdesktop/internal/env"
	"github.com/snowdreamtech/unibootdesktop/internal/sysinfo"
	"github.com/snowdreamtech/unibootdesktop/pkg/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
		return filepath.Join(localAppData, "UniBootDesktop", "webview2")
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "UniBootDesktop", "webview2")
	}
	return filepath.Join(env.GetDataDir(), "webview2")
}

// ensureDarwinLocalizations ensures macOS App bundle has necessary .lproj directories
// so that native system dialogs (e.g. NSOpenPanel, NSSavePanel) automatically localize to the user's OS language.
func ensureDarwinLocalizations() {
	if goRuntime.GOOS != "darwin" {
		return
	}
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	macosDir := filepath.Dir(exePath)
	if filepath.Base(macosDir) != "MacOS" {
		return
	}
	contentsDir := filepath.Dir(macosDir)
	if filepath.Base(contentsDir) != "Contents" {
		return
	}
	resourcesDir := filepath.Join(contentsDir, "Resources")
	locales := []string{"zh-Hans", "zh_CN", "zh-Hant", "zh_TW", "en", "ja", "ko", "de", "fr", "es", "ru", "pt", "it"}
	for _, loc := range locales {
		lprojDir := filepath.Join(resourcesDir, loc+".lproj")
		_ = os.MkdirAll(lprojDir, 0o755)
		stringsFile := filepath.Join(lprojDir, "InfoPlist.strings")
		if _, err := os.Stat(stringsFile); os.IsNotExist(err) {
			_ = os.WriteFile(stringsFile, []byte("/* Localized versions of Info.plist keys */\n"), 0o644)
		}
	}
}

// RunWails initializes and launches the Wails v2 desktop GUI application.
func RunWails() error {
	ensureDarwinLocalizations()
	fmt.Println(">>> Starting Wails GUI Runtime...")
	app := NewApp()

	// Determine native window appearance and background color from saved user theme preference,
	// or dynamically resolve against OS system appearance when set to "system" or empty.
	isDark := false
	winTheme := windows.SystemDefault
	if cfg, err := config.Load(); err == nil && cfg != nil {
		switch cfg.Theme {
		case "dark":
			isDark = true
			winTheme = windows.Dark
		case "light":
			isDark = false
			winTheme = windows.Light
		default:
			isDark = sysinfo.IsSystemDarkTheme()
			winTheme = windows.SystemDefault
		}
	} else {
		isDark = sysinfo.IsSystemDarkTheme()
		winTheme = windows.SystemDefault
	}

	winBackdrop := windows.Auto
	if sysinfo.SupportsMicaBackdrop() {
		winBackdrop = windows.Mica
	}

	macAppearance := mac.NSAppearanceNameAqua
	backgroundColour := &options.RGBA{R: 248, G: 250, B: 252, A: 255}
	if isDark {
		macAppearance = mac.NSAppearanceNameDarkAqua
		backgroundColour = &options.RGBA{R: 11, G: 15, B: 25, A: 255}
	}

	return wails.Run(&options.App{
		Title:       "UniBootDesktop",
		Width:       1180,
		Height:      820,
		MinWidth:    1024,
		MinHeight:   728,
		StartHidden: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: backgroundColour,
		Menu:             BuildAppMenu(app, "auto"),
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
		},
		OnStartup: app.startup,
		OnDomReady: func(ctx context.Context) {
			time.AfterFunc(50*time.Millisecond, func() {
				wailsRuntime.Show(ctx)
				wailsRuntime.WindowShow(ctx)
				if cfg, err := config.Load(); err == nil && cfg != nil {
					sysinfo.SyncTitleBarTheme(cfg.Theme)
				} else {
					sysinfo.SyncTitleBarTheme("system")
				}
			})
		},
		OnShutdown:    app.shutdown,
		OnBeforeClose: app.beforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "com.snowdreamtech.unibootdesktop",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewUserDataPath:  resolveWindowsUserDataPath(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  sysinfo.SupportsMicaBackdrop(),
			DisableWindowIcon:    false,
			Theme:                winTheme,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:           windows.RGB(11, 15, 25),
				DarkModeTitleBarInactive:   windows.RGB(15, 23, 42),
				DarkModeTitleText:          windows.RGB(241, 245, 249),
				DarkModeTitleTextInactive:  windows.RGB(148, 163, 184),
				DarkModeBorder:             windows.RGB(30, 41, 59),
				DarkModeBorderInactive:     windows.RGB(30, 41, 59),
				LightModeTitleBar:          windows.RGB(248, 250, 252),
				LightModeTitleBarInactive:  windows.RGB(241, 245, 249),
				LightModeTitleText:         windows.RGB(15, 23, 42),
				LightModeTitleTextInactive: windows.RGB(100, 116, 139),
				LightModeBorder:            windows.RGB(226, 232, 240),
				LightModeBorderInactive:    windows.RGB(226, 232, 240),
			},
			BackdropType:         winBackdrop,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           macAppearance,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "UniBootDesktop",
				Message: fmt.Sprintf("Universal Go Desktop Suite\nVersion %s", env.GitTag),
				Icon:    appIcon,
			},
		},
		Linux: &linux.Options{
			Icon:                appIcon,
			WindowIsTranslucent: false,
			ProgramName:         "unibootdesktop",
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
		},
	})
}

func init() {
	cmd.WailsRunner = RunWails
}
