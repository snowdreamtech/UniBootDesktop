// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"

	"github.com/snowdreamtech/unigodesktop/internal/config"
	"github.com/snowdreamtech/unigodesktop/internal/desktop"
	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the UniGoDesktop Wails graphical user interface",
	Long:  `Launch the UniGoDesktop interactive Wails / Webview graphical desktop interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
