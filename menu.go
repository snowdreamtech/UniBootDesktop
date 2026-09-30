// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"runtime"

	"github.com/snowdreamtech/unibootdesktop/internal/i18n"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// BuildAppMenu constructs localized application menu based on given language code.
func BuildAppMenu(app *App, lang string) *menu.Menu {
	mt := i18n.GetMenuTranslations(lang)
	appMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		appSubMenu := appMenu.AddSubmenu(mt.App)
		appSubMenu.AddText(mt.About, keys.CmdOrCtrl("i"), func(cd *menu.CallbackData) {
			if app != nil && app.ctx != nil {
				wailsRuntime.EventsEmit(app.ctx, "open-about-modal")
			}
		})
		appSubMenu.AddSeparator()
		appSubMenu.AddText(mt.Hide, keys.CmdOrCtrl("h"), func(cd *menu.CallbackData) {
			if app != nil && app.ctx != nil {
				wailsRuntime.WindowHide(app.ctx)
			}
		})
		appSubMenu.AddText(mt.ShowAll, nil, func(cd *menu.CallbackData) {
			if app != nil && app.ctx != nil {
				wailsRuntime.WindowShow(app.ctx)
			}
		})
		appSubMenu.AddSeparator()
		appSubMenu.AddText(mt.Quit, keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
			if app != nil && app.ctx != nil {
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
			if app != nil && app.ctx != nil {
				wailsRuntime.WindowMinimise(app.ctx)
			}
		})
		windowMenu.AddText(mt.Zoom, nil, func(cd *menu.CallbackData) {
			if app != nil && app.ctx != nil {
				wailsRuntime.WindowToggleMaximise(app.ctx)
			}
		})

		helpMenu := appMenu.AddSubmenu(mt.Help)
		helpMenu.AddText(mt.About, nil, func(cd *menu.CallbackData) {
			if app != nil && app.ctx != nil {
				wailsRuntime.EventsEmit(app.ctx, "open-about-modal")
			}
		})
	}
	return appMenu
}
