// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build unix

package privilege

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

// RunWorkerFromArgs parses command-line arguments and runs the worker server loop.
func RunWorkerFromArgs(args []string) error {
	var socketPath, token string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == "--socket" || arg == "-socket") && i+1 < len(args) {
			socketPath = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--socket=") {
			socketPath = strings.TrimPrefix(arg, "--socket=")
		} else if (arg == "--token" || arg == "-token") && i+1 < len(args) {
			token = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--token=") {
			token = strings.TrimPrefix(arg, "--token=")
		}
	}
	return RunWorkerServer(socketPath, token)
}

// RunWorkerServer starts the Unix domain socket server loop for the privileged worker.
func RunWorkerServer(socketPath, token string) error {
	if socketPath == "" || token == "" {
		return fmt.Errorf("socket path and token are required")
	}

	// Clean up stale socket if present
	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket %s: %w", socketPath, err)
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	// Ensure socket is accessible by current process and user
	_ = os.Chmod(socketPath, 0666)

	snapshots := make(map[string]diskSnapshot)
	var mu sync.Mutex

	// Clean up snapshots on termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigChan
		mu.Lock()
		restoreAllSnapshots(snapshots)
		mu.Unlock()
		_ = listener.Close()
		_ = os.Remove(socketPath)
		os.Exit(0)
	}()

	logger.Info("Privileged worker listening on unix socket", "socket", socketPath)

	idleTimer := time.NewTimer(defaultWorkerIdleTimeout)
	go func() {
		<-idleTimer.C
		logger.Info("Privileged worker idle timeout reached, shutting down")
		mu.Lock()
		restoreAllSnapshots(snapshots)
		mu.Unlock()
		_ = listener.Close()
		_ = os.Remove(socketPath)
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		// Reset idle timer on connection
		idleTimer.Reset(defaultWorkerIdleTimeout)
		go handleWorkerConnection(conn, token, snapshots, &mu)
	}

	mu.Lock()
	restoreAllSnapshots(snapshots)
	mu.Unlock()
	return nil
}

// StartOrConnectWorker launches the privileged worker with administrator elevation and connects to it.
func StartOrConnectWorker(prompt string) (*WorkerClient, error) {
	if client := GetActiveWorkerClient(); client != nil {
		return client, nil
	}

	token, err := GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	socketDir := "/tmp/unigo-ipc"
	_ = os.MkdirAll(socketDir, 0700)
	socketPath := filepath.Join(socketDir, fmt.Sprintf("w-%d.sock", os.Getpid()))
	_ = os.Remove(socketPath)

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to determine executable path: %w", err)
	}

	if prompt == "" {
		prompt = "UniGoDesktop requires administrator privileges to access raw storage devices and verify boot partitions."
	}

	// Launch worker in background with elevation
	switch runtime.GOOS {
	case "darwin":
		escapedExe := strings.ReplaceAll(exe, "'", "'\"'\"'")
		escapedSocket := strings.ReplaceAll(socketPath, "'", "'\"'\"'")
		escapedToken := strings.ReplaceAll(token, "'", "'\"'\"'")
		escapedPrompt := strings.ReplaceAll(prompt, `"`, `\"`)

		// Use nohup and redirect stdout/stderr to background it so AppleScript returns once daemon is spawned
		bgCmd := fmt.Sprintf("nohup '%s' --privileged-worker --socket '%s' --token '%s' >/dev/null 2>&1 &",
			escapedExe, escapedSocket, escapedToken)
		appleScript := fmt.Sprintf(`do shell script "%s" with prompt "%s" with administrator privileges`,
			bgCmd, escapedPrompt)

		cmd := exec.Command("osascript", "-e", appleScript)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("elevation failed: %s (%w)", strings.TrimSpace(string(out)), err)
		}

	case "linux":
		// On Linux, use pkexec with background execution
		bgCmd := fmt.Sprintf("nohup %s --privileged-worker --socket %s --token %s >/dev/null 2>&1 &",
			exe, socketPath, token)
		cmd := exec.Command("pkexec", "sh", "-c", bgCmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("elevation failed: %s (%w)", strings.TrimSpace(string(out)), err)
		}

	default:
		return nil, fmt.Errorf("unsupported unix platform: %s", runtime.GOOS)
	}

	// Retry connection until socket appears or timeout (120 seconds to allow ample time for user authentication)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var conn net.Conn
	for {
		select {
		case <-ctx.Done():
			_ = os.Remove(socketPath)
			return nil, fmt.Errorf("timed out waiting for privileged worker to start")
		default:
			c, err := net.Dial("unix", socketPath)
			if err == nil {
				conn = c
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if conn != nil {
			break
		}
	}

	client := &WorkerClient{
		conn:  conn,
		token: token,
	}

	if !client.IsAlive() {
		_ = client.Close()
		_ = os.Remove(socketPath)
		return nil, fmt.Errorf("privileged worker failed handshake")
	}

	SetActiveWorkerClient(client)
	return client, nil
}
