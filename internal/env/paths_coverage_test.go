// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"errors"
	"os"
	"testing"
)

func TestEnvPaths_CoverageMore(t *testing.T) {
	// 1. Test GetConfigDir windows path
	oldOS := RuntimeGOOS
	RuntimeGOOS = "windows"

	os.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", "")
	os.Setenv("UNIBOOTDESKTOP_XDG_CONFIG_HOME", "")

	oldOsUserConfigDir := OsUserConfigDir
	OsUserConfigDir = func() (string, error) { return "C:\\Users\\test\\AppData\\Roaming", nil }
	GetConfigDir()

	OsUserConfigDir = func() (string, error) { return "", errors.New("err") }
	GetConfigDir()

	// 2. Test GetDataDir windows
	os.Setenv("UNIBOOTDESKTOP_DATA_DIR", "")
	os.Setenv("UNIBOOTDESKTOP_XDG_DATA_HOME", "")
	os.Setenv("UNIBOOTDESKTOP_LOCALAPPDATA", "C:\\AppData\\Local")
	GetDataDir()
	os.Unsetenv("UNIBOOTDESKTOP_LOCALAPPDATA")
	GetDataDir()

	// 3. Test GetCacheDir darwin
	os.Setenv("UNIBOOTDESKTOP_CACHE_DIR", "")
	os.Setenv("UNIBOOTDESKTOP_XDG_CACHE_HOME", "")
	RuntimeGOOS = "darwin"
	GetCacheDir()

	// 4. Test GetCacheDir windows
	RuntimeGOOS = "windows"
	GetCacheDir()

	// 5. Test GetLockFilePath error
	os.Setenv("UNIBOOTDESKTOP_LOCK_FILE", "")
	oldOsGetwd := OsGetwd
	OsGetwd = func() (string, error) { return "", errors.New("err") }
	GetLockFilePath()

	// Reset
	RuntimeGOOS = oldOS
	OsUserConfigDir = oldOsUserConfigDir
	OsGetwd = oldOsGetwd

	// 6. Test OsUserHomeDir error paths
	oldOsUserHomeDir := OsUserHomeDir
	OsUserHomeDir = func() (string, error) { return "", errors.New("err") }
	os.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", "")
	os.Setenv("UNIBOOTDESKTOP_XDG_CONFIG_HOME", "")
	GetConfigDir()

	os.Setenv("UNIBOOTDESKTOP_DATA_DIR", "")
	os.Setenv("UNIBOOTDESKTOP_XDG_DATA_HOME", "")
	GetDataDir()

	os.Setenv("UNIBOOTDESKTOP_CACHE_DIR", "")
	os.Setenv("UNIBOOTDESKTOP_XDG_CACHE_HOME", "")
	GetCacheDir()

	OsUserHomeDir = oldOsUserHomeDir

	// 7. Test direct UNIBOOTDESKTOP_DIR env vars
	os.Setenv("UNIBOOTDESKTOP_CONFIG_DIR", "/custom/config")
	GetConfigDir()
	os.Unsetenv("UNIBOOTDESKTOP_CONFIG_DIR")

	os.Setenv("UNIBOOTDESKTOP_DATA_DIR", "/custom/data")
	GetDataDir()
	os.Unsetenv("UNIBOOTDESKTOP_DATA_DIR")

	os.Setenv("UNIBOOTDESKTOP_CACHE_DIR", "/custom/cache")
	GetCacheDir()
	os.Unsetenv("UNIBOOTDESKTOP_CACHE_DIR")

	// 8. Test GetCacheDir linux fallback
	RuntimeGOOS = "linux"
	GetCacheDir()
	RuntimeGOOS = oldOS
}
