package guide

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestServiceCreatesAndPublishesGuide(t *testing.T) {
	service, repo := newTestService(t)
	ctx := context.Background()
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(48 * time.Hour)
	item, err := service.Create(ctx, GuideInput{
		Title: "Japan", Country: "Japan", City: "Tokyo", Days: 3,
		StartDate: &start, EndDate: &end, Published: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	public, err := repo.List(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(public) != 1 || public[0].ID != item.ID {
		t.Fatalf("public guides = %#v", public)
	}
}

func TestServiceRejectsInvalidGuideDates(t *testing.T) {
	service, _ := newTestService(t)
	start := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	end := start.Add(-time.Hour)
	_, err := service.Create(context.Background(), GuideInput{
		Title: "Invalid", StartDate: &start, EndDate: &end,
	})
	if err != ErrInvalidDates {
		t.Fatalf("Create() error = %v, want ErrInvalidDates", err)
	}
}

func TestRepositorySoftDeleteHidesGuide(t *testing.T) {
	service, repo := newTestService(t)
	item, err := service.Create(context.Background(), GuideInput{Title: "Delete me", Published: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(context.Background(), item.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(context.Background(), item.ID, false); err != ErrNotFound {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestRepositorySortsGuidesTransactionally(t *testing.T) {
	service, repo := newTestService(t)
	first, _ := service.Create(context.Background(), GuideInput{Title: "First"})
	second, _ := service.Create(context.Background(), GuideInput{Title: "Second"})
	if err := repo.Sort(context.Background(), []string{second.ID, first.ID}); err != nil {
		t.Fatal(err)
	}
	items, err := repo.List(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if items[0].ID != second.ID || items[1].ID != first.ID {
		t.Fatalf("ordered guides = %#v", items)
	}
	if err := repo.Sort(context.Background(), []string{first.ID, "missing"}); err != ErrInvalidOwnership {
		t.Fatalf("Sort() error = %v, want ErrInvalidOwnership", err)
	}
}

func newTestService(t *testing.T) (*Service, *Repository) {
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
	return NewService(repo), repo
}
