// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/internal/updater"
)

// UpdateStatus represents release update metadata.
type UpdateStatus struct {
	HasUpdate     bool   `json:"hasUpdate"`
	CurrentTag    string `json:"currentTag"`
	LatestTag     string `json:"latestTag"`
	DownloadURL   string `json:"downloadUrl"`
}

// CheckUpdate queries GitHub Releases for newer release versions.
func CheckUpdate(ctx context.Context) *UpdateStatus {
	currentTag := env.GitTag
	if currentTag == "" || currentTag == "N/A" {
		currentTag = "v0.1.0"
	}

	info, err := updater.FetchLatestReleaseInfo(ctx)
	if err == nil && info != nil {
		return &UpdateStatus{
			HasUpdate:   true,
			CurrentTag:  currentTag,
			LatestTag:   info.TagName,
			DownloadURL: "https://github.com/snowdreamtech/UniGoDesktop/releases/tag/" + info.TagName,
		}
	}

	return &UpdateStatus{
		HasUpdate:   false,
		CurrentTag:  currentTag,
		LatestTag:   currentTag,
		DownloadURL: "",
	}
}

// BuildProxyURL formats a URL with the given GitHub proxy prefix if configured.
func BuildProxyURL(rawURL string, proxyPrefix string) string {
	proxyPrefix = strings.TrimSpace(proxyPrefix)
	if proxyPrefix == "" || strings.EqualFold(proxyPrefix, "direct") {
		return rawURL
	}
	if !strings.HasSuffix(proxyPrefix, "/") {
		proxyPrefix += "/"
	}
	return proxyPrefix + rawURL
}

// DownloadFileWithProxy downloads a remote URL to destPath using optional proxy prefix and retries.
func DownloadFileWithProxy(ctx context.Context, rawURL string, destPath string, proxyPrefix string) error {
	finalURL := BuildProxyURL(rawURL, proxyPrefix)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	var resp *http.Response
	var downloadErr error
	for i := 0; i < 3; i++ {
		resp, downloadErr = client.Do(req)
		if downloadErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	if downloadErr != nil {
		return fmt.Errorf("failed after retries: %w", downloadErr)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		return fmt.Errorf("unexpected status code: %d", status)
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save file contents: %w", err)
	}

	return nil
}
