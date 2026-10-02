// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/snowdreamtech/unibootdesktop/internal/env"
)

func TestGetEffectiveVentoyDir(t *testing.T) {
	defaultDir := env.GetVentoyDir()
	if got := GetEffectiveVentoyDir(""); got != defaultDir {
		t.Fatalf("expected default %q, got %q", defaultDir, got)
	}

	custom := filepath.Join(os.TempDir(), "custom-ventoy-dir")
	if got := GetEffectiveVentoyDir(custom); got != custom {
		t.Fatalf("expected custom %q, got %q", custom, got)
	}
}

func TestSelectVentoyAssetForOS(t *testing.T) {
	assets := []VentoyReleaseAsset{
		{Name: "sha256.txt"},
		{Name: "ventoy-1.1.17-linux.tar.gz"},
		{Name: "ventoy-1.1.17-livecd.iso"},
		{Name: "ventoy-1.1.17-windows.zip"},
	}

	winAsset := selectVentoyAssetForOS("windows", assets)
	if winAsset != "ventoy-1.1.17-windows.zip" {
		t.Errorf("expected ventoy-1.1.17-windows.zip for windows, got %s", winAsset)
	}

	linuxAsset := selectVentoyAssetForOS("linux", assets)
	if linuxAsset != "ventoy-1.1.17-linux.tar.gz" {
		t.Errorf("expected ventoy-1.1.17-linux.tar.gz for linux, got %s", linuxAsset)
	}

	darwinAsset := selectVentoyAssetForOS("darwin", assets)
	if darwinAsset != "" {
		t.Errorf("expected empty asset for darwin, got %s", darwinAsset)
	}
}

func TestExtractZipStripRoot(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test-ventoy.zip")
	destDir := filepath.Join(tempDir, "extracted")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	files := map[string]string{
		"ventoy-1.1.17/Ventoy2Disk.exe": "mock-windows-binary",
		"ventoy-1.1.17/ventoy/boot.img": "mock-boot-image",
	}

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zw.Create error: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("w.Write error: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zw.Close error: %v", err)
	}

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write zip file error: %v", err)
	}

	if err := extractZipStripRoot(zipPath, destDir); err != nil {
		t.Fatalf("extractZipStripRoot failed: %v", err)
	}

	exePath := filepath.Join(destDir, "Ventoy2Disk.exe")
	if data, err := os.ReadFile(exePath); err != nil || string(data) != "mock-windows-binary" {
		t.Errorf("expected extracted Ventoy2Disk.exe with correct content, err: %v", err)
	}

	bootPath := filepath.Join(destDir, "ventoy", "boot.img")
	if data, err := os.ReadFile(bootPath); err != nil || string(data) != "mock-boot-image" {
		t.Errorf("expected extracted ventoy/boot.img with correct content, err: %v", err)
	}
}

func TestExtractTarGzStripRoot(t *testing.T) {
	tempDir := t.TempDir()
	tarPath := filepath.Join(tempDir, "test-ventoy.tar.gz")
	destDir := filepath.Join(tempDir, "extracted-tar")

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	files := map[string]string{
		"ventoy-1.1.17/Ventoy2Disk.sh": "echo 'mock-linux-script'",
		"ventoy-1.1.17/tool/worker":    "mock-tool-binary",
	}

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tw.WriteHeader error: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("tw.Write error: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tw.Close error: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("gzw.Close error: %v", err)
	}

	if err := os.WriteFile(tarPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write tar.gz file error: %v", err)
	}

	if err := extractTarGzStripRoot(tarPath, destDir); err != nil {
		t.Fatalf("extractTarGzStripRoot failed: %v", err)
	}

	shPath := filepath.Join(destDir, "Ventoy2Disk.sh")
	if data, err := os.ReadFile(shPath); err != nil || string(data) != "echo 'mock-linux-script'" {
		t.Errorf("expected extracted Ventoy2Disk.sh with correct content, err: %v", err)
	}
}

func TestParseVentoySha256File(t *testing.T) {
	tempDir := t.TempDir()
	shaPath := filepath.Join(tempDir, "sha256.txt")

	content := `
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 *ventoy-1.1.17-windows.zip
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad  ventoy-1.1.17-linux.tar.gz
`
	if err := os.WriteFile(shaPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write sha256 file: %v", err)
	}

	m, err := parseVentoySha256File(shaPath)
	if err != nil {
		t.Fatalf("parseVentoySha256File error: %v", err)
	}

	if m["ventoy-1.1.17-windows.zip"] != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("unexpected hash for windows zip: %v", m["ventoy-1.1.17-windows.zip"])
	}
	if m["ventoy-1.1.17-linux.tar.gz"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("unexpected hash for linux tar.gz: %v", m["ventoy-1.1.17-linux.tar.gz"])
	}
}

func TestVerifyFileSha256(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.dat")
	data := []byte("hello ventoy sha256 test")
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	sum := sha256.Sum256(data)
	expectedHex := hex.EncodeToString(sum[:])

	if err := verifyFileSha256(testFile, expectedHex); err != nil {
		t.Errorf("verifyFileSha256 failed on valid hash: %v", err)
	}

	if err := verifyFileSha256(testFile, "0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Errorf("verifyFileSha256 should fail on invalid hash")
	}
}

func TestReadLocalVentoyVersion(t *testing.T) {
	tempDir := t.TempDir()
	versionFile := filepath.Join(tempDir, "version.json")
	if err := os.WriteFile(versionFile, []byte(`{"tagName": "v1.1.17"}`), 0644); err != nil {
		t.Fatalf("write version.json error: %v", err)
	}

	ver := ReadLocalVentoyVersion(tempDir)
	if ver != "v1.1.17" {
		t.Errorf("expected v1.1.17, got %q", ver)
	}
}

func TestFetchLatestVentoyReleaseOnDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("only run on darwin")
	}

	info, err := FetchLatestVentoyRelease(t.Context(), "", "")
	if err != nil {
		t.Fatalf("unexpected error on darwin: %v", err)
	}
	if info.IsSupportedOS {
		t.Errorf("expected IsSupportedOS to be false on darwin")
	}
	if info.TargetAsset != "" {
		t.Errorf("expected empty target asset on darwin")
	}
}
