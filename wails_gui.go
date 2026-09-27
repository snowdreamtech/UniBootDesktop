// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"embed"
	"fmt"

	"github.com/snowdreamtech/unigodesktop/cmd"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

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
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                 false,
				HideTitleBar:              false,
				FullSizeContent:           false,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "UniGoDesktop",
				Message: "Universal Go Desktop Suite",
			},
		},
	})
}

func init() {
	cmd.WailsRunner = RunWails
}
