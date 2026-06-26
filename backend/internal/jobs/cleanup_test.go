package jobs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/asset"
	"github.com/besscroft/bophotos/backend/internal/repository"
	"github.com/besscroft/bophotos/backend/internal/storage"
)

func TestCleanupRemovesStaleStagingAndPurgesExpiredDeletedAssets(t *testing.T) {
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("repository.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatalf("repository.Migrate() error = %v", err)
	}
	assetRepo := asset.NewRepository(db)
	local, err := storage.NewLocal(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("storage.NewLocal() error = %v", err)
	}

	staleStaging := writeCleanupFile(t, local, "staging/stale.upload")
	freshStaging := writeCleanupFile(t, local, "staging/fresh.upload")
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(staleStaging, old, old); err != nil {
		t.Fatalf("Chtimes(stale) error = %v", err)
	}

	deletedAt := time.Now().UTC().Add(-31 * 24 * time.Hour)
	expired := cleanupAsset("expired", deletedAt)
	if err := assetRepo.Create(context.Background(), expired); err != nil {
		t.Fatalf("Create(expired) error = %v", err)
	}
	writeCleanupFile(t, local, expired.OriginalKey)

	recentAt := time.Now().UTC().Add(-10 * 24 * time.Hour)
	recent := cleanupAsset("recent", recentAt)
	if err := assetRepo.Create(context.Background(), recent); err != nil {
		t.Fatalf("Create(recent) error = %v", err)
	}
	writeCleanupFile(t, local, recent.OriginalKey)

	cleanup := NewCleanup(assetRepo, local, time.Hour, 30*24*time.Hour)
	if err := cleanup.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}

	if _, err := os.Stat(staleStaging); !os.IsNotExist(err) {
		t.Errorf("stale staging error = %v, want not exist", err)
	}
	if _, err := os.Stat(freshStaging); err != nil {
		t.Errorf("fresh staging error = %v", err)
	}
	expiredResult, err := assetRepo.Get(context.Background(), expired.ID)
	if err != nil {
		t.Fatalf("Get(expired) error = %v", err)
	}
	if expiredResult.Status != asset.StatusPurged {
		t.Errorf("expired status = %q, want purged", expiredResult.Status)
	}
	recentResult, err := assetRepo.Get(context.Background(), recent.ID)
	if err != nil {
		t.Fatalf("Get(recent) error = %v", err)
	}
	if recentResult.Status != asset.StatusDeleted {
		t.Errorf("recent status = %q, want deleted", recentResult.Status)
	}
}

func cleanupAsset(id string, deletedAt time.Time) asset.Asset {
	return asset.Asset{
		ID: id, Status: asset.StatusDeleted, OriginalName: id + ".jpg",
		OriginalKey: "originals/" + id + ".jpg", SHA256: id, MIMEType: "image/jpeg",
		CreatedAt: deletedAt, UpdatedAt: deletedAt, DeletedAt: &deletedAt,
	}
}

func writeCleanupFile(t *testing.T, local *storage.Local, key string) string {
	t.Helper()
	path, err := local.Resolve(key)
	if err != nil {
		t.Fatalf("Resolve(%q) error = %v", key, err)
	}
	if err := os.WriteFile(path, []byte("file"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", key, err)
	}
	return path
}
