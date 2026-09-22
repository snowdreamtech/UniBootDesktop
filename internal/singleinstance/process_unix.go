//go:build !windows

package singleinstance

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func terminateStaleProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal stale instance: %w", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := process.Kill(); err != nil && processAlive(pid) {
		return fmt.Errorf("kill stale instance: %w", err)
	}
	return nil
}
