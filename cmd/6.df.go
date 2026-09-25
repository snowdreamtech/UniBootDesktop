// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/internal/utils"
	"github.com/spf13/cobra"
)

var dfHumanReadable bool

var dfCmd = &cobra.Command{
	Use:   "df",
	Short: "Display the disk usage of unigodesktop data directories",
	Long:  `Display the disk usage of various folders within the unigodesktop data directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := env.GetDataDir()

		entries, err := os.ReadDir(dataDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Data directory does not exist yet.")
				return nil
			}
			return fmt.Errorf("failed to read data directory: %w", err)
		}

		tableData := pterm.TableData{
			{"Directory", "Size"},
		}

		var totalSize int64

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirPath := filepath.Join(dataDir, entry.Name())
			size, err := utils.CalculateDirectorySize(dirPath)
			if err != nil {
				continue
			}

			totalSize += size

			sizeStr := fmt.Sprintf("%d", size)
			if dfHumanReadable {
				sizeStr = utils.FormatBytes(size)
			}

			tableData = append(tableData, []string{entry.Name(), sizeStr})
		}

		totalStr := fmt.Sprintf("%d", totalSize)
		if dfHumanReadable {
			totalStr = utils.FormatBytes(totalSize)
		}
		tableData = append(tableData, []string{"TOTAL", totalStr})

		fmt.Printf("Data Directory: %s\n\n", dataDir)
		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

		return nil
	},
}

func init() {
	dfCmd.Flags().BoolVarP(&dfHumanReadable, "human-readable", "H", true, "print sizes in human readable format (e.g., 1K 234M 2G)")
	rootCmd.AddCommand(dfCmd)
}
