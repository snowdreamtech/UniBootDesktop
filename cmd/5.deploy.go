// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/installer"
	"github.com/spf13/cobra"
)

var (
	deployDisk       string
	deployDisks      string
	deployAllUsb     bool
	deployMode       string
	deployFs         string
	deployIsoPaths   []string
	deployVentoyPath string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy UniBoot bootable engine to target USB drive(s)",
	Long: `Deploy UniBoot bootable firmware and ISO images to one or multiple USB drives.

Modes:
  A - Ventoy MultiBoot Hybrid Pro (Default: supports ISO/WIM/VHD multi-boot)
  B - 1-Sec Cloud Disk (macOS-native friendly iPXE cloud network boot)

Examples:
  # Deploy Mode A to a single USB drive with exFAT filesystem
  unigodesktop deploy --disk /dev/disk2 --mode A --fs exfat -y

  # Deploy Mode A with ISO image auto-copy
  unigodesktop deploy --disk /dev/disk2 -i ~/Downloads/Ubuntu.iso -y

  # Batch Deploy Mode B (1-Sec Cloud) to multiple USB drives in parallel
  unigodesktop deploy --disks /dev/disk2,/dev/disk3 --mode B -y

  # Deploy to ALL detected removable USB drives
  unigodesktop deploy --all-usb --mode A -y`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Gather target disks
		var targetDisks []string
		availableDisks, err := disk.GetRemovableDisks()
		if err != nil {
			return fmt.Errorf("failed to scan removable disks: %w", err)
		}

		if deployAllUsb {
			for _, d := range availableDisks {
				targetDisks = append(targetDisks, d.Device)
			}
			if len(targetDisks) == 0 {
				pterm.Warning.Println("No removable USB drives detected for --all-usb deployment.")
				return nil
			}
		} else if deployDisks != "" {
			parts := strings.Split(deployDisks, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					targetDisks = append(targetDisks, p)
				}
			}
		} else if deployDisk != "" {
			targetDisks = append(targetDisks, strings.TrimSpace(deployDisk))
		} else {
			return fmt.Errorf("must specify --disk, --disks, or --all-usb. Use --help for usage details")
		}

		// 2. Validate mode and filesystem
		modeUpper := strings.ToUpper(strings.TrimSpace(deployMode))
		if modeUpper != "A" && modeUpper != "B" {
			return fmt.Errorf("invalid mode '%s': must be 'A' or 'B'", deployMode)
		}

		if deployFs == "" {
			deployFs = "exFAT"
		}

		// 3. Prompt for confirmation if --yes is false
		if !yes {
			pterm.DefaultHeader.WithFullWidth().Println("⚠️  DANGER ZONE: USB FORMAT CONFIRMATION")
			pterm.Warning.Println("Target USB drive(s) WILL BE FORMATTED. All existing data will be permanently erased!")
			pterm.Println(fmt.Sprintf("Target Disk(s) : %s", strings.Join(targetDisks, ", ")))
			pterm.Println(fmt.Sprintf("Deployment Mode: Mode %s (%s)", modeUpper, map[string]string{"A": "Ventoy MultiBoot", "B": "1-Sec Cloud iPXE"}[modeUpper]))
			pterm.Println(fmt.Sprintf("Filesystem     : %s", deployFs))
			if len(deployIsoPaths) > 0 {
				pterm.Println(fmt.Sprintf("ISO Images     : %s", strings.Join(deployIsoPaths, ", ")))
			}
			pterm.Println()

			confirm, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Do you want to proceed with formatting and deployment?").Show()
			if !confirm {
				pterm.Info.Println("Deployment cancelled by user.")
				return nil
			}
		}

		pterm.Success.Println(fmt.Sprintf("🚀 Starting UniBoot deployment to %d target drive(s)...", len(targetDisks)))
		start := time.Now()

		// 4. Multi-device Parallel Execution
		ctx := context.Background()
		var wg sync.WaitGroup
		results := make([]*installer.DeployResult, len(targetDisks))
		errs := make([]error, len(targetDisks))

		pb, _ := pterm.DefaultProgressbar.WithTotal(len(targetDisks)).WithTitle("Deploying USB Drives").Start()

		for i, dev := range targetDisks {
			wg.Add(1)
			go func(idx int, target string) {
				defer wg.Done()

				progressCallback := func(p installer.IsoCopyProgress) {
					pb.UpdateTitle(fmt.Sprintf("[%s] Copying %s (%.1f%%)", target, filepath.Base(p.CurrentFile), p.Progress))
				}

				var res *installer.DeployResult
				var err error

				if modeUpper == "A" {
					res, err = installer.DeployModeAWithIsoAndVentoyPath(ctx, target, deployFs, deployVentoyPath, deployIsoPaths, progressCallback)
				} else {
					res, err = installer.DeployModeB(ctx, target, deployFs)
				}

				results[idx] = res
				errs[idx] = err
				pb.Increment()
			}(i, dev)
		}

		wg.Wait()
		pb.Stop()

		// 5. Output Summary Table
		tableData := pterm.TableData{
			{"Device", "Mode", "Status", "Details"},
		}

		successCount := 0
		for i, dev := range targetDisks {
			err := errs[i]
			res := results[i]

			if err == nil && res != nil && res.Success {
				successCount++
				tableData = append(tableData, []string{dev, fmt.Sprintf("Mode %s", modeUpper), "🟢 SUCCESS", res.Message})
			} else {
				errMsg := "failed"
				if err != nil {
					errMsg = err.Error()
				} else if res != nil {
					errMsg = res.Message
				}
				tableData = append(tableData, []string{dev, fmt.Sprintf("Mode %s", modeUpper), "❌ FAILED", errMsg})
			}
		}

		fmt.Println()
		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		fmt.Println()

		pterm.Success.Println(fmt.Sprintf("Deployment completed: %d/%d succeeded in %.2fs", successCount, len(targetDisks), time.Since(start).Seconds()))
		return nil
	},
}

func init() {
	deployCmd.Flags().StringVarP(&deployDisk, "disk", "d", "", "target USB disk device (e.g., /dev/disk2 or PhysicalDrive1)")
	deployCmd.Flags().StringVar(&deployDisks, "disks", "", "comma-separated target USB disks for batch parallel deployment")
	deployCmd.Flags().BoolVar(&deployAllUsb, "all-usb", false, "automatically deploy to ALL detected removable USB drives")
	deployCmd.Flags().StringVarP(&deployMode, "mode", "m", "A", "deployment mode: 'A' (Ventoy MultiBoot) or 'B' (1-Sec Cloud iPXE)")
	deployCmd.Flags().StringVarP(&deployFs, "fs", "f", "exFAT", "partition filesystem type: exFAT, FAT32, NTFS, or ext4")
	deployCmd.Flags().StringSliceVarP(&deployIsoPaths, "iso", "i", nil, "source ISO/image file paths to automatically copy")
	deployCmd.Flags().StringVar(&deployVentoyPath, "ventoy-path", "", "custom path to Ventoy CLI binary")

	rootCmd.AddCommand(deployCmd)
}
