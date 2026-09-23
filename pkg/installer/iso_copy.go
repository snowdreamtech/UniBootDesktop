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

// IsoCopyPlanEntry describes one user-approved image copy decision.
type IsoCopyPlanEntry struct {
	SourcePath string `json:"sourcePath"`
	TargetName string `json:"targetName"`
	Action     string `json:"action"` // replace, skip, or rename
}

// IsoCopyConflict describes an existing target filename found during deployment preflight.
type IsoCopyConflict struct {
	TargetDisk    string `json:"targetDisk"`
	SourcePath    string `json:"sourcePath"`
	FileName      string `json:"fileName"`
	SuggestedName string `json:"suggestedName"`
	ConflictType  string `json:"conflictType"` // source_duplicate, target_exists, or source_duplicate_target_exists
}

// IsoCopyDiskPlan contains approved image copy decisions for one target disk.
type IsoCopyDiskPlan struct {
	TargetDisk string             `json:"targetDisk"`
	Entries    []IsoCopyPlanEntry `json:"entries"`
}

// CopyIsoFilesToDisk copies selected local ISO/IMG files into <mountPoint>/iso/ directory on target drive.
func CopyIsoFilesToDisk(mountPoint string, isoPaths []string, progressCb CopyIsoProgressCallback) error {
	return CopyIsoFilesToDiskWithContext(context.Background(), mountPoint, isoPaths, progressCb)
}

// CopyIsoFilesToDiskWithContext copies selected local ISO/IMG files with cancellation context support and live speed/ETA tracking.
func CopyIsoFilesToDiskWithContext(ctx context.Context, mountPoint string, isoPaths []string, progressCb CopyIsoProgressCallback) error {
	entries := make([]IsoCopyPlanEntry, 0, len(isoPaths))
	for _, sourcePath := range isoPaths {
		entries = append(entries, IsoCopyPlanEntry{SourcePath: sourcePath, Action: "rename"})
	}
	return CopyIsoFilesToDiskWithPlan(ctx, mountPoint, entries, progressCb)
}

