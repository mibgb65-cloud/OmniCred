//go:build !windows

package desktop

import (
	"os/exec"
	"path/filepath"
)

func startElevatedUninstaller(path string) error {
	command := exec.Command(path)
	command.Dir = filepath.Dir(path)
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
