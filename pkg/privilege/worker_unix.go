// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build unix

package privilege

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

// isProcessAlive checks whether a process with given PID exists in the system.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// cleanStaleSockets removes leftover Unix domain sockets from dead GUI processes.
func cleanStaleSockets(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "w-") && strings.HasSuffix(name, ".sock") {
			pidStr := strings.TrimSuffix(strings.TrimPrefix(name, "w-"), ".sock")
			if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
				if !isProcessAlive(pid) {
					_ = os.Remove(filepath.Join(dir, name))
				}
			}
		}
	}
}

// RunWorkerFromArgs parses command-line arguments and runs the worker server loop.
func RunWorkerFromArgs(args []string) error {
	var socketPath, token, tokenFile string
	var parentPID int
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
		} else if (arg == "--token-file" || arg == "-token-file") && i+1 < len(args) {
			tokenFile = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--token-file=") {
			tokenFile = strings.TrimPrefix(arg, "--token-file=")
		} else if (arg == "--parent-pid" || arg == "-parent-pid") && i+1 < len(args) {
			parentPID, _ = strconv.Atoi(args[i+1])
			i++
		} else if strings.HasPrefix(arg, "--parent-pid=") {
			parentPID, _ = strconv.Atoi(strings.TrimPrefix(arg, "--parent-pid="))
		}
	}
	if token == "" && tokenFile != "" {
		data, err := os.ReadFile(tokenFile)
		if err != nil {
			return fmt.Errorf("failed to read token file: %w", err)
		}
		token = strings.TrimSpace(string(data))
		_ = os.Remove(tokenFile)
	}
	return RunWorkerServer(socketPath, token, parentPID)
}

// RunWorkerServer starts the Unix domain socket server loop for the privileged worker.
func RunWorkerServer(socketPath, token string, parentPID int) error {
	if socketPath == "" || token == "" {
		return fmt.Errorf("socket path and token are required")
	}

	// Ensure SIGHUP from terminating parent shell does not kill the daemon worker
	signal.Ignore(syscall.SIGHUP)

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

	// Clean up snapshots on termination signal (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		mu.Lock()
		restoreAllSnapshots(snapshots)
		mu.Unlock()
		_ = listener.Close()
		_ = os.Remove(socketPath)
		os.Exit(0)
	}()

	// Parent process watchdog: if GUI app crashes or is force-killed, auto-restore and exit
	if parentPID > 0 {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if !isProcessAlive(parentPID) {
					logger.Info("Parent process exited, privileged worker auto-terminating", "parent_pid", parentPID)
					mu.Lock()
					restoreAllSnapshots(snapshots)
					mu.Unlock()
					_ = listener.Close()
					_ = os.Remove(socketPath)
					os.Exit(0)
				}
			}
		}()
	}

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

func dialWorkerSocket(socketPath string) (net.Conn, error) {
	return net.DialTimeout("unix", socketPath, 2*time.Second)
}

