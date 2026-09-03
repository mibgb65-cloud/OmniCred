package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"omnicred/internal/identity"
)

const identityColumns = `id, country, full_name, localized_name, first_name, middle_name, last_name,
	gender, birth_date, street_address, city, region, postal_code, phone, email, password, created_at, updated_at`

func (store *Store) CreateIdentity(ctx context.Context, item identity.Profile) (identity.Profile, error) {
	result, err := store.db.ExecContext(ctx, `
		INSERT INTO identity_profiles (
			country, full_name, localized_name, first_name, middle_name, last_name, gender, birth_date,
			street_address, city, region, postal_code, phone, email, password, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Country, item.FullName, item.LocalizedName, item.FirstName, item.MiddleName, item.LastName,
		item.Gender, item.BirthDate, item.StreetAddress, item.City, item.Region, item.PostalCode,
		item.Phone, item.Email, item.Password, formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	if err != nil {
		return identity.Profile{}, fmt.Errorf("insert identity profile: %w", err)
	}
	item.ID, err = result.LastInsertId()
	if err != nil {
		return identity.Profile{}, fmt.Errorf("read inserted identity id: %w", err)
	}
	return item, nil
}

func (store *Store) GetIdentity(ctx context.Context, id int64) (identity.Profile, error) {
	row := store.db.QueryRowContext(ctx, "SELECT "+identityColumns+" FROM identity_profiles WHERE id = ?", id)
	item, err := scanIdentity(row)
	if errors.Is(err, sql.ErrNoRows) {
		return identity.Profile{}, identity.ErrNotFound
	}
	if err != nil {
		return identity.Profile{}, fmt.Errorf("select identity profile: %w", err)
	}
	return item, nil
}

func (store *Store) ListIdentities(ctx context.Context, filter identity.Filter) ([]identity.Profile, error) {
	query := "SELECT " + identityColumns + " FROM identity_profiles WHERE 1 = 1"
	args := make([]any, 0, 2)
	if filter.Country != "" {
		query += " AND country = ? COLLATE NOCASE"
		args = append(args, filter.Country)
	}
	if filter.Query != "" {
		query += ` AND LOWER(country || ' ' || full_name || ' ' || localized_name || ' ' || first_name || ' ' ||
			middle_name || ' ' || last_name || ' ' || city || ' ' || region || ' ' || phone || ' ' || email) LIKE LOWER(?)`
		args = append(args, "%"+filter.Query+"%")
	}
	query += " ORDER BY id ASC"

	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list identity profiles: %w", err)
	}
	defer rows.Close()

	items := make([]identity.Profile, 0)
	for rows.Next() {
		item, err := scanIdentity(rows)
		if err != nil {
			return nil, fmt.Errorf("scan identity profile list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity profiles: %w", err)
	}
	return items, nil
}

func (store *Store) UpdateIdentity(ctx context.Context, item identity.Profile) (identity.Profile, error) {
	result, err := store.db.ExecContext(ctx, `
		UPDATE identity_profiles SET
			country = ?, full_name = ?, localized_name = ?, first_name = ?, middle_name = ?, last_name = ?,
			gender = ?, birth_date = ?, street_address = ?, city = ?, region = ?, postal_code = ?,
			phone = ?, email = ?, password = ?, updated_at = ?
		WHERE id = ?`,
		item.Country, item.FullName, item.LocalizedName, item.FirstName, item.MiddleName, item.LastName,
		item.Gender, item.BirthDate, item.StreetAddress, item.City, item.Region, item.PostalCode,
		item.Phone, item.Email, item.Password, formatTime(item.UpdatedAt), item.ID,
	)
	if err != nil {
		return identity.Profile{}, fmt.Errorf("update identity profile: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return identity.Profile{}, fmt.Errorf("read affected identity rows: %w", err)
	}
	if count == 0 {
		return identity.Profile{}, identity.ErrNotFound
	}
	return item, nil
}

func (store *Store) DeleteIdentity(ctx context.Context, id int64) error {
	result, err := store.db.ExecContext(ctx, "DELETE FROM identity_profiles WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete identity profile: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected identity rows: %w", err)
	}
	if count == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func scanIdentity(row scanner) (identity.Profile, error) {
	var item identity.Profile
	var createdAt, updatedAt string
	err := row.Scan(
		&item.ID, &item.Country, &item.FullName, &item.LocalizedName, &item.FirstName, &item.MiddleName,
		&item.LastName, &item.Gender, &item.BirthDate, &item.StreetAddress, &item.City, &item.Region,
		&item.PostalCode, &item.Phone, &item.Email, &item.Password, &createdAt, &updatedAt,
	)
	if err != nil {
		return identity.Profile{}, err
	}
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return identity.Profile{}, fmt.Errorf("parse identity created_at: %w", err)
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return identity.Profile{}, fmt.Errorf("parse identity updated_at: %w", err)
	}
	return item, nil
}

var _ identity.Store = (*Store)(nil)
