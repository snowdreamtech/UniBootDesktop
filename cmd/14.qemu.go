// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigodesktop/pkg/hypervisor"
	"github.com/spf13/cobra"
)

var (
	qemuDisk string
	qemuMem  string
)

var qemuCmd = &cobra.Command{
	Use:   "qemu",
	Short: "Launch virtual machine to test target bootable disk drive",
	Long: `Launch an isolated virtual machine window (QEMU or other available hypervisors) to test the bootability of a disk drive or ISO image without rebooting your computer.

Examples:
  # Test disk drive /dev/disk2 in virtual machine
  unigodesktop qemu --disk /dev/disk2

  # Test disk drive with 4GB RAM
  unigodesktop qemu -d /dev/disk2 -m 4096`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target := strings.TrimSpace(qemuDisk)
		if target == "" {
			return fmt.Errorf("must specify --disk or -d target drive path. Use --help for usage details")
		}

		pterm.DefaultHeader.WithFullWidth().Println("🖥️  VIRTUAL MACHINE BOOT SIMULATOR")

		// 1. Detect hypervisor status
		mgr := hypervisor.GetManager()
		status := mgr.DetectBest()
		if status == nil || !status.Installed {
			pterm.Error.Println("No supported virtual machine (QEMU, UTM, VMware, VirtualBox) is installed on your system.")
			pterm.Info.Println("Please install QEMU using Homebrew (macOS: 'brew install qemu') or your system package manager.")
			return fmt.Errorf("no supported hypervisor found on system PATH")
		}

		pterm.Success.Println(fmt.Sprintf("%s Detected: %s (%s)", status.Name, status.Version, status.Path))
		pterm.Info.Println(fmt.Sprintf("Launching VM test window for target: %s ...", target))

		// 2. Launch hypervisor VM
		ctx := context.Background()
		err := mgr.LaunchBest(ctx, target, hypervisor.BootModeAuto)
		if err != nil {
			return fmt.Errorf("failed to launch VM simulator: %w", err)
		}

		pterm.Success.Println("🚀 Virtual machine simulator launched successfully!")
		return nil
	},
}

func init() {
	qemuCmd.Flags().StringVarP(&qemuDisk, "disk", "d", "", "target disk device or ISO image path")
	qemuCmd.Flags().StringVarP(&qemuMem, "mem", "m", "2048", "allocated memory in MB for virtual machine")

	rootCmd.AddCommand(qemuCmd)
}
