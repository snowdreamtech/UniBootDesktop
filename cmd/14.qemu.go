// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigodesktop/pkg/qemu"
	"github.com/spf13/cobra"
)

var (
	qemuDisk string
	qemuMem  string
)

var qemuCmd = &cobra.Command{
	Use:   "qemu",
	Short: "Launch QEMU virtual machine to test target USB bootable drive",
	Long: `Launch an isolated QEMU simulator virtual machine window to test the bootability of a USB drive or ISO image without rebooting your computer.

Examples:
  # Test USB drive /dev/disk2 in QEMU
  unigodesktop qemu --disk /dev/disk2

  # Test USB drive with 4GB RAM
  unigodesktop qemu -d /dev/disk2 -m 4096`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target := strings.TrimSpace(qemuDisk)
		if target == "" {
			return fmt.Errorf("must specify --disk or -d target drive path. Use --help for usage details")
		}

		pterm.DefaultHeader.WithFullWidth().Println("🖥️  QEMU VIRTUAL MACHINE BOOT SIMULATOR")

		// 1. Detect QEMU status
		status := qemu.Detect()
		if !status.Installed {
			pterm.Error.Println("QEMU is not installed on your system.")
			pterm.Info.Println("Please install QEMU using Homebrew (macOS: 'brew install qemu') or your Linux package manager (e.g., 'apt install qemu-system-x86').")
			return fmt.Errorf("qemu binary not found on system PATH")
		}

		pterm.Success.Println(fmt.Sprintf("QEMU Detected: %s (%s)", status.Version, status.Path))
		pterm.Info.Println(fmt.Sprintf("Launching VM test window for target: %s ...", target))

		// 2. Launch QEMU VM
		ctx := context.Background()
		err := qemu.LaunchTest(ctx, target)
		if err != nil {
			return fmt.Errorf("failed to launch QEMU simulator: %w", err)
		}

		pterm.Success.Println("🚀 QEMU simulator launched successfully!")
		return nil
	},
}

func init() {
	qemuCmd.Flags().StringVarP(&qemuDisk, "disk", "d", "", "target USB drive device or ISO image path")
	qemuCmd.Flags().StringVarP(&qemuMem, "mem", "m", "2048", "allocated memory in MB for virtual machine")

	rootCmd.AddCommand(qemuCmd)
}
