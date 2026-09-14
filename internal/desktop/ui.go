// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package desktop

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

// UIRunner manages the web backend server bridge and UI window binding.
type UIRunner struct {
	app      *App
	server   *http.Server
	listener net.Listener
	port     int
	mu       sync.Mutex
}

// NewUIRunner constructs a new UIRunner instance.
func NewUIRunner(app *App) *UIRunner {
	return &UIRunner{
		app: app,
	}
}

// Start boots the embedded HTTP server for web assets and backend API bridge.
func (r *UIRunner) Start(ctx context.Context) error {
	r.mu.Lock()
	mux := http.NewServeMux()

	// Serve embedded static UI assets
	mux.Handle("/", http.FileServer(http.FS(GetAssetsFS())))

	// Provide health/status API endpoint for desktop UI
	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","app":"UniGoDesktop","state":"%s","uptime":"%s"}`,
			r.app.GetState(), r.app.Uptime().String())
	})

	// Find available local port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		r.mu.Unlock()
		return fmt.Errorf("failed to bind local desktop port: %w", err)
	}

	r.listener = ln
	r.port = ln.Addr().(*net.TCPAddr).Port

	r.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	r.mu.Unlock()

	logger.Info("Desktop Web Bridge Server started", "url", fmt.Sprintf("http://127.0.0.1:%d", r.port))

	// Listen and serve
	if err := r.server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("desktop web server error: %w", err)
	}

	return nil
}

// Stop cleanly terminates the embedded desktop web bridge server.
func (r *UIRunner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.server.Shutdown(ctx)
	r.server = nil
	return err
}

// GetPort returns the active local port for the desktop bridge server.
func (r *UIRunner) GetPort() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.port
}
