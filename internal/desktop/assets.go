// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package desktop

import (
	"embed"
	"io/fs"
	"path"
)

// assetsFS embeds the static web assets for the desktop interface.
//
//go:embed static/*
var assetsFS embed.FS

// GetAssetsFS returns the filesystem subtree containing static web assets.
func GetAssetsFS() fs.FS {
	sub, err := fs.Sub(assetsFS, "static")
	if err != nil {
		return assetsFS
	}
	return sub
}

// ReadAsset reads a static asset file by relative path.
func ReadAsset(name string) ([]byte, error) {
	return assetsFS.ReadFile(path.Join("static", name))
}
