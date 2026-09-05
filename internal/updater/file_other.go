//go:build !windows

package updater

import (
	"os"
	"runtime"
)

func nativeArchitecture() string { return runtime.GOARCH }

func openInstaller(path string) (*os.File, error) { return os.Open(path) }
