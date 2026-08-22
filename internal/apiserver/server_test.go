package apiserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestServerStartAndShutdown(t *testing.T) {
	handler := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := New("127.0.0.1:0", handler, logger)
	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	response, err := http.Get("http://" + server.Address())
	if err != nil {
		t.Fatalf("GET server: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("GET status = %d", response.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
