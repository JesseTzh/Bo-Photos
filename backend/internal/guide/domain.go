package guide

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

type Module struct {
	ID             string          `json:"id"`
	GuideID        string          `json:"guide_id"`
	Name           string          `json:"name"`
	Kind           string          `json:"kind"`
	Template       ModuleTemplate  `json:"template,omitempty"`
	DataVersion    int             `json:"data_version"`
	Sort           int             `json:"sort"`
	StructuredData json.RawMessage `json:"structured_data,omitempty"`
	Hidden         bool            `json:"hidden"`
	Blocks         []Block         `json:"blocks,omitempty"`
}
type ModuleInput struct {
	Name           string          `json:"name"`
	Kind           string          `json:"kind"`
	Template       ModuleTemplate  `json:"template"`
	DataVersion    int             `json:"data_version"`
	StructuredData json.RawMessage `json:"structured_data"`
	Hidden         bool            `json:"hidden"`
}
type Block struct {
	ID          string          `json:"id"`
	ModuleID    string          `json:"module_id"`
	Type        BlockType       `json:"type"`
	DataVersion int             `json:"data_version"`
	Sort        int             `json:"sort"`
	Data        json.RawMessage `json:"data"`
}
type BlockInput struct {
	Type        BlockType       `json:"type"`
	DataVersion int             `json:"data_version"`
	Data        json.RawMessage `json:"data"`
}
type TOCItem struct {
	ID             string `json:"id"`
	GuideID        string `json:"guide_id"`
	Title          string `json:"title"`
	TargetModuleID string `json:"target_module_id"`
	Level          int    `json:"level"`
	Sort           int    `json:"sort"`
	Hidden         bool   `json:"hidden"`
}
type TOCInput struct {
	Title          string `json:"title"`
	Level          int    `json:"level"`
	TargetModuleID string `json:"target_module_id"`
	Hidden         bool   `json:"hidden"`
}
type GuideAlbum struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	CoverURL string `json:"cover_url"`
	Sort     int    `json:"sort"`
}

func (r *Repository) CreateModule(ctx context.Context, guideID string, input ModuleInput) (Module, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return Module{}, ErrInvalidName
	}
	if _, err := r.Get(ctx, guideID, false); err != nil {
		return Module{}, err
	}
	if input.DataVersion == 0 {
		input.DataVersion = 1
	}
	var data any
	if input.Kind == "structured" {
		normalized, err := ValidateStructuredData(input.Template, input.DataVersion, input.StructuredData)
		if err != nil {
			return Module{}, err
		}
		data = string(normalized)
	} else if input.Kind != "content" {
		return Module{}, ErrInvalidContent
	}
	var sort int
	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort),-1)+1 FROM guide_modules WHERE guide_id=?`, guideID).Scan(&sort)
	item := Module{ID: newID(), GuideID: guideID, Name: input.Name, Kind: input.Kind, Template: input.Template, DataVersion: input.DataVersion, StructuredData: input.StructuredData, Sort: sort, Hidden: input.Hidden}
	_, err := r.db.ExecContext(ctx, `INSERT INTO guide_modules(id,guide_id,name,kind,template,data_version,structured_data,sort,hidden,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, guideID, input.Name, input.Kind, nullableString(string(input.Template)), input.DataVersion, data, sort, input.Hidden, formatTime(time.Now()), formatTime(time.Now()))
	return item, err
}

