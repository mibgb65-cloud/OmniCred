package appsettings

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	db                 *sql.DB
	currentPath        string
	configPath         string
	version            string
	apiAddress         string
	repositoryURL      string
	releasesEndpoint   string
	startedAt          time.Time
	client             *http.Client
	mu                 sync.RWMutex
	pendingPath        string
	uninstallAvailable bool
}

func NewService(db *sql.DB, databasePath, configPath, version, apiAddress, repositoryURL string, uninstallAvailable bool) *Service {
	return &Service{
		db: db, currentPath: databasePath, configPath: configPath, version: version,
		apiAddress: apiAddress, repositoryURL: repositoryURL,
		releasesEndpoint: "https://api.github.com/repos/mibgb65-cloud/OmniCred/releases/latest",
		startedAt:        time.Now().UTC(), client: &http.Client{Timeout: 12 * time.Second},
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
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, service.releasesEndpoint, nil)
	if err != nil {
		return UpdateInfo{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "OmniCred/"+service.version)
	response, err := service.client.Do(request)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("check GitHub release: %w", err)
	}
	defer response.Body.Close()
	now := time.Now().UTC()
	if response.StatusCode == http.StatusNotFound {
		return UpdateInfo{
			CurrentVersion: service.version, LatestVersion: service.version, ReleaseURL: service.repositoryURL + "/releases",
			CheckedAt: now, Status: "no_releases",
		}, nil
	}
	if response.StatusCode != http.StatusOK {
		return UpdateInfo{}, fmt.Errorf("GitHub release API returned %d", response.StatusCode)
	}
	var release struct {
		TagName     string    `json:"tag_name"`
		HTMLURL     string    `json:"html_url"`
		PublishedAt time.Time `json:"published_at"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return UpdateInfo{}, fmt.Errorf("decode GitHub release: %w", err)
	}
	return UpdateInfo{
		CurrentVersion: service.version, LatestVersion: release.TagName,
		UpdateAvailable: compareVersion(release.TagName, service.version) > 0,
		ReleaseURL:      release.HTMLURL, PublishedAt: &release.PublishedAt, CheckedAt: now, Status: "ok",
	}, nil
}

func samePath(left, right string) bool {
	leftPath, _ := filepath.Abs(left)
	rightPath, _ := filepath.Abs(right)
	return strings.EqualFold(filepath.Clean(leftPath), filepath.Clean(rightPath))
}

func compareVersion(left, right string) int {
	parse := func(value string) [3]int {
		value = strings.TrimPrefix(strings.TrimSpace(value), "v")
		value = strings.SplitN(value, "-", 2)[0]
		parts := strings.Split(value, ".")
		var result [3]int
		for index := 0; index < len(parts) && index < 3; index++ {
			result[index], _ = strconv.Atoi(parts[index])
		}
		return result
	}
	a, b := parse(left), parse(right)
	for index := range a {
		if a[index] > b[index] {
			return 1
		}
		if a[index] < b[index] {
			return -1
		}
	}
	return 0
}
