package asset

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestRepositoryListsOnlyPublicReadyAssets(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	readyVisible := testAsset("ready-visible", StatusReady)
	readyVisible.Visible = true
	readyVisible.ShowOnHomepage = true
	mustCreate(t, repo, readyVisible)

	processing := testAsset("processing", StatusProcessing)
	processing.Visible = true
	processing.ShowOnHomepage = true
	mustCreate(t, repo, processing)

	hidden := testAsset("hidden", StatusReady)
	hidden.Visible = false
	hidden.ShowOnHomepage = true
	mustCreate(t, repo, hidden)

	items, total, err := repo.ListPublic(ctx, PublicFilter{HomepageOnly: true, Page: 1, PageSize: 16})
	if err != nil {
		t.Fatalf("ListPublic() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != readyVisible.ID {
		t.Fatalf("ListPublic() = %#v total=%d, want only %q", items, total, readyVisible.ID)
	}
}

func TestRepositoryFiltersByCameraLensFeaturedAndShootTime(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	older := testAsset("older", StatusReady)
	older.Visible = true
	older.Camera = "Fujifilm X-T5"
	older.Lens = "XF 23mm F1.4"
	older.Featured = true
	older.ShootAt = timePtr(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC))
	mustCreate(t, repo, older)

	newer := testAsset("newer", StatusReady)
	newer.Visible = true
	newer.Camera = "Fujifilm X-T5"
	newer.Lens = "XF 23mm F1.4"
	newer.Featured = true
	newer.ShootAt = timePtr(time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC))
	mustCreate(t, repo, newer)

	other := testAsset("other", StatusReady)
	other.Visible = true
	other.Camera = "Sony A7 IV"
	other.Lens = "FE 35mm F1.4"
	other.Featured = true
	mustCreate(t, repo, other)

	items, total, err := repo.ListPublic(ctx, PublicFilter{
		Cameras:       []string{"Fujifilm X-T5"},
		Lenses:        []string{"XF 23mm F1.4"},
		Featured:      boolPtr(true),
		ShootTimeSort: SortAscending,
		Page:          1,
		PageSize:      16,
	})
	if err != nil {
		t.Fatalf("ListPublic() error = %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("ListPublic() total=%d len=%d, want 2", total, len(items))
	}
	if items[0].ID != older.ID || items[1].ID != newer.ID {
		t.Fatalf("ordered IDs = %q, %q; want %q, %q", items[0].ID, items[1].ID, older.ID, newer.ID)
	}
}

