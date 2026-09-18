// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package utils

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
	"time"
)

// ChecksumResult represents the calculated file hash metadata.
type ChecksumResult struct {
	FilePath   string `json:"filePath"`
	Algorithm  string `json:"algorithm"`
	Hash       string `json:"hash"`
	FileSize   int64  `json:"fileSize"`
	DurationMs int64  `json:"durationMs"`
}

// CalculateFileChecksum calculates the hash string for a file given the algorithm choice (md5, sha256, sha512).
func CalculateFileChecksum(ctx context.Context, filePath string, algo string) (*ChecksumResult, error) {
	startTime := time.Now()
	cleanAlgo := strings.ToLower(strings.TrimSpace(algo))
	if cleanAlgo == "" {
		cleanAlgo = "sha256"
	}

	var hasher hash.Hash
	switch cleanAlgo {
	case "md5":
		hasher = md5.New()
	case "sha256", "sha-256":
		hasher = sha256.New()
		cleanAlgo = "sha256"
	case "sha512", "sha-512":
		hasher = sha512.New()
		cleanAlgo = "sha512"
	default:
		return nil, fmt.Errorf("unsupported checksum algorithm: %s (supported: md5, sha256, sha512)", algo)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file for checksum: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, readErr := file.Read(buf)
		if n > 0 {
			hasher.Write(buf[:n])
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, fmt.Errorf("read chunk for checksum: %w", readErr)
		}
	}

	hashStr := hex.EncodeToString(hasher.Sum(nil))
	durationMs := time.Since(startTime).Milliseconds()

	return &ChecksumResult{
		FilePath:   filePath,
		Algorithm:  cleanAlgo,
		Hash:       hashStr,
		FileSize:   stat.Size(),
		DurationMs: durationMs,
	}, nil
}
