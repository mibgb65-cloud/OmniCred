package apiserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	address  string
	handler  http.Handler
	logger   *slog.Logger
	mu       sync.Mutex
	listener net.Listener
	server   *http.Server
}

func New(address string, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{address: address, handler: handler, logger: logger}
}

func (server *Server) Start() error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.listener != nil {
		return nil
	}

	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", server.address, err)
	}
	server.listener = listener
	server.server = &http.Server{
		Handler:           server.handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	go func() {
		if err := server.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			server.logger.Error("local API stopped unexpectedly", "error", err)
		}
	}()
	server.logger.Info("local API is ready", "url", "http://"+listener.Addr().String())
	return nil
}

func (server *Server) Shutdown(ctx context.Context) error {
	server.mu.Lock()
	httpServer := server.server
	server.mu.Unlock()
	if httpServer == nil {
		return nil
	}
	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown local API: %w", err)
	}
	return nil
}

func (server *Server) Address() string {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.listener != nil {
		return server.listener.Addr().String()
	}
	return server.address
}
