// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

const (
	WorkerActionPing        = "ping"
	WorkerActionAcquireDisk = "acquire_disk"
	WorkerActionReleaseDisk = "release_disk"
	WorkerActionExit        = "exit"

	defaultWorkerIdleTimeout = 30 * time.Minute
)

// WorkerRequest represents an RPC request from main application to privileged worker.
type WorkerRequest struct {
	Token  string   `json:"token"`
	Action string   `json:"action"`
	Paths  []string `json:"paths,omitempty"`
	UID    int      `json:"uid,omitempty"`
	GID    int      `json:"gid,omitempty"`
}

// WorkerResponse represents an RPC response from privileged worker.
type WorkerResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// diskSnapshot stores original ownership and permissions for automatic restoration.
type diskSnapshot struct {
	path      string
	perm      os.FileMode
	uid       int
	gid       int
	haveOwner bool
}

// WorkerClient maintains a persistent RPC connection to the background privileged worker.
type WorkerClient struct {
	mu     sync.Mutex
	conn   net.Conn
	token  string
	closed bool
}

var (
	globalWorkerMutex  sync.RWMutex
	globalWorkerClient *WorkerClient
)

// GenerateRandomToken generates a secure hex-encoded 256-bit token.
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GetActiveWorkerClient returns the active connected WorkerClient if available and healthy.
func GetActiveWorkerClient() *WorkerClient {
	globalWorkerMutex.RLock()
	client := globalWorkerClient
	globalWorkerMutex.RUnlock()

	if client != nil && client.IsAlive() {
		return client
	}
	return nil
}

// SetActiveWorkerClient sets or resets the active WorkerClient.
func SetActiveWorkerClient(client *WorkerClient) {
	globalWorkerMutex.Lock()
	defer globalWorkerMutex.Unlock()
	globalWorkerClient = client
}

// IsAlive checks whether the client connection is currently active and responding.
func (c *WorkerClient) IsAlive() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.conn == nil {
		return false
	}

	// Non-blocking ping with short deadline
	_ = c.conn.SetDeadline(time.Now().Add(1 * time.Second))
	defer func() {
		if c.conn != nil {
			_ = c.conn.SetDeadline(time.Time{})
		}
	}()

	req := WorkerRequest{
		Token:  c.token,
		Action: WorkerActionPing,
	}
	if err := json.NewEncoder(c.conn).Encode(req); err != nil {
		return false
	}

	var resp WorkerResponse
	if err := json.NewDecoder(c.conn).Decode(&resp); err != nil {
		return false
	}
	return resp.Success
}

