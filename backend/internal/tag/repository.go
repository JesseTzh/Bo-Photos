package tag

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

func (r *Repository) Create(ctx context.Context, item Tag) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO tags(id,name,category,parent_id,detail,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?)`, item.ID, item.Name, nullable(item.Category), nullable(item.ParentID),
		nullable(item.Detail), formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	return err
}

func (r *Repository) Update(ctx context.Context, id string, input UpdateInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE tags SET name=?,category=?,detail=?,updated_at=? WHERE id=?`,
		input.Name, nullable(input.Category), nullable(input.Detail), formatTime(time.Now()), id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tags SET category=?,updated_at=? WHERE parent_id=?`,
		input.Name, formatTime(time.Now()), id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Move(ctx context.Context, id, parentID, category string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE tags SET parent_id=?,category=?,updated_at=? WHERE id=?`,
		nullable(parentID), category, formatTime(time.Now()), id)
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
	result, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id string) (Tag, error) {
	return scan(r.db.QueryRowContext(ctx, `SELECT id,name,COALESCE(category,''),COALESCE(parent_id,''),
		COALESCE(detail,''),created_at,updated_at FROM tags WHERE id=?`, id))
}

func (r *Repository) GetByName(ctx context.Context, name string) (Tag, error) {
	return scan(r.db.QueryRowContext(ctx, `SELECT id,name,COALESCE(category,''),COALESCE(parent_id,''),
		COALESCE(detail,''),created_at,updated_at FROM tags WHERE name=?`, name))
}

func (r *Repository) List(ctx context.Context) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,COALESCE(category,''),COALESCE(parent_id,''),
		COALESCE(detail,''),created_at,updated_at FROM tags ORDER BY name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Tag{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) HasChildren(ctx context.Context, id string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tags WHERE parent_id=?`, id).Scan(&count)
	return count > 0, err
}

func (r *Repository) AncestorIDs(ctx context.Context, id string) ([]string, error) {
	ids := []string{}
	current := id
	for current != "" {
		item, err := r.Get(ctx, current)
		if err != nil {
			return nil, err
		}
		ids = append(ids, item.ID)
		current = item.ParentID
	}
	for left, right := 0, len(ids)-1; left < right; left, right = left+1, right-1 {
		ids[left], ids[right] = ids[right], ids[left]
	}
	return ids, nil
}

func (r *Repository) AssetTagIDs(ctx context.Context, assetID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT at.tag_id FROM asset_tags at JOIN tags t ON t.id=at.tag_id
		WHERE at.asset_id=? ORDER BY CASE WHEN t.parent_id IS NULL THEN 0 ELSE 1 END,t.name,t.id`, assetID)
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

func (r *Repository) ReplaceAssetTags(ctx context.Context, assetID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM asset_tags WHERE asset_id=?`, assetID); err != nil {
		return err
	}
	now := formatTime(time.Now())
	for _, id := range ids {
		if _, err = tx.ExecContext(ctx, `INSERT INTO asset_tags(asset_id,tag_id,created_at) VALUES(?,?,?)`, assetID, id, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) RemoveAssociationsFromAssetsWithTag(ctx context.Context, tagID string, removeIDs []string) error {
	if len(removeIDs) == 0 {
		return nil
	}
	args := make([]any, 0, len(removeIDs)+1)
	for _, id := range removeIDs {
		args = append(args, id)
	}
	args = append(args, tagID)
	_, err := r.db.ExecContext(ctx, `DELETE FROM asset_tags
		WHERE tag_id IN (`+placeholders(len(removeIDs))+`)
		AND asset_id IN (SELECT asset_id FROM asset_tags WHERE tag_id=?)`, args...)
	return err
}

func (r *Repository) ResyncAllAssetAncestors(ctx context.Context) error {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT asset_id FROM asset_tags`)
	if err != nil {
		return err
	}
	var assets []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		assets = append(assets, id)
	}
	rows.Close()
	for _, assetID := range assets {
		current, err := r.AssetTagIDs(ctx, assetID)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		var expanded []string
		for _, id := range current {
			ancestors, err := r.AncestorIDs(ctx, id)
			if err != nil {
				return err
			}
			for _, ancestor := range ancestors {
				if !seen[ancestor] {
					seen[ancestor] = true
					expanded = append(expanded, ancestor)
				}
			}
		}
		if err := r.ReplaceAssetTags(ctx, assetID, expanded); err != nil {
			return err
		}
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Tag, error) {
	var item Tag
	var created, updated string
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.ParentID, &item.Detail, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Tag{}, ErrNotFound
	}
	if err != nil {
		return Tag{}, fmt.Errorf("scan tag: %w", err)
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return item, nil
}

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func placeholders(count int) string {
	values := make([]string, count)
	for index := range values {
		values[index] = "?"
	}
	return strings.Join(values, ",")
}
