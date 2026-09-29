// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows && !nogui

package main

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = 0x00000008 // DETACHED_PROCESS
}
