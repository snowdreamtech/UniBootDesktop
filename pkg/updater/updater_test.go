// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"testing"
)

func TestBuildProxyURL(t *testing.T) {
	rawURL := "https://github.com/snowdreamtech/unigodesktop/releases/download/v1.0.0/app.tar.gz"

	tests := []struct {
		proxyPrefix string
		expected    string
	}{
		{"", rawURL},
		{"direct", rawURL},
		{"DIRECT", rawURL},
		{"https://proxy.example.com", "https://proxy.example.com/" + rawURL},
		{"https://proxy.example.com/", "https://proxy.example.com/" + rawURL},
		{"https://my-custom-proxy.org/", "https://my-custom-proxy.org/" + rawURL},
	}

	for _, tt := range tests {
		got := BuildProxyURL(rawURL, tt.proxyPrefix)
		if got != tt.expected {
			t.Errorf("BuildProxyURL(%q, %q) = %q; want %q", rawURL, tt.proxyPrefix, got, tt.expected)
		}
	}
}

func TestHasNewVersion(t *testing.T) {
	tests := []struct {
		name       string
		currentTag string
		latestTag  string
		want       bool
	}{
		{
			name:       "identical version v0.3.4",
			currentTag: "v0.3.4",
			latestTag:  "v0.3.4",
			want:       false,
		},
		{
			name:       "identical version without v prefix",
			currentTag: "0.3.4",
			latestTag:  "v0.3.4",
			want:       false,
		},
		{
			name:       "remote is newer minor",
			currentTag: "v0.3.4",
			latestTag:  "v0.4.0",
			want:       true,
		},
		{
			name:       "remote is newer patch",
			currentTag: "v0.3.4",
			latestTag:  "v0.3.5",
			want:       true,
		},
		{
			name:       "remote is older",
			currentTag: "v0.3.4",
			latestTag:  "v0.3.3",
			want:       false,
		},
		{
			name:       "current is N/A (dev build)",
			currentTag: "N/A",
			latestTag:  "v0.3.4",
			want:       false,
		},
		{
			name:       "current is dev",
			currentTag: "dev",
			latestTag:  "v0.3.4",
			want:       false,
		},
		{
			name:       "current is empty",
			currentTag: "",
			latestTag:  "v0.3.4",
			want:       false,
		},
		{
			name:       "latest is empty",
			currentTag: "v0.3.4",
			latestTag:  "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasNewVersion(tt.currentTag, tt.latestTag)
			if got != tt.want {
				t.Errorf("HasNewVersion(%q, %q) = %v; want %v", tt.currentTag, tt.latestTag, got, tt.want)
			}
		})
	}
}
