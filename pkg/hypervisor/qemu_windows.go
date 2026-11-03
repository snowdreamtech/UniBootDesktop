// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package hypervisor

import (
	"os"
	"os/exec"
)

func canAccessDeviceNode(path string) bool {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err == nil {
		_ = f.Close()
		return true
	}
	return false
}

func setQEMUSysProcAttr(cmd *exec.Cmd) {
	// On Windows, Setsid is not supported; cmd runs without Unix session decoupling.
}
