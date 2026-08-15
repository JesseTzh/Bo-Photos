package repository

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAppliesSQLitePragmasToConnections(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if got := strings.ToLower(journalMode); got != "wal" {
		t.Errorf("journal_mode = %q, want wal", got)
	}

	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}

	var busyTimeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("query busy_timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busyTimeout)
	}

	if got := db.Stats().MaxOpenConnections; got != 4 {
		t.Errorf("MaxOpenConnections = %d, want 4", got)
	}
}

func TestMigrateCreatesFoundationTables(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	for _, table := range []string{
		"administrators", "sessions", "configs",
		"visit_logs",
	} {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %q not created: %v", table, err)
		}
	}

	for _, index := range []string{
		"album_assets_album_sort_idx",
		"album_assets_asset_idx",
		"asset_tags_tag_asset_idx",
		"tags_parent_idx",
		"visit_logs_created_idx",
	} {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?`,
			index,
		).Scan(&name)
		if err != nil {
			t.Fatalf("index %q not created: %v", index, err)
		}
	}
}

func TestMigrateUpgradesLegacyVersionSix(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	legacySchema := `
		DROP TABLE visit_logs;
		CREATE TABLE visit_logs (
			id TEXT PRIMARY KEY,
			path TEXT NOT NULL,
			page_type TEXT NOT NULL CHECK (page_type IN ('home','gallery','album','guide','about','other')),
			ip_hash TEXT,
			user_agent TEXT,
			referrer TEXT,
			source TEXT NOT NULL CHECK (source IN ('direct','referer','search','other')),
			created_at TEXT NOT NULL
		);
		CREATE TABLE guides (id TEXT PRIMARY KEY);
		CREATE TABLE guide_modules (id TEXT PRIMARY KEY);
		CREATE TABLE guide_content_blocks (id TEXT PRIMARY KEY);
		CREATE TABLE guide_toc (id TEXT PRIMARY KEY);
		CREATE TABLE guide_albums (guide_id TEXT, album_id TEXT);
		INSERT INTO guides(id) VALUES ('legacy-guide');
		INSERT INTO visit_logs(id,path,page_type,source,created_at) VALUES
			('home-visit','/','home','direct','2026-08-15T00:00:00Z'),
			('guide-visit','/guides','guide','direct','2026-08-15T00:00:00Z');
		UPDATE schema_migrations SET version=6, dirty=0;
	`
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("upgrade legacy version 6: %v", err)
	}

	var version int
	if err := db.QueryRow(`SELECT version FROM schema_migrations`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 7 {
		t.Fatalf("migration version = %d, want 7", version)
	}

	for _, table := range []string{"guides", "guide_modules", "guide_content_blocks", "guide_toc", "guide_albums"} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("legacy table %q still exists", table)
		}
	}

	var visits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM visit_logs`).Scan(&visits); err != nil {
		t.Fatal(err)
	}
	if visits != 1 {
		t.Fatalf("visit count = %d, want 1", visits)
	}
}
