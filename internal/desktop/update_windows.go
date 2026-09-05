//go:build windows

package desktop

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func updateParameters(directory string, pid int) (string, error) {
	if !filepath.IsAbs(directory) || strings.ContainsAny(directory, "\"\r\n\x00") || len(directory) > 240 || pid <= 0 {
		return "", errors.New("当前安装位置不支持静默更新")
	}
	// NSIS 要求 /D 最后出现，路径含空格时也不能添加引号。
	return fmt.Sprintf("/S /UPDATEPID=%d /D=%s", pid, filepath.Clean(directory)), nil
}