// CopyIsoFilesToDiskWithPlan executes a pre-approved image copy plan without prompting during writes.
func CopyIsoFilesToDiskWithPlan(ctx context.Context, mountPoint string, entries []IsoCopyPlanEntry, progressCb CopyIsoProgressCallback) error {
	if len(entries) == 0 {
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

	for idx, entry := range entries {
		srcPath := entry.SourcePath
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
		targetName := filepath.Base(entry.TargetName)
		if targetName == "." || targetName == string(filepath.Separator) || targetName == "" {
			targetName = fileName
		}
		if entry.Action == "skip" {
			continue
		}
		var destPath string
		if entry.Action == "replace" {
			destPath = filepath.Join(targetIsoDir, targetName)
		} else if entry.TargetName != "" {
			destPath = filepath.Join(targetIsoDir, targetName)
		} else {
			destPath, err = reserveIsoDestination(targetIsoDir, targetName)
			if err != nil {
				return fmt.Errorf("failed to reserve target image file for %s: %w", srcPath, err)
			}
		}

		// Open source file
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open source image %s: %w", srcPath, err)
		}

		// Create destination file
		flags := os.O_WRONLY | os.O_CREATE
		if entry.Action == "replace" {
			flags |= os.O_TRUNC
		} else {
			flags |= os.O_EXCL
		}
		destFile, err := os.OpenFile(destPath, flags, info.Mode().Perm())
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create target image file %s: %w", destPath, err)
		}

		totalSize := info.Size()
		var copiedTotal int64
		startTime := time.Now()

		windowStartTime := startTime
		windowStartBytes := int64(0)
		currentInstantSpeed := 0.0
		currentEtaSec := int64(0)

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

					// Update instantaneous speed using a 500ms sliding window
					windowDuration := time.Since(windowStartTime)
					if windowDuration >= 500*time.Millisecond {
						windowBytes := copiedTotal - windowStartBytes
						instantSpeed := (float64(windowBytes) / (1024 * 1024)) / windowDuration.Seconds()

						// Lightly smooth with previous reading to prevent wild spikes while keeping it responsive
						if currentInstantSpeed > 0 {
							currentInstantSpeed = 0.7*instantSpeed + 0.3*currentInstantSpeed
						} else {
							currentInstantSpeed = instantSpeed
						}

						if currentInstantSpeed > 0 {
							remainingBytes := totalSize - copiedTotal
							currentEtaSec = int64((float64(remainingBytes) / (1024 * 1024)) / currentInstantSpeed)
						}

						windowStartTime = time.Now()
						windowStartBytes = copiedTotal
					} else if currentInstantSpeed == 0 && elapsedDuration.Seconds() > 0.1 {
						// Initial fallback before first 500ms window completes
						currentInstantSpeed = (float64(copiedTotal) / (1024 * 1024)) / elapsedDuration.Seconds()
						if currentInstantSpeed > 0 {
							remainingBytes := totalSize - copiedTotal
							currentEtaSec = int64((float64(remainingBytes) / (1024 * 1024)) / currentInstantSpeed)
						}
					}

					progressCb(IsoCopyProgress{
						CurrentFile:   fileName,
						FileIndex:     idx + 1,
						TotalFiles:    len(entries),
						CopiedBytes:   copiedTotal,
						FileSizeBytes: totalSize,
						Progress:      pct,
						SpeedMBps:     currentInstantSpeed,
						ElapsedSec:    elapsedSec,
						EtaSec:        currentEtaSec,
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

// FindIsoCopyConflictsInDirectory returns filename conflicts without reading file contents.
func FindIsoCopyConflictsInDirectory(targetDisk string, targetIsoDir string, isoPaths []string) ([]IsoCopyConflict, error) {
	entries, err := os.ReadDir(targetIsoDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read target iso directory %s: %w", targetIsoDir, err)
	}
	existing := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			existing[entry.Name()] = struct{}{}
		}
	}
	targetFiles := make(map[string]struct{}, len(existing))
	for fileName := range existing {
		targetFiles[fileName] = struct{}{}
	}

	conflicts := make([]IsoCopyConflict, 0)
	sourceNames := make(map[string]int, len(isoPaths))
	for _, sourcePath := range isoPaths {
		sourceNames[filepath.Base(sourcePath)]++
	}
	for _, sourcePath := range isoPaths {
		fileName := filepath.Base(sourcePath)
		_, targetExists := targetFiles[fileName]
		if !targetExists && sourceNames[fileName] < 2 {
			continue
		}
		if sourceNames[fileName] > 1 {
			existing[fileName] = struct{}{}
		}
		suggestedName := fileName
		for suffix := 1; ; suffix++ {
			if _, ok := existing[suggestedName]; !ok {
				break
			}
			extension := filepath.Ext(fileName)
			baseName := strings.TrimSuffix(fileName, extension)
			suggestedName = fmt.Sprintf("%s (%d)%s", baseName, suffix, extension)
		}
		conflictType := "source_duplicate"
		if targetExists {
			conflictType = "source_duplicate_target_exists"
		}
		conflicts = append(conflicts, IsoCopyConflict{
			TargetDisk:    targetDisk,
			SourcePath:    sourcePath,
			FileName:      fileName,
			SuggestedName: suggestedName,
			ConflictType:  conflictType,
		})
		existing[suggestedName] = struct{}{}
	}
	return conflicts, nil
}

func reserveIsoDestination(targetDir string, fileName string) (string, error) {
	for suffix := 0; ; suffix++ {
		candidateName := fileName
		if suffix > 0 {
			extension := filepath.Ext(fileName)
			baseName := strings.TrimSuffix(fileName, extension)
			candidateName = fmt.Sprintf("%s (%d)%s", baseName, suffix, extension)
		}
		candidatePath := filepath.Join(targetDir, candidateName)
		if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
			return candidatePath, nil
		} else if err != nil {
			return "", err
		}
	}
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
