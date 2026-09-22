package singleinstance

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrAlreadyRunning = errors.New("another UniGoDesktop instance is already running")

type lockMetadata struct {
	PID   int    `json:"pid"`
	Port  string `json:"port"`
	Token string `json:"token"`
}

type Guard struct {
	path     string
	listener net.Listener
	token    string
	mu       sync.RWMutex
	activate func()
	closeOne sync.Once
}

func Acquire(uniqueID string) (*Guard, error) {
	path := filepath.Join(os.TempDir(), uniqueID+".json")
	for attempt := 0; attempt < 2; attempt++ {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, fmt.Errorf("listen for single-instance activation: %w", err)
		}
		guard := &Guard{path: path, listener: listener, token: tokenFor(uniqueID)}
		metadata := lockMetadata{PID: os.Getpid(), Port: listener.Addr().String(), Token: guard.token}
		if err := createMetadata(path, metadata); err == nil {
			go guard.serve()
			return guard, nil
		} else if !errors.Is(err, os.ErrExist) {
			listener.Close()
			return nil, err
		}
		listener.Close()

		metadata, err = readMetadata(path)
		if err == nil && notify(metadata) {
			return nil, ErrAlreadyRunning
		}
		if err == nil && processAlive(metadata.PID) {
			if err := terminateStaleProcess(metadata.PID); err != nil {
				return nil, fmt.Errorf("existing instance is unresponsive: %w", err)
			}
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("remove stale single-instance state: %w", err)
		}
	}
	return nil, fmt.Errorf("acquire single-instance guard after recovery")
}

func (g *Guard) SetActivate(callback func()) {
	g.mu.Lock()
	g.activate = callback
	g.mu.Unlock()
}

func (g *Guard) Close() {
	g.closeOne.Do(func() {
		_ = g.listener.Close()
		_ = os.Remove(g.path)
	})
}

func (g *Guard) serve() {
	for {
		connection, err := g.listener.Accept()
		if err != nil {
			return
		}
		go g.handle(connection)
	}
}

func (g *Guard) handle(connection net.Conn) {
	defer connection.Close()
	connection.SetDeadline(time.Now().Add(500 * time.Millisecond))
	request, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil || strings.TrimSpace(request) != g.token {
		return
	}
	g.mu.RLock()
	callback := g.activate
	g.mu.RUnlock()
	if callback != nil {
		callback()
	}
	_, _ = connection.Write([]byte("ok\n"))
}

func notify(metadata lockMetadata) bool {
	connection, err := net.DialTimeout("tcp", metadata.Port, 300*time.Millisecond)
	if err != nil {
		return false
	}
	defer connection.Close()
	connection.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := fmt.Fprintf(connection, "%s\n", metadata.Token); err != nil {
		return false
	}
	response, err := bufio.NewReader(connection).ReadString('\n')
	return err == nil && strings.TrimSpace(response) == "ok"
}

func readMetadata(path string) (lockMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return lockMetadata{}, err
	}
	var metadata lockMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return lockMetadata{}, fmt.Errorf("parse single-instance state: %w", err)
	}
	return metadata, nil
}

func createMetadata(path string, metadata lockMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode single-instance state: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write single-instance state: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("close single-instance state: %w", err)
	}
	return nil
}

func tokenFor(uniqueID string) string {
	hash := sha256.Sum256([]byte(uniqueID))
	return hex.EncodeToString(hash[:])
}
