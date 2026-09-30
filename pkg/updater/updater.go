// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/snowdreamtech/unibootdesktop/internal/archive"
	"github.com/snowdreamtech/unibootdesktop/internal/env"
	pkgHttp "github.com/snowdreamtech/unibootdesktop/internal/http"
	"github.com/snowdreamtech/unibootdesktop/internal/updater"
	"github.com/snowdreamtech/unibootdesktop/internal/version"
)

// UpdateStatus represents release update metadata.
type UpdateStatus struct {
	HasUpdate   bool   `json:"hasUpdate"`
	CurrentTag  string `json:"currentTag"`
	LatestTag   string `json:"latestTag"`
	DownloadURL string `json:"downloadUrl"`
}

// GuiUpdateResult represents the outcome of an in-app GUI update operation.
type GuiUpdateResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// UpdateProgress provides structured progress events for internationalization and native UI display.
type UpdateProgress struct {
	Percentage  int    `json:"percentage"`
	Status      string `json:"status"`
	Stage       string `json:"stage"`
	Detail      string `json:"detail,omitempty"`
	LoadedBytes int64  `json:"loadedBytes,omitempty"`
	TotalBytes  int64  `json:"totalBytes,omitempty"`
}

// ProgressFunc reports update progress.
type ProgressFunc func(p UpdateProgress)

