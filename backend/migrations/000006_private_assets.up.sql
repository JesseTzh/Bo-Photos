ALTER TABLE assets ADD COLUMN private INTEGER NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS assets_public_idx;
CREATE INDEX assets_public_idx
    ON assets(status, visible, private, show_on_homepage, sort, created_at, id);
