// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package privilege

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestGenerateRandomToken(t *testing.T) {
	tok1, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken failed: %v", err)
	}
	if len(tok1) != 64 {
		t.Fatalf("expected 64 hex chars (32 bytes), got %d: %s", len(tok1), tok1)
	}

	tok2, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken failed: %v", err)
	}
	if tok1 == tok2 {
		t.Fatal("tokens must be uniquely generated")
	}
}

func TestWorkerClientServerRoundTrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix domain socket tests run on unix-like platforms")
	}

	socketPath := fmt.Sprintf("/tmp/test-worker-%d-1.sock", os.Getpid())
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)
	token := "test-secret-token-1234567890abcdef"

	// Start worker server in background
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- RunWorkerServer(socketPath, token, 0)
	}()

	// Wait for socket to be ready
	var conn net.Conn
	var dialErr error
	for i := 0; i < 20; i++ {
		conn, dialErr = net.Dial("unix", socketPath)
		if dialErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if dialErr != nil {
		t.Fatalf("failed to connect to test worker server: %v", dialErr)
	}

	client := &WorkerClient{
		conn:  conn,
		token: token,
	}
	defer client.Close()

	if !client.IsAlive() {
		t.Fatal("expected worker client to be alive after handshake")
	}

	// Verify IsElevated reflects active worker status
	SetActiveWorkerClient(client)
	if !IsElevated() {
		t.Fatal("expected IsElevated to return true when active worker client is connected")
	}

	// Test ReleaseDiskAccess on safe dummy path
	if err := client.ReleaseDiskAccess([]string{"/dev/sdb"}); err != nil {
		t.Fatalf("ReleaseDiskAccess failed: %v", err)
	}

	// Test ReadSector on blocked path
	if _, err := client.ReadSector("/dev/disk0", 512); err == nil {
		t.Fatal("expected ReadSector on /dev/disk0 to be rejected as system disk")
	}

	// Reset worker client and verify IsElevated reverts
	SetActiveWorkerClient(nil)
	ResetElevationCache()
	if os.Geteuid() != 0 && IsElevated() {
		t.Fatal("expected IsElevated to return false after active worker client is detached")
	}
}

func TestWorkerClientRejectsInvalidToken(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix domain socket tests run on unix-like platforms")
	}

	socketPath := fmt.Sprintf("/tmp/test-worker-%d-2.sock", os.Getpid())
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)
	token := "correct-token"

	go func() {
		_ = RunWorkerServer(socketPath, token, 0)
	}()

	var conn net.Conn
	for i := 0; i < 20; i++ {
		c, err := net.Dial("unix", socketPath)
		if err == nil {
			conn = c
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if conn == nil {
		t.Fatal("failed to connect to test worker")
	}

	badClient := &WorkerClient{
		conn:  conn,
		token: "wrong-token",
	}
	defer badClient.Close()

	if badClient.IsAlive() {
		t.Fatal("expected worker to reject invalid token")
	}
}

func TestWorkerAcquireDiskRejectsBlockedDevices(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix domain socket tests run on unix-like platforms")
	}

	socketPath := fmt.Sprintf("/tmp/test-worker-%d-3.sock", os.Getpid())
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)
	token := "security-token"

	go func() {
		_ = RunWorkerServer(socketPath, token, 0)
	}()

	var conn net.Conn
	for i := 0; i < 20; i++ {
		c, err := net.Dial("unix", socketPath)
		if err == nil {
			conn = c
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if conn == nil {
		t.Fatal("failed to connect to test worker")
	}

	client := &WorkerClient{
		conn:  conn,
		token: token,
	}
	defer client.Close()

	// Ensure system root disk paths are rejected
	err := client.AcquireDiskAccess([]string{"/dev/disk0"})
	if err == nil {
		t.Fatal("expected AcquireDiskAccess to reject system disk /dev/disk0")
	}
	fmt.Printf("Successfully rejected blocked disk: %v\n", err)
}

func TestWorkerClientRunCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix domain socket tests run on unix-like platforms")
	}

	socketPath := fmt.Sprintf("/tmp/test-worker-%d-4.sock", os.Getpid())
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)
	token := "runcmd-token"

	go func() {
		_ = RunWorkerServer(socketPath, token, 0)
	}()

	var conn net.Conn
	for i := 0; i < 20; i++ {
		c, err := net.Dial("unix", socketPath)
		if err == nil {
			conn = c
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if conn == nil {
		t.Fatal("failed to connect to test worker")
	}

	client := &WorkerClient{
		conn:  conn,
		token: token,
	}
	defer client.Close()

	// Test 1: Disallowed command should be rejected immediately
	_, err := client.RunCommand("rm", "-rf", "/tmp/nonexistent")
	if err == nil {
		t.Fatal("expected RunCommand with 'rm' to be rejected by allowlist")
	}

	// Test 2: Shell injection or dangerous argument characters should be rejected
	_, err = client.RunCommand("mount", ";", "rm", "-rf")
	if err == nil {
		t.Fatal("expected RunCommand with ';' to be rejected by argument validator")
	}

	// Test 3: Disallowed disk path in arguments should be rejected
	_, err = client.RunCommand("chmod", "666", "/dev/disk0")
	if err == nil {
		t.Fatal("expected RunCommand on /dev/disk0 to be rejected by device path validator")
	}
}

func TestWorkerClientAutoRecoveryOnBrokenConnection(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	client := &WorkerClient{
		conn:  clientConn,
		token: "test-token",
	}
	SetActiveWorkerClient(client)

	// Close server side to simulate sudden worker termination/crash
	_ = serverConn.Close()

	// IsAlive should detect broken pipe, call markFailedLocked, and reset global client
	if client.IsAlive() {
		t.Fatal("expected IsAlive to return false on dead connection")
	}

	if GetActiveWorkerClient() != nil {
		t.Fatal("expected GetActiveWorkerClient to be reset to nil after failure")
	}

	if !client.closed {
		t.Fatal("expected client to be marked closed after failure")
	}
}

func TestIsProcessAlive(t *testing.T) {
	if !isProcessAlive(os.Getpid()) {
		t.Fatalf("expected current process %d to be alive", os.Getpid())
	}
	if isProcessAlive(99999999) {
		t.Fatal("expected non-existent process 99999999 to be reported dead")
	}
}


