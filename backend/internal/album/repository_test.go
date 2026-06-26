package album

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestRepositoryCRUDAndCoverFallback(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	createAsset(t, db, "first", true)
	createAsset(t, db, "second", true)
	createAsset(t, db, "hidden", false)

	item := Album{ID: "album-1", Name: "Japan", Value: "japan", Visible: true, RandomShow: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(ctx, item); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceAssets(ctx, item.ID, []string{"second", "hidden", "first"}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByValue(ctx, "japan", true)
	if err != nil {
		t.Fatal(err)
	}
	if got.CoverAssetID != "" || got.EffectiveCoverAssetID != "second" ||
		len(got.AssetIDs) != 2 || got.AssetIDs[0] != "second" || got.AssetIDs[1] != "first" {
		t.Fatalf("album = %#v", got)
	}
}

func TestRepositoryRejectsDuplicateValue(t *testing.T) {
	repo, _ := newTestRepository(t)
	now := time.Now()
	for _, id := range []string{"one", "two"} {
		err := repo.Create(context.Background(), Album{ID: id, Name: id, Value: "same", CreatedAt: now, UpdatedAt: now})
		if id == "two" && err == nil {
			t.Fatal("duplicate album value was accepted")
		}
	}
}

func TestAdminListIncludesOrderedAssetIDs(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	createAsset(t, db, "first", true)
	createAsset(t, db, "second", true)
	now := time.Now()
	if err := repo.Create(ctx, Album{ID: "album", Name: "Album", Value: "album", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceAssets(ctx, "album", []string{"second", "first"}); err != nil {
		t.Fatal(err)
	}
	items, err := repo.List(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].AssetIDs) != 2 ||
		items[0].AssetIDs[0] != "second" || items[0].AssetIDs[1] != "first" {
		t.Fatalf("albums = %#v", items)
	}
}

func newTestRepository(t *testing.T) (*Repository, *sql.DB) {
	t.Helper()
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return NewRepository(db), db
}

func createAsset(t *testing.T, db *sql.DB, id string, visible bool) {
	t.Helper()
	visibleInt := 0
	if visible {
		visibleInt = 1
	}
	_, err := db.Exec(`INSERT INTO assets (
		id,status,original_name,original_key,sha256,mime_type,visible,created_at,updated_at
	) VALUES (?,?,?,?,?,?,?,?,?)`, id, "ready", id+".jpg", id+".jpg", id, "image/jpeg", visibleInt, time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
}