// StartOrConnectWorker launches the privileged worker with administrator elevation and connects to it.
func StartOrConnectWorker(prompt string) (*WorkerClient, error) {
	StartWorkerMutex.Lock()
	defer StartWorkerMutex.Unlock()

	if client := GetActiveWorkerClient(); client != nil {
		return client, nil
	}

	socketDir := filepath.Join(os.TempDir(), fmt.Sprintf("unigo-ipc-%d", os.Getuid()))
	_ = os.MkdirAll(socketDir, 0700)
	_ = os.Chmod(socketDir, 0700)
	cleanStaleSockets(socketDir)

	socketPath := filepath.Join(socketDir, fmt.Sprintf("w-%d.sock", os.Getpid()))

	// If an existing worker is already running and listening for this process, connect silently
	if lastWorkerToken != "" {
		if c, err := dialWorkerSocket(socketPath); err == nil {
			testClient := &WorkerClient{
				conn:       c,
				token:      lastWorkerToken,
				socketPath: socketPath,
				lastPing:   time.Now(),
			}
			if testClient.IsAlive() {
				SetActiveWorkerClient(testClient)
				return testClient, nil
			}
			_ = testClient.Close()
		}
	}

	token, err := GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	_ = os.Remove(socketPath)
	workerLog := filepath.Join(socketDir, "worker.log")
	_ = os.Remove(workerLog)

	tokenFile := filepath.Join(socketDir, fmt.Sprintf("t-%d.tok", os.Getpid()))
	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		return nil, fmt.Errorf("failed to write token file: %w", err)
	}
	defer func() {
		_ = os.Remove(tokenFile)
	}()

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to determine executable path: %w", err)
	}

	if prompt == "" {
		prompt = "UniGoDesktop requires administrator privileges to access raw storage devices and verify boot partitions."
	}

	elevDone := make(chan error, 1)

	switch runtime.GOOS {
	case "darwin":
		escapedExe := strings.ReplaceAll(exe, "'", "'\"'\"'")
		escapedSocket := strings.ReplaceAll(socketPath, "'", "'\"'\"'")
		escapedTokenFile := strings.ReplaceAll(tokenFile, "'", "'\"'\"'")
		escapedPrompt := strings.ReplaceAll(prompt, `"`, `\"`)
		escapedLog := strings.ReplaceAll(workerLog, "'", "'\"'\"'")
		escapedDir := strings.ReplaceAll(socketDir, "'", "'\"'\"'")

		// Wrap in parentheses subshell with full I/O redirection so AppleScript 'do shell script'
		// detaches cleanly and returns in under 200ms without holding inherited pipe descriptors.
		bgCmd := fmt.Sprintf("(cd '%s' && '%s' --privileged-worker --socket '%s' --token-file '%s' --parent-pid '%d') </dev/null >'%s' 2>&1 &",
			escapedDir, escapedExe, escapedSocket, escapedTokenFile, os.Getpid(), escapedLog)
		appleScript := fmt.Sprintf(`do shell script "%s" with prompt "%s" with administrator privileges`,
			bgCmd, escapedPrompt)

		go func() {
			cmd := exec.Command("osascript", "-e", appleScript)
			out, err := cmd.CombinedOutput()
			if err != nil {
				elevDone <- fmt.Errorf("elevation failed: %s (%w)", strings.TrimSpace(string(out)), err)
				return
			}
			elevDone <- nil
		}()

	case "linux":
		bgCmd := fmt.Sprintf("(cd %s && %s --privileged-worker --socket %s --token-file %s --parent-pid %d) </dev/null >%s 2>&1 &",
			socketDir, exe, socketPath, tokenFile, os.Getpid(), workerLog)
		go func() {
			cmd := exec.Command("pkexec", "sh", "-c", bgCmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				elevDone <- fmt.Errorf("elevation failed: %s (%w)", strings.TrimSpace(string(out)), err)
				return
			}
			elevDone <- nil
		}()

	default:
		return nil, fmt.Errorf("unsupported unix platform: %s", runtime.GOOS)
	}

	// Retry connection until socket appears or timeout (60 seconds for user authorization)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var conn net.Conn
	for {
		// If user explicitly dismissed or cancelled the authorization prompt, abort immediately
		select {
		case elevErr := <-elevDone:
			if elevErr != nil {
				_ = os.Remove(socketPath)
				return nil, elevErr
			}
		default:
		}

		select {
		case <-ctx.Done():
			_ = os.Remove(socketPath)
			if logData, lErr := os.ReadFile(workerLog); lErr == nil && len(logData) > 0 {
				return nil, fmt.Errorf("privileged worker failed to start: %s", strings.TrimSpace(string(logData)))
			}
			return nil, fmt.Errorf("timed out waiting for privileged worker to start")
		default:
			c, err := net.Dial("unix", socketPath)
			if err == nil {
				conn = c
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if conn != nil {
			break
		}
	}

	client := &WorkerClient{
		conn:       conn,
		token:      token,
		socketPath: socketPath,
		lastPing:   time.Now(),
	}

	if !client.IsAlive() {
		_ = client.Close()
		_ = os.Remove(socketPath)
		return nil, fmt.Errorf("privileged worker failed handshake")
	}

	SetActiveWorkerClient(client)
	return client, nil
}
