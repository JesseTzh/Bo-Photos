package media

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/asset"
	"github.com/besscroft/bophotos/backend/internal/repository"
	"github.com/besscroft/bophotos/backend/internal/storage"
)

func TestPreviewVisibilityAndOriginalRange(t *testing.T) {
	repo, local := newMediaTestDependencies(t)
	item := asset.Asset{
		ID: "media", Status: asset.StatusReady, OriginalName: "旅行 照片.jpg",
		OriginalKey: "originals/media.jpg", PreviewKey: "previews/media-v1.webp",
		ThumbnailKey: "thumbnails/media-v1.webp", SHA256: "media", MIMEType: "image/jpeg",
		Visible: true, CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	writeMedia(t, local, item.OriginalKey, []byte("0123456789"))
	writeMedia(t, local, item.PreviewKey, []byte("preview"))
	writeMedia(t, local, item.ThumbnailKey, []byte("thumbnail"))

	handler := NewHandler(repo, local, func(context.Context) bool { return false }, true)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/media/preview", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("preview status = %d", response.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/assets/media/original", nil)
	request.Header.Set("Range", "bytes=2-5")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusPartialContent || response.Body.String() != "2345" {
		t.Fatalf("range status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestHiddenMediaRequiresAdministratorAndOriginalSetting(t *testing.T) {
	repo, local := newMediaTestDependencies(t)
	item := asset.Asset{
		ID: "hidden-media", Status: asset.StatusReady, OriginalName: "hidden.jpg",
		OriginalKey: "originals/hidden.jpg", PreviewKey: "previews/hidden-v1.webp",
		SHA256: "hidden-media", MIMEType: "image/jpeg", Visible: false,
		CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	writeMedia(t, local, item.OriginalKey, []byte("original"))
	writeMedia(t, local, item.PreviewKey, []byte("preview"))

	public := NewHandler(repo, local, func(context.Context) bool { return false }, false)
	response := httptest.NewRecorder()
	public.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/hidden-media/preview", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("public hidden preview status = %d, want 404", response.Code)
	}

	admin := NewHandler(repo, local, func(context.Context) bool { return true }, false)
	response = httptest.NewRecorder()
	admin.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/hidden-media/preview", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("admin hidden preview status = %d, want 200", response.Code)
	}

	visible := item
	visible.ID = "visible-original-disabled"
	visible.Visible = true
	visible.OriginalKey = "originals/visible-disabled.jpg"
	visible.PreviewKey = ""
	visible.SHA256 = visible.ID
	if err := repo.Create(context.Background(), visible); err != nil {
		t.Fatalf("Create(visible) error = %v", err)
	}
	writeMedia(t, local, visible.OriginalKey, []byte("original"))
	response = httptest.NewRecorder()
	public.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/visible-original-disabled/original", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("disabled original status = %d, want 404", response.Code)
	}
}

func TestPrivateMediaRequiresAdministrator(t *testing.T) {
	repo, local := newMediaTestDependencies(t)
	item := asset.Asset{
		ID: "private-media", Status: asset.StatusReady, OriginalName: "private.jpg",
		OriginalKey: "originals/private.jpg", PreviewKey: "previews/private-v1.webp",
		SHA256: "private-media", MIMEType: "image/jpeg", Visible: true, Private: true,
		CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	writeMedia(t, local, item.PreviewKey, []byte("preview"))

	public := NewHandler(repo, local, func(context.Context) bool { return false }, true)
	response := httptest.NewRecorder()
	public.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/private-media/preview", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("public private preview status = %d, want 404", response.Code)
	}

	admin := NewHandler(repo, local, func(context.Context) bool { return true }, true)
	response = httptest.NewRecorder()
	admin.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/private-media/preview", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("admin private preview status = %d, want 200", response.Code)
	}
}

func TestVideoContentIsPublicInlineAndSupportsRange(t *testing.T) {
	repo, local := newMediaTestDependencies(t)
	item := asset.Asset{
		ID: "intro-video", Status: asset.StatusReady, OriginalName: "intro.mp4",
		OriginalKey: "originals/intro-video.mp4", SHA256: "intro-video",
		MIMEType: "video/mp4", Visible: true, CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	writeMedia(t, local, item.OriginalKey, []byte("0123456789"))

	handler := NewHandler(repo, local, func(context.Context) bool { return false }, false)
	request := httptest.NewRequest(http.MethodGet, "/assets/intro-video/content", nil)
	request.Header.Set("Range", "bytes=3-6")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusPartialContent || response.Body.String() != "3456" {
		t.Fatalf("video range status=%d body=%q", response.Code, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "video/mp4" {
		t.Fatalf("Content-Type = %q, want video/mp4", contentType)
	}
}

func newMediaTestDependencies(t *testing.T) (*asset.Repository, *storage.Local) {
	t.Helper()
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("repository.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatalf("repository.Migrate() error = %v", err)
	}
	repo := asset.NewRepository(db)
	local, err := storage.NewLocal(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("storage.NewLocal() error = %v", err)
	}
	return repo, local
}

func writeMedia(t *testing.T, local *storage.Local, key string, data []byte) {
	t.Helper()
	path, err := local.Resolve(key)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func testTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}
