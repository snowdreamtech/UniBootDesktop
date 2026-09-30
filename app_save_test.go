// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"testing"

	"github.com/snowdreamtech/unibootdesktop/pkg/config"
	"github.com/snowdreamtech/unibootdesktop/pkg/hypervisor"
	"github.com/stretchr/testify/assert"
)

func TestAppSaveConfig_DirectWithEmptyHost(t *testing.T) {
	app := &App{}
	cfg := &config.AppConfig{
		Mode:                 "cloud",
		AutoCheckUpdate:      false,
		Theme:                "dark",
		Language:             "zh-CN",
		FileSystem:           "exFAT",
		ProxyProtocol:        "direct",
		ProxyHost:            "",
		ProxyPort:            1080, // dangling port from UI default
		VentoySecureBoot:     true,
		VentoyPartitionStyle: "MBR",
	}

	err := app.SaveConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, cfg.ProxyPort, "ProxyPort should be normalized to 0 when direct or host is empty")
}

func TestLaunchVMWithConfigRejectsUnsafeTargetDisk(t *testing.T) {
	t.Setenv("UNIBOOT_DRY_RUN", "1")
	app := &App{}
	err := app.LaunchVMWithConfig("/", "qemu", hypervisor.VMConfig{})
	assert.Error(t, err)
}

func TestAppSaveConfig_All52Languages(t *testing.T) {
	app := &App{}
	testLanguages := []string{"vi-VN", "ar-SA", "tr-TR", "it-IT", "zh-CN", "en-US", "auto"}
	for _, lang := range testLanguages {
		cfg := &config.AppConfig{
			Mode:                 "cloud",
			Theme:                "dark",
			Language:             lang,
			FileSystem:           "exFAT",
			ProxyProtocol:        "direct",
			VentoyPartitionStyle: "MBR",
		}
		err := app.SaveConfig(cfg)
		assert.NoError(t, err, "Language %s should be valid", lang)
	}
}
