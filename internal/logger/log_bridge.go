// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"context"
	"fmt"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogEntry represents a single structured log line sent to the UI or stored in history
type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
}

var (
	logBufferMutex sync.RWMutex
	logBuffer      []LogEntry
	maxBufferSize  = 500
	globalLogID    int64
	wailsCtx       context.Context
	wailsCtxMutex  sync.RWMutex
)

// SetWailsContext registers the Wails runtime context for broadcasting real-time logs to the UI.
func SetWailsContext(ctx context.Context) {
	wailsCtxMutex.Lock()
	defer wailsCtxMutex.Unlock()
	wailsCtx = ctx
}

// RecordLog records a log entry into the memory buffer and emits it to Wails if attached.
func RecordLog(level string, msg string, args ...any) {
	logBufferMutex.Lock()
	globalLogID++
	entry := LogEntry{
		ID:        globalLogID,
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	if len(args) > 0 {
		entry.Details = fmt.Sprintf("%v", args)
	}

	logBuffer = append(logBuffer, entry)
	if len(logBuffer) > maxBufferSize {
		logBuffer = logBuffer[len(logBuffer)-maxBufferSize:]
	}
	logBufferMutex.Unlock()

	// Broadcast to Wails UI
	wailsCtxMutex.RLock()
	ctx := wailsCtx
	wailsCtxMutex.RUnlock()

	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, "log:entry", entry)
	}
}

// GetRecentLogs returns a copy of recent log entries from the memory buffer.
func GetRecentLogs() []LogEntry {
	logBufferMutex.RLock()
	defer logBufferMutex.RUnlock()
	result := make([]LogEntry, len(logBuffer))
	copy(result, logBuffer)
	return result
}

// ClearLogs clears the in-memory log buffer.
func ClearLogs() {
	logBufferMutex.Lock()
	defer logBufferMutex.Unlock()
	logBuffer = make([]LogEntry, 0)
}
