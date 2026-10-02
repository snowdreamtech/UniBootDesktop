// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/snowdreamtech/unibootdesktop/internal/env"
	pkgHttp "github.com/snowdreamtech/unibootdesktop/internal/http"
	"github.com/snowdreamtech/unibootdesktop/pkg/updater"
)

// VentoyReleaseAsset represents an asset attached to a Ventoy GitHub release.
type VentoyReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// VentoyReleaseInfo holds metadata regarding the latest Ventoy release and local status.
type VentoyReleaseInfo struct {
	TagName       string               `json:"tagName"`
	Name          string               `json:"name"`
	PublishedAt   string               `json:"publishedAt"`
	Body          string               `json:"body"`
	Assets        []VentoyReleaseAsset `json:"assets"`
	LocalVersion  string               `json:"localVersion"`
	HasUpdate     bool                 `json:"hasUpdate"`
	IsSupportedOS bool                 `json:"isSupportedOS"`
	TargetOS      string               `json:"targetOS"`
	TargetAsset   string               `json:"targetAsset"`
	InstallDir    string               `json:"installDir"`
	Message       string               `json:"message"`
}

// GetEffectiveVentoyDir returns the designated directory for Ventoy installation.
// If customized path is provided and non-empty, it returns that path.
// Otherwise, it returns the standard application Ventoy directory from env.GetVentoyDir().
func GetEffectiveVentoyDir(customPath string) string {
	clean := strings.TrimSpace(customPath)
	if clean != "" {
		return clean
	}
	return env.GetVentoyDir()
}

// ReadLocalVentoyVersion attempts to retrieve the installed Ventoy version.
// It first inspects version.json in targetDir, then falls back to ValidateVentoyCli.
func ReadLocalVentoyVersion(targetDir string) string {
	cleanDir := GetEffectiveVentoyDir(targetDir)
	if cleanDir == "" {
		return ""
	}

	versionFile := filepath.Join(cleanDir, "version.json")
	if data, err := os.ReadFile(versionFile); err == nil {
		var ver struct {
			TagName string `json:"tagName"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &ver); err == nil {
			if ver.TagName != "" {
				return ver.TagName
			}
			if ver.Version != "" {
				return "v" + strings.TrimPrefix(ver.Version, "v")
			}
		}
	}

	val := ValidateVentoyCli(cleanDir)
	if val != nil && val.Valid && val.Version != "" {
		if !strings.HasPrefix(val.Version, "v") && !strings.HasPrefix(val.Version, "V") {
			return "v" + val.Version
		}
		return val.Version
	}

	return ""
}

// FetchLatestVentoyRelease retrieves release metadata from official Ventoy GitHub releases.
func FetchLatestVentoyRelease(ctx context.Context, proxyPrefix string, localVentoyDir string) (*VentoyReleaseInfo, error) {
	targetOS := runtime.GOOS
	effectiveDir := GetEffectiveVentoyDir(localVentoyDir)
	localVer := ReadLocalVentoyVersion(effectiveDir)

	info := &VentoyReleaseInfo{
		TargetOS:      targetOS,
		InstallDir:    effectiveDir,
		LocalVersion:  localVer,
		IsSupportedOS: targetOS != "darwin",
	}

	if targetOS == "darwin" {
		info.Message = "macOS does not have official native Ventoy CLI formatting binaries. Please use Cloud Mode or prepare Ventoy on Windows/Linux."
		return info, nil
	}

	apiURL := "https://api.github.com/repos/ventoy/Ventoy/releases/latest"
	finalURL := updater.BuildProxyURL(apiURL, proxyPrefix)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ventoy release request: %w", err)
	}
	req.Header.Set("User-Agent", "UniBootDesktop-VentoyDetector/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := pkgHttp.NewClientWithTimeout(15 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Ventoy release from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d for Ventoy release", resp.StatusCode)
	}

	var raw struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
		Body        string `json:"body"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode Ventoy release JSON: %w", err)
	}

	info.TagName = raw.TagName
	info.Name = raw.Name
	info.PublishedAt = raw.PublishedAt
	info.Body = raw.Body

	for _, a := range raw.Assets {
		info.Assets = append(info.Assets, VentoyReleaseAsset{
			Name:               a.Name,
			BrowserDownloadURL: a.BrowserDownloadURL,
			Size:               a.Size,
		})
	}

	// Match required asset for current OS
	info.TargetAsset = selectVentoyAssetForOS(targetOS, info.Assets)
	if localVer == "" {
		info.HasUpdate = true
	} else {
		cleanLocal := strings.TrimPrefix(strings.TrimPrefix(localVer, "v"), "V")
		cleanRemote := strings.TrimPrefix(strings.TrimPrefix(raw.TagName, "v"), "V")
		info.HasUpdate = cleanLocal != cleanRemote
	}

	return info, nil
}

func selectVentoyAssetForOS(targetOS string, assets []VentoyReleaseAsset) string {
	switch targetOS {
	case "windows":
		for _, a := range assets {
			if strings.HasSuffix(a.Name, "-windows.zip") {
				return a.Name
			}
		}
	case "linux":
		for _, a := range assets {
			if strings.HasSuffix(a.Name, "-linux.tar.gz") {
				return a.Name
			}
		}
	}
	return ""
}

// DownloadAndExtractVentoy downloads and extracts the official Ventoy toolchain for the current platform
// into the strictly designated target directory.
func DownloadAndExtractVentoy(ctx context.Context, proxyPrefix string, targetDir string) (*VentoyReleaseInfo, error) {
	if runtime.GOOS == "darwin" {
		return nil, fmt.Errorf("official Ventoy CLI is not supported on macOS")
	}

	destDir := GetEffectiveVentoyDir(targetDir)
	if destDir == "" {
		return nil, fmt.Errorf("ventoy target directory cannot be empty")
	}

	rel, err := FetchLatestVentoyRelease(ctx, proxyPrefix, destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect latest Ventoy release: %w", err)
	}

	if rel.TargetAsset == "" {
		return nil, fmt.Errorf("no matching Ventoy release asset found for %s in release %s", runtime.GOOS, rel.TagName)
	}

	var targetAsset *VentoyReleaseAsset
	var sha256Asset *VentoyReleaseAsset
	for i := range rel.Assets {
		if rel.Assets[i].Name == rel.TargetAsset {
			targetAsset = &rel.Assets[i]
		}
		if rel.Assets[i].Name == "sha256.txt" {
			sha256Asset = &rel.Assets[i]
		}
	}

	if targetAsset == nil {
		return nil, fmt.Errorf("asset %s not found in release", rel.TargetAsset)
	}

	// Strictly store temporary files in designated cache directory
	cacheDir := env.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to prepare cache directory: %w", err)
	}

	archivePath := filepath.Join(cacheDir, targetAsset.Name)
	defer os.Remove(archivePath)

	if err := updater.DownloadFileWithProxy(ctx, targetAsset.BrowserDownloadURL, archivePath, proxyPrefix); err != nil {
		return nil, fmt.Errorf("failed to download Ventoy asset %s: %w", targetAsset.Name, err)
	}

	// Validate SHA256 if sha256.txt is available
	if sha256Asset != nil {
		sha256Path := filepath.Join(cacheDir, "ventoy_sha256.txt")
		if err := updater.DownloadFileWithProxy(ctx, sha256Asset.BrowserDownloadURL, sha256Path, proxyPrefix); err == nil {
			defer os.Remove(sha256Path)
			if checksumMap, err := parseVentoySha256File(sha256Path); err == nil {
				if expected, ok := checksumMap[targetAsset.Name]; ok {
					if err := verifyFileSha256(archivePath, expected); err != nil {
						return nil, fmt.Errorf("ventoy asset checksum verification failed: %w", err)
					}
				}
			}
		}
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create Ventoy target directory %s: %w", destDir, err)
	}

	// Extract with root directory flattening
	if strings.HasSuffix(targetAsset.Name, ".zip") {
		if err := extractZipStripRoot(archivePath, destDir); err != nil {
			return nil, fmt.Errorf("failed to extract Ventoy zip archive: %w", err)
		}
	} else if strings.HasSuffix(targetAsset.Name, ".tar.gz") {
		if err := extractTarGzStripRoot(archivePath, destDir); err != nil {
			return nil, fmt.Errorf("failed to extract Ventoy tar.gz archive: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported Ventoy archive format: %s", targetAsset.Name)
	}

	// Fix Linux/Unix executable permissions
	if runtime.GOOS != "windows" {
		_ = fixVentoyPermissions(destDir)
	}

	// Persist version record
	versionRecord := map[string]interface{}{
		"tagName":     rel.TagName,
		"assetName":   targetAsset.Name,
		"installedAt": time.Now().Format(time.RFC3339),
		"os":          runtime.GOOS,
	}
	if vData, err := json.MarshalIndent(versionRecord, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(destDir, "version.json"), vData, 0644)
	}

	rel.LocalVersion = rel.TagName
	rel.HasUpdate = false

	return rel, nil
}