// AcquireDiskAccess requests the privileged worker to grant the current user access to disk nodes.
func (c *WorkerClient) AcquireDiskAccess(paths []string) error {
	if c == nil {
		return errors.New("worker client is not connected")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.conn == nil {
		return errors.New("worker connection is closed")
	}

	req := WorkerRequest{
		Token:  c.token,
		Action: WorkerActionAcquireDisk,
		Paths:  paths,
		UID:    os.Getuid(),
		GID:    os.Getgid(),
	}

	_ = c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer func() {
		if c.conn != nil {
			_ = c.conn.SetDeadline(time.Time{})
		}
	}()

	if err := json.NewEncoder(c.conn).Encode(req); err != nil {
		return fmt.Errorf("failed to send acquire disk request: %w", err)
	}

	var resp WorkerResponse
	if err := json.NewDecoder(c.conn).Decode(&resp); err != nil {
		return fmt.Errorf("failed to read acquire disk response: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("worker failed to acquire disk access: %s", resp.Error)
	}
	return nil
}

// ReleaseDiskAccess requests the privileged worker to restore original disk node permissions.
func (c *WorkerClient) ReleaseDiskAccess(paths []string) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.conn == nil {
		return nil
	}

	req := WorkerRequest{
		Token:  c.token,
		Action: WorkerActionReleaseDisk,
		Paths:  paths,
	}

	_ = c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer func() {
		if c.conn != nil {
			_ = c.conn.SetDeadline(time.Time{})
		}
	}()

	if err := json.NewEncoder(c.conn).Encode(req); err != nil {
		return fmt.Errorf("failed to send release disk request: %w", err)
	}

	var resp WorkerResponse
	if err := json.NewDecoder(c.conn).Decode(&resp); err != nil {
		return fmt.Errorf("failed to read release disk response: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("worker failed to release disk access: %s", resp.Error)
	}
	return nil
}

// Close closes the connection to the worker.
func (c *WorkerClient) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	if c.conn != nil {
		_ = c.conn.Close()
	}
	return nil
}

// handleWorkerConnection handles RPC requests from the main application.
func handleWorkerConnection(conn net.Conn, expectedToken string, snapshots map[string]diskSnapshot, mu *sync.Mutex) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req WorkerRequest
		if err := decoder.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("Privileged worker client disconnected")
			} else {
				logger.Warn("Privileged worker decode error", "error", err)
			}
			return
		}

		if req.Token != expectedToken {
			_ = encoder.Encode(WorkerResponse{Success: false, Error: "invalid token"})
			return
		}

		switch req.Action {
		case WorkerActionPing:
			_ = encoder.Encode(WorkerResponse{Success: true})

		case WorkerActionAcquireDisk:
			mu.Lock()
			err := applyDiskAcquisition(req.Paths, req.UID, req.GID, snapshots)
			mu.Unlock()
			if err != nil {
				_ = encoder.Encode(WorkerResponse{Success: false, Error: err.Error()})
			} else {
				_ = encoder.Encode(WorkerResponse{Success: true})
			}

		case WorkerActionReleaseDisk:
			mu.Lock()
			applyDiskRelease(req.Paths, snapshots)
			mu.Unlock()
			_ = encoder.Encode(WorkerResponse{Success: true})

		case WorkerActionExit:
			mu.Lock()
			restoreAllSnapshots(snapshots)
			mu.Unlock()
			_ = encoder.Encode(WorkerResponse{Success: true})
			os.Exit(0)

		default:
			_ = encoder.Encode(WorkerResponse{Success: false, Error: fmt.Sprintf("unknown action: %s", req.Action)})
		}
	}
}

func applyDiskAcquisition(paths []string, targetUID, targetGID int, snapshots map[string]diskSnapshot) error {
	for _, raw := range paths {
		path := raw
		if path == "" {
			continue
		}
		if err := ValidateRawDevicePath(path); err != nil {
			return fmt.Errorf("unsafe device path %q: %w", path, err)
		}

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if _, exists := snapshots[path]; !exists {
			origPerm := info.Mode().Perm()
			uid, gid, haveOwner := snapshotOwner(info)
			snapshots[path] = diskSnapshot{
				path:      path,
				perm:      origPerm,
				uid:       uid,
				gid:       gid,
				haveOwner: haveOwner,
			}
		}

		if targetUID >= 0 && targetGID >= 0 {
			_ = chownPath(path, targetUID, targetGID)
		}
		_ = chmodPath(path, temporaryRawDiskPerm)
		logger.Info("Privileged worker granted disk access", "path", path, "uid", targetUID, "gid", targetGID)
	}
	return nil
}

func applyDiskRelease(paths []string, snapshots map[string]diskSnapshot) {
	for _, raw := range paths {
		path := raw
		if snap, ok := snapshots[path]; ok {
			if snap.haveOwner {
				_ = chownPath(snap.path, snap.uid, snap.gid)
			}
			_ = chmodPath(snap.path, snap.perm)
			delete(snapshots, path)
			logger.Info("Privileged worker restored disk permissions", "path", path)
		}
	}
}

func restoreAllSnapshots(snapshots map[string]diskSnapshot) {
	for path, snap := range snapshots {
		if snap.haveOwner {
			_ = chownPath(snap.path, snap.uid, snap.gid)
		}
		_ = chmodPath(snap.path, snap.perm)
		logger.Info("Privileged worker emergency restored disk permissions", "path", path)
	}
}
