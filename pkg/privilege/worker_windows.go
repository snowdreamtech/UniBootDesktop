// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package privilege

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"golang.org/x/sys/windows"
)

// isProcessAlive checks whether a process with given PID exists on Windows.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)

	var exitCode uint32
	if err := windows.GetExitCodeProcess(h, &exitCode); err != nil {
		return false
	}
	// STILL_ACTIVE is 259
	return exitCode == 259
}

// RunWorkerFromArgs parses command-line arguments and runs the worker server loop.
func RunWorkerFromArgs(args []string) error {
	var portFile, token, tokenFile string
	var parentPID int
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == "--port-file" || arg == "-port-file") && i+1 < len(args) {
			portFile = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--port-file=") {
			portFile = strings.TrimPrefix(arg, "--port-file=")
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
	return RunWorkerServer(portFile, token, parentPID)
}

// RunWorkerServer starts the TCP loopback server loop for the privileged worker on Windows.
func RunWorkerServer(portFile, token string, parentPID int) error {
	if portFile == "" || token == "" {
		return fmt.Errorf("port file and token are required")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to listen on tcp loopback: %w", err)
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(portFile)
	}()

	addr := listener.Addr().(*net.TCPAddr)
	if err := os.WriteFile(portFile, []byte(strconv.Itoa(addr.Port)), 0600); err != nil {
		return fmt.Errorf("failed to write port file: %w", err)
	}

	snapshots := make(map[string]diskSnapshot)
	var mu sync.Mutex

	logger.Info("Privileged worker listening on loopback", "port", addr.Port)

	// Parent process watchdog: if GUI app exits or is terminated, auto-clean and exit
	if parentPID > 0 {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if !isProcessAlive(parentPID) {
					logger.Info("Parent process exited, Windows privileged worker terminating", "parent_pid", parentPID)
					mu.Lock()
					restoreAllSnapshots(snapshots)
					mu.Unlock()
					_ = listener.Close()
					_ = os.Remove(portFile)
					os.Exit(0)
				}
			}
		}()
	}

	idleTimer := time.NewTimer(defaultWorkerIdleTimeout)
	go func() {
		<-idleTimer.C
		logger.Info("Privileged worker idle timeout reached, shutting down")
		mu.Lock()
		restoreAllSnapshots(snapshots)
		mu.Unlock()
		_ = listener.Close()
		_ = os.Remove(portFile)
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		idleTimer.Reset(defaultWorkerIdleTimeout)
		go handleWorkerConnection(conn, token, snapshots, &mu)
	}

	mu.Lock()
	restoreAllSnapshots(snapshots)
	mu.Unlock()
	return nil
}

func dialWorkerSocket(addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, 2*time.Second)
}

// StartOrConnectWorker launches the privileged worker with administrator elevation and connects to it on Windows.
func StartOrConnectWorker(prompt string) (*WorkerClient, error) {
	StartWorkerMutex.Lock()
	defer StartWorkerMutex.Unlock()

	if client := GetActiveWorkerClient(); client != nil {
		return client, nil
	}

	token, err := GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	portFile := filepath.Join(os.TempDir(), fmt.Sprintf("unigo-worker-%d.port", os.Getpid()))
	_ = os.Remove(portFile)

	tokenFile := filepath.Join(os.TempDir(), fmt.Sprintf("unigo-worker-%d.tok", os.Getpid()))
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

	args := []string{
		"--privileged-worker",
		"--port-file", portFile,
		"--token-file", tokenFile,
		"--parent-pid", strconv.Itoa(os.Getpid()),
	}

	psCmd := buildPowerShellStartDaemonCommand(exe, args)
	elevDone := make(chan error, 1)
	go func() {
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", psCmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			elevDone <- fmt.Errorf("elevation failed: %s (%w)", strings.TrimSpace(string(out)), err)
			return
		}
		elevDone <- nil
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var port int
	for {
		select {
		case elevErr := <-elevDone:
			if elevErr != nil {
				_ = os.Remove(portFile)
				return nil, elevErr
			}
		default:
		}

		select {
		case <-ctx.Done():
			_ = os.Remove(portFile)
			return nil, fmt.Errorf("timed out waiting for privileged worker port file")
		default:
			data, err := os.ReadFile(portFile)
			if err == nil && len(data) > 0 {
				p, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
				if parseErr == nil && p > 0 {
					port = p
					break
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		if port > 0 {
			break
		}
	}

	targetAddr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", targetAddr, 3*time.Second)
	if err != nil {
		_ = os.Remove(portFile)
		return nil, fmt.Errorf("failed to connect to worker on %s: %w", targetAddr, err)
	}

	client := &WorkerClient{
		conn:       conn,
		token:      token,
		socketPath: targetAddr,
		lastPing:   time.Now(),
	}

	if !client.IsAlive() {
		_ = client.Close()
		_ = os.Remove(portFile)
		return nil, fmt.Errorf("privileged worker failed handshake")
	}

	SetActiveWorkerClient(client)
	return client, nil
}
