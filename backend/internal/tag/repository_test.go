package tag

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestServiceRejectsCyclesAndParentDeletion(t *testing.T) {
	service, repo, _ := newTestService(t)
	ctx := context.Background()
	if _, err := service.Create(ctx, CreateInput{Name: "Place"}); err != nil {
		t.Fatal(err)
	}
	parent, _ := repo.GetByName(ctx, "Place")
	if _, err := service.Create(ctx, CreateInput{Name: "Japan", ParentID: parent.ID}); err != nil {
		t.Fatal(err)
	}
	child, _ := repo.GetByName(ctx, "Japan")
	if err := service.Move(ctx, parent.ID, child.ID); err != ErrCycle {
		t.Fatalf("Move() error = %v, want ErrCycle", err)
	}
	if err := service.Delete(ctx, parent.ID); err != ErrHasChildren {
		t.Fatalf("Delete() error = %v, want ErrHasChildren", err)
	}
}

func TestAssignAssetIncludesAncestors(t *testing.T) {
	service, repo, db := newTestService(t)
	ctx := context.Background()
	parent, _ := service.Create(ctx, CreateInput{Name: "Place"})
	child, _ := service.Create(ctx, CreateInput{Name: "Japan", ParentID: parent.ID})
	createReadyAsset(t, db, "asset-1")
	if err := service.AssignAsset(ctx, "asset-1", []string{child.ID}); err != nil {
		t.Fatal(err)
	}
	ids, err := repo.AssetTagIDs(ctx, "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != parent.ID || ids[1] != child.ID {
		t.Fatalf("tag ids = %#v", ids)
	}
}

func TestMoveRemovesObsoleteParentAssociation(t *testing.T) {
	service, repo, db := newTestService(t)
	ctx := context.Background()
	oldParent, _ := service.Create(ctx, CreateInput{Name: "Old"})
	newParent, _ := service.Create(ctx, CreateInput{Name: "New"})
	child, _ := service.Create(ctx, CreateInput{Name: "Child", ParentID: oldParent.ID})
	createReadyAsset(t, db, "asset-1")
	if err := service.AssignAsset(ctx, "asset-1", []string{child.ID}); err != nil {
		t.Fatal(err)
	}
	if err := service.Move(ctx, child.ID, newParent.ID); err != nil {
		t.Fatal(err)
	}
	ids, err := repo.AssetTagIDs(ctx, "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if contains(ids, oldParent.ID) || !contains(ids, newParent.ID) || !contains(ids, child.ID) {
		t.Fatalf("tag ids after move = %#v", ids)
	}
}

func TestServiceMaintainsCategoryWhenCreatingAndMoving(t *testing.T) {
	service, repo, _ := newTestService(t)
	ctx := context.Background()
	first, err := service.Create(ctx, CreateInput{Name: "Places"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(ctx, CreateInput{Name: "People"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := service.Create(ctx, CreateInput{Name: "Japan", ParentID: first.ID})
	if err != nil {
		t.Fatal(err)
	}
	if child.Category != first.Name {
		t.Fatalf("child category = %q, want %q", child.Category, first.Name)
	}
	if err := service.Move(ctx, child.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	moved, err := repo.Get(ctx, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Category != second.Name {
		t.Fatalf("moved category = %q, want %q", moved.Category, second.Name)
	}
}

func TestServiceRejectsEmptyTagName(t *testing.T) {
	service, _, _ := newTestService(t)
	if _, err := service.Create(context.Background(), CreateInput{Name: "  "}); err != ErrInvalidName {
		t.Fatalf("Create() error = %v, want ErrInvalidName", err)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func newTestService(t *testing.T) (*Service, *Repository, *sql.DB) {
	t.Helper()
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	return NewService(repo), repo, db
}

func createReadyAsset(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.Exec(`INSERT INTO assets (id,status,original_name,original_key,sha256,mime_type,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?)`, id, "ready", id+".jpg", id+".jpg", id, "image/jpeg", now, now)
	if err != nil {
		t.Fatal(err)
	}
}
