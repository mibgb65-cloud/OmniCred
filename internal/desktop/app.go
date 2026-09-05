package desktop

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"omnicred/internal/apiserver"
	"omnicred/internal/updater"
)

type App struct {
	server           *apiserver.Server
	logger           *slog.Logger
	mu               sync.RWMutex
	ctx              context.Context
	databasePath     string
	uninstallerPath  string
	startUninstaller func(string) error
	updater          *updater.Manager
	startInstaller   func(string) error
}

func New(server *apiserver.Server, logger *slog.Logger, databasePath, uninstallerPath string, updates *updater.Manager) *App {
	return &App{
		server: server, logger: logger, databasePath: databasePath,
		uninstallerPath: uninstallerPath, startUninstaller: startElevatedUninstaller,
		updater: updates, startInstaller: startUpdateInstaller,
	}
}

func DetectUninstaller() (string, bool) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", false
	}
	uninstallerPath := filepath.Join(filepath.Dir(executablePath), "uninstall.exe")
	return uninstallerPath, isUninstallerAvailable(uninstallerPath)
}

func isUninstallerAvailable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func (app *App) ChooseDatabasePath() (string, error) {
	app.mu.RLock()
	ctx := app.ctx
	app.mu.RUnlock()
	if ctx == nil {
		return "", nil
	}
	return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		Title:                "选择新的数据文件位置 / Choose database location",
		DefaultDirectory:     filepath.Dir(app.databasePath),
		DefaultFilename:      "omnicred.db",
		CanCreateDirectories: true,
		Filters:              []runtime.FileFilter{{DisplayName: "SQLite Database (*.db)", Pattern: "*.db"}},
	})
}

func (app *App) Uninstall() error {
	app.mu.RLock()
	ctx := app.ctx
	uninstallerPath := app.uninstallerPath
	startUninstaller := app.startUninstaller
	app.mu.RUnlock()
	if !isUninstallerAvailable(uninstallerPath) {
		return fmt.Errorf("uninstaller is not available")
	}
	if err := startUninstaller(uninstallerPath); err != nil {
		return fmt.Errorf("start uninstaller: %w", err)
	}
	if ctx != nil {
		go func() {
			time.Sleep(300 * time.Millisecond)
			runtime.Quit(ctx)
		}()
	}
	return nil
}

func (app *App) Startup(ctx context.Context) {
	app.mu.Lock()
	app.ctx = ctx
	app.mu.Unlock()
	if err := app.server.Start(); err != nil {
		app.logger.Error("unable to start local API", "error", err)
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "OmniCred 无法启动",
			Message: "本地 API 端口 8787 无法使用。请关闭占用该端口的程序后重试。",
		})
		runtime.Quit(ctx)
	}
}

func (app *App) Shutdown(_ context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("unable to stop local API cleanly", "error", err)
	}
}

func (app *App) SecondInstance() {
	app.mu.RLock()
	ctx := app.ctx
	app.mu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.WindowUnminimise(ctx)
	runtime.WindowShow(ctx)
}
