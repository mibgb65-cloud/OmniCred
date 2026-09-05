package updater

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func waitDownload(t *testing.T, manager *Manager) State {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		state := manager.Status()
		if state.Phase != "downloading" && state.Phase != "verifying" {
			return state
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("download did not complete")
	return State{}
}

func TestDownloadVerifyAndInstall(t *testing.T) {
	manager, latest, manifest, data := testManager(t)
	serveRelease(manager, latest, manifest, data)
	if _, err := manager.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Download(); err != nil {
		t.Fatal(err)
	}
	state := waitDownload(t, manager)
	if state.Phase != "ready" || state.Downloaded != int64(len(data)) {
		t.Fatalf("state = %+v", state)
	}
	launchError := errors.New("cancelled permission")
	if err := manager.Install(func(string) error { return launchError }); !errors.Is(err, launchError) {
		t.Fatal(err)
	}
	if manager.Status().Phase != "ready" {
		t.Fatal("cancelled launch must remain retryable")
	}
	var launched string
	if err := manager.Install(func(path string) error { launched = path; return nil }); err != nil {
		t.Fatal(err)
	}
	if launched == "" || manager.Status().Phase != "installing" {
		t.Fatal("installer not launched")
	}
	if err := manager.Install(func(string) error { t.Fatal("double launch"); return nil }); err == nil {
		t.Fatal("double install accepted")
	}
}

func TestDownloadFailuresNeverBecomeInstallable(t *testing.T) {
	for _, kind := range []string{"corrupt", "truncated", "oversized", "http_error"} {
		t.Run(kind, func(t *testing.T) {
			manager, latest, manifest, data := testManager(t)
			payload := append([]byte(nil), data...)
			switch kind {
			case "corrupt":
				payload[0] = '!'
			case "truncated":
				payload = payload[:3]
			case "oversized":
				payload = append(payload, '!')
			}
			serveRelease(manager, latest, manifest, payload)
			if _, err := manager.Check(context.Background()); err != nil {
				t.Fatal(err)
			}
			if kind == "http_error" {
				manager.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 503, Body: http.NoBody}, nil
				})
			}
			if _, err := manager.Download(); err != nil {
				t.Fatal(err)
			}
			state := waitDownload(t, manager)
			if state.Phase != "error" || state.Error == "" {
				t.Fatalf("state = %+v", state)
			}
			if err := manager.Install(func(string) error { t.Fatal("unsafe installer launched"); return nil }); err == nil {
				t.Fatal("install accepted")
			}
			entries, err := os.ReadDir(manager.cacheRoot)
			if err != nil || len(entries) != 0 {
				t.Fatalf("partial file retained: %v", err)
			}
			serveRelease(manager, latest, manifest, data)
			if _, err := manager.Download(); err != nil {
				t.Fatal(err)
			}
			if state := waitDownload(t, manager); state.Phase != "ready" {
				t.Fatalf("retry = %+v", state)
			}
		})
	}
}

func TestReverifyBeforeLaunch(t *testing.T) {
	manager, latest, manifest, data := testManager(t)
	serveRelease(manager, latest, manifest, data)
	_, _ = manager.Check(context.Background())
	_, _ = manager.Download()
	if state := waitDownload(t, manager); state.Phase != "ready" {
		t.Fatal(state)
	}
	data[0] = '!'
	if err := os.WriteFile(manager.installerPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := manager.Install(func(string) error { t.Fatal("modified installer launched"); return nil }); err == nil {
		t.Fatal("tampered installer accepted")
	}
	if manager.Status().Phase != "error" {
		t.Fatal("tampered installer still ready")
	}
}

func TestCancellationAndDuplicateDownloads(t *testing.T) {
	manager, latest, manifest, data := testManager(t)
	serveRelease(manager, latest, manifest, data)
	_, _ = manager.Check(context.Background())
	started := make(chan struct{})
	var requests atomic.Int32
	manager.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		close(started)
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	_, _ = manager.Download()
	<-started
	for range 3 {
		_, _ = manager.Download()
	}
	manager.Cancel()
	if state := waitDownload(t, manager); state.Phase != "idle" {
		t.Fatalf("cancel = %+v", state)
	}
	if requests.Load() != 1 {
		t.Fatal("duplicate download requests")
	}
	entries, _ := os.ReadDir(manager.cacheRoot)
	if len(entries) != 0 {
		t.Fatal("cancelled download not cleaned")
	}
}

func TestCloseCancelsActiveResponseAndPreventsRestart(t *testing.T) {
	manager, latest, manifest, data := testManager(t)
	serveRelease(manager, latest, manifest, data)
	_, _ = manager.Check(context.Background())
	started := make(chan struct{})
	manager.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		reader, writer := io.Pipe()
		go func() { <-request.Context().Done(); _ = writer.CloseWithError(request.Context().Err()) }()
		close(started)
		return &http.Response{StatusCode: 200, ContentLength: -1, Body: reader}, nil
	})
	_, _ = manager.Download()
	<-started
	manager.Close()
	if _, err := manager.Download(); err == nil {
		t.Fatal("closed manager restarted")
	}
}

func TestSignedRedirectErrorsDoNotLeakURL(t *testing.T) {
	manager, _, _, _ := testManager(t)
	manager.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("https://release-assets.githubusercontent.com/file?secret=do-not-log")
	})
	_, err := manager.Check(context.Background())
	if err == nil || strings.Contains(err.Error(), "do-not-log") {
		t.Fatalf("unsafe error: %v", err)
	}
}
