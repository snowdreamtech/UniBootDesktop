// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package utils

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateFileChecksum(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.iso")

	content := []byte("Hello UniBoot ISO Checksum Test Data")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx := context.Background()

	// Test SHA256
	shaRes, err := CalculateFileChecksum(ctx, testFile, "sha256")
	if err != nil {
		t.Fatalf("CalculateFileChecksum sha256 failed: %v", err)
	}
	if shaRes.Hash == "" {
		t.Errorf("expected non-empty sha256 hash")
	}
	if shaRes.Algorithm != "sha256" {
		t.Errorf("expected algo sha256, got %s", shaRes.Algorithm)
	}

	// Test MD5
	md5Res, err := CalculateFileChecksum(ctx, testFile, "md5")
	if err != nil {
		t.Fatalf("CalculateFileChecksum md5 failed: %v", err)
	}
	if md5Res.Hash == "" {
		t.Errorf("expected non-empty md5 hash")
	}
	if md5Res.Algorithm != "md5" {
		t.Errorf("expected algo md5, got %s", md5Res.Algorithm)
	}

	// Test SHA512
	sha512Res, err := CalculateFileChecksum(ctx, testFile, "sha512")
	if err != nil {
		t.Fatalf("CalculateFileChecksum sha512 failed: %v", err)
	}
	if sha512Res.Hash == "" {
		t.Errorf("expected non-empty sha512 hash")
	}

	// Test unsupported algorithm
	_, err = CalculateFileChecksum(ctx, testFile, "unknown_algo")
	if err == nil {
		t.Errorf("expected error for unsupported algo, got nil")
	}
}
