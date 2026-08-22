//go:build !windows

package appsettings

func installerDatabasePath() string {
	return ""
}
