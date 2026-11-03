// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows

package hypervisor

import (
	"os/exec"
	"syscall"
)

func canAccessDeviceNode(path string) bool {
	// 6 = R_OK (4) | W_OK (2)
	return syscall.Access(path, 6) == nil
}

func setQEMUSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
