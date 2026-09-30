// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package config

import (
	"fmt"
	"os"
	"syscall"
)

func acquireConfigLock(cfgPath string) (*os.File, func(), error) {
	lockPath := cfgPath + ".lock"
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("open config lock: %w", err)
	}
	overlapped := &syscall.Overlapped{}
	if err := syscall.LockFileEx(syscall.Handle(file.Fd()), syscall.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, overlapped); err != nil {
		_ = file.Close()
		return nil, nil, fmt.Errorf("lock config file: %w", err)
	}
	unlock := func() {
		_ = syscall.UnlockFileEx(syscall.Handle(file.Fd()), 0, 1, 0, overlapped)
		_ = file.Close()
		_ = os.Remove(lockPath)
	}
	return file, unlock, nil
}
