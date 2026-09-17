// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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

var (
	urlCredRegex    = regexp.MustCompile(`(?i)(https?://[^:]+:)[^@]+(@)`)
	authHeaderRegex = regexp.MustCompile(`(?i)(bearer|basic)\s+\S+`)
	kvPairRegex     = regexp.MustCompile(`(?i)\b(password|passwd|pass|pwd|secret|token|apikey|api_key|access_key|secret_key|private_key|key|auth|credential|credentials)=[^&\s,;]+`)
)

// SetWailsContext registers the Wails runtime context for broadcasting real-time logs to the UI.
func SetWailsContext(ctx context.Context) {
	wailsCtxMutex.Lock()
	defer wailsCtxMutex.Unlock()
	wailsCtx = ctx
}

// isSensitiveKey checks if a log argument key represents a sensitive credential/secret/key.
func isSensitiveKey(keyStr string) bool {
	k := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(keyStr), "_", ""), "-", ""))
	if k == "" {
		return false
	}
	// 1. Exact match for short keywords
	exactKeys := []string{"key", "pass", "pwd", "cred", "creds", "sig", "auth"}
	for _, e := range exactKeys {
		if k == e {
			return true
		}
	}
	// 2. Substring match for explicit security term keywords
	substringKeys := []string{
		"password", "passwd", "secret", "token", "credential", "authorization",
		"privatekey", "apikey", "accesskey", "secretkey", "publickey", "authkey",
		"clientkey", "userkey", "bearer", "signature",
	}
	for _, s := range substringKeys {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

// sanitizeString masks embedded credentials, tokens, bearer headers and query parameters inside plain text strings.
func sanitizeString(str string) string {
	if str == "" {
		return str
	}
	str = urlCredRegex.ReplaceAllString(str, "${1}******${2}")
	str = authHeaderRegex.ReplaceAllString(str, "${1} ******")
	str = kvPairRegex.ReplaceAllString(str, "${1}=******")
	return str
}

// sanitizeArgs automatically redacts sensitive parameters like passwords, tokens, secrets, API keys, and credentials.
func sanitizeArgs(args ...any) []any {
	if len(args) == 0 {
		return args
	}
	sanitized := make([]any, len(args))
	copy(sanitized, args)

	for i := 0; i < len(sanitized); i++ {
		if keyStr, ok := sanitized[i].(string); ok {
			// Slog key-value pairs sit at even indices (0, 2, 4...)
			if i%2 == 0 && isSensitiveKey(keyStr) && i+1 < len(sanitized) {
				sanitized[i+1] = "******"
				i++ // Skip value
				continue
			}
			sanitized[i] = sanitizeString(keyStr)
		}
	}
	return sanitized
}

// RecordLog records a log entry into the memory buffer and emits it to Wails if attached.
func RecordLog(level string, msg string, args ...any) {
	cleanArgs := sanitizeArgs(args...)

	logBufferMutex.Lock()
	globalLogID++
	entry := LogEntry{
		ID:        globalLogID,
		Timestamp: time.Now(),
		Level:     level,
		Message:   sanitizeString(msg),
	}
	if len(cleanArgs) > 0 {
		entry.Details = fmt.Sprintf("%v", cleanArgs)
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
