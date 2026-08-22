package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"omnicred/internal/platform"
)

func (store *Store) CreatePlatform(ctx context.Context, item platform.Platform) (platform.Platform, error) {
	result, err := store.db.ExecContext(ctx, `
		INSERT INTO platforms (name, created_at, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(name) DO NOTHING`, item.Name, formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	if err != nil {
		return platform.Platform{}, fmt.Errorf("insert platform: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return platform.Platform{}, fmt.Errorf("read platform insert result: %w", err)
	}
	if count == 0 {
		return platform.Platform{}, platform.ErrAlreadyExists
	}
	item.ID, err = result.LastInsertId()
	if err != nil {
		return platform.Platform{}, fmt.Errorf("read platform id: %w", err)
	}
	return item, nil
}

func (store *Store) ListPlatforms(ctx context.Context) ([]platform.Platform, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.created_at, p.updated_at, COUNT(c.id)
		FROM platforms p
		LEFT JOIN credentials c ON LOWER(c.provider) = LOWER(p.name)
		GROUP BY p.id, p.name, p.created_at, p.updated_at
		ORDER BY LOWER(p.name), p.id`)
	if err != nil {
		return nil, fmt.Errorf("list platforms: %w", err)
	}
	defer rows.Close()

	items := make([]platform.Platform, 0)
	for rows.Next() {
		item, err := scanPlatform(rows)
		if err != nil {
			return nil, fmt.Errorf("scan platform list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platforms: %w", err)
	}
	return items, nil
}

func (store *Store) UpdatePlatform(ctx context.Context, id int64, name string, updatedAt time.Time) (platform.Platform, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return platform.Platform{}, fmt.Errorf("begin platform update: %w", err)
	}
	defer tx.Rollback()

	var oldName, createdAt string
	if err := tx.QueryRowContext(ctx, "SELECT name, created_at FROM platforms WHERE id = ?", id).Scan(&oldName, &createdAt); errors.Is(err, sql.ErrNoRows) {
		return platform.Platform{}, platform.ErrNotFound
	} else if err != nil {
		return platform.Platform{}, fmt.Errorf("select platform: %w", err)
	}
	var duplicate int
	err = tx.QueryRowContext(ctx, "SELECT 1 FROM platforms WHERE name = ? COLLATE NOCASE AND id <> ?", name, id).Scan(&duplicate)
	if err == nil {
		return platform.Platform{}, platform.ErrAlreadyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return platform.Platform{}, fmt.Errorf("check platform name: %w", err)
	}

	formatted := formatTime(updatedAt)
	if _, err := tx.ExecContext(ctx, "UPDATE platforms SET name = ?, updated_at = ? WHERE id = ?", name, formatted, id); err != nil {
		return platform.Platform{}, fmt.Errorf("update platform: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE credentials SET provider = ?, updated_at = ? WHERE LOWER(provider) = LOWER(?)", name, formatted, oldName); err != nil {
		return platform.Platform{}, fmt.Errorf("rename credential platform: %w", err)
	}
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM credentials WHERE LOWER(provider) = LOWER(?)", name).Scan(&count); err != nil {
		return platform.Platform{}, fmt.Errorf("count platform credentials: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return platform.Platform{}, fmt.Errorf("parse platform created_at: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return platform.Platform{}, fmt.Errorf("commit platform update: %w", err)
	}
	return platform.Platform{ID: id, Name: name, CredentialCount: count, CreatedAt: created, UpdatedAt: updatedAt}, nil
}

func (store *Store) DeletePlatform(ctx context.Context, id int64) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin platform delete: %w", err)
	}
	defer tx.Rollback()

	var name string
	if err := tx.QueryRowContext(ctx, "SELECT name FROM platforms WHERE id = ?", id).Scan(&name); errors.Is(err, sql.ErrNoRows) {
		return platform.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("select platform for delete: %w", err)
	}
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM credentials WHERE LOWER(provider) = LOWER(?)", name).Scan(&count); err != nil {
		return fmt.Errorf("count platform credentials: %w", err)
	}
	if count > 0 {
		return platform.ErrInUse
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM platforms WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete platform: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit platform delete: %w", err)
	}
	return nil
}

func scanPlatform(row scanner) (platform.Platform, error) {
	var item platform.Platform
	var createdAt, updatedAt string
	if err := row.Scan(&item.ID, &item.Name, &createdAt, &updatedAt, &item.CredentialCount); err != nil {
		return platform.Platform{}, err
	}
	var err error
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return platform.Platform{}, fmt.Errorf("parse created_at: %w", err)
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return platform.Platform{}, fmt.Errorf("parse updated_at: %w", err)
	}
	return item, nil
}

var _ platform.Store = (*Store)(nil)
