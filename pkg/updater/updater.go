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
	pkgHttp "github.com/snowdreamtech/unigodesktop/internal/http"
	"github.com/snowdreamtech/unigodesktop/internal/updater"
	"github.com/snowdreamtech/unigodesktop/internal/version"
)

// UpdateStatus represents release update metadata.
type UpdateStatus struct {
	HasUpdate   bool   `json:"hasUpdate"`
	CurrentTag  string `json:"currentTag"`
	LatestTag   string `json:"latestTag"`
	DownloadURL string `json:"downloadUrl"`
}

// HasNewVersion checks whether latestTag is semantically newer than currentTag.
// Returns false if currentTag is empty, "N/A", "dev", or if latestTag is not newer.
func HasNewVersion(currentTag, latestTag string) bool {
	cleanCur := strings.TrimSpace(currentTag)
	cleanLatest := strings.TrimSpace(latestTag)
	if cleanCur == "" || cleanCur == "N/A" || cleanCur == "dev" {
		return false
	}
	if cleanLatest == "" {
		return false
	}
	return version.CompareVersions(cleanLatest, cleanCur) > 0
}

// CheckUpdate queries GitHub Releases for newer release versions.
func CheckUpdate(ctx context.Context) *UpdateStatus {
	currentTag := env.GitTag

	info, err := updater.FetchLatestReleaseInfo(ctx)
	if err == nil && info != nil {
		latestTag := info.TagName
		hasUpdate := HasNewVersion(currentTag, latestTag)
		var downloadURL string
		if hasUpdate {
			downloadURL = "https://github.com/snowdreamtech/UniGoDesktop/releases/tag/" + latestTag
		}
		return &UpdateStatus{
			HasUpdate:   hasUpdate,
			CurrentTag:  currentTag,
			LatestTag:   latestTag,
			DownloadURL: downloadURL,
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

	// Per-attempt timeout of 15s to prevent hanging endlessly on blocked networks
	client := pkgHttp.NewClientWithTimeout(15 * time.Second)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
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
	}

	finalURL := BuildProxyURL(rawURL, proxyPrefix)

	var lastErr error
	for i := 0; i < 2; i++ {
		reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, finalURL, nil)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create download request: %w", err)
		}
		req.Header.Set("User-Agent", "UniGoDesktop/1.0")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				resp.Body.Close()
				cancel()
				return fmt.Errorf("failed to create target directory: %w", err)
			}

			tmpPath := destPath + ".tmp"
			out, err := os.Create(tmpPath)
			if err != nil {
				resp.Body.Close()
				cancel()
				return fmt.Errorf("failed to create temp destination file: %w", err)
			}

			_, copyErr := io.Copy(out, resp.Body)
			resp.Body.Close()
			out.Close()
			cancel()

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
			cancel()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}

	if proxyPrefix == "" || strings.EqualFold(proxyPrefix, "direct") {
		return fmt.Errorf("direct GitHub connection failed (%v). Please configure GitHub proxy prefix in Settings and try again", lastErr)
	}

	return fmt.Errorf("download failed (%s): %w", rawURL, lastErr)
}
