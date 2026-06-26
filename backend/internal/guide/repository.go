package guide

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, item Guide) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO guides
		(id,title,country,city,days,start_date,end_date,cover_asset_id,published,sort,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Title, item.Country, item.City, item.Days, nullableTime(item.StartDate),
		nullableTime(item.EndDate), nullableString(item.CoverAssetID), item.Published, item.Sort,
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	return err
}

func (r *Repository) Get(ctx context.Context, id string, public bool) (Guide, error) {
	where := "id=? AND deleted_at IS NULL"
	if public {
		where += " AND published=1"
	}
	item, err := scanGuide(r.db.QueryRowContext(ctx, `SELECT id,title,country,city,days,start_date,end_date,
		COALESCE(cover_asset_id,''),published,sort,created_at,updated_at FROM guides WHERE `+where, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Guide{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) List(ctx context.Context, public bool) ([]Guide, error) {
	where := "deleted_at IS NULL"
	if public {
		where += " AND published=1"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,title,country,city,days,start_date,end_date,
		COALESCE(cover_asset_id,''),published,sort,created_at,updated_at
		FROM guides WHERE `+where+` ORDER BY sort,created_at DESC,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Guide{}
	for rows.Next() {
		item, err := scanGuide(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id string, item Guide) error {
	result, err := r.db.ExecContext(ctx, `UPDATE guides SET title=?,country=?,city=?,days=?,start_date=?,
		end_date=?,cover_asset_id=?,published=?,sort=?,updated_at=? WHERE id=? AND deleted_at IS NULL`,
		item.Title, item.Country, item.City, item.Days, nullableTime(item.StartDate), nullableTime(item.EndDate),
		nullableString(item.CoverAssetID), item.Published, item.Sort, formatTime(time.Now()), id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE guides SET deleted_at=?,updated_at=?
		WHERE id=? AND deleted_at IS NULL`, formatTime(time.Now()), formatTime(time.Now()), id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Sort(ctx context.Context, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for index, id := range ids {
		result, err := tx.ExecContext(ctx, `UPDATE guides SET sort=?,updated_at=?
			WHERE id=? AND deleted_at IS NULL`, index, formatTime(time.Now()), id)
		if err != nil {
			return err
		}
		count, _ := result.RowsAffected()
		if count != 1 {
			return ErrInvalidOwnership
		}
	}
	return tx.Commit()
}

func (r *Repository) CoverAvailable(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return true, nil
	}
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM assets
		WHERE id=? AND status!='purged' AND purged_at IS NULL`, id).Scan(&count)
	return count == 1, err
}

type scanner interface{ Scan(...any) error }

func scanGuide(row scanner) (Guide, error) {
	var item Guide
	var start, end sql.NullString
	var created, updated string
	err := row.Scan(&item.ID, &item.Title, &item.Country, &item.City, &item.Days, &start, &end,
		&item.CoverAssetID, &item.Published, &item.Sort, &created, &updated)
	if err != nil {
		return Guide{}, err
	}
	item.StartDate = parseNullableTime(start)
	item.EndDate = parseNullableTime(end)
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return item, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}
func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
func wrap(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
