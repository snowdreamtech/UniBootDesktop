// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unibootdesktop/cmd"
	"github.com/snowdreamtech/unibootdesktop/pkg/privilege"
)

func main() {
	for i, arg := range os.Args {
		if arg == "--privileged-worker" {
			if err := privilege.RunWorkerFromArgs(os.Args[i+1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Privileged worker error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}
	// When built with GUI support (default, no build tags), WailsRunner is
	// set by wails_gui.go's init(). In CLI-only builds (-tags nogui),
	// WailsRunner remains nil and we fall through to the CLI.
	if len(os.Args) <= 1 || (len(os.Args) > 1 && (os.Args[1] == "gui" || os.Args[1] == "desktop")) {
		if cmd.WailsRunner != nil {
			if err := cmd.WailsRunner(); err != nil {
				fmt.Fprintf(os.Stderr, "Error launching Wails application: %v\n", err)
				os.Exit(1)
			}
			return
		}
		// CLI-only build: fall through to cobra which handles "gui" subcommand
	}

	cmd.Execute()
}
