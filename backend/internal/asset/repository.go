package asset

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/besscroft/bophotos/backend/internal/imageproc"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item Asset) error {
	if item.DerivativeVersion == 0 {
		item.DerivativeVersion = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO assets (
			id, status, original_name, original_key, preview_key, thumbnail_key,
			sha256, mime_type, byte_size, width, height, title, description,
			longitude, latitude, blurhash, exif_json, shoot_at, camera, lens,
			exposure_time, aperture, iso, focal_length, error_code, visible,
			show_on_homepage, featured, sort, derivative_version, created_at,
			updated_at, deleted_at, purged_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.ID, item.Status, item.OriginalName, item.OriginalKey,
		nullString(item.PreviewKey), nullString(item.ThumbnailKey), item.SHA256,
		item.MIMEType, item.ByteSize, item.Width, item.Height, nullString(item.Title),
		nullString(item.Description), item.Longitude, item.Latitude,
		nullString(item.BlurHash), nullString(item.EXIFJSON), nullableTime(item.ShootAt),
		nullString(item.Camera), nullString(item.Lens), nullString(item.ExposureTime),
		nullString(item.Aperture), nullString(item.ISO), nullString(item.FocalLength),
		nullString(item.ErrorCode), item.Visible, item.ShowOnHomepage, item.Featured,
		item.Sort, item.DerivativeVersion, formatTime(item.CreatedAt),
		formatTime(item.UpdatedAt), nullableTime(item.DeletedAt), nullableTime(item.PurgedAt),
	)
	if err != nil {
		return fmt.Errorf("create asset: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id string) (Asset, error) {
	item, err := scanAsset(r.db.QueryRowContext(ctx, `SELECT `+assetColumns+` FROM assets WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) FindBySHA256(ctx context.Context, hash string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM assets WHERE sha256 = ? AND status != 'purged' ORDER BY created_at, id
	`, hash)
	if err != nil {
		return nil, fmt.Errorf("find duplicate assets: %w", err)
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan duplicate asset: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) MarkReady(
	ctx context.Context,
	id string,
	previewKey string,
	thumbnailKey string,
	metadata imageproc.Metadata,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE assets SET
			status = 'ready', preview_key = ?, thumbnail_key = ?, width = ?, height = ?,
			exif_json = ?, shoot_at = ?, camera = ?, lens = ?, exposure_time = ?,
			aperture = ?, iso = ?, focal_length = ?, longitude = ?, latitude = ?,
			error_code = NULL, updated_at = ?
		WHERE id = ? AND status = 'processing'
	`,
		previewKey, thumbnailKey, metadata.Width, metadata.Height,
		nullString(metadata.EXIFJSON), nullableTime(metadata.ShootAt),
		nullString(metadata.Camera), nullString(metadata.Lens),
		nullString(metadata.ExposureTime), nullString(metadata.Aperture),
		nullString(metadata.ISO), nullString(metadata.FocalLength),
		metadata.Longitude, metadata.Latitude, formatTime(time.Now().UTC()), id,
	)
	if err != nil {
		return fmt.Errorf("mark asset ready: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return ErrInvalidTransition
	}
	return nil
}

func (r *Repository) FailStaleProcessing(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE assets
		SET status = 'failed', error_code = 'ASSET_PROCESSING_INTERRUPTED', updated_at = ?
		WHERE status = 'processing' AND updated_at < ?
	`, formatTime(time.Now().UTC()), formatTime(before))
	if err != nil {
		return 0, fmt.Errorf("fail stale processing assets: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read stale processing count: %w", err)
	}
	return count, nil
}

func (r *Repository) DeletedBefore(ctx context.Context, before time.Time) ([]Asset, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+assetColumns+` FROM assets
		WHERE status = 'deleted' AND deleted_at IS NOT NULL AND deleted_at < ?
		ORDER BY deleted_at, id
	`, formatTime(before))
	if err != nil {
		return nil, fmt.Errorf("list expired deleted assets: %w", err)
	}
	defer rows.Close()
	items := make([]Asset, 0)
	for rows.Next() {
		item, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListPublic(ctx context.Context, filter PublicFilter) ([]Asset, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 16
	}

	fromSQL := "assets"
	selectColumns := assetColumns
	where := []string{"assets.status = 'ready'", "assets.visible = 1", "assets.deleted_at IS NULL"}
	args := make([]any, 0)
	if filter.HomepageOnly {
		where = append(where, "assets.show_on_homepage = 1")
	}
	if len(filter.Cameras) > 0 {
		where = append(where, "assets.camera IN ("+placeholders(len(filter.Cameras))+")")
		for _, value := range filter.Cameras {
			args = append(args, value)
		}
	}
	if len(filter.Lenses) > 0 {
		where = append(where, "assets.lens IN ("+placeholders(len(filter.Lenses))+")")
		for _, value := range filter.Lenses {
			args = append(args, value)
		}
	}
	if filter.Featured != nil {
		where = append(where, "assets.featured = ?")
		args = append(args, *filter.Featured)
	}
	if filter.Album != "" {
		fromSQL = "assets JOIN album_assets aa ON aa.asset_id = assets.id JOIN albums a ON a.id = aa.album_id"
		selectColumns = qualifiedAssetColumns("assets")
		where = append(where, "a.album_value = ?", "a.visible = 1", "a.deleted_at IS NULL")
		args = append(args, filter.Album)
	}
	if len(filter.Tags) > 0 {
		if filter.TagsOperator == "or" {
			where = append(where, `EXISTS (
				SELECT 1 FROM asset_tags at JOIN tags t ON t.id=at.tag_id
				WHERE at.asset_id=assets.id AND t.id IN (`+placeholders(len(filter.Tags))+`)
			)`)
			for _, value := range filter.Tags {
				args = append(args, value)
			}
		} else {
			for _, value := range filter.Tags {
				where = append(where, `EXISTS (
					SELECT 1 FROM asset_tags at WHERE at.asset_id=assets.id AND at.tag_id=?
				)`)
				args = append(args, value)
			}
		}
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+fromSQL+" WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count public assets: %w", err)
	}

	order := "assets.sort ASC, assets.created_at DESC, assets.id ASC"
	if filter.Album != "" {
		var randomShow bool
		var imageSorting int
		err := r.db.QueryRowContext(ctx, `SELECT random_show,image_sorting FROM albums
			WHERE album_value=? AND visible=1 AND deleted_at IS NULL`, filter.Album).Scan(&randomShow, &imageSorting)
		if errors.Is(err, sql.ErrNoRows) {
			return []Asset{}, 0, nil
		}
		if err != nil {
			return nil, 0, fmt.Errorf("read album sorting: %w", err)
		}
		if randomShow {
			order = "random()"
		} else {
			switch imageSorting {
			case 1:
				order = `aa.sort ASC, assets.created_at DESC, assets.updated_at DESC, assets.id ASC`
			case 2:
				order = `aa.sort ASC, assets.shoot_at IS NULL ASC, assets.shoot_at DESC, assets.created_at DESC, assets.id ASC`
			case 3:
				order = `aa.sort ASC, assets.created_at ASC, assets.updated_at ASC, assets.id ASC`
			case 4:
				order = `aa.sort ASC, assets.shoot_at IS NULL ASC, assets.shoot_at ASC, assets.created_at ASC, assets.id ASC`
			default:
				order = `aa.sort ASC, assets.created_at DESC, assets.id ASC`
			}
		}
	}
	if filter.Album == "" && filter.ShootTimeSort == SortAscending {
		order = "assets.shoot_at IS NULL ASC, assets.shoot_at ASC, assets.created_at ASC, assets.id ASC"
	} else if filter.Album == "" && filter.ShootTimeSort == SortDescending {
		order = "assets.shoot_at IS NULL ASC, assets.shoot_at DESC, assets.created_at DESC, assets.id ASC"
	}

	queryArgs := append(append([]any{}, args...), filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+selectColumns+` FROM `+fromSQL+`
		WHERE `+whereSQL+`
		ORDER BY `+order+`
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list public assets: %w", err)
	}
	defer rows.Close()

	items := make([]Asset, 0)
	for rows.Next() {
		item, err := scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate public assets: %w", err)
	}
	return items, total, nil
}

func quoteSQL(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func (r *Repository) ListAdmin(ctx context.Context, filter AdminFilter) ([]Asset, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 20
	}
	where := []string{"1 = 1"}
	args := make([]any, 0)
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Visible != nil {
		where = append(where, "visible = ?")
		args = append(args, *filter.Visible)
	}
	if filter.Featured != nil {
		where = append(where, "featured = ?")
		args = append(args, *filter.Featured)
	}
	if filter.Album != "" {
		where = append(where, `EXISTS (
			SELECT 1 FROM album_assets aa JOIN albums a ON a.id=aa.album_id
			WHERE aa.asset_id=assets.id AND a.album_value=? AND a.deleted_at IS NULL
		)`)
		args = append(args, filter.Album)
	}
	for column, value := range map[string]string{
		"camera": filter.Camera, "lens": filter.Lens, "exposure_time": filter.ExposureTime,
		"aperture": filter.Aperture, "iso": filter.ISO,
	} {
		if value != "" {
			where = append(where, column+" = ?")
			args = append(args, value)
		}
	}
	if len(filter.TagIDs) > 0 {
		if filter.TagsOperator == "or" {
			where = append(where, `EXISTS (
				SELECT 1 FROM asset_tags at WHERE at.asset_id=assets.id AND at.tag_id IN (`+placeholders(len(filter.TagIDs))+`)
			)`)
			for _, id := range filter.TagIDs {
				args = append(args, id)
			}
		} else {
			for _, id := range filter.TagIDs {
				where = append(where, `EXISTS (
					SELECT 1 FROM asset_tags at WHERE at.asset_id=assets.id AND at.tag_id=?
				)`)
				args = append(args, id)
			}
		}
	}
	if filter.Title != "" {
		where = append(where, "title LIKE ?")
		args = append(args, "%"+filter.Title+"%")
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assets WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin assets: %w", err)
	}
	queryArgs := append(append([]any{}, args...), filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+assetColumns+` FROM assets WHERE `+whereSQL+`
		ORDER BY created_at DESC, id ASC LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin assets: %w", err)
	}
	defer rows.Close()
	items := make([]Asset, 0)
	for rows.Next() {
		item, err := scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id string, input UpdateInput) error {
	sets := make([]string, 0)
	args := make([]any, 0)
	add := func(column string, value any) {
		sets = append(sets, column+" = ?")
		args = append(args, value)
	}
	if input.Title != nil {
		add("title", nullString(*input.Title))
	}
	if input.Description != nil {
		add("description", nullString(*input.Description))
	}
	if input.Width != nil {
		if *input.Width < 0 {
			return errors.New("width must be non-negative")
		}
		add("width", *input.Width)
	}
	if input.Height != nil {
		if *input.Height < 0 {
			return errors.New("height must be non-negative")
		}
		add("height", *input.Height)
	}
	if input.Longitude.Set {
		add("longitude", input.Longitude.Value)
	}
	if input.Latitude.Set {
		add("latitude", input.Latitude.Value)
	}
	if input.ShootAt != nil {
		value := strings.TrimSpace(*input.ShootAt)
		if value == "" {
			add("shoot_at", nil)
		} else {
			parsed, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return errors.New("shoot_at must be RFC3339")
			}
			add("shoot_at", formatTime(parsed))
		}
	}
	if input.Camera != nil {
		add("camera", nullString(*input.Camera))
	}
	if input.Lens != nil {
		add("lens", nullString(*input.Lens))
	}
	if input.ExposureTime != nil {
		add("exposure_time", nullString(*input.ExposureTime))
	}
	if input.Aperture != nil {
		add("aperture", nullString(*input.Aperture))
	}
	if input.ISO != nil {
		add("iso", nullString(*input.ISO))
	}
	if input.FocalLength != nil {
		add("focal_length", nullString(*input.FocalLength))
	}
	if input.EXIFJSON != nil {
		add("exif_json", nullString(*input.EXIFJSON))
	}
	if input.Visible != nil {
		add("visible", *input.Visible)
	}
	if input.ShowOnHomepage != nil {
		add("show_on_homepage", *input.ShowOnHomepage)
	}
	if input.Featured != nil {
		add("featured", *input.Featured)
	}
	if input.Sort != nil {
		add("sort", *input.Sort)
	}
	if len(sets) == 0 {
		return nil
	}
	add("updated_at", formatTime(time.Now().UTC()))
	args = append(args, id)
	result, err := r.db.ExecContext(ctx, "UPDATE assets SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
	if err != nil {
		return fmt.Errorf("update asset: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateSort(ctx context.Context, orders map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sort update: %w", err)
	}
	defer tx.Rollback()
	for id, sortValue := range orders {
		if _, err := tx.ExecContext(ctx, `UPDATE assets SET sort = ?, updated_at = ? WHERE id = ?`, sortValue, formatTime(time.Now().UTC()), id); err != nil {
			return fmt.Errorf("update asset sort: %w", err)
		}
	}
	return tx.Commit()
}

type AdminFilterOptions struct {
	Cameras       []string `json:"cameras"`
	Lenses        []string `json:"lenses"`
	ExposureTimes []string `json:"exposure_times"`
	Apertures     []string `json:"apertures"`
	ISOs          []string `json:"isos"`
}

func (r *Repository) AdminFilterOptions(ctx context.Context) (AdminFilterOptions, error) {
	read := func(column string) ([]string, error) {
		rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT `+column+` FROM assets WHERE status!='purged' AND `+column+` IS NOT NULL AND `+column+` != '' ORDER BY `+column)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		values := make([]string, 0)
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, rows.Err()
	}
	var out AdminFilterOptions
	var err error
	if out.Cameras, err = read("camera"); err != nil {
		return out, fmt.Errorf("read camera filters: %w", err)
	}
	if out.Lenses, err = read("lens"); err != nil {
		return out, fmt.Errorf("read lens filters: %w", err)
	}
	if out.ExposureTimes, err = read("exposure_time"); err != nil {
		return out, fmt.Errorf("read exposure filters: %w", err)
	}
	if out.Apertures, err = read("aperture"); err != nil {
		return out, fmt.Errorf("read aperture filters: %w", err)
	}
	if out.ISOs, err = read("iso"); err != nil {
		return out, fmt.Errorf("read iso filters: %w", err)
	}
	return out, nil
}

func (r *Repository) FilterOptions(ctx context.Context) ([]string, []string, error) {
	options, err := r.AdminFilterOptions(ctx)
	if err != nil {
		return nil, nil, err
	}
	return options.Cameras, options.Lenses, nil
}

func (r *Repository) AlbumIDsForAsset(ctx context.Context, assetID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT album_id FROM album_assets WHERE asset_id=? ORDER BY album_id`, assetID)
	if err != nil {
		return nil, fmt.Errorf("list asset albums: %w", err)
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

func (r *Repository) ReplaceAssetAlbums(ctx context.Context, assetID string, albumIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin asset album replace: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM album_assets WHERE asset_id=?`, assetID); err != nil {
		return fmt.Errorf("clear asset albums: %w", err)
	}
	for _, albumID := range albumIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO album_assets(album_id,asset_id,sort)
			VALUES(?,?,COALESCE((SELECT MAX(sort)+1 FROM album_assets WHERE album_id=?),0))
		`, albumID, assetID, albumID); err != nil {
			return fmt.Errorf("insert asset album: %w", err)
		}
	}
	return tx.Commit()
}

func (r *Repository) Transition(ctx context.Context, id string, to Status, errorCode string) error {
	var from Status
	if err := r.db.QueryRowContext(ctx, `SELECT status FROM assets WHERE id = ?`, id).Scan(&from); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("read asset status: %w", err)
	}
	if !CanTransition(from, to) {
		return ErrInvalidTransition
	}

	now := time.Now().UTC()
	var deletedAt any
	var purgedAt any
	if to == StatusDeleted {
		deletedAt = formatTime(now)
	}
	if to == StatusPurged {
		purgedAt = formatTime(now)
	}
	if to == StatusReady {
		deletedAt = nil
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE assets
		SET status = ?, error_code = ?, updated_at = ?,
		    derivative_version = CASE WHEN ? = 'processing' THEN derivative_version + 1 ELSE derivative_version END,
		    deleted_at = CASE WHEN ? IS NOT NULL OR ? = 'ready' THEN ? ELSE deleted_at END,
		    purged_at = COALESCE(?, purged_at)
		WHERE id = ?
	`, to, nullString(errorCode), formatTime(now), to, deletedAt, to, deletedAt, purgedAt, id)
	if err != nil {
		return fmt.Errorf("transition asset: %w", err)
	}
	return nil
}

const assetColumns = `
	id, status, original_name, original_key, preview_key, thumbnail_key,
	sha256, mime_type, byte_size, width, height, title, description,
	longitude, latitude, blurhash, exif_json, shoot_at, camera, lens,
	exposure_time, aperture, iso, focal_length, error_code, visible,
	show_on_homepage, featured, sort, derivative_version, created_at,
	updated_at, deleted_at, purged_at
`

type scanner interface {
	Scan(dest ...any) error
}

func scanAsset(row scanner) (Asset, error) {
	var item Asset
	var previewKey, thumbnailKey, title, description sql.NullString
	var longitude, latitude sql.NullFloat64
	var blurhash, exifJSON, shootAt, camera, lens sql.NullString
	var exposure, aperture, iso, focal, errorCode sql.NullString
	var createdAt, updatedAt string
	var deletedAt, purgedAt sql.NullString
	err := row.Scan(
		&item.ID, &item.Status, &item.OriginalName, &item.OriginalKey,
		&previewKey, &thumbnailKey, &item.SHA256, &item.MIMEType, &item.ByteSize,
		&item.Width, &item.Height, &title, &description, &longitude, &latitude,
		&blurhash, &exifJSON, &shootAt, &camera, &lens, &exposure, &aperture,
		&iso, &focal, &errorCode, &item.Visible, &item.ShowOnHomepage,
		&item.Featured, &item.Sort, &item.DerivativeVersion, &createdAt,
		&updatedAt, &deletedAt, &purgedAt,
	)
	if err != nil {
		return Asset{}, fmt.Errorf("scan asset: %w", err)
	}
	item.PreviewKey = previewKey.String
	item.ThumbnailKey = thumbnailKey.String
	item.Title = title.String
	item.Description = description.String
	if longitude.Valid {
		item.Longitude = &longitude.Float64
	}
	if latitude.Valid {
		item.Latitude = &latitude.Float64
	}
	item.BlurHash = blurhash.String
	item.EXIFJSON = exifJSON.String
	item.ShootAt = parseNullableTime(shootAt)
	item.Camera = camera.String
	item.Lens = lens.String
	item.ExposureTime = exposure.String
	item.Aperture = aperture.String
	item.ISO = iso.String
	item.FocalLength = focal.String
	item.ErrorCode = errorCode.String
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	item.DeletedAt = parseNullableTime(deletedAt)
	item.PurgedAt = parseNullableTime(purgedAt)
	return item, nil
}

func placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func qualifiedAssetColumns(prefix string) string {
	parts := strings.Split(assetColumns, ",")
	for index, part := range parts {
		parts[index] = prefix + "." + strings.TrimSpace(part)
	}
	return strings.Join(parts, ", ")
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
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
