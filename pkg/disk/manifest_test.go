// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package disk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniBootManifest_Lifecycle(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Initially no manifest
	assert.False(t, HasUniBootManifest(tmpDir))
	_, err := ReadUniBootManifest(tmpDir)
	assert.Error(t, err)

	// 2. Write cloud mode manifest
	err = WriteUniBootManifest(tmpDir, "cloud", "1.2.0")
	assert.NoError(t, err)
	assert.True(t, HasUniBootManifest(tmpDir))

	// 3. Read and verify content
	m, err := ReadUniBootManifest(tmpDir)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, MagicUniBootDisk, m.Magic)
	assert.Equal(t, "cloud", m.Mode)
	assert.Equal(t, "1.2.0", m.Version)
	assert.Equal(t, "ipxe", m.Engine.Name)
	assert.NotEmpty(t, m.UUID)
	assert.Greater(t, m.CreatedAt, int64(0))

	// 4. Overwrite with hybrid mode manifest
	err = WriteUniBootManifest(tmpDir, "hybrid", "2.0.0")
	assert.NoError(t, err)

	m2, err := ReadUniBootManifest(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "hybrid", m2.Mode)
	assert.Equal(t, "2.0.0", m2.Version)
	assert.Equal(t, "ventoy+ipxe", m2.Engine.Name)

	// 5. Test invalid magic
	manifestPath := filepath.Join(tmpDir, "ipxe", "uniboot.json")
	_ = os.WriteFile(manifestPath, []byte(`{"magic":"WRONG_MAGIC","mode":"cloud"}`), 0644)
	assert.False(t, HasUniBootManifest(tmpDir))

	_, errWrong := ReadUniBootManifest(tmpDir)
	assert.Error(t, errWrong)
	assert.Contains(t, errWrong.Error(), "magic mismatch")
}
