// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployModeA(t *testing.T) {
	ctx := context.Background()

	res, err := DeployModeA(ctx, "/Volumes/MyUSB", "exFAT")
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "/Volumes/MyUSB", res.Target)
	assert.Contains(t, res.Message, "Successfully deployed")

	_, err = DeployModeA(ctx, "/", "exFAT")
	assert.Error(t, err)
}

func TestDeployModeB(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	res, err := DeployModeB(ctx, "dummy_usb_disk")
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "dummy_usb_disk", res.Target)
	assert.Contains(t, res.Message, "Successfully deployed Cloud Pure Mode B")

	_, err = DeployModeB(ctx, "/")
	assert.Error(t, err)
}

func TestDeployModeABatch(t *testing.T) {
	ctx := context.Background()

	// Empty list
	_, err := DeployModeABatch(ctx, []string{}, "exFAT")
	assert.Error(t, err)

	// Valid targets
	results, err := DeployModeABatch(ctx, []string{"/Volumes/USB1", "/Volumes/USB2"}, "exFAT")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)

	// System drive included -> validation error
	_, err = DeployModeABatch(ctx, []string{"/Volumes/USB1", "/"}, "exFAT")
	assert.Error(t, err)
}

func TestDeployModeBBatch(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	// Empty list
	_, err := DeployModeBBatch(ctx, []string{})
	assert.Error(t, err)

	// Valid targets
	results, err := DeployModeBBatch(ctx, []string{"dummy_usb_1", "dummy_usb_2"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)

	// System drive included -> validation error
	_, err = DeployModeBBatch(ctx, []string{"dummy_usb_1", "/"})
	assert.Error(t, err)
}
