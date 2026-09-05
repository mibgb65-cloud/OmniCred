//go:build windows

package desktop

import "testing"

func TestUpdateParameters(t *testing.T) {
	got, err := updateParameters(`C:\Program Files\中文目录\OmniCred`, 1234)
	if err != nil || got != `/S /UPDATEPID=1234 /D=C:\Program Files\中文目录\OmniCred` {
		t.Fatalf("parameters = %q, %v", got, err)
	}
	for _, path := range []string{"relative", "C:\\bad\" /S", "C:\\bad\npath"} {
		if _, err := updateParameters(path, 1234); err == nil {
			t.Fatalf("accepted %q", path)
		}
	}
	if _, err := updateParameters(`C:\OmniCred`, 0); err == nil {
		t.Fatal("invalid PID accepted")
	}
}
