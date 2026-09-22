// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	pkgHttp "github.com/snowdreamtech/unigodesktop/internal/http"
	"github.com/snowdreamtech/unigodesktop/internal/updater"
	internalVersion "github.com/snowdreamtech/unigodesktop/internal/version"
)

const maxDownloadSize int64 = 512 * 1024 * 1024

// UpdateStatus represents release update metadata.
type UpdateStatus struct {
	HasUpdate   bool   `json:"hasUpdate"`
	CurrentTag  string `json:"currentTag"`
	LatestTag   string `json:"latestTag"`
	DownloadURL string `json:"downloadUrl"`
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
			HasUpdate:   isNewerVersion(currentTag, info.TagName),
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

func isNewerVersion(currentTag string, latestTag string) bool {
	current, currentErr := internalVersion.ParseSemVer(currentTag)
	latest, latestErr := internalVersion.ParseSemVer(latestTag)
	if currentErr != nil || latestErr != nil {
		return false
	}
	return latest.Compare(current) > 0
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
		req.Header.Set("User-Agent", "UniBootDesktop/1.0")
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

			_, copyErr := copyWithLimit(out, resp.Body, maxDownloadSize)
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

// GuiUpdateProgress represents realtime download progress for GUI updates.
type GuiUpdateProgress struct {
	Percentage  int    `json:"percentage"`
	Transferred int64  `json:"transferred"`
	Total       int64  `json:"total"`
	Status      string `json:"status"`
	TargetFile  string `json:"targetFile"`
}

// GuiUpdateResult represents the GUI update operation outcome.
type GuiUpdateResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	TargetFile     string `json:"targetFile"`
	RequireRestart bool   `json:"requireRestart"`
}

type progressWriter struct {
	total       int64
	transferred int64
	onProgress  func(transferred int64, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.transferred += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.transferred, pw.total)
	}
	return n, nil
}

// PerformGuiUpdate fetches latest GUI release asset and downloads with progress reporting.
func PerformGuiUpdate(ctx context.Context, proxyPrefix string, progressCb func(p GuiUpdateProgress)) (*GuiUpdateResult, error) {
	rel, err := updater.FetchLatestReleaseInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	var bestAsset *updater.ReleaseAsset
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for _, asset := range rel.Assets {
		nameLower := strings.ToLower(asset.Name)
		if strings.Contains(nameLower, "gui") || strings.Contains(nameLower, "desktop") || strings.HasSuffix(nameLower, ".dmg") || strings.HasSuffix(nameLower, ".exe") || strings.HasSuffix(nameLower, ".appimage") {
			if goos == "darwin" && (strings.HasSuffix(nameLower, ".dmg") || strings.HasSuffix(nameLower, ".zip")) {
				bestAsset = &asset
				break
			} else if goos == "windows" && (strings.HasSuffix(nameLower, ".exe") || strings.HasSuffix(nameLower, ".zip")) {
				bestAsset = &asset
				break
			} else if goos == "linux" && (strings.HasSuffix(nameLower, ".appimage") || strings.HasSuffix(nameLower, ".deb") || strings.HasSuffix(nameLower, ".tar.gz")) {
				bestAsset = &asset
				break
			}
		}
	}

	if bestAsset == nil && len(rel.Assets) > 0 {
		bestAsset = &rel.Assets[0]
	}

	if bestAsset == nil {
		return nil, fmt.Errorf("no suitable GUI release package found for %s/%s", goos, goarch)
	}

	var checksumAsset *updater.ReleaseAsset
	for _, asset := range rel.Assets {
		if strings.EqualFold(asset.Name, "checksums.txt") {
			checksumAsset = &asset
			break
		}
	}
	if checksumAsset == nil {
		return nil, fmt.Errorf("release does not contain checksums.txt")
	}

	downloadDir := filepath.Join(env.GetDataDir(), "downloads")
	_ = os.MkdirAll(downloadDir, 0755)
	destPath := filepath.Join(downloadDir, bestAsset.Name)
	checksumPath := destPath + ".checksums"
	defer os.Remove(checksumPath)
	if err := DownloadFileWithProxy(ctx, checksumAsset.BrowserDownloadURL, checksumPath, proxyPrefix); err != nil {
		return nil, fmt.Errorf("download checksums: %w", err)
	}
	checksumData, err := os.ReadFile(checksumPath)
	if err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}
	expectedChecksum, err := findSHA256Checksum(checksumData, bestAsset.Name)
	if err != nil {
		return nil, err
	}

	if progressCb != nil {
		progressCb(GuiUpdateProgress{
			Percentage: 0,
			Status:     fmt.Sprintf("Downloading %s...", bestAsset.Name),
			TargetFile: destPath,
		})
	}

	err = DownloadFileWithProgress(ctx, bestAsset.BrowserDownloadURL, destPath, proxyPrefix, func(transferred, total int64) {
		pct := 0
		if total > 0 {
			pct = int((float64(transferred) / float64(total)) * 100)
		}
		if progressCb != nil {
			progressCb(GuiUpdateProgress{
				Percentage:  pct,
				Transferred: transferred,
				Total:       total,
				Status:      fmt.Sprintf("Downloading %s (%d%%)", bestAsset.Name, pct),
				TargetFile:  destPath,
			})
		}
	})

	if err != nil {
		return nil, err
	}
	if err := verifySHA256File(destPath, expectedChecksum); err != nil {
		_ = os.Remove(destPath)
		return nil, fmt.Errorf("verify GUI update package: %w", err)
	}

	return &GuiUpdateResult{
		Success:        true,
		Message:        fmt.Sprintf("Successfully downloaded %s (%s)", rel.TagName, bestAsset.Name),
		TargetFile:     destPath,
		RequireRestart: true,
	}, nil
}

// DownloadFileWithProgress downloads a file with progress reporting and optional proxy.
func DownloadFileWithProgress(ctx context.Context, rawURL string, destPath string, proxyPrefix string, onProgress func(transferred, total int64)) error {
	finalURL := BuildProxyURL(rawURL, proxyPrefix)
	client := pkgHttp.NewClientWithTimeout(60 * time.Second)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "UniGoDesktop/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to download URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	pw := &progressWriter{
		total:      resp.ContentLength,
		onProgress: onProgress,
	}

	if _, err := copyWithLimit(out, io.TeeReader(resp.Body, pw), maxDownloadSize); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed during download: %w", err)
	}

	_ = os.Remove(destPath)
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to place downloaded file: %w", err)
	}

	return nil
}

func copyWithLimit(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	if maxBytes <= 0 {
		return 0, fmt.Errorf("download size limit must be positive")
	}

	bytesCopied, err := io.Copy(dst, io.LimitReader(src, maxBytes+1))
	if err != nil {
		return bytesCopied, err
	}
	if bytesCopied > maxBytes {
		return bytesCopied, fmt.Errorf("download exceeds maximum size of %d bytes", maxBytes)
	}
	return bytesCopied, nil
}

func findSHA256Checksum(data []byte, assetName string) (string, error) {
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.TrimPrefix(fields[1], "*") != assetName {
			continue
		}
		return fields[0], nil
	}
	return "", fmt.Errorf("checksum for %s not found", assetName)
}

func verifySHA256File(filePath string, expected string) error {
	expected = strings.TrimSpace(expected)
	if len(expected) != sha256.Size*2 {
		return fmt.Errorf("invalid SHA-256 checksum length")
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return fmt.Errorf("invalid SHA-256 checksum: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}
