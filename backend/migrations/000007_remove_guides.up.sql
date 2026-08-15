DROP TABLE IF EXISTS guide_albums;
DROP TABLE IF EXISTS guide_toc;
DROP TABLE IF EXISTS guide_content_blocks;
DROP TABLE IF EXISTS guide_modules;
DROP TABLE IF EXISTS guides;

CREATE TABLE visit_logs_without_guides (
    id TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    page_type TEXT NOT NULL CHECK (page_type IN ('home','gallery','album','about','other')),
    ip_hash TEXT,
    user_agent TEXT,
    referrer TEXT,
    source TEXT NOT NULL CHECK (source IN ('direct','referer','search','other')),
    created_at TEXT NOT NULL
);

INSERT INTO visit_logs_without_guides (
    id, path, page_type, ip_hash, user_agent, referrer, source, created_at
)
SELECT id, path, page_type, ip_hash, user_agent, referrer, source, created_at
FROM visit_logs
WHERE page_type != 'guide';

DROP TABLE visit_logs;
ALTER TABLE visit_logs_without_guides RENAME TO visit_logs;
CREATE INDEX visit_logs_created_idx ON visit_logs(created_at);
CREATE INDEX visit_logs_path_idx ON visit_logs(path);
CREATE INDEX visit_logs_ip_hash_idx ON visit_logs(ip_hash);
