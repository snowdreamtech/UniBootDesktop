// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unibootdesktop/pkg/hypervisor"
	"github.com/spf13/cobra"
)

var (
	vmDisk     string
	vmEngine   string
	vmMem      string
	vmBootMode string
)

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Launch virtual machine to test target bootable disk drive",
	Long: `Launch an isolated virtual machine window (QEMU, UTM, VMware, VirtualBox, KVM, Hyper-V, Parallels)
to test the bootability of a disk drive or ISO image without rebooting your computer.

Examples:
  # Test disk drive /dev/disk2 in best available virtual machine
  unibootdesktop vm --disk /dev/disk2

  # Test disk drive specifically in QEMU with 4GB RAM in UEFI mode
  unibootdesktop vm -d /dev/disk2 -e qemu -m 4096 -b uefi`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target := strings.TrimSpace(vmDisk)
		if target == "" {
			return fmt.Errorf("must specify --disk or -d target drive path. Use --help for usage details")
		}

		pterm.DefaultHeader.WithFullWidth().Println("🖥️  VIRTUAL MACHINE BOOT SIMULATOR")

		mgr := hypervisor.GetManager()
		bootMode := strings.ToLower(strings.TrimSpace(vmBootMode))
		if bootMode == "" {
			bootMode = hypervisor.BootModeAuto
		}

		memMB := hypervisor.GetRecommendedVMMemoryMB()
		if trimmedMem := strings.TrimSpace(vmMem); trimmedMem != "" {
			if parsed, err := strconv.Atoi(trimmedMem); err == nil && parsed > 0 {
				memMB = parsed
			}
		}

		vmCfg := hypervisor.VMConfig{
			CpuCores:     hypervisor.GetRecommendedVCPUs(),
			MemoryMB:     memMB,
			BootMode:     bootMode,
			DisplayAccel: true,
			SecureBoot:   false,
		}

		ctx := context.Background()
		engine := strings.ToLower(strings.TrimSpace(vmEngine))

		if engine != "" && engine != "auto" {
			hType := hypervisor.HypervisorType(engine)
			pterm.Info.Println(fmt.Sprintf("Requesting specified hypervisor engine: %s (RAM: %dMB, Boot: %s)...", engine, memMB, bootMode))
			if err := mgr.LaunchSpecifiedConfigured(ctx, target, hType, vmCfg); err != nil {
				return fmt.Errorf("failed to launch specified hypervisor %s: %w", engine, err)
			}
		} else {
			status := mgr.DetectBest()
			if status == nil || !status.Installed {
				pterm.Error.Println("No supported virtual machine (QEMU, UTM, VMware, VirtualBox, KVM, Hyper-V, Parallels) is installed on your system.")
				pterm.Info.Println("Please install QEMU using Homebrew (macOS: 'brew install qemu') or your system package manager.")
				return fmt.Errorf("no supported hypervisor found on system PATH")
			}

			pterm.Success.Println(fmt.Sprintf("Selected Best Hypervisor: %s (%s, %s)", status.Name, status.Version, status.Path))
			pterm.Info.Println(fmt.Sprintf("Launching VM test window for target: %s (RAM: %dMB, Boot: %s)...", target, memMB, bootMode))

			if err := mgr.LaunchBestConfigured(ctx, target, vmCfg); err != nil {
				return fmt.Errorf("failed to launch VM simulator: %w", err)
			}
		}

		pterm.Success.Println("🚀 Virtual machine simulator launched successfully!")
		return nil
	},
}

func init() {
	vmCmd.Flags().StringVarP(&vmDisk, "disk", "d", "", "target disk device or ISO image path")
	vmCmd.Flags().StringVarP(&vmEngine, "engine", "e", "auto", "hypervisor engine (auto, qemu, utm, vmware, virtualbox, kvm, hyperv, parallels)")
	vmCmd.Flags().StringVarP(&vmMem, "mem", "m", "2048", "allocated memory in MB for virtual machine")
	vmCmd.Flags().StringVarP(&vmBootMode, "boot", "b", "auto", "boot mode: auto, uefi, bios")

	rootCmd.AddCommand(vmCmd)
}
