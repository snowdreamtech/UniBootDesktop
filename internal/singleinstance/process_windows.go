//go:build windows

package singleinstance

import (
	"fmt"
	"os"
)

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	return err == nil && process != nil
}

func terminateStaleProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if err := process.Kill(); err != nil {
		return fmt.Errorf("kill stale instance: %w", err)
	}
	return nil
}
