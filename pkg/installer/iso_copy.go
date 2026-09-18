// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package installer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IsoCopyProgress holds real-time progress details for copying system images to U-disk.
type IsoCopyProgress struct {
	CurrentFile   string  `json:"currentFile"`
	FileIndex     int     `json:"fileIndex"`
	TotalFiles    int     `json:"totalFiles"`
	CopiedBytes   int64   `json:"copiedBytes"`
	FileSizeBytes int64   `json:"fileSizeBytes"`
	Progress      float64 `json:"progress"` // 0.0 to 100.0
	SpeedMBps     float64 `json:"speedMBps"`
	ElapsedSec    int64   `json:"elapsedSec"`
	EtaSec        int64   `json:"etaSec"`
}

// CopyIsoProgressCallback defines the function signature for reporting ISO copy progress.
type CopyIsoProgressCallback func(progress IsoCopyProgress)

// CopyIsoFilesToDisk copies selected local ISO/IMG files into <mountPoint>/iso/ directory on target drive.
func CopyIsoFilesToDisk(mountPoint string, isoPaths []string, progressCb CopyIsoProgressCallback) error {
	return CopyIsoFilesToDiskWithContext(context.Background(), mountPoint, isoPaths, progressCb)
}

// CopyIsoFilesToDiskWithContext copies selected local ISO/IMG files with cancellation context support and live speed/ETA tracking.
func CopyIsoFilesToDiskWithContext(ctx context.Context, mountPoint string, isoPaths []string, progressCb CopyIsoProgressCallback) error {
	if len(isoPaths) == 0 {
		return nil
	}

	if mountPoint == "" {
		return fmt.Errorf("mount point cannot be empty")
	}

	targetIsoDir := filepath.Join(mountPoint, "iso")
	if err := os.MkdirAll(targetIsoDir, 0755); err != nil {
		return fmt.Errorf("failed to create iso directory at %s: %w", targetIsoDir, err)
	}

	buffer := make([]byte, 1024*1024) // 1MB buffer for high throughput U-disk write

	for idx, srcPath := range isoPaths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("failed to stat source image file %s: %w", srcPath, err)
		}

		if info.IsDir() {
			continue
		}

		fileName := filepath.Base(srcPath)
		destPath := filepath.Join(targetIsoDir, fileName)

		// Open source file
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open source image %s: %w", srcPath, err)
		}

		// Create destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create target image file %s: %w", destPath, err)
		}

		totalSize := info.Size()
		var copiedTotal int64
		startTime := time.Now()

		for {
			select {
			case <-ctx.Done():
				srcFile.Close()
				destFile.Close()
				os.Remove(destPath)
				return ctx.Err()
			default:
			}

			n, readErr := srcFile.Read(buffer)
			if n > 0 {
				written, writeErr := destFile.Write(buffer[:n])
				if writeErr != nil {
					srcFile.Close()
					destFile.Close()
					return fmt.Errorf("error writing to %s: %w", destPath, writeErr)
				}
				copiedTotal += int64(written)

				if progressCb != nil && totalSize > 0 {
					pct := (float64(copiedTotal) / float64(totalSize)) * 100.0
					elapsedDuration := time.Since(startTime)
					elapsedSec := int64(elapsedDuration.Seconds())
					
					speedMBps := 0.0
					etaSec := int64(0)
					if elapsedDuration.Seconds() > 0.1 {
						speedMBps = (float64(copiedTotal) / (1024 * 1024)) / elapsedDuration.Seconds()
						if speedMBps > 0 {
							remainingBytes := totalSize - copiedTotal
							etaSec = int64((float64(remainingBytes) / (1024 * 1024)) / speedMBps)
						}
					}

					progressCb(IsoCopyProgress{
						CurrentFile:   fileName,
						FileIndex:     idx + 1,
						TotalFiles:    len(isoPaths),
						CopiedBytes:   copiedTotal,
						FileSizeBytes: totalSize,
						Progress:      pct,
						SpeedMBps:     speedMBps,
						ElapsedSec:    elapsedSec,
						EtaSec:        etaSec,
					})
				}
			}

			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				srcFile.Close()
				destFile.Close()
				return fmt.Errorf("error reading from %s: %w", srcPath, readErr)
			}
		}

		srcFile.Close()
		destFile.Sync()
		destFile.Close()
	}

	return nil
}

// ValidateIsoPath checks if a given file path is a valid supported system image file by Ventoy (.iso, .wim, .img, .vhd, .vhdx, .vti, .efi, .bin, .xz, .gz, .raw).
func ValidateIsoPath(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".iso", ".wim", ".img", ".vhd", ".vhdx", ".vti", ".efi", ".bin", ".xz", ".gz", ".raw":
		return true
	default:
		return false
	}
}
