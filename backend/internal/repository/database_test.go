package repository

import (
	"context"
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
		"guides", "guide_modules", "guide_content_blocks", "guide_toc", "guide_albums",
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
		"guides_public_idx",
		"guide_modules_guide_sort_idx",
		"guide_content_blocks_module_sort_idx",
		"guide_toc_guide_sort_idx",
		"guide_albums_guide_sort_idx",
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
