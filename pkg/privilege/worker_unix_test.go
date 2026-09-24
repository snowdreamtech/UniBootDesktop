// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build unix

package privilege

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStaleSocketCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a socket belonging to a dead PID
	deadSocket := filepath.Join(tmpDir, "w-99999999.sock")
	if err := os.WriteFile(deadSocket, []byte(""), 0666); err != nil {
		t.Fatal(err)
	}

	// Create a socket belonging to current (alive) PID
	aliveSocket := filepath.Join(tmpDir, fmt.Sprintf("w-%d.sock", os.Getpid()))
	if err := os.WriteFile(aliveSocket, []byte(""), 0666); err != nil {
		t.Fatal(err)
	}

	cleanStaleSockets(tmpDir)

	if _, err := os.Stat(deadSocket); !os.IsNotExist(err) {
		t.Fatalf("expected dead socket %s to be cleaned up", deadSocket)
	}

	if _, err := os.Stat(aliveSocket); err != nil {
		t.Fatalf("expected alive socket %s to be preserved", aliveSocket)
	}
}
