//go:build windows

package appsettings

import "golang.org/x/sys/windows/registry"

func installerDatabasePath() string {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\OmniCred`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()

	path, _, err := key.GetStringValue("InstallerDatabasePath")
	if err != nil {
		return ""
	}
	return path
}
