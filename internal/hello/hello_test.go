// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hello

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPrintHello(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintHello()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if len(output) == 0 {
		t.Errorf("expected output from PrintHello, got empty string")
	}
}
