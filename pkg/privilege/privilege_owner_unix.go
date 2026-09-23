// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build unix

package privilege

import (
	"fmt"
	"os"
	"syscall"
)

func snapshotOwner(info os.FileInfo) (uid, gid int, ok bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false
	}
	return int(stat.Uid), int(stat.Gid), true
}

func chownPath(path string, uid, gid int) error {
	if uid < 0 || gid < 0 {
		return fmt.Errorf("invalid owner %d:%d", uid, gid)
	}
	if err := os.Chown(path, uid, gid); err == nil {
		return nil
	}
	spec := fmt.Sprintf("%d:%d", uid, gid)
	_, err := RunElevated("Restore or adjust raw disk ownership", "chown "+spec+" "+path)
	return err
}