func (r *Repository) CreateBlock(ctx context.Context, guideID, moduleID string, input BlockInput) (Block, error) {
	var kind string
	if err := r.db.QueryRowContext(ctx, `SELECT kind FROM guide_modules WHERE id=? AND guide_id=?`, moduleID, guideID).Scan(&kind); err != nil {
		return Block{}, ErrInvalidOwnership
	}
	if kind != "content" {
		return Block{}, ErrInvalidContent
	}
	if input.DataVersion == 0 {
		input.DataVersion = 1
	}
	normalized, err := ValidateBlockData(input.Type, input.DataVersion, input.Data)
	if err != nil {
		return Block{}, err
	}
	if input.Type == BlockImage {
		var data imageData
		if json.Unmarshal(normalized, &data) != nil {
			return Block{}, ErrInvalidContent
		}
		available, err := r.CoverAvailable(ctx, data.AssetID)
		if err != nil {
			return Block{}, err
		}
		if !available {
			return Block{}, ErrInvalidCover
		}
	}
	var sort int
	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort),-1)+1 FROM guide_content_blocks WHERE module_id=?`, moduleID).Scan(&sort)
	item := Block{ID: newID(), ModuleID: moduleID, Type: input.Type, DataVersion: input.DataVersion, Sort: sort, Data: normalized}
	_, err = r.db.ExecContext(ctx, `INSERT INTO guide_content_blocks(id,module_id,type,data_version,data,sort,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`, item.ID, moduleID, item.Type, item.DataVersion, string(normalized), sort, formatTime(time.Now()), formatTime(time.Now()))
	return item, err
}

func (r *Repository) ListModules(ctx context.Context, guideID string, public bool) ([]Module, error) {
	where := "guide_id=?"
	if public {
		where += " AND hidden=0"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,guide_id,name,kind,COALESCE(template,''),data_version,COALESCE(structured_data,''),sort,hidden FROM guide_modules WHERE `+where+` ORDER BY sort,id`, guideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Module
	for rows.Next() {
		var m Module
		var data string
		if err := rows.Scan(&m.ID, &m.GuideID, &m.Name, &m.Kind, &m.Template, &m.DataVersion, &data, &m.Sort, &m.Hidden); err != nil {
			return nil, err
		}
		m.StructuredData = []byte(data)
		m.Blocks, _ = r.ListBlocks(ctx, m.ID)
		items = append(items, m)
	}
	return items, rows.Err()
}
func (r *Repository) ListBlocks(ctx context.Context, moduleID string) ([]Block, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,module_id,type,data_version,data,sort FROM guide_content_blocks WHERE module_id=? ORDER BY sort,id`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Block
	for rows.Next() {
		var b Block
		var data string
		if err := rows.Scan(&b.ID, &b.ModuleID, &b.Type, &b.DataVersion, &data, &b.Sort); err != nil {
			return nil, err
		}
		b.Data = []byte(data)
		items = append(items, b)
	}
	return items, rows.Err()
}

func (r *Repository) ReplaceTOC(ctx context.Context, guideID string, inputs []TOCInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM guide_toc WHERE guide_id=?`, guideID); err != nil {
		return err
	}
	for i, input := range inputs {
		if input.Level != 1 && input.Level != 2 {
			return ErrInvalidContent
		}
		var count int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM guide_modules WHERE id=? AND guide_id=?`, input.TargetModuleID, guideID).Scan(&count); err != nil || count != 1 {
			return ErrInvalidOwnership
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO guide_toc(id,guide_id,title,level,target_module_id,sort,hidden,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, newID(), guideID, input.Title, input.Level, input.TargetModuleID, i, input.Hidden, formatTime(time.Now()), formatTime(time.Now()))
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *Repository) AutoGenerateTOC(ctx context.Context, guideID string) error {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name FROM guide_modules WHERE guide_id=? AND hidden=0 ORDER BY sort,id`, guideID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var inputs []TOCInput
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		inputs = append(inputs, TOCInput{Title: name, Level: 1, TargetModuleID: id})
	}
	return r.ReplaceTOC(ctx, guideID, inputs)
}
func (r *Repository) ListTOC(ctx context.Context, guideID string, public bool) ([]TOCItem, error) {
	where := "guide_id=?"
	if public {
		where += " AND hidden=0"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,guide_id,title,level,target_module_id,sort,hidden FROM guide_toc WHERE `+where+` ORDER BY sort,id`, guideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TOCItem
	for rows.Next() {
		var x TOCItem
		if err := rows.Scan(&x.ID, &x.GuideID, &x.Title, &x.Level, &x.TargetModuleID, &x.Sort, &x.Hidden); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateModule(ctx context.Context, guideID, id string, input ModuleInput) error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return ErrInvalidName
	}
	if input.DataVersion == 0 {
		input.DataVersion = 1
	}
	var data any
	if input.Kind == "structured" {
		normalized, err := ValidateStructuredData(input.Template, input.DataVersion, input.StructuredData)
		if err != nil {
			return err
		}
		data = string(normalized)
	} else if input.Kind != "content" {
		return ErrInvalidContent
	}
	result, err := r.db.ExecContext(ctx, `UPDATE guide_modules SET name=?,kind=?,template=?,data_version=?,structured_data=?,hidden=?,updated_at=? WHERE id=? AND guide_id=?`, input.Name, input.Kind, nullableString(string(input.Template)), input.DataVersion, data, input.Hidden, formatTime(time.Now()), id, guideID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return ErrInvalidOwnership
	}
	if input.Kind == "structured" {
		if _, err := r.db.ExecContext(ctx, `DELETE FROM guide_content_blocks WHERE module_id=?`, id); err != nil {
			return err
		}
	}
	return nil
}
func (r *Repository) DeleteModule(ctx context.Context, guideID, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM guide_modules WHERE id=? AND guide_id=?`, id, guideID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return ErrInvalidOwnership
	}
	return nil
}
func (r *Repository) SortModules(ctx context.Context, guideID string, ids []string) error {
	return r.sortChildren(ctx, "guide_modules", "guide_id", guideID, ids)
}
func (r *Repository) UpdateBlock(ctx context.Context, guideID, moduleID, id string, input BlockInput) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM guide_modules WHERE id=? AND guide_id=? AND kind='content'`, moduleID, guideID).Scan(&count); err != nil || count != 1 {
		return ErrInvalidOwnership
	}
	normalized, err := ValidateBlockData(input.Type, input.DataVersion, input.Data)
	if err != nil {
		return err
	}
	if input.Type == BlockImage {
		var data imageData
		if json.Unmarshal(normalized, &data) != nil {
			return ErrInvalidContent
		}
		available, err := r.CoverAvailable(ctx, data.AssetID)
		if err != nil {
			return err
		}
		if !available {
			return ErrInvalidCover
		}
	}
	result, err := r.db.ExecContext(ctx, `UPDATE guide_content_blocks SET type=?,data_version=?,data=?,updated_at=? WHERE id=? AND module_id=?`, input.Type, input.DataVersion, string(normalized), formatTime(time.Now()), id, moduleID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return ErrInvalidOwnership
	}
	return nil
}
func (r *Repository) DeleteBlock(ctx context.Context, moduleID, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM guide_content_blocks WHERE id=? AND module_id=?`, id, moduleID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return ErrInvalidOwnership
	}
	return nil
}
func (r *Repository) SortBlocks(ctx context.Context, moduleID string, ids []string) error {
	return r.sortChildren(ctx, "guide_content_blocks", "module_id", moduleID, ids)
}
func (r *Repository) sortChildren(ctx context.Context, table, parentColumn, parentID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		result, err := tx.ExecContext(ctx, `UPDATE `+table+` SET sort=?,updated_at=? WHERE id=? AND `+parentColumn+`=?`, i, formatTime(time.Now()), id, parentID)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return ErrInvalidOwnership
		}
	}
	return tx.Commit()
}
func (r *Repository) ReplaceAlbums(ctx context.Context, guideID string, ids []string) error {
	if _, err := r.Get(ctx, guideID, false); err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM guide_albums WHERE guide_id=?`, guideID); err != nil {
		return err
	}
	for i, id := range ids {
		if _, err = tx.ExecContext(ctx, `INSERT INTO guide_albums(guide_id,album_id,sort) VALUES(?,?,?)`, guideID, id, i); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *Repository) ListAlbums(ctx context.Context, guideID string, public bool) ([]GuideAlbum, error) {
	where := "ga.guide_id=? AND a.deleted_at IS NULL"
	if public {
		where += " AND a.visible=1"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT a.id,a.name,a.album_value,ga.sort,COALESCE(
		(SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
		 WHERE aa.album_id=a.id AND aa.asset_id=a.cover_asset_id AND x.status='ready' AND x.visible=1 LIMIT 1),
		(SELECT aa.asset_id FROM album_assets aa JOIN assets x ON x.id=aa.asset_id
		 WHERE aa.album_id=a.id AND x.status='ready' AND x.visible=1 ORDER BY aa.sort LIMIT 1),'')
		FROM guide_albums ga JOIN albums a ON a.id=ga.album_id WHERE `+where+` ORDER BY ga.sort,a.id`, guideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GuideAlbum
	for rows.Next() {
		var x GuideAlbum
		var cover string
		if err := rows.Scan(&x.ID, &x.Name, &x.Value, &x.Sort, &cover); err != nil {
			return nil, err
		}
		if cover != "" {
			x.CoverURL = "/media/assets/" + cover + "/preview"
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

var _ = sql.ErrNoRows
