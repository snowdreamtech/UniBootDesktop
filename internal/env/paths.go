// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Exported for testing override
var (
	OsUserHomeDir   = os.UserHomeDir
	OsUserConfigDir = os.UserConfigDir
	OsGetwd         = os.Getwd
	RuntimeGOOS     = runtime.GOOS
	RuntimeGOARCH   = runtime.GOARCH
)

// GetFSToolName returns a sanitized tool name for use in filesystem paths.
// It implements Scheme B: provider-tool-name, replacing colons and slashes with hyphens.
func GetFSToolName(tool, backend string) string {
	name := tool
	// If tool already contains the backend as a prefix (followed by a hyphen), don't double it.
	// For 'native' backend, we don't add a prefix to align with mise core tools layout.
	if backend != "" && backend != "native" && !strings.HasPrefix(tool, backend+"-") && !strings.HasPrefix(tool, backend+":") {
		name = backend + "-" + tool
	}

	// Replace colons and slashes with hyphens, and remove @ for consistency with Mise
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "@", "")

	return name
}

// GetConfigDir returns the root configuration directory for UniGoDesktop.
// It uses UNIGODESKTOP_CONFIG_DIR if set, otherwise falls back to XDG config directory.
func GetConfigDir() string {
	if configDir := Get("CONFIG_DIR"); configDir != "" {
		return configDir
	}

	if configHome := Get("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "unigodesktop")
	}

	homeDir, err := OsUserHomeDir()
	if err != nil {
		return "./unigodesktop_config"
	}

	if RuntimeGOOS == "windows" {
		if appData, err := OsUserConfigDir(); err == nil {
			return filepath.Join(appData, "unigodesktop")
		}
	}

	// For macOS and Linux, we unify on the standard XDG ~/.config
	return filepath.Join(homeDir, ".config", "unigodesktop")
}

// GetDataDir returns the root data directory for UniGoDesktop.
// It uses UNIGODESKTOP_DATA_DIR if set, otherwise falls back to appropriate OS directories.
func GetDataDir() string {
	if dataDir := Get("DATA_DIR"); dataDir != "" {
		return dataDir
	}

	// Follow XDG Base Directory Specification for data home if XDG_DATA_HOME is set
	if dataHome := Get("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "unigodesktop")
	}

	homeDir, err := OsUserHomeDir()
	if err != nil {
		return "./unigodesktop_data" // Fallback if home directory cannot be determined
	}

	if RuntimeGOOS == "windows" {
		// Windows stores data in Local AppData
		if localAppData := Get("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "unigodesktop")
		}
		return filepath.Join(homeDir, "AppData", "Local", "unigodesktop")
	}

	// For macOS and Linux, we unify on the standard XDG ~/.local/share
	return filepath.Join(homeDir, ".local", "share", "unigodesktop")
}

// GetDatabasePath returns the path to the UniGoDesktop SQLite database.
func GetDatabasePath() string {
	return filepath.Join(GetDataDir(), "unigodesktop.db")
}

// GetShimsDir returns the directory where UniGoDesktop shims are stored.
func GetShimsDir() string {
	return filepath.Join(GetDataDir(), "shims")
}

// GetInstallsDir returns the directory where tools are installed.
func GetInstallsDir() string {
	return filepath.Join(GetDataDir(), "installs")
}

// GetDownloadsDir returns the directory where artifacts are downloaded before extraction.
func GetDownloadsDir() string {
	return filepath.Join(GetDataDir(), "downloads")
}

// GetPluginsDir returns the directory where plugins (e.g., asdf plugins) are stored.
func GetPluginsDir() string {
	return filepath.Join(GetDataDir(), "plugins")
}

// GetCacheDir returns the directory where cache files are stored.
// It follows XDG Base Directory Specification for cache home.
func GetCacheDir() string {
	if cacheDir := Get("CACHE_DIR"); cacheDir != "" {
		return cacheDir
	}

	if cacheHome := Get("XDG_CACHE_HOME"); cacheHome != "" {
		return filepath.Join(cacheHome, "unigodesktop")
	}

	homeDir, err := OsUserHomeDir()
	if err != nil {
		return "./unigodesktop_cache"
	}

	if RuntimeGOOS == "darwin" {
		// macOS standard cache directory
		return filepath.Join(homeDir, "Library", "Caches", "unigodesktop")
	}

	if RuntimeGOOS == "windows" {
		// Windows uses Local AppData for cache too, but usually in a 'cache' subfolder
		return filepath.Join(GetDataDir(), "cache")
	}

	// Default for Linux and others (XDG standard)
	return filepath.Join(homeDir, ".cache", "unigodesktop")
}

// GetLockFilePath returns the path of the unigodesktop.lock file.
func GetLockFilePath() string {
	if custom := Get("LOCK_FILE"); custom != "" {
		return custom
	}
	wd, err := OsGetwd()
	if err != nil {
		return "unigodesktop.lock"
	}
	return filepath.Join(wd, "unigodesktop.lock")
}

// GetGlobalConfigPath returns the path to the global unigodesktop.toml configuration file.
func GetGlobalConfigPath() string {
	return filepath.Join(GetConfigDir(), "unigodesktop.toml")
}
