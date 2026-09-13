// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	proxyPrefix = strings.TrimSpace(proxyPrefix)

	client := &http.Client{
		Timeout: 120 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			// If proxyPrefix is configured, ensure redirect target URL also goes through proxy
			if proxyPrefix != "" && !strings.EqualFold(proxyPrefix, "direct") {
				targetURL := req.URL.String()
				if !strings.HasPrefix(targetURL, proxyPrefix) {
					newURL := BuildProxyURL(targetURL, proxyPrefix)
					if parsedURL, err := url.Parse(newURL); err == nil {
						req.URL = parsedURL
					}
				}
			}
			return nil
		},
	}

	finalURL := BuildProxyURL(rawURL, proxyPrefix)

	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create download request: %w", err)
		}
		req.Header.Set("User-Agent", "UniBootDesktop/1.0")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create target directory: %w", err)
			}

			tmpPath := destPath + ".tmp"
			out, err := os.Create(tmpPath)
			if err != nil {
				return fmt.Errorf("failed to create temp destination file: %w", err)
			}

			_, copyErr := io.Copy(out, resp.Body)
			out.Close()

			if copyErr != nil {
				os.Remove(tmpPath)
				lastErr = fmt.Errorf("failed to save file contents: %w", copyErr)
			} else {
				if err := os.Rename(tmpPath, destPath); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("failed to replace destination file: %w", err)
				}
				return nil
			}
		} else {
			if resp != nil {
				lastErr = fmt.Errorf("HTTP status %d", resp.StatusCode)
				resp.Body.Close()
			} else {
				lastErr = err
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	return fmt.Errorf("download failed for %s after retries: %w", rawURL, lastErr)
}

