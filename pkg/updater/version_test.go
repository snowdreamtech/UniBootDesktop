// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import "testing"

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "newer release", current: "v1.2.3", latest: "v1.3.0", want: true},
		{name: "same release", current: "v1.2.3", latest: "v1.2.3", want: false},
		{name: "latest is older", current: "v1.3.0", latest: "v1.2.3", want: false},
		{name: "release prefix is optional", current: "1.2.3", latest: "v1.2.4", want: true},
		{name: "invalid latest version", current: "v1.2.3", latest: "nightly", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNewerVersion(tt.current, tt.latest); got != tt.want {
				t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}
