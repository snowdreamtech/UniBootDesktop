// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"embed"
	"fmt"
	"os"
	"runtime"

	"github.com/snowdreamtech/unigodesktop/cmd"
	"github.com/snowdreamtech/unigodesktop/internal/i18n"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// RunWails initializes and launches the Wails v2 desktop GUI application.
func RunWails() error {
	fmt.Println(">>> Starting Wails GUI Runtime...")
	app := NewApp()

	mt := i18n.GetMenuTranslations("auto")
	appMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		appSubMenu := appMenu.AddSubmenu(mt.App)
		appSubMenu.AddText(mt.About, keys.CmdOrCtrl("i"), func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.EventsEmit(app.ctx, "open-about-modal")
			}
		})
		appSubMenu.AddSeparator()
		appSubMenu.AddText(mt.Hide, keys.CmdOrCtrl("h"), func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.WindowHide(app.ctx)
			}
		})
		appSubMenu.AddText(mt.ShowAll, nil, func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.WindowShow(app.ctx)
			}
		})
		appSubMenu.AddSeparator()
		appSubMenu.AddText(mt.Quit, keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.Quit(app.ctx)
			}
		})

		editMenu := appMenu.AddSubmenu(mt.Edit)
		editMenu.AddText(mt.Undo, keys.CmdOrCtrl("z"), nil)
		editMenu.AddText(mt.Redo, keys.CmdOrCtrl("Z"), nil)
		editMenu.AddSeparator()
		editMenu.AddText(mt.Cut, keys.CmdOrCtrl("x"), nil)
		editMenu.AddText(mt.Copy, keys.CmdOrCtrl("c"), nil)
		editMenu.AddText(mt.Paste, keys.CmdOrCtrl("v"), nil)
		editMenu.AddText(mt.SelectAll, keys.CmdOrCtrl("a"), nil)

		windowMenu := appMenu.AddSubmenu(mt.Window)
		windowMenu.AddText(mt.Minimize, keys.CmdOrCtrl("m"), func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.WindowMinimise(app.ctx)
			}
		})
		windowMenu.AddText(mt.Zoom, nil, func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.WindowToggleMaximise(app.ctx)
			}
		})

		helpMenu := appMenu.AddSubmenu(mt.Help)
		helpMenu.AddText(mt.About, nil, func(cd *menu.CallbackData) {
			if app.ctx != nil {
				wailsRuntime.EventsEmit(app.ctx, "open-about-modal")
			}
		})
	}

	return wails.Run(&options.App{
		Title:  "UniGoDesktop",
		Width:  1180,
		Height: 820,
		MinWidth: 1024,
		MinHeight: 728,
		Menu:   appMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "d6f1a8c0-87a4-4a24-9b57-unigodesktop-single-instance",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				if app.ctx != nil {
					wailsRuntime.WindowUnminimise(app.ctx)
					wailsRuntime.WindowShow(app.ctx)
				}
			},
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
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

func main() {
	if len(os.Args) <= 1 || (len(os.Args) > 1 && (os.Args[1] == "gui" || os.Args[1] == "desktop")) {
		if err := RunWails(); err != nil {
			fmt.Fprintf(os.Stderr, "Error launching Wails application: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cmd.Execute()
}
