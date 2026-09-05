//go:build windows

package desktop

import (
	"errors"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

type shellExecuteInfo struct {
	Size       uint32
	Mask       uint32
	Window     windows.Handle
	Verb       *uint16
	File       *uint16
	Parameters *uint16
	Directory  *uint16
	Show       int32
	Instance   windows.Handle
	IDList     uintptr
	Class      *uint16
	ClassKey   windows.Handle
	HotKey     uint32
	Icon       windows.Handle
	Process    windows.Handle
}

func launchElevatedInstaller(path, parameters string) (windows.Handle, error) {
	return launchInstallerProcess(path, parameters, "runas")
}

func launchInstallerProcess(path, parameters, verbName string) (windows.Handle, error) {
	verb, err := windows.UTF16PtrFromString(verbName)
	if err != nil {
		return 0, err
	}
	file, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	arguments, err := windows.UTF16PtrFromString(parameters)
	if err != nil {
		return 0, err
	}
	directory, err := windows.UTF16PtrFromString(filepath.Dir(path))
	if err != nil {
		return 0, err
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ole := windows.NewLazySystemDLL("ole32.dll")
	result, _, _ := ole.NewProc("CoInitializeEx").Call(0, 2)
	if int32(result) >= 0 {
		defer ole.NewProc("CoUninitialize").Call()
	}
	info := shellExecuteInfo{
		Mask: 0x00000040 | 0x00000100 | 0x00000400, // 保留进程句柄、同步启动、由应用展示错误。
		Verb: verb, File: file, Parameters: arguments, Directory: directory, Show: windows.SW_HIDE,
	}
	info.Size = uint32(unsafe.Sizeof(info))
	ok, _, callErr := windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteExW").Call(uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return 0, callErr
	}
	if info.Process == 0 {
		return 0, errors.New("installer process handle unavailable")
	}
	return info.Process, nil
}

func showUpdateError(message string) {
	title, _ := windows.UTF16PtrFromString("OmniCred 更新")
	content, _ := windows.UTF16PtrFromString(message)
	_, _ = windows.MessageBox(0, content, title, windows.MB_OK|windows.MB_ICONERROR)
}
