// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/internal/utils"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/spf13/cobra"
)

var (
	dfHumanReadable bool
	dfUsb           bool
)

var dfCmd = &cobra.Command{
	Use:   "df",
	Short: "Display USB drives and UniBoot directory disk usage",
	Long:  `Display information about removable USB storage devices and data directory disk usage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. If --usb flag or default run, display USB drive inspector table
		if dfUsb || !cmd.Flags().Changed("human-readable") {
			disks, err := disk.GetRemovableDisks()
			if err == nil && len(disks) > 0 {
				if jsonOutput {
					b, _ := json.MarshalIndent(disks, "", "  ")
					fmt.Println(string(b))
					return nil
				}

				pterm.DefaultHeader.WithFullWidth().Println("🔌 REMOVABLE USB DRIVES & HARDWARE SPECS")
				tableData := pterm.TableData{
					{"Device", "Name / Model", "Capacity", "FileSystem", "Partition", "USB Protocol & Speed"},
				}

				for _, d := range disks {
					sizeStr := d.Formatted
					if sizeStr == "" {
						sizeStr = utils.FormatBytes(int64(d.Size))
					}
					fsStr := d.FileSystem
					if fsStr == "" {
						fsStr = "Unknown"
					}
					schemeStr := d.PartitionScheme
					if schemeStr == "" {
						schemeStr = "MBR/GPT"
					}
					speedStr := fmt.Sprintf("%s (%s)", d.UsbVersion, d.UsbSpeed)
					if d.UsbVersion == "" {
						speedStr = "USB 2.0 / 3.0"
					}
					tableData = append(tableData, []string{
						d.Device,
						d.Name,
						sizeStr,
						fsStr,
						schemeStr,
						speedStr,
					})
				}

				_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
				fmt.Println()
			}
		}

		// 2. Display UniBoot App Data Directory Storage Usage
		dataDir := env.GetDataDir()
		entries, err := os.ReadDir(dataDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Data directory does not exist yet.")
				return nil
			}
			return fmt.Errorf("failed to read data directory: %w", err)
		}

		dirTableData := pterm.TableData{
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
			if dfHumanReadable || true {
				sizeStr = utils.FormatBytes(size)
			}

			dirTableData = append(dirTableData, []string{entry.Name(), sizeStr})
		}

		totalStr := utils.FormatBytes(totalSize)
		dirTableData = append(dirTableData, []string{"TOTAL", totalStr})

		pterm.DefaultSection.Println("📦 UniBoot App Data Directory Storage Usage")
		fmt.Printf("Path: %s\n\n", dataDir)
		_ = pterm.DefaultTable.WithHasHeader().WithData(dirTableData).Render()

		return nil
	},
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(dfCmd)
	}
	dfCmd.Flags().BoolVarP(&dfHumanReadable, "human-readable", "h", true, "print sizes in human readable format (e.g., 1023M, 14.8G)")
	dfCmd.Flags().BoolVarP(&dfUsb, "usb", "u", false, "display removable USB drives and hardware specifications")
}
