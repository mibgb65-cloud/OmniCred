//go:build windows

package updater

import (
	"context"
	"os"
	"testing"
)

func TestVerifiedInstallerRemainsLockedDuringLaunch(t *testing.T) {
	manager, latest, manifest, data := testManager(t)
	serveRelease(manager, latest, manifest, data)
	_, _ = manager.Check(context.Background())
	_, _ = manager.Download()
	if state := waitDownload(t, manager); state.Phase != "ready" {
		t.Fatal(state)
	}
	if err := manager.Install(func(path string) error {
		if file, err := os.OpenFile(path, os.O_WRONLY, 0); err == nil {
			file.Close()
			t.Fatal("installer can be modified between verification and launch")
		}
		if os.Remove(path) == nil {
			t.Fatal("locked installer was deleted")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
