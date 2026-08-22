package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"omnicred/internal/credential"
)

const credentialColumns = "id, provider, account, username, password, created_at, updated_at"

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (store *Store) Create(ctx context.Context, item credential.Credential) (credential.Credential, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return credential.Credential{}, fmt.Errorf("begin credential insert: %w", err)
	}
	defer tx.Rollback()
	if err := ensurePlatform(ctx, tx, item.Provider, item.CreatedAt); err != nil {
		return credential.Credential{}, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO credentials (provider, account, username, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		item.Provider, item.Account, item.Username, item.Password,
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	if err != nil {
		return credential.Credential{}, fmt.Errorf("insert credential: %w", err)
	}
	item.ID, err = result.LastInsertId()
	if err != nil {
		return credential.Credential{}, fmt.Errorf("read inserted id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return credential.Credential{}, fmt.Errorf("commit credential insert: %w", err)
	}
	return item, nil
}

func (store *Store) Get(ctx context.Context, id int64) (credential.Credential, error) {
	row := store.db.QueryRowContext(ctx, "SELECT "+credentialColumns+" FROM credentials WHERE id = ?", id)
	item, err := scanCredential(row)
	if errors.Is(err, sql.ErrNoRows) {
		return credential.Credential{}, credential.ErrNotFound
	}
	if err != nil {
		return credential.Credential{}, fmt.Errorf("select credential: %w", err)
	}
	return item, nil
}

func (store *Store) List(ctx context.Context, filter credential.Filter) ([]credential.Credential, error) {
	query := "SELECT " + credentialColumns + " FROM credentials WHERE 1 = 1"
	args := make([]any, 0, 3)
	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
	}
	if filter.Query != "" {
		query += " AND (LOWER(account) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?))"
		pattern := "%" + filter.Query + "%"
		args = append(args, pattern, pattern)
	}
	query += " ORDER BY id ASC"

	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list credentials: %w", err)
	}
	defer rows.Close()

	items := make([]credential.Credential, 0)
	for rows.Next() {
		item, err := scanCredential(rows)
		if err != nil {
			return nil, fmt.Errorf("scan credential list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credentials: %w", err)
	}
	return items, nil
}

func (store *Store) Update(ctx context.Context, item credential.Credential) (credential.Credential, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return credential.Credential{}, fmt.Errorf("begin credential update: %w", err)
	}
	defer tx.Rollback()
	if err := ensurePlatform(ctx, tx, item.Provider, item.UpdatedAt); err != nil {
		return credential.Credential{}, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE credentials
		SET provider = ?, account = ?, username = ?, password = ?, updated_at = ?
		WHERE id = ?`,
		item.Provider, item.Account, item.Username, item.Password, formatTime(item.UpdatedAt), item.ID,
	)
	if err != nil {
		return credential.Credential{}, fmt.Errorf("update credential: %w", err)
	}
	if err := requireAffectedRow(result); err != nil {
		return credential.Credential{}, err
	}
	if err := tx.Commit(); err != nil {
		return credential.Credential{}, fmt.Errorf("commit credential update: %w", err)
	}
	return item, nil
}

func ensurePlatform(ctx context.Context, tx *sql.Tx, name string, timestamp time.Time) error {
	formatted := formatTime(timestamp)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platforms (name, created_at, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(name) DO NOTHING`, name, formatted, formatted)
	if err != nil {
		return fmt.Errorf("ensure credential platform: %w", err)
	}
	return nil
}

func (store *Store) Delete(ctx context.Context, id int64) error {
	result, err := store.db.ExecContext(ctx, "DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete credential: %w", err)
	}
	return requireAffectedRow(result)
}

type scanner interface {
	Scan(...any) error
}

func scanCredential(row scanner) (credential.Credential, error) {
	var item credential.Credential
	var createdAt, updatedAt string
	err := row.Scan(&item.ID, &item.Provider, &item.Account, &item.Username, &item.Password, &createdAt, &updatedAt)
	if err != nil {
		return credential.Credential{}, err
	}
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return credential.Credential{}, fmt.Errorf("parse created_at: %w", err)
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return credential.Credential{}, fmt.Errorf("parse updated_at: %w", err)
	}
	return item, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func requireAffectedRow(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if count == 0 {
		return credential.ErrNotFound
	}
	return nil
}

var _ credential.Store = (*Store)(nil)
