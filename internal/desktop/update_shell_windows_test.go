//go:build windows

package desktop

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestInstallerShellReturnsWaitableProcessHandle(t *testing.T) {
	// 使用测试程序模拟安装器，仅验证 Win32 调用和进程句柄，不请求 UAC 或执行安装。
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("OMNICRED_UPDATE_TEST_CHILD", "1")
	handle, err := launchInstallerProcess(executable, "-test.run=^TestInstallerLaunchChildProcess$", "open")
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	status, err := windows.WaitForSingleObject(handle, 10000)
	if err != nil || status != windows.WAIT_OBJECT_0 {
		t.Fatalf("wait = %d, %v", status, err)
	}
	var code uint32
	if err := windows.GetExitCodeProcess(handle, &code); err != nil || code != 0 {
		t.Fatalf("exit = %d, %v", code, err)
	}
}

func TestInstallerLaunchChildProcess(t *testing.T) {
	if os.Getenv("OMNICRED_UPDATE_TEST_CHILD") != "1" {
		return
	}
	// 测试子进程正常退出即验证原用户进程可被正确启动和等待。
}

func TestHelperRejectsIncompleteInvocation(t *testing.T) {
	if RunUpdateHelper([]string{"-db", "test.db"}) {
		t.Fatal("normal startup intercepted")
	}
	for _, arguments := range [][]string{{helperFlag}, {helperFlag, "invalid"}, {helperFlag, "0", "target", "installer", "hash", "directory"}} {
		if !RunUpdateHelper(arguments) {
			t.Fatal("internal mode must not open the application")
		}
	}
}
