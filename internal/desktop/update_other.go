//go:build !windows

package desktop

import "errors"

func startUpdateInstaller(string) error { return errors.New("当前系统不支持静默安装") }

func RunUpdateHelper([]string) bool { return false }
