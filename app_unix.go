// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows && !nogui

package main

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
}

// newDetachedCmd creates a new exec.Cmd configured to run detached from the current process group.
func newDetachedCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	detachProcess(cmd)
	return cmd
}
