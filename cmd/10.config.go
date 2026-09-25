// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(configCmd)
		configCmd.AddCommand(configGetCmd)
		configCmd.AddCommand(configSetCmd)
		configCmd.AddCommand(configListCmd)
	}
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Query or modify the UniGoDesktop configuration settings stored in unigodesktop.toml.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return configListCmd.RunE(cmd, args)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		tableData := pterm.TableData{
			{"Key", "Value"},
			{"debug", fmt.Sprintf("%t", cfg.Debug)},
			{"autoCheckUpdate", fmt.Sprintf("%t", cfg.AutoCheckUpdate)},
			{"theme", cfg.Theme},
			{"language", cfg.Language},
			{"githubProxy", cfg.GithubProxy},
			{"proxyProtocol", cfg.ProxyProtocol},
			{"proxyHost", cfg.ProxyHost},
			{"proxyPort", strconv.Itoa(cfg.ProxyPort)},
			{"proxyUser", cfg.ProxyUser},
		}

		pterm.DefaultSection.Println("⚙️ UniGoDesktop Configuration")
		return pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		logger.Debug("Getting config value", "key", key)

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		switch key {
		case "debug":
			fmt.Println(cfg.Debug)
		case "autocheckupdate":
			fmt.Println(cfg.AutoCheckUpdate)
		case "theme":
			fmt.Println(cfg.Theme)
		case "language":
			fmt.Println(cfg.Language)
		case "githubproxy":
			fmt.Println(cfg.GithubProxy)
		case "proxyprotocol":
			fmt.Println(cfg.ProxyProtocol)
		case "proxyhost":
			fmt.Println(cfg.ProxyHost)
		case "proxyport":
			fmt.Println(cfg.ProxyPort)
		case "proxyuser":
			fmt.Println(cfg.ProxyUser)
		case "proxypassword":
			fmt.Println(cfg.ProxyPassword)
		default:
			pterm.Warning.Printf("Key '%s' not recognized in configuration\n", args[0])
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		val := args[1]
		logger.Debug("Setting config value", "key", key, "value", val)

		cfg, err := config.Load()
		if err != nil {
			cfg = config.GetDefaultConfig()
		}

		switch key {
		case "debug":
			parsed, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("invalid boolean value for debug: %w", err)
			}
			cfg.Debug = parsed
		case "autocheckupdate":
			parsed, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("invalid boolean value for autoCheckUpdate: %w", err)
			}
			cfg.AutoCheckUpdate = parsed
		case "theme":
			cfg.Theme = val
		case "language":
			cfg.Language = val
		case "githubproxy":
			cfg.GithubProxy = val
		case "proxyprotocol":
			cfg.ProxyProtocol = val
		case "proxyhost":
			cfg.ProxyHost = val
		case "proxyport":
			port, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("invalid port integer: %w", err)
			}
			cfg.ProxyPort = port
		case "proxyuser":
			cfg.ProxyUser = val
		case "proxypassword":
			cfg.ProxyPassword = val
		default:
			return fmt.Errorf("unknown configuration key '%s'", args[0])
		}

		if err := cfg.Save(); err != nil {
			pterm.Error.Printf("Failed to save configuration: %v\n", err)
			return err
		}

		pterm.Success.Printf("Set '%s' to '%s' in unigodesktop.toml\n", args[0], val)
		return nil
	},
}