func parseVentoySha256File(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := strings.ToLower(parts[0])
			fileName := strings.TrimPrefix(parts[1], "*")
			result[filepath.Base(fileName)] = hash
		}
	}
	return result, scanner.Err()
}

func verifyFileSha256(filePath string, expectedHash string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if strings.ToLower(actual) != strings.ToLower(expectedHash) {
		return fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedHash, actual)
	}
	return nil
}

func extractZipStripRoot(zipPath string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	rootPrefix := detectZipCommonPrefix(r.File)

	for _, f := range r.File {
		relPath := f.Name
		if rootPrefix != "" {
			relPath = strings.TrimPrefix(relPath, rootPrefix)
			relPath = strings.TrimPrefix(relPath, "/")
		}
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(destDir, filepath.FromSlash(relPath))
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal zip path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

func detectZipCommonPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}
	var firstSegment string
	for _, f := range files {
		clean := strings.Trim(filepath.ToSlash(f.Name), "/")
		parts := strings.Split(clean, "/")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		if firstSegment == "" {
			firstSegment = parts[0]
		} else if firstSegment != parts[0] {
			return ""
		}
	}
	return firstSegment
}

func extractTarGzStripRoot(tarGzPath string, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Since tar stream cannot easily be rewound without reading everything,
	// we identify the root prefix by inspecting the first entry or common patterns.
	var rootPrefix string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cleanHeaderName := strings.Trim(filepath.ToSlash(header.Name), "/")
		parts := strings.Split(cleanHeaderName, "/")
		if rootPrefix == "" && len(parts) > 1 && strings.HasPrefix(parts[0], "ventoy-") {
			rootPrefix = parts[0]
		}

		relPath := cleanHeaderName
		if rootPrefix != "" && strings.HasPrefix(relPath, rootPrefix) {
			relPath = strings.TrimPrefix(relPath, rootPrefix)
			relPath = strings.TrimPrefix(relPath, "/")
		}
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(destDir, filepath.FromSlash(relPath))
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal tar path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			mode := os.FileMode(header.Mode)
			if mode == 0 {
				mode = 0644
			}
			outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func fixVentoyPermissions(destDir string) error {
	return filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			name := strings.ToLower(info.Name())
			if strings.HasSuffix(name, ".sh") || strings.Contains(path, filepath.Join("tool", "")) {
				_ = os.Chmod(path, 0755)
			}
		}
		return nil
	})
}
