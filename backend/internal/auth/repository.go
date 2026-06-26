package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAlreadyInitialized = errors.New("administrator already initialized")
	ErrNotInitialized     = errors.New("administrator not initialized")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IsInitialized(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM administrators WHERE id = 1)`,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check administrator: %w", err)
	}
	return exists, nil
}

func (r *Repository) Create(ctx context.Context, passwordHash string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO administrators (id, password_hash, created_at, updated_at)
		 SELECT 1, ?, ?, ?
		 WHERE NOT EXISTS (SELECT 1 FROM administrators WHERE id = 1)`,
		passwordHash,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("create administrator: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read administrator insert result: %w", err)
	}
	if rows != 1 {
		return ErrAlreadyInitialized
	}
	return nil
}

func (r *Repository) PasswordHash(ctx context.Context) (string, error) {
	var hash string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT password_hash FROM administrators WHERE id = 1`,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotInitialized
	}
	if err != nil {
		return "", fmt.Errorf("read administrator password: %w", err)
	}
	return hash, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, passwordHash string) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE administrators SET password_hash = ?, updated_at = ? WHERE id = 1`,
		passwordHash,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("update administrator password: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read administrator update result: %w", err)
	}
	if rows != 1 {
		return ErrNotInitialized
	}
	return nil
}

func (r *Repository) DeleteAllSessions(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions`); err != nil {
		return fmt.Errorf("delete all sessions: %w", err)
	}
	return nil
}
