package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	closeWithError := func(cause error) (*sql.DB, error) {
		_ = db.Close()
		return nil, cause
	}
	if err := db.PingContext(ctx); err != nil {
		return closeWithError(fmt.Errorf("ping database: %w", err))
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return closeWithError(fmt.Errorf("set busy timeout: %w", err))
	}
	if err := migrate(ctx, db); err != nil {
		return closeWithError(err)
	}
	if err := os.Chmod(path, 0o600); err != nil && !os.IsNotExist(err) {
		return closeWithError(fmt.Errorf("restrict database permissions: %w", err))
	}
	return db, nil
}
