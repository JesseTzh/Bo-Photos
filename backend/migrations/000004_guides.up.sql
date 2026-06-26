CREATE TABLE guides (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    country TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    days INTEGER NOT NULL DEFAULT 0 CHECK (days >= 0),
    start_date TEXT,
    end_date TEXT,
    cover_asset_id TEXT REFERENCES assets(id) ON DELETE SET NULL,
    published INTEGER NOT NULL DEFAULT 0,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT,
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);

CREATE INDEX guides_public_idx
    ON guides(published, deleted_at, sort, created_at, id);

CREATE TABLE guide_modules (
    id TEXT PRIMARY KEY,
    guide_id TEXT NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('content', 'structured')),
    template TEXT CHECK (template IN ('itinerary', 'expense', 'checklist', 'transport', 'photo', 'tips')),
    data_version INTEGER NOT NULL DEFAULT 1,
    structured_data TEXT,
    sort INTEGER NOT NULL DEFAULT 0,
    hidden INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (
        (kind = 'content' AND template IS NULL AND structured_data IS NULL)
        OR
        (kind = 'structured' AND template IS NOT NULL AND structured_data IS NOT NULL)
    )
);

CREATE INDEX guide_modules_guide_sort_idx
    ON guide_modules(guide_id, sort, id);

CREATE TABLE guide_content_blocks (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL REFERENCES guide_modules(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('markdown', 'image', 'video', 'link', 'tasks', 'warning', 'divider')),
    data_version INTEGER NOT NULL DEFAULT 1,
    data TEXT NOT NULL,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX guide_content_blocks_module_sort_idx
    ON guide_content_blocks(module_id, sort, id);

CREATE TABLE guide_toc (
    id TEXT PRIMARY KEY,
    guide_id TEXT NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1 CHECK (level IN (1, 2)),
    target_module_id TEXT NOT NULL REFERENCES guide_modules(id) ON DELETE CASCADE,
    sort INTEGER NOT NULL DEFAULT 0,
    hidden INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX guide_toc_guide_sort_idx
    ON guide_toc(guide_id, sort, id);

CREATE TABLE guide_albums (
    guide_id TEXT NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    album_id TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    sort INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (guide_id, album_id)
);

CREATE INDEX guide_albums_guide_sort_idx
    ON guide_albums(guide_id, sort, album_id);
