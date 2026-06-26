package album

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, item Album) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO albums
		(id,name,album_value,detail,theme,visible,sort,random_show,license,cover_asset_id,image_sorting,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Name, item.Value, nullable(item.Detail), defaultString(item.Theme, "0"), item.Visible,
		item.Sort, item.RandomShow, nullable(item.License), nullable(item.CoverAssetID),
		defaultInt(item.ImageSorting, 1), formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return ErrDuplicateValue
		}
		return fmt.Errorf("create album: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, id string, item Album) error {
	result, err := r.db.ExecContext(ctx, `UPDATE albums SET
		name=?, album_value=?, detail=?, theme=?, visible=?, sort=?, random_show=?, license=?,
		cover_asset_id=?, image_sorting=?, updated_at=? WHERE id=? AND deleted_at IS NULL`,
		item.Name, item.Value, nullable(item.Detail), defaultString(item.Theme, "0"), item.Visible,
		item.Sort, item.RandomShow, nullable(item.License), nullable(item.CoverAssetID),
		defaultInt(item.ImageSorting, 1), formatTime(time.Now()), id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return ErrDuplicateValue
		}
		return fmt.Errorf("update album: %w", err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE albums SET deleted_at=?, updated_at=? WHERE id=? AND deleted_at IS NULL`,
		formatTime(time.Now()), formatTime(time.Now()), id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id string, public bool) (Album, error) {
	return r.get(ctx, "a.id = ?", id, public)
}

func (r *Repository) GetByValue(ctx context.Context, value string, public bool) (Album, error) {
	return r.get(ctx, "a.album_value = ?", value, public)
}

func (r *Repository) get(ctx context.Context, condition string, value any, public bool) (Album, error) {
	where := condition + " AND a.deleted_at IS NULL"
	if public {
		where += " AND a.visible = 1"
	}
	row := r.db.QueryRowContext(ctx, `SELECT a.id,a.name,a.album_value,COALESCE(a.detail,''),a.theme,
		a.visible,a.sort,a.random_show,COALESCE(a.license,''),COALESCE(a.cover_asset_id,''),
		a.image_sorting,a.created_at,a.updated_at,
		COALESCE(
			(SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
				WHERE aa.album_id=a.id AND aa.asset_id=a.cover_asset_id AND x.status='ready'
				AND x.visible=1 AND x.deleted_at IS NULL LIMIT 1),
			(SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
				WHERE aa.album_id=a.id AND x.status='ready' AND x.visible=1 AND x.deleted_at IS NULL
				ORDER BY aa.sort,x.id LIMIT 1),''),
		(SELECT COUNT(*) FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
			WHERE aa.album_id=a.id AND x.status='ready' AND x.deleted_at IS NULL)
		FROM albums a WHERE `+where, value)
	item, err := scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Album{}, ErrNotFound
	}
	if err != nil {
		return Album{}, err
	}
	item.AssetIDs, err = r.AssetIDs(ctx, item.ID, public)
	if public {
		item.AssetCount = len(item.AssetIDs)
	}
	return item, err
}

func (r *Repository) List(ctx context.Context, public bool) ([]Album, error) {
	where := "a.deleted_at IS NULL"
	if public {
		where += " AND a.visible=1"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT a.id,a.name,a.album_value,COALESCE(a.detail,''),a.theme,
		a.visible,a.sort,a.random_show,COALESCE(a.license,''),COALESCE(a.cover_asset_id,''),
		a.image_sorting,a.created_at,a.updated_at,
		COALESCE(
			(SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
				WHERE aa.album_id=a.id AND aa.asset_id=a.cover_asset_id AND x.status='ready'
				AND x.visible=1 AND x.deleted_at IS NULL LIMIT 1),
			(SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
				WHERE aa.album_id=a.id AND x.status='ready' AND x.visible=1 AND x.deleted_at IS NULL
				ORDER BY aa.sort,x.id LIMIT 1),''),
		(SELECT COUNT(*) FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
			WHERE aa.album_id=a.id AND x.status='ready' AND x.deleted_at IS NULL)
		FROM albums a WHERE `+where+` ORDER BY a.sort,a.created_at,a.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Album{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		if public {
			item.AssetCount, err = r.PublicAssetCount(ctx, item.ID)
			if err != nil {
				return nil, err
			}
		} else {
			item.AssetIDs, err = r.AssetIDs(ctx, item.ID, false)
			if err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ReplaceAssets(ctx context.Context, albumID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM album_assets WHERE album_id=?`, albumID); err != nil {
		return err
	}
	for sort, id := range ids {
		if _, err = tx.ExecContext(ctx, `INSERT INTO album_assets(album_id,asset_id,sort) VALUES(?,?,?)`, albumID, id, sort); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) UpdateOrder(ctx context.Context, orderedIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin album order update: %w", err)
	}
	defer tx.Rollback()
	now := formatTime(time.Now())
	for index, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx, `UPDATE albums SET sort=?, updated_at=? WHERE id=? AND deleted_at IS NULL`, index, now, id); err != nil {
			return fmt.Errorf("update album order: %w", err)
		}
	}
	return tx.Commit()
}

