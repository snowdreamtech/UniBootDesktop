// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package privilege

import "os"

func snapshotOwner(os.FileInfo) (uid, gid int, ok bool) {
	return 0, 0, false
}

func chownPath(string, int, int) error {
	return nil
}
