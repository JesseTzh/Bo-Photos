CREATE TABLE assets (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL CHECK (status IN ('processing', 'ready', 'failed', 'deleted', 'purged')),
    original_name TEXT NOT NULL,
    original_key TEXT NOT NULL,
    preview_key TEXT,
    thumbnail_key TEXT,
    sha256 TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    byte_size INTEGER NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    title TEXT,
    description TEXT,
    longitude REAL,
    latitude REAL,
    blurhash TEXT,
    exif_json TEXT,
    shoot_at TEXT,
    camera TEXT,
    lens TEXT,
    exposure_time TEXT,
    aperture TEXT,
    iso TEXT,
    focal_length TEXT,
    error_code TEXT,
    visible INTEGER NOT NULL DEFAULT 1,
    show_on_homepage INTEGER NOT NULL DEFAULT 1,
    featured INTEGER NOT NULL DEFAULT 0,
    sort INTEGER NOT NULL DEFAULT 0,
    derivative_version INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT,
    purged_at TEXT
);

CREATE INDEX assets_public_idx
    ON assets(status, visible, show_on_homepage, sort, created_at, id);
CREATE INDEX assets_camera_idx ON assets(camera);
CREATE INDEX assets_lens_idx ON assets(lens);
CREATE INDEX assets_featured_idx ON assets(featured);
CREATE INDEX assets_shoot_at_idx ON assets(shoot_at);
CREATE INDEX assets_sha256_idx ON assets(sha256);

CREATE TABLE albums (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    album_value TEXT NOT NULL UNIQUE,
    detail TEXT,
    theme TEXT NOT NULL DEFAULT '0',
    visible INTEGER NOT NULL DEFAULT 1,
    sort INTEGER NOT NULL DEFAULT 0,
    random_show INTEGER NOT NULL DEFAULT 1,
    license TEXT,
    cover_asset_id TEXT REFERENCES assets(id) ON DELETE SET NULL,
    image_sorting INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT
);

CREATE TABLE album_assets (
    album_id TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    sort INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (album_id, asset_id)
);

CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    category TEXT,
    parent_id TEXT REFERENCES tags(id) ON DELETE RESTRICT,
    detail TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE asset_tags (
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL,
    PRIMARY KEY (asset_id, tag_id)
);
