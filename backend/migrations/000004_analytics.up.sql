CREATE TABLE visit_logs (
    id TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    page_type TEXT NOT NULL CHECK (page_type IN ('home','gallery','album','about','other')),
    ip_hash TEXT,
    user_agent TEXT,
    referrer TEXT,
    source TEXT NOT NULL CHECK (source IN ('direct','referer','search','other')),
    created_at TEXT NOT NULL
);
CREATE INDEX visit_logs_created_idx ON visit_logs(created_at);
CREATE INDEX visit_logs_path_idx ON visit_logs(path);
CREATE INDEX visit_logs_ip_hash_idx ON visit_logs(ip_hash);