func (r *Repository) SetCover(ctx context.Context, albumID string, assetID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE albums SET cover_asset_id=?, updated_at=? WHERE id=? AND deleted_at IS NULL`,
		nullable(assetID), formatTime(time.Now()), albumID)
	if err != nil {
		return fmt.Errorf("update album cover: %w", err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

type AlbumAsset struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	Title        string `json:"title,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Visible      bool   `json:"visible"`
	Featured     bool   `json:"featured"`
	Sort         int    `json:"sort"`
	PreviewURL   string `json:"preview_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

func (r *Repository) Assets(ctx context.Context, albumID string) ([]AlbumAsset, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT x.id,x.original_name,COALESCE(x.title,''),x.width,x.height,
		x.visible,x.featured,aa.sort,COALESCE(x.preview_key,''),COALESCE(x.thumbnail_key,''),x.derivative_version
		FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
		WHERE aa.album_id=? AND x.status!='purged' AND x.deleted_at IS NULL
		ORDER BY aa.sort ASC,x.created_at DESC,x.id ASC`, albumID)
	if err != nil {
		return nil, fmt.Errorf("list album assets: %w", err)
	}
	defer rows.Close()
	items := []AlbumAsset{}
	for rows.Next() {
		var item AlbumAsset
		var previewKey, thumbnailKey string
		var version int
		if err := rows.Scan(&item.ID, &item.OriginalName, &item.Title, &item.Width, &item.Height,
			&item.Visible, &item.Featured, &item.Sort, &previewKey, &thumbnailKey, &version); err != nil {
			return nil, err
		}
		base := "/media/assets/" + item.ID
		if previewKey != "" {
			item.PreviewURL = base + "/preview?v=" + fmt.Sprint(version)
		}
		if thumbnailKey != "" {
			item.ThumbnailURL = base + "/thumbnail?v=" + fmt.Sprint(version)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateAssetSort(ctx context.Context, albumID string, orders map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin album asset sort update: %w", err)
	}
	defer tx.Rollback()
	for assetID, sortValue := range orders {
		if _, err := tx.ExecContext(ctx, `UPDATE album_assets SET sort=? WHERE album_id=? AND asset_id=?`, sortValue, albumID, assetID); err != nil {
			return fmt.Errorf("update album asset sort: %w", err)
		}
	}
	return tx.Commit()
}

func (r *Repository) ResetAssetSort(ctx context.Context, albumID string) error {
	rows, err := r.db.QueryContext(ctx, `SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
		WHERE aa.album_id=? ORDER BY x.created_at DESC,x.id ASC`, albumID)
	if err != nil {
		return fmt.Errorf("read album assets for reset: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	orders := make(map[string]int, len(ids))
	for index, id := range ids {
		orders[id] = index
	}
	return r.UpdateAssetSort(ctx, albumID, orders)
}

func (r *Repository) AssetIDs(ctx context.Context, albumID string, public bool) ([]string, error) {
	query := `SELECT aa.asset_id FROM album_assets aa`
	if public {
		query += ` JOIN assets x ON x.id=aa.asset_id`
	}
	query += ` WHERE aa.album_id=?`
	if public {
		query += ` AND x.status='ready' AND x.visible=1 AND x.deleted_at IS NULL`
	}
	query += ` ORDER BY aa.sort,aa.asset_id`
	rows, err := r.db.QueryContext(ctx, query, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) PublicAssetCount(ctx context.Context, albumID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM album_assets aa
		JOIN assets x ON x.id=aa.asset_id
		WHERE aa.album_id=? AND x.status='ready' AND x.visible=1 AND x.deleted_at IS NULL`, albumID).Scan(&count)
	return count, err
}

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Album, error) {
	var item Album
	var visible, storedRandom bool
	var created, updated, effectiveCover string
	err := row.Scan(&item.ID, &item.Name, &item.Value, &item.Detail, &item.Theme, &visible,
		&item.Sort, &storedRandom, &item.License, &item.CoverAssetID, &item.ImageSorting,
		&created, &updated, &effectiveCover, &item.AssetCount)
	item.Visible = visible
	item.RandomShow = storedRandom
	item.EffectiveCoverAssetID = effectiveCover
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return item, err
}

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}
func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
