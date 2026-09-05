//go:build windows

package desktop

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

const helperFlag = "--omnicred-update-helper"

func startUpdateInstaller(path string) error {
	executable, err := os.Executable()
	if err != nil {
		return errors.New("无法确定当前安装位置")
	}
	if _, err := updateParameters(filepath.Dir(executable), os.Getpid()); err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Base(executable), "OmniCred.exe") {
		return errors.New("当前程序已重命名，请从发布页手动安装更新")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return errors.New("无法读取当前工作目录")
	}
	// 辅助进程必须使用副本，否则 Windows 会继续锁住待替换的主程序。
	helperPath := filepath.Join(filepath.Dir(path), "update-helper.exe")
	_ = os.Remove(helperPath)
	if err := copyUpdateHelper(executable, helperPath); err != nil {
		return errors.New("无法准备更新辅助程序，请检查磁盘空间")
	}
	digest, err := installerDigest(path)
	if err != nil {
		return errors.New("无法读取待安装的更新包")
	}
	resultPath := filepath.Join(filepath.Dir(path), "launch-result")
	_ = os.Remove(resultPath)
	arguments := []string{helperFlag, strconv.Itoa(os.Getpid()), executable, path, digest, workingDirectory}
	arguments = append(arguments, os.Args[1:]...)
	command := exec.Command(helperPath, arguments...)
	command.Dir = filepath.Dir(helperPath)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS}
	if err := command.Start(); err != nil {
		return errors.New("无法启动更新辅助程序")
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait() }()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.NewTimer(3 * time.Minute)
	defer timeout.Stop()
	for {
		select {
		case <-ticker.C:
			if result, err := os.ReadFile(resultPath); err == nil {
				switch string(result) {
				case "started":
					return nil
				case "cancelled":
					return errors.New("已取消管理员授权，应用尚未退出，可再次点击重启并更新")
				default:
					return errors.New("无法启动更新安装器，应用尚未退出，请重试")
				}
			}
		case <-finished:
			if result, _ := os.ReadFile(resultPath); string(result) == "cancelled" {
				return errors.New("已取消管理员授权，应用尚未退出，可再次点击重启并更新")
			}
			return errors.New("更新辅助程序未能启动安装，应用尚未退出")
		case <-timeout.C:
			_ = command.Process.Kill()
			return errors.New("等待管理员授权超时，应用尚未退出，请重试")
		}
	}
}

// RunUpdateHelper 在初始化数据库、单实例锁和 WebView 之前处理内部更新模式。
func RunUpdateHelper(arguments []string) bool {
	if len(arguments) == 0 || arguments[0] != helperFlag {
		return false
	}
	if len(arguments) < 6 {
		return true
	}
	pid, err := strconv.Atoi(arguments[1])
	if err != nil || pid <= 0 {
		return true
	}
	target, installer, expected := arguments[2], arguments[3], arguments[4]
	executable, err := os.Executable()
	if err != nil || filepath.Base(executable) != "update-helper.exe" || filepath.Base(installer) != "installer.exe" ||
		!strings.EqualFold(filepath.Dir(executable), filepath.Dir(installer)) || !filepath.IsAbs(target) || len(expected) != 64 || !filepath.IsAbs(arguments[5]) {
		return true
	}
	resultPath := filepath.Join(filepath.Dir(executable), "launch-result")
	writeResult := func(result string) bool {
		temporary := resultPath + ".tmp"
		if os.WriteFile(temporary, []byte(result), 0o600) != nil {
			return false
		}
		return os.Rename(temporary, resultPath) == nil
	}
	parameters, err := updateParameters(filepath.Dir(target), pid)
	if err != nil {
		writeResult("failed")
		return true
	}
	name, _ := windows.UTF16PtrFromString(installer)
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		writeResult("failed")
		return true
	}
	locked := os.NewFile(uintptr(handle), installer)
	defer locked.Close()
	digest, err := installerDigest(installer)
	if err != nil || digest != expected {
		writeResult("failed")
		return true
	}
	process, err := launchElevatedInstaller(installer, parameters)
	if err != nil {
		if errors.Is(err, windows.ERROR_CANCELLED) {
			writeResult("cancelled")
		} else {
			writeResult("failed")
		}
		return true
	}
	defer windows.CloseHandle(process)
	if !writeResult("started") {
		_ = windows.TerminateProcess(process, 1)
		return true
	}
	status, err := windows.WaitForSingleObject(process, 10*60*1000)
	if err == nil && status == uint32(windows.WAIT_TIMEOUT) {
		// 安装器可能仍在写入文件，不能因等待超时强杀，否则会破坏覆盖升级。
		showUpdateError("更新安装耗时较长，请等待安装结束，暂勿再次启动 OmniCred。")
		status, err = windows.WaitForSingleObject(process, windows.INFINITE)
	}
	if err != nil || status != windows.WAIT_OBJECT_0 {
		showUpdateError("无法确认更新安装结果，请等待安装器退出后再打开 OmniCred。")
		return true
	}
	var code uint32
	if windows.GetExitCodeProcess(process, &code) != nil || code != 0 {
		showUpdateError("更新安装失败，已尝试恢复旧版程序。请在应用中重新下载更新，或从发布页安装。")
	}
	// 辅助进程始终以原用户权限运行，因此重启后的应用不会继承安装器的管理员权限。
	command := exec.Command(target, arguments[6:]...)
	command.Dir = arguments[5]
	if command.Start() != nil {
		showUpdateError("更新安装已结束，但无法自动打开应用。请通过桌面快捷方式启动 OmniCred。")
	} else {
		_ = command.Process.Release()
	}
	return true
}

func installerDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, io.LimitReader(file, (512<<20)+1)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyUpdateHelper(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o700)
	if err != nil {
		return err
	}
	_, err = io.Copy(output, input)
	closeErr := output.Close()
	return errors.Join(err, closeErr)
}
