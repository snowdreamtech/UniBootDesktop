// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sysinfo

import "testing"

func TestIsSystemDarkTheme(t *testing.T) {
	// IsSystemDarkTheme should execute safely without crashing on any supported platform
	_ = IsSystemDarkTheme()
}

func TestSupportsMicaBackdrop(t *testing.T) {
	// SupportsMicaBackdrop should execute safely without crashing on any supported platform
	_ = SupportsMicaBackdrop()
}
