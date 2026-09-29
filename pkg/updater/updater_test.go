// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unigodesktop/internal/updater"
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

func TestGetAppBundlePath(t *testing.T) {
	tests := []struct {
		name     string
		execPath string
		expected string
	}{
		{
			name:     "macOS standard Applications app bundle",
			execPath: "/Applications/UniGoDesktop.app/Contents/MacOS/unigodesktop",
			expected: "/Applications/UniGoDesktop.app",
		},
		{
			name:     "macOS user Applications bundle",
			execPath: "/Users/alice/Applications/UniGoDesktop.app/Contents/MacOS/UniGoDesktop",
			expected: "/Users/alice/Applications/UniGoDesktop.app",
		},
		{
			name:     "macOS build directory bundle",
			execPath: "/workspace/build/bin/UniGoDesktop.app/Contents/MacOS/unigodesktop",
			expected: "/workspace/build/bin/UniGoDesktop.app",
		},
		{
			name:     "direct .app folder without trailing path",
			execPath: "/Applications/UniGoDesktop.app",
			expected: "/Applications/UniGoDesktop.app",
		},
		{
			name:     "standalone Linux/Unix binary",
			execPath: "/usr/local/bin/unigodesktop",
			expected: "",
		},
		{
			name:     "subfolder containing app word but not .app",
			execPath: "/opt/application/bin/unigodesktop",
			expected: "",
		},
		{
			name:     "folder containing .app_data",
			execPath: "/Users/bob/.app_data/bin/unigodesktop",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAppBundlePath(tt.execPath)
			if got != tt.expected {
				t.Errorf("GetAppBundlePath(%q) = %q; want %q", tt.execPath, got, tt.expected)
			}
		})
	}
}

func TestFindGuiReleaseAsset_Darwin(t *testing.T) {
	assets := []updater.ReleaseAsset{
		{Name: "unigodesktop-cli_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/cli.tar.gz"},
		{Name: "unigodesktop-cli_Darwin_arm64.tar.gz.sbom.json", BrowserDownloadURL: "https://example.com/cli.sbom"},
		{Name: "unigodesktop-gui_windows_amd64_portable.zip", BrowserDownloadURL: "https://example.com/win.zip"},
		{Name: "UniGoDesktop.dmg", BrowserDownloadURL: "https://example.com/UniGoDesktop.dmg"},
		{Name: "UniGoDesktop.dmg.sbom.spdx.json", BrowserDownloadURL: "https://example.com/dmg.sbom"},
		{Name: "UniGoDesktop.dmg.sigstore.json", BrowserDownloadURL: "https://example.com/dmg.sig"},
	}

	asset, err := FindGuiReleaseAsset(assets, "darwin", "arm64")
	if err != nil {
		t.Fatalf("unexpected error finding asset: %v", err)
	}
	if asset.Name != "UniGoDesktop.dmg" {
		t.Errorf("expected UniGoDesktop.dmg, got %s", asset.Name)
	}

	// Guaranteed to never select Windows or Linux or CLI archives on Darwin
	if asset.Name != "UniGoDesktop.dmg" {
		t.Errorf("asset matching violated Darwin integrity: %s", asset.Name)
	}
}

func TestFindGuiReleaseAsset_Windows(t *testing.T) {
	assets := []updater.ReleaseAsset{
		{Name: "unigodesktop-cli_Windows_x86_64.zip", BrowserDownloadURL: "https://example.com/cli.zip"},
		{Name: "UniGoDesktop.dmg", BrowserDownloadURL: "https://example.com/UniGoDesktop.dmg"},
		{Name: "unigodesktop-gui_windows_amd64_installer.exe", BrowserDownloadURL: "https://example.com/installer.exe"},
		{Name: "unigodesktop-gui_windows_amd64_portable.zip", BrowserDownloadURL: "https://example.com/portable.zip"},
	}

	asset, err := FindGuiReleaseAsset(assets, "windows", "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Prefers portable zip
	if asset.Name != "unigodesktop-gui_windows_amd64_portable.zip" {
		t.Errorf("expected portable.zip, got %s", asset.Name)
	}
}

func TestFindGuiReleaseAsset_Linux(t *testing.T) {
	assets := []updater.ReleaseAsset{
		{Name: "unigodesktop-cli_Linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/cli.tar.gz"},
		{Name: "UniGoDesktop.dmg", BrowserDownloadURL: "https://example.com/UniGoDesktop.dmg"},
		{Name: "unigodesktop-gui_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/gui.tar.gz"},
		{Name: "unigodesktop-gui_0.4.0_linux_amd64.AppImage", BrowserDownloadURL: "https://example.com/gui.AppImage"},
	}

	asset, err := FindGuiReleaseAsset(assets, "linux", "amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Prefers AppImage
	if asset.Name != "unigodesktop-gui_0.4.0_linux_amd64.AppImage" {
		t.Errorf("expected AppImage, got %s", asset.Name)
	}
}

func TestFindGuiReleaseAsset_NotFound(t *testing.T) {
	assets := []updater.ReleaseAsset{
		{Name: "unigodesktop-cli_Linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/cli.tar.gz"},
	}

	_, err := FindGuiReleaseAsset(assets, "darwin", "arm64")
	if err == nil {
		t.Error("expected error when no matching GUI asset found, got nil")
	}
}

func TestPendingUpdate_SaveGetClear(t *testing.T) {
	tmpDir := t.TempDir()

	pending := &PendingUpdate{
		Shell:      "/bin/bash",
		ScriptPath: filepath.Join(tmpDir, "apply_update.sh"),
		Target:     "/Applications/UniGoDesktop.app",
		Staged:     filepath.Join(tmpDir, "UniGoDesktop.app"),
	}

	// Save
	if err := SavePendingUpdate(tmpDir, pending); err != nil {
		t.Fatalf("SavePendingUpdate failed: %v", err)
	}

	// Get
	got, err := GetPendingUpdate(tmpDir)
	if err != nil {
		t.Fatalf("GetPendingUpdate failed: %v", err)
	}
	if got.Shell != pending.Shell || got.ScriptPath != pending.ScriptPath || got.Target != pending.Target || got.Staged != pending.Staged {
		t.Errorf("GetPendingUpdate mismatch: got %+v, want %+v", got, pending)
	}

	// Clear
	if err := ClearPendingUpdate(tmpDir); err != nil {
		t.Fatalf("ClearPendingUpdate failed: %v", err)
	}

	// Get after clear should fail
	_, err = GetPendingUpdate(tmpDir)
	if err == nil {
		t.Error("expected error getting pending update after clear, got nil")
	}
}

func TestDownloadWithProgress(t *testing.T) {
	content := "test update payload data"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "update.bin")

	var progressCalled bool
	onProgress := func(pct int, status string) {
		progressCalled = true
	}

	err := DownloadWithProgress(context.Background(), server.URL, destFile, "", onProgress, 0, 100)
	if err != nil {
		t.Fatalf("DownloadWithProgress failed: %v", err)
	}

	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
	if !progressCalled {
		t.Error("expected onProgress to be called")
	}
}
