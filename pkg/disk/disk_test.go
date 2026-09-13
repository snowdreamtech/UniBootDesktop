// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIgnoredVolume(t *testing.T) {
	tests := []struct {
		name     string
		volName  string
		expected bool
	}{
		{"Macintosh HD system disk", "Macintosh HD", true},
		{"System volume", "System", true},
		{"VTOYEFI upper", "VTOYEFI", true},
		{"vtoyefi lower", "vtoyefi", true},
		{"VTOYEFI with prefix", "VTOYEFI_BOOT", true},
		{"Normal Ventoy volume", "Ventoy", false},
		{"Normal USB volume", "MyUSBKey", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIgnoredVolume(tt.volName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTargetDisk(t *testing.T) {
	assert.Error(t, ValidateTargetDisk(""))
	assert.Error(t, ValidateTargetDisk("/"))
	assert.Error(t, ValidateTargetDisk("C:"))
	assert.NoError(t, ValidateTargetDisk("/Volumes/MyUSBKey"))
}
