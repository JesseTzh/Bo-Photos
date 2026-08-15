DROP INDEX IF EXISTS assets_public_idx;
CREATE INDEX assets_public_idx
    ON assets(status, visible, show_on_homepage, sort, created_at, id);