// PendingUpdate holds the staged update state for application restart.
type PendingUpdate struct {
	Shell      string `json:"shell"`
	ScriptPath string `json:"scriptPath"`
	Target     string `json:"target"`
	Staged     string `json:"staged"`
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
	if ctx == nil {
		ctx = context.Background()
	}
	currentTag := env.GitTag

	info, err := updater.FetchLatestReleaseInfo(ctx)
	if err == nil && info != nil {
		latestTag := info.TagName
		hasUpdate := HasNewVersion(currentTag, latestTag)
		var downloadURL string
		if hasUpdate {
			downloadURL = "https://github.com/snowdreamtech/UniBootDesktop/releases/tag/" + latestTag
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

// GetAppBundlePath finds the enclosing .app bundle path if execPath is inside a macOS .app bundle.
// Returns empty string if not inside an .app bundle.
func GetAppBundlePath(execPath string) string {
	idx := strings.Index(execPath, ".app")
	if idx == -1 {
		return ""
	}
	after := execPath[idx+4:]
	if len(after) == 0 || after[0] == '/' || after[0] == filepath.Separator {
		return execPath[:idx+4]
	}
	return ""
}

// FindGuiReleaseAsset locates the appropriate GUI release asset for the given targetOS and targetArch.
func FindGuiReleaseAsset(assets []updater.ReleaseAsset, targetOS, targetArch string) (*updater.ReleaseAsset, error) {
	cleanOS := strings.ToLower(strings.TrimSpace(targetOS))
	cleanArch := strings.ToLower(strings.TrimSpace(targetArch))

	var matched []updater.ReleaseAsset
	for _, a := range assets {
		name := strings.ToLower(a.Name)
		// Skip metadata, checksums, signatures, SBOMs
		if strings.HasSuffix(name, ".sha256") ||
			strings.HasSuffix(name, ".sigstore.json") ||
			strings.HasSuffix(name, ".sbom.spdx.json") ||
			strings.HasSuffix(name, ".sbom.json") ||
			strings.Contains(name, "checksums.txt") {
			continue
		}
		// Skip CLI archives
		if strings.Contains(name, "-cli_") || strings.Contains(name, "-cli-") {
			continue
		}

		switch cleanOS {
		case "darwin":
			// macOS release asset is a DMG image (e.g. UniBootDesktop.dmg)
			if strings.HasSuffix(name, ".dmg") {
				matched = append(matched, a)
			}
		case "windows":
			// Windows GUI assets: portable zip or installer exe
			if strings.Contains(name, "windows") || strings.Contains(name, "win") {
				if strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".exe") {
					matched = append(matched, a)
				}
			}
		case "linux":
			// Linux GUI assets: AppImage, tar.gz, deb, rpm
			if strings.Contains(name, "linux") && (strings.Contains(name, cleanArch) || cleanArch == "") {
				if strings.HasSuffix(name, ".appimage") || strings.HasSuffix(name, ".tar.gz") {
					matched = append(matched, a)
				}
			}
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no GUI release asset found for %s/%s", cleanOS, cleanArch)
	}

	// Prioritization
	for _, a := range matched {
		n := strings.ToLower(a.Name)
		if cleanOS == "windows" && strings.HasSuffix(n, ".zip") {
			return &a, nil
		}
		if cleanOS == "linux" && strings.HasSuffix(n, ".appimage") {
			return &a, nil
		}
	}

	return &matched[0], nil
}

type progressTrackingReader struct {
	reader     io.Reader
	totalBytes int64
	readBytes  int64
	startPct   int
	endPct     int
	lastPct    int
	lastUpdate time.Time
	onProgress ProgressFunc
}

func (pt *progressTrackingReader) Read(p []byte) (int, error) {
	n, err := pt.reader.Read(p)
	if n > 0 {
		pt.readBytes += int64(n)
		if pt.onProgress != nil {
			var currentPct int
			if pt.totalBytes > 0 {
				fraction := float64(pt.readBytes) / float64(pt.totalBytes)
				if fraction > 1.0 {
					fraction = 1.0
				}
				currentPct = pt.startPct + int(fraction*float64(pt.endPct-pt.startPct))
			} else {
				currentPct = pt.startPct
			}

			now := time.Now()
			if currentPct != pt.lastPct && (currentPct-pt.lastPct >= 1 || now.Sub(pt.lastUpdate) >= 200*time.Millisecond) {
				pt.lastPct = currentPct
				pt.lastUpdate = now
				var status string
				if pt.totalBytes > 0 {
					status = fmt.Sprintf("Downloading update: %.1f MB / %.1f MB (%d%%)",
						float64(pt.readBytes)/1048576.0, float64(pt.totalBytes)/1048576.0, currentPct)
				} else {
					status = fmt.Sprintf("Downloading update: %.1f MB", float64(pt.readBytes)/1048576.0)
				}
				pt.onProgress(UpdateProgress{
					Percentage:  currentPct,
					Status:      status,
					Stage:       "downloading",
					LoadedBytes: pt.readBytes,
					TotalBytes:  pt.totalBytes,
				})
			}
		}
	}
	return n, err
}

// DownloadWithProgress downloads a remote URL to destPath with retry and progress streaming.
func DownloadWithProgress(
	ctx context.Context,
	rawURL string,
	destPath string,
	proxyPrefix string,
	onProgress ProgressFunc,
	startPct, endPct int,
) error {
	proxyPrefix = strings.TrimSpace(proxyPrefix)

	client := pkgHttp.NewClientWithTimeout(60 * time.Second)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
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
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create download request: %w", err)
		}
		req.Header.Set("User-Agent", "UniBootDesktop/1.0")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				resp.Body.Close()
				return fmt.Errorf("failed to create target directory: %w", err)
			}

			tmpPath := destPath + ".tmp"
			out, err := os.Create(tmpPath)
			if err != nil {
				resp.Body.Close()
				return fmt.Errorf("failed to create temp destination file: %w", err)
			}

			tracker := &progressTrackingReader{
				reader:     resp.Body,
				totalBytes: resp.ContentLength,
				startPct:   startPct,
				endPct:     endPct,
				onProgress: onProgress,
			}

			copyBuf := make([]byte, 64*1024)
			_, copyErr := io.CopyBuffer(out, tracker, copyBuf)
			resp.Body.Close()
			out.Close()

			if copyErr != nil {
				os.Remove(tmpPath)
				lastErr = fmt.Errorf("failed to save file contents: %w", copyErr)
			} else {
				if err := os.Rename(tmpPath, destPath); err != nil {
					os.Remove(tmpPath)
					return fmt.Errorf("failed to replace destination file: %w", err)
				}
				if onProgress != nil {
					onProgress(UpdateProgress{
						Percentage: endPct,
						Status:     "Download completed",
						Stage:      "download_completed",
					})
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

	if proxyPrefix == "" || strings.EqualFold(proxyPrefix, "direct") {
		return fmt.Errorf("direct GitHub connection failed (%v). Please configure GitHub proxy prefix in Settings and try again", lastErr)
	}

	return fmt.Errorf("download failed (%s): %w", rawURL, lastErr)
}

// DownloadFileWithProxy downloads a remote URL to destPath using optional proxy prefix and retries.
func DownloadFileWithProxy(ctx context.Context, rawURL string, destPath string, proxyPrefix string) error {
	return DownloadWithProgress(ctx, rawURL, destPath, proxyPrefix, nil, 0, 100)
}

// StageMacDmgUpdate mounts the downloaded macOS DMG, extracts the .app bundle to staging, and writes apply_update.sh.
func StageMacDmgUpdate(
	ctx context.Context,
	dmgPath string,
	updatesDir string,
	onProgress ProgressFunc,
) (*PendingUpdate, error) {
	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 75,
			Status:     "Mounting disk image...",
			Stage:      "mounting",
		})
	}

	mountDir, err := os.MkdirTemp("", "unibootdesktop-mount-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary mount point: %w", err)
	}
	defer func() {
		_ = exec.Command("hdiutil", "detach", mountDir, "-force").Run()
		_ = os.RemoveAll(mountDir)
	}()

	cmd := exec.CommandContext(ctx, "hdiutil", "attach", dmgPath, "-mountpoint", mountDir, "-nobrowse", "-readonly", "-noautoopen")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to mount DMG: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	entries, err := os.ReadDir(mountDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read mounted DMG directory: %w", err)
	}

	var appName string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".app") {
			appName = e.Name()
			break
		}
	}
	if appName == "" {
		return nil, fmt.Errorf("no .app bundle found inside DMG")
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 82,
			Status:     fmt.Sprintf("Extracting %s...", appName),
			Stage:      "extracting",
			Detail:     appName,
		})
	}

	stagingDir := filepath.Join(updatesDir, "staging")
	_ = os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create staging directory: %w", err)
	}

	stagedApp := filepath.Join(stagingDir, appName)
	sourceApp := filepath.Join(mountDir, appName)

	cpCmd := exec.CommandContext(ctx, "cp", "-R", sourceApp, stagedApp)
	if out, err := cpCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to copy application bundle: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	// Detach DMG early and delete DMG file to free space
	_ = exec.Command("hdiutil", "detach", mountDir, "-force").Run()
	_ = os.Remove(dmgPath)

	execPath, err := os.Executable()
	if err == nil {
		execPath, _ = filepath.EvalSymlinks(execPath)
	}
	targetBundle := GetAppBundlePath(execPath)
	if targetBundle == "" {
		if _, err := os.Stat("/Applications/" + appName); err == nil {
			targetBundle = "/Applications/" + appName
		} else {
			targetBundle = execPath
		}
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 92,
			Status:     "Preparing update apply script...",
			Stage:      "preparing_script",
		})
	}

	elevatedScriptPath := filepath.Join(updatesDir, "elevated_replace.sh")
	elevatedScriptContent := `#!/bin/bash
TARGET="$1"
STAGED="$2"
rm -rf "$TARGET"
cp -R "$STAGED" "$TARGET"
xattr -dr com.apple.quarantine "$TARGET" 2>/dev/null || true
`
	if err := os.WriteFile(elevatedScriptPath, []byte(elevatedScriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write elevated helper script: %w", err)
	}

	scriptPath := filepath.Join(updatesDir, "apply_update.sh")
	scriptContent := `#!/bin/bash
PID=$1
TARGET="$2"
STAGED="$3"
UPDATE_DIR="$4"

while kill -0 "$PID" 2>/dev/null; do
    sleep 0.1
done
sleep 0.5

if [[ "$TARGET" == *".app"* ]]; then
    if rm -rf "$TARGET" 2>/dev/null && cp -R "$STAGED" "$TARGET" 2>/dev/null; then
        xattr -dr com.apple.quarantine "$TARGET" 2>/dev/null || true
        open "$TARGET"
    else
        osascript -e "do shell script \"/bin/bash \\\"$UPDATE_DIR/elevated_replace.sh\\\" \\\"$TARGET\\\" \\\"$STAGED\\\"\" with administrator privileges" 2>/dev/null || true
        open "$TARGET"
    fi
else
    if [ -f "$STAGED/Contents/MacOS/unibootdesktop" ]; then
        cp -f "$STAGED/Contents/MacOS/unibootdesktop" "$TARGET" 2>/dev/null || true
        chmod +x "$TARGET" 2>/dev/null || true
        "$TARGET" &
    elif [ -f "$STAGED" ]; then
        cp -f "$STAGED" "$TARGET" 2>/dev/null || true
        chmod +x "$TARGET" 2>/dev/null || true
        "$TARGET" &
    fi
fi

rm -rf "$UPDATE_DIR" 2>/dev/null || true
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write update script: %w", err)
	}

	pending := &PendingUpdate{
		Shell:      "/bin/bash",
		ScriptPath: scriptPath,
		Target:     targetBundle,
		Staged:     stagedApp,
	}

	if err := SavePendingUpdate(env.GetDataDir(), pending); err != nil {
		return nil, fmt.Errorf("failed to save pending update metadata: %w", err)
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 100,
			Status:     "Update ready! Restart application to apply.",
			Stage:      "ready",
		})
	}

	return pending, nil
}

// StageWindowsUpdate extracts the Windows update archive into staging and prepares apply_update.bat.
func StageWindowsUpdate(
	ctx context.Context,
	archivePath string,
	updatesDir string,
	onProgress ProgressFunc,
) (*PendingUpdate, error) {
	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 75,
			Status:     "Extracting Windows update package...",
			Stage:      "extracting",
			Detail:     filepath.Base(archivePath),
		})
	}

	stagingDir := filepath.Join(updatesDir, "staging")
	_ = os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create staging directory: %w", err)
	}

	isInstaller := false
	var stagedExe string

	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		if err := archive.ExtractArchiveFromFile(archivePath, stagingDir); err != nil {
			return nil, fmt.Errorf("failed to extract Windows archive: %w", err)
		}
		_ = os.Remove(archivePath)

		stagedExe = filepath.Join(stagingDir, "unibootdesktop.exe")
		if _, err := os.Stat(stagedExe); err != nil {
			// Look for any .exe in stagingDir
			entries, _ := os.ReadDir(stagingDir)
			for _, e := range entries {
				if strings.HasSuffix(strings.ToLower(e.Name()), ".exe") {
					stagedExe = filepath.Join(stagingDir, e.Name())
					break
				}
			}
		}
	} else {
		// Single installer executable: move to staging
		dest := filepath.Join(stagingDir, filepath.Base(archivePath))
		if err := os.Rename(archivePath, dest); err != nil {
			return nil, fmt.Errorf("failed to stage installer executable: %w", err)
		}
		stagedExe = dest
		isInstaller = true
	}

	execPath, err := os.Executable()
	if err == nil {
		execPath, _ = filepath.EvalSymlinks(execPath)
	}

	scriptPath := filepath.Join(updatesDir, "apply_update.bat")
	var scriptContent string
	if isInstaller {
		scriptContent = `@echo off
set PID=%~1
set TARGET=%~2
set STAGED=%~3
set UPDATE_DIR=%~4

:WAIT_LOOP
tasklist /FI "PID eq %PID%" 2>NUL | find /I "%PID%" >NUL
if not errorlevel 1 (
    timeout /T 1 /NOBREAK >NUL
    goto WAIT_LOOP
)

start "" "%STAGED%" /SILENT
rmdir /S /Q "%UPDATE_DIR%" 2>NUL
`
	} else {
		scriptContent = `@echo off
set PID=%~1
set TARGET=%~2
set STAGED=%~3
set UPDATE_DIR=%~4

:WAIT_LOOP
tasklist /FI "PID eq %PID%" 2>NUL | find /I "%PID%" >NUL
if not errorlevel 1 (
    timeout /T 1 /NOBREAK >NUL
    goto WAIT_LOOP
)

copy /Y "%STAGED%" "%TARGET%" >NUL
start "" "%TARGET%"
rmdir /S /Q "%UPDATE_DIR%" 2>NUL
`
	}

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write update batch script: %w", err)
	}

	pending := &PendingUpdate{
		Shell:      "cmd.exe",
		ScriptPath: scriptPath,
		Target:     execPath,
		Staged:     stagedExe,
	}

	if err := SavePendingUpdate(env.GetDataDir(), pending); err != nil {
		return nil, fmt.Errorf("failed to save pending update metadata: %w", err)
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 100,
			Status:     "Update ready! Restart application to apply.",
			Stage:      "ready",
		})
	}

	return pending, nil
}

// StageLinuxUpdate stages an AppImage or tar.gz package on Linux and writes apply_update.sh.
func StageLinuxUpdate(
	ctx context.Context,
	filePath string,
	updatesDir string,
	arch string,
	onProgress ProgressFunc,
) (*PendingUpdate, error) {
	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 75,
			Status:     "Staging Linux update package...",
			Stage:      "staging",
			Detail:     filepath.Base(filePath),
		})
	}

	stagingDir := filepath.Join(updatesDir, "staging")
	_ = os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create staging directory: %w", err)
	}

	var stagedPath string
	lowerPath := strings.ToLower(filePath)

	if strings.HasSuffix(lowerPath, ".tar.gz") || strings.HasSuffix(lowerPath, ".tgz") {
		if err := archive.ExtractArchiveFromFile(filePath, stagingDir); err != nil {
			return nil, fmt.Errorf("failed to extract Linux archive: %w", err)
		}
		_ = os.Remove(filePath)

		stagedBinary := filepath.Join(stagingDir, "unibootdesktop")
		if _, err := os.Stat(stagedBinary); err != nil {
			entries, _ := os.ReadDir(stagingDir)
			for _, e := range entries {
				if !e.IsDir() {
					stagedBinary = filepath.Join(stagingDir, e.Name())
					break
				}
			}
		}
		stagedPath = stagedBinary
	} else {
		// AppImage or standalone binary
		dest := filepath.Join(stagingDir, filepath.Base(filePath))
		if err := os.Rename(filePath, dest); err != nil {
			return nil, fmt.Errorf("failed to stage Linux package: %w", err)
		}
		stagedPath = dest
	}

	_ = os.Chmod(stagedPath, 0755)

	// In AppImage environments, os.Getenv("APPIMAGE") points to the original .AppImage file on disk,
	// whereas os.Executable() points to the extracted /tmp/.mount_*/ binary.
	targetPath := os.Getenv("APPIMAGE")
	if targetPath == "" {
		execPath, err := os.Executable()
		if err == nil {
			execPath, _ = filepath.EvalSymlinks(execPath)
		}
		targetPath = execPath
	}

	scriptPath := filepath.Join(updatesDir, "apply_update.sh")
	scriptContent := `#!/bin/sh
PID=$1
TARGET="$2"
STAGED="$3"
UPDATE_DIR="$4"

while kill -0 "$PID" 2>/dev/null; do
    sleep 0.1
done
sleep 0.5

cp -f "$STAGED" "$TARGET" 2>/dev/null || true
chmod +x "$TARGET" 2>/dev/null || true
"$TARGET" &

rm -rf "$UPDATE_DIR" 2>/dev/null || true
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write update shell script: %w", err)
	}

	pending := &PendingUpdate{
		Shell:      "/bin/sh",
		ScriptPath: scriptPath,
		Target:     targetPath,
		Staged:     stagedPath,
	}

	if err := SavePendingUpdate(env.GetDataDir(), pending); err != nil {
		return nil, fmt.Errorf("failed to save pending update metadata: %w", err)
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 100,
			Status:     "Update ready! Restart application to apply.",
			Stage:      "ready",
		})
	}

	return pending, nil
}

