package appsettings

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"omnicred/internal/updater"
)

type Service struct {
	db                 *sql.DB
	currentPath        string
	configPath         string
	version            string
	apiAddress         string
	repositoryURL      string
	startedAt          time.Time
	updater            *updater.Manager
	mu                 sync.RWMutex
	pendingPath        string
	uninstallAvailable bool
}

func NewService(db *sql.DB, databasePath, configPath, version, apiAddress, repositoryURL string, uninstallAvailable bool, updates *updater.Manager) *Service {
	if updates == nil {
		updates = updater.New(version, false)
	}
	return &Service{
		db: db, currentPath: databasePath, configPath: configPath, version: version,
		apiAddress: apiAddress, repositoryURL: repositoryURL,
		startedAt: time.Now().UTC(), updater: updates,
		uninstallAvailable: uninstallAvailable,
	}
}

func (service *Service) Status(ctx context.Context) (RuntimeStatus, error) {
	var count int
	databaseErr := service.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM credentials").Scan(&count)
	service.mu.RLock()
	pendingPath := service.pendingPath
	service.mu.RUnlock()
	now := time.Now().UTC()
	return RuntimeStatus{
		Version: service.version, DatabasePath: service.currentPath, PendingDatabasePath: pendingPath,
		ConfigPath: service.configPath, APIAddress: service.apiAddress, RepositoryURL: service.repositoryURL,
		StartedAt: service.startedAt, UptimeSeconds: int64(now.Sub(service.startedAt).Seconds()),
		DatabaseOK: databaseErr == nil, CredentialCount: count, UninstallAvailable: service.uninstallAvailable,
	}, nil
}

func (service *Service) MigrateStorage(ctx context.Context, input StorageInput) (StorageResult, error) {
	target := strings.TrimSpace(input.DatabasePath)
	if target == "" {
		return StorageResult{}, newValidationError("database_path", "is required")
	}
	absolute, err := filepath.Abs(target)
	if err != nil {
		return StorageResult{}, newValidationError("database_path", "is invalid")
	}
	if !strings.EqualFold(filepath.Ext(absolute), ".db") {
		return StorageResult{}, newValidationError("database_path", "must use the .db extension")
	}
	if samePath(absolute, service.currentPath) {
		return StorageResult{}, ErrSameDatabasePath
	}
	if _, err := os.Stat(absolute); err == nil {
		return StorageResult{}, ErrTargetExists
	} else if !os.IsNotExist(err) {
		return StorageResult{}, fmt.Errorf("inspect target database: %w", err)
	}

	service.mu.Lock()
	defer service.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(absolute), 0o700); err != nil {
		return StorageResult{}, fmt.Errorf("create target directory: %w", err)
	}
	statement := "VACUUM INTO '" + strings.ReplaceAll(filepath.Clean(absolute), "'", "''") + "'"
	if _, err := service.db.ExecContext(ctx, statement); err != nil {
		return StorageResult{}, fmt.Errorf("copy database to new location: %w", err)
	}
	if err := os.Chmod(absolute, 0o600); err != nil {
		_ = os.Remove(absolute)
		return StorageResult{}, fmt.Errorf("restrict target database permissions: %w", err)
	}
	if err := saveDatabasePath(service.configPath, absolute); err != nil {
		_ = os.Remove(absolute)
		return StorageResult{}, err
	}
	service.pendingPath = absolute
	return StorageResult{DatabasePath: absolute, RestartRequired: true}, nil
}

func (service *Service) CheckUpdate(ctx context.Context) (UpdateInfo, error) {
	return service.updater.Check(ctx)
}

func samePath(left, right string) bool {
	leftPath, _ := filepath.Abs(left)
	rightPath, _ := filepath.Abs(right)
	return strings.EqualFold(filepath.Clean(leftPath), filepath.Clean(rightPath))
}
