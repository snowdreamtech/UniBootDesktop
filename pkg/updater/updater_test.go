// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"testing"
)

func TestBuildProxyURL(t *testing.T) {
	rawURL := "https://github.com/snowdreamtech/UniBoot/releases/download/v1.0.0/undionly.kpxe"

	tests := []struct {
		proxyPrefix string
		expected    string
	}{
		{"", rawURL},
		{"direct", rawURL},
		{"DIRECT", rawURL},
		{"https://ghproxy.net", "https://ghproxy.net/" + rawURL},
		{"https://ghproxy.net/", "https://ghproxy.net/" + rawURL},
		{"https://mirror.ghproxy.com/", "https://mirror.ghproxy.com/" + rawURL},
	}

	for _, tt := range tests {
		got := BuildProxyURL(rawURL, tt.proxyPrefix)
		if got != tt.expected {
			t.Errorf("BuildProxyURL(%q, %q) = %q; want %q", rawURL, tt.proxyPrefix, got, tt.expected)
		}
	}
}
