// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_LifecycleAndAPIs(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	ctx := context.Background()
	app.startup(ctx)

	// Greet
	assert.Equal(t, "Hello World, Welcome to UniGoDesktop!", app.Greet(""))
	assert.Equal(t, "Hello Alice, Welcome to UniGoDesktop!", app.Greet("Alice"))

	// HelloInfo
	info := app.GetHelloInfo()
	assert.NotNil(t, info)
	assert.Contains(t, info.Greeting, "Hello World")

	// SystemInfo
	sys := app.GetSystemInfo()
	assert.NotNil(t, sys)
	assert.Equal(t, "UniGoDesktop", sys.AppName)

	// Config
	cfg, err := app.GetConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	err = app.SaveConfig(cfg)
	assert.NoError(t, err)
}