// VerifyFileSHA256 computes and verifies the SHA-256 checksum of filePath against expectedHash.
func VerifyFileSHA256(filePath string, expectedHash string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum verification: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	buf := make([]byte, 64*1024)
	if _, err := io.CopyBuffer(hasher, f, buf); err != nil {
		return fmt.Errorf("failed to compute file hash: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actualHash, strings.TrimSpace(expectedHash)) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}
	return nil
}

// PerformGuiUpdate executes end-to-end download, staging, and preparation for atomic restart replacement.
func PerformGuiUpdate(
	ctx context.Context,
	proxyPrefix string,
	onProgress ProgressFunc,
) (*GuiUpdateResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 5,
			Status:     "Checking for latest release...",
			Stage:      "checking",
		})
	}

	info, err := updater.FetchLatestReleaseInfo(ctx)
	if err != nil {
		return &GuiUpdateResult{Success: false, Message: fmt.Sprintf("Failed to query GitHub release: %v", err)}, err
	}

	asset, err := FindGuiReleaseAsset(info.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return &GuiUpdateResult{Success: false, Message: err.Error()}, err
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Percentage: 10,
			Status:     fmt.Sprintf("Found release asset: %s", asset.Name),
			Stage:      "found_asset",
			Detail:     asset.Name,
		})
	}

	updatesDir := filepath.Join(env.GetDataDir(), "updates")
	downloadDir := filepath.Join(updatesDir, "download")
	destFile := filepath.Join(downloadDir, asset.Name)

	if err := DownloadWithProgress(ctx, asset.BrowserDownloadURL, destFile, proxyPrefix, onProgress, 10, 70); err != nil {
		return &GuiUpdateResult{Success: false, Message: fmt.Sprintf("Download failed: %v", err)}, err
	}

	// Verify SHA-256 checksum if checksums file exists in release assets
	var checksumsAsset *updater.ReleaseAsset
	for _, a := range info.Assets {
		if strings.EqualFold(a.Name, "checksums.txt") || strings.HasSuffix(strings.ToLower(a.Name), "_checksums.txt") {
			checksumsAsset = &a
			break
		}
	}

	if checksumsAsset != nil {
		if onProgress != nil {
			onProgress(UpdateProgress{
				Percentage: 72,
				Status:     "Verifying download checksum...",
				Stage:      "verifying_checksum",
				Detail:     checksumsAsset.Name,
			})
		}
		checksumPath := filepath.Join(downloadDir, checksumsAsset.Name)
		if err := DownloadFileWithProxy(ctx, checksumsAsset.BrowserDownloadURL, checksumPath, proxyPrefix); err == nil {
			chkData, readErr := os.ReadFile(checksumPath)
			if readErr == nil {
				var expectedHash string
				for _, line := range strings.Split(string(chkData), "\n") {
					parts := strings.Fields(line)
					if len(parts) >= 2 && parts[1] == asset.Name {
						expectedHash = parts[0]
						break
					}
				}
				if expectedHash != "" {
					if err := VerifyFileSHA256(destFile, expectedHash); err != nil {
						_ = os.Remove(destFile)
						return &GuiUpdateResult{
							Success: false,
							Message: fmt.Sprintf("Integrity verification failed: %v", err),
						}, err
					}
				}
			}
		}
	}

	var stageErr error
	switch runtime.GOOS {
	case "darwin":
		_, stageErr = StageMacDmgUpdate(ctx, destFile, updatesDir, onProgress)
	case "windows":
		_, stageErr = StageWindowsUpdate(ctx, destFile, updatesDir, onProgress)
	case "linux":
		_, stageErr = StageLinuxUpdate(ctx, destFile, updatesDir, runtime.GOARCH, onProgress)
	default:
		stageErr = fmt.Errorf("unsupported operating system for automatic update: %s", runtime.GOOS)
	}

	if stageErr != nil {
		return &GuiUpdateResult{Success: false, Message: fmt.Sprintf("Failed to stage update: %v", stageErr)}, stageErr
	}

	// Invalidate update cache so subsequent checks reflect current
	_ = updater.ClearCache()

	return &GuiUpdateResult{
		Success: true,
		Message: "Update downloaded and prepared successfully. Restart to apply.",
	}, nil
}

// GetPendingUpdate reads the pending update metadata from disk if it exists.
func GetPendingUpdate(dataDir string) (*PendingUpdate, error) {
	p := filepath.Join(dataDir, "updates", "pending_update.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var pending PendingUpdate
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, err
	}
	return &pending, nil
}

// SavePendingUpdate saves the pending update metadata to disk.
func SavePendingUpdate(dataDir string, pending *PendingUpdate) error {
	dir := filepath.Join(dataDir, "updates")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	p := filepath.Join(dir, "pending_update.json")
	data, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// ClearPendingUpdate removes pending update metadata and directory.
func ClearPendingUpdate(dataDir string) error {
	dir := filepath.Join(dataDir, "updates")
	return os.RemoveAll(dir)
}
