// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"

	"github.com/snowdreamtech/unibootdesktop/internal/desktop"
	"github.com/snowdreamtech/unibootdesktop/pkg/config"
	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:     "gui",
	Aliases: []string{"desktop"},
	Short:   "Launch the UniBootDesktop graphical user interface",
	Long:    `Launch the UniBootDesktop interactive Wails / Webview graphical desktop interface and system tray.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if WailsRunner != nil {
			return WailsRunner()
		}
		cfg, err := config.Load()
		if err != nil {
			cfg = config.GetDefaultConfig()
		}
		app := desktop.NewApp(cfg)
		return app.Start(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
