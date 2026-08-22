//go:build windows

package desktop

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func startElevatedUninstaller(path string) error {
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return fmt.Errorf("prepare elevation request: %w", err)
	}
	executable, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("prepare uninstaller path: %w", err)
	}
	directory, err := windows.UTF16PtrFromString(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("prepare uninstaller directory: %w", err)
	}
	if err := windows.ShellExecute(0, verb, executable, nil, directory, windows.SW_SHOWNORMAL); err != nil {
		if errors.Is(err, windows.ERROR_CANCELLED) {
			return errors.New("administrator permission request was cancelled")
		}
		return fmt.Errorf("request administrator permission: %w", err)
	}
	return nil
}
