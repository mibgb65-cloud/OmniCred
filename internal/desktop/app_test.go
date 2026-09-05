package desktop

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestIsUninstallerAvailable(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "uninstall.exe")
	if isUninstallerAvailable(path) {
		t.Fatal("missing uninstaller must not be available")
	}
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write test uninstaller: %v", err)
	}
	if !isUninstallerAvailable(path) {
		t.Fatal("regular uninstaller file must be available")
	}
	if isUninstallerAvailable(directory) {
		t.Fatal("directory must not be treated as an uninstaller")
	}
}

func TestUninstallUsesConfiguredLauncher(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "uninstall.exe")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write test uninstaller: %v", err)
	}

	app := New(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), "", path, nil)
	var launchedPath string
	app.startUninstaller = func(path string) error {
		launchedPath = path
		return nil
	}
	if err := app.Uninstall(); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if launchedPath != path {
		t.Fatalf("launched path = %q; want %q", launchedPath, path)
	}

	launchErr := errors.New("launch failed")
	app.startUninstaller = func(string) error { return launchErr }
	if err := app.Uninstall(); !errors.Is(err, launchErr) {
		t.Fatalf("uninstall error = %v; want wrapped launch error", err)
	}
}
