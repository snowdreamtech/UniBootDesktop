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

func TestDeployHybridMode(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	res, err := DeployHybridMode(ctx, "dummy_usb_disk", "exFAT")
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "dummy_usb_disk", res.Target)
	assert.Contains(t, res.Message, "Hybrid Mode")

	_, err = DeployHybridMode(ctx, "/", "exFAT")
	assert.Error(t, err)
}

func TestDeployCloudMode(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	res, err := DeployCloudMode(ctx, "dummy_usb_disk", "exFAT")
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "dummy_usb_disk", res.Target)
	assert.Contains(t, res.Message, "Cloud Mode")

	_, err = DeployCloudMode(ctx, "/", "exFAT")
	assert.Error(t, err)
}

func TestDeployHybridModeBatch(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	// Empty list
	_, err := DeployHybridModeBatch(ctx, []string{}, "exFAT")
	assert.Error(t, err)

	// Valid targets
	results, err := DeployHybridModeBatch(ctx, []string{"dummy_usb_1", "dummy_usb_2"}, "exFAT")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)

	// System drive included -> validation error
	_, err = DeployHybridModeBatch(ctx, []string{"dummy_usb_1", "/"}, "exFAT")
	assert.Error(t, err)
}

func TestDeployCloudModeBatch(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	// Empty list
	_, err := DeployCloudModeBatch(ctx, []string{}, "exFAT")
	assert.Error(t, err)

	// Valid targets
	results, err := DeployCloudModeBatch(ctx, []string{"dummy_usb_1", "dummy_usb_2"}, "exFAT")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)

	// System drive included -> validation error
	_, err = DeployCloudModeBatch(ctx, []string{"dummy_usb_1", "/"}, "exFAT")
	assert.Error(t, err)
}

func TestDeployHybridModeBatchPartialFailure(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	results, err := DeployHybridModeBatchWithVentoyAndIso(ctx, []string{"dummy_usb_1", "dummy_usb_fail"}, "exFAT", "", nil, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.False(t, results[1].Success)
	assert.NotNil(t, results[1].Diagnostics)
}

func TestDeployCloudModeBatchPartialFailure(t *testing.T) {
	os.Setenv("UNIBOOT_DRY_RUN", "true")
	defer os.Unsetenv("UNIBOOT_DRY_RUN")

	ctx := context.Background()

	results, err := DeployCloudModeBatch(ctx, []string{"dummy_usb_1", "dummy_usb_fail"}, "exFAT")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.False(t, results[1].Success)
	assert.NotNil(t, results[1].Diagnostics)
}
