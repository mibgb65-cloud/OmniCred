package desktop

import (
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
