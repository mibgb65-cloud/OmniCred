package appsettings

import "time"

var (
	ErrSameDatabasePath = newValidationError("database_path", "must be different from the current path")
	ErrTargetExists     = newValidationError("database_path", "already exists")
)

type RuntimeStatus struct {
	Version             string    `json:"version"`
	DatabasePath        string    `json:"database_path"`
	PendingDatabasePath string    `json:"pending_database_path,omitempty"`
	ConfigPath          string    `json:"config_path"`
	APIAddress          string    `json:"api_address"`
	RepositoryURL       string    `json:"repository_url"`
	StartedAt           time.Time `json:"started_at"`
	UptimeSeconds       int64     `json:"uptime_seconds"`
	DatabaseOK          bool      `json:"database_ok"`
	CredentialCount     int       `json:"credential_count"`
	UninstallAvailable  bool      `json:"uninstall_available"`
}

type StorageInput struct {
	DatabasePath string `json:"database_path"`
}

type StorageResult struct {
	DatabasePath    string `json:"database_path"`
	RestartRequired bool   `json:"restart_required"`
}

type UpdateInfo struct {
	CurrentVersion  string     `json:"current_version"`
	LatestVersion   string     `json:"latest_version"`
	UpdateAvailable bool       `json:"update_available"`
	ReleaseURL      string     `json:"release_url"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	CheckedAt       time.Time  `json:"checked_at"`
	Status          string     `json:"status"`
}

type ValidationError struct {
	Field   string
	Message string
}

func newValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

func (err *ValidationError) Error() string {
	return err.Field + ": " + err.Message
}
