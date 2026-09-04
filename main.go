package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"omnicred/internal/apiserver"
	"omnicred/internal/appsettings"
	"omnicred/internal/credential"
	"omnicred/internal/desktop"
	"omnicred/internal/httpapi"
	"omnicred/internal/identity"
	"omnicred/internal/platform"
	"omnicred/internal/sqlite"
)

//go:embed all:frontend/dist
var assets embed.FS

var appVersion = "0.3.2"

const (
	repositoryURL = "https://github.com/mibgb65-cloud/OmniCred"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("OmniCred stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseOverride := flag.String("db", "", "SQLite database path (overrides saved settings)")
	flag.Parse()
	databasePath, configPath, err := appsettings.ResolveDatabasePath(*databaseOverride, developmentBuild)
	if err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	db, err := sqlite.Open(initCtx, databasePath)
	cancel()
	if err != nil {
		return err
	}
	defer db.Close()

	store := sqlite.NewStore(db)
	credentialService := credential.NewService(store)
	identityService := identity.NewService(store)
	platformService := platform.NewService(store)
	uninstallerPath, uninstallAvailable := desktop.DetectUninstaller()
	settingsService := appsettings.NewService(db, databasePath, configPath, appVersion, apiAddress, repositoryURL, uninstallAvailable)
	api := httpapi.New(credentialService, identityService, platformService, settingsService, logger)
	server := apiserver.New(apiAddress, api, logger)
	app := desktop.New(server, logger, databasePath, uninstallerPath)

	err = wails.Run(&options.App{
		Title:            applicationTitle,
		Width:            1280,
		Height:           820,
		MinWidth:         760,
		MinHeight:        560,
		DisableResize:    false,
		Frameless:        true,
		WindowStartState: options.Normal,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: options.NewRGB(7, 18, 15),
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind:             []interface{}{app},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: singleInstanceID,
			OnSecondInstanceLaunch: func(_ options.SecondInstanceData) {
				app.SecondInstance()
			},
		},
		EnableDefaultContextMenu: false,
	})
	if err != nil {
		return fmt.Errorf("run desktop window: %w", err)
	}
	return nil
}