func TestRepositoryUsesStableOffsetPagination(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	for _, id := range []string{"asset-c", "asset-a", "asset-b"} {
		item := testAsset(id, StatusReady)
		item.Visible = true
		item.Sort = 10
		item.CreatedAt = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		mustCreate(t, repo, item)
	}

	first, total, err := repo.ListPublic(ctx, PublicFilter{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	second, _, err := repo.ListPublic(ctx, PublicFilter{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if total != 3 || len(first) != 2 || len(second) != 1 {
		t.Fatalf("pagination total=%d first=%d second=%d", total, len(first), len(second))
	}
	if first[0].ID != "asset-a" || first[1].ID != "asset-b" || second[0].ID != "asset-c" {
		t.Fatalf("unstable order: %#v %#v", first, second)
	}
}

func TestRepositoryRejectsInvalidStatusTransition(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	item := testAsset("transition", StatusProcessing)
	mustCreate(t, repo, item)

	if err := repo.Transition(ctx, item.ID, StatusReady, ""); err != nil {
		t.Fatalf("processing -> ready error = %v", err)
	}
	if err := repo.Transition(ctx, item.ID, StatusPurged, ""); err != ErrInvalidTransition {
		t.Fatalf("ready -> purged error = %v, want ErrInvalidTransition", err)
	}
}

func TestRepositoryAllowsDuplicateSHA256ForWarningOnly(t *testing.T) {
	repo := newTestRepository(t)

	first := testAsset("duplicate-first", StatusProcessing)
	first.SHA256 = "same-content"
	mustCreate(t, repo, first)

	second := testAsset("duplicate-second", StatusProcessing)
	second.SHA256 = "same-content"
	if err := repo.Create(context.Background(), second); err != nil {
		t.Fatalf("Create() duplicate SHA256 error = %v, want upload allowed", err)
	}
}

func TestRepositoryMarksStaleProcessingAssetsFailed(t *testing.T) {
	repo := newTestRepository(t)
	stale := testAsset("stale", StatusProcessing)
	stale.UpdatedAt = time.Now().UTC().Add(-2 * time.Hour)
	mustCreate(t, repo, stale)
	fresh := testAsset("fresh", StatusProcessing)
	fresh.UpdatedAt = time.Now().UTC()
	mustCreate(t, repo, fresh)

	count, err := repo.FailStaleProcessing(context.Background(), time.Now().UTC().Add(-time.Hour))
	if err != nil {
		t.Fatalf("FailStaleProcessing() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("updated count = %d, want 1", count)
	}
	item, err := repo.Get(context.Background(), stale.ID)
	if err != nil {
		t.Fatalf("Get(stale) error = %v", err)
	}
	if item.Status != StatusFailed || item.ErrorCode != "ASSET_PROCESSING_INTERRUPTED" {
		t.Fatalf("stale asset = %#v", item)
	}
}

func TestRepositoryFiltersPublicAssetsByAlbumAndTags(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	for _, id := range []string{"both", "one", "outside"} {
		item := testAsset(id, StatusReady)
		item.Visible = true
		mustCreate(t, repo, item)
	}
	db := repo.db
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.Exec(`INSERT INTO albums(id,name,album_value,created_at,updated_at) VALUES(?,?,?,?,?)`,
		"album", "Album", "album", now, now)
	if err != nil {
		t.Fatal(err)
	}
	for index, id := range []string{"both", "one"} {
		if _, err := db.Exec(`INSERT INTO album_assets(album_id,asset_id,sort) VALUES(?,?,?)`, "album", id, index); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"tag-a", "tag-b"} {
		if _, err := db.Exec(`INSERT INTO tags(id,name,created_at,updated_at) VALUES(?,?,?,?)`, id, id, now, now); err != nil {
			t.Fatal(err)
		}
	}
	for _, relation := range [][2]string{{"both", "tag-a"}, {"both", "tag-b"}, {"one", "tag-a"}, {"outside", "tag-b"}} {
		if _, err := db.Exec(`INSERT INTO asset_tags(asset_id,tag_id,created_at) VALUES(?,?,?)`,
			relation[0], relation[1], now); err != nil {
			t.Fatal(err)
		}
	}

	items, total, err := repo.ListPublic(ctx, PublicFilter{
		Album: "album", Tags: []string{"tag-a", "tag-b"}, TagsOperator: "and", Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != "both" {
		t.Fatalf("AND filter items=%#v total=%d", items, total)
	}
	items, total, err = repo.ListPublic(ctx, PublicFilter{
		Album: "album", Tags: []string{"tag-b"}, TagsOperator: "or", Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != "both" {
		t.Fatalf("OR album filter items=%#v total=%d", items, total)
	}
}

func newTestRepository(t *testing.T) *Repository {
	t.Helper()
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("repository.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatalf("repository.Migrate() error = %v", err)
	}
	return NewRepository(db)
}

func testAsset(id string, status Status) Asset {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return Asset{
		ID:           id,
		Status:       status,
		OriginalName: id + ".jpg",
		OriginalKey:  "originals/" + id + ".jpg",
		MIMEType:     "image/jpeg",
		SHA256:       id,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func mustCreate(t *testing.T, repo *Repository, item Asset) {
	t.Helper()
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("Create(%q) error = %v", item.ID, err)
	}
}

func boolPtr(value bool) *bool           { return &value }
func timePtr(value time.Time) *time.Time { return &value }
