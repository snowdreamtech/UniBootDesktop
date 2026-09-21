//go:build !windows

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
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		return nil, nil, fmt.Errorf("lock config file: %w", err)
	}
	unlock := func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
		_ = os.Remove(lockPath)
	}
	return file, unlock, nil
}
