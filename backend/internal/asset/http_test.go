package asset

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/besscroft/bophotos/backend/internal/imageproc"
	"github.com/besscroft/bophotos/backend/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func TestAdminUploadReturnsAcceptedAndDuplicateWarning(t *testing.T) {
	handler, _, _ := newTestAssetHTTP(t)
	content := append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("duplicate")...)

	first := uploadRequest(t, handler, "first.jpg", content)
	if first.Code != http.StatusAccepted {
		t.Fatalf("first status = %d body=%s", first.Code, first.Body.String())
	}
	second := uploadRequest(t, handler, "second.jpg", content)
	if second.Code != http.StatusAccepted {
		t.Fatalf("second status = %d body=%s", second.Code, second.Body.String())
	}
	var envelope struct {
		Data struct {
			Status            Status   `json:"status"`
			DuplicateAssetIDs []string `json:"duplicate_asset_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if envelope.Data.Status != StatusProcessing || len(envelope.Data.DuplicateAssetIDs) != 1 {
		t.Fatalf("payload = %#v", envelope.Data)
	}
}

func TestPublicAssetsExcludeHiddenAndProcessing(t *testing.T) {
	handler, repo, _ := newTestAssetHTTP(t)
	visible := testAsset("visible", StatusReady)
	visible.Visible = true
	mustCreate(t, repo, visible)
	hidden := testAsset("hidden", StatusReady)
	hidden.Visible = false
	mustCreate(t, repo, hidden)
	processing := testAsset("processing-api", StatusProcessing)
	processing.Visible = true
	mustCreate(t, repo, processing)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets?page=1&page_size=16", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data struct {
			Items []AssetDTO `json:"items"`
			Total int        `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if envelope.Data.Total != 1 || len(envelope.Data.Items) != 1 || envelope.Data.Items[0].ID != visible.ID {
		t.Fatalf("payload = %#v", envelope.Data)
	}
}

func TestAssetErrorsIncludeRequestID(t *testing.T) {
	handler, _, _ := newTestAssetHTTP(t)
	handler = middleware.RequestID(handler)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
	var payload struct {
		Error struct {
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Error.RequestID == "" {
		t.Fatal("request_id is empty")
	}
}

func TestAdminUpdateMapsSnakeCaseFields(t *testing.T) {
	handler, repo, _ := newTestAssetHTTP(t)
	item := testAsset("update-json", StatusReady)
	mustCreate(t, repo, item)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/admin/assets/update-json",
		bytes.NewBufferString(`{"title":"New title","show_on_homepage":false,"featured":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	updated, err := repo.Get(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if updated.Title != "New title" || updated.ShowOnHomepage || !updated.Featured {
		t.Fatalf("updated asset = %#v", updated)
	}
}

func TestAdminPurgeDeletedAsset(t *testing.T) {
	handler, repo, local := newTestAssetHTTP(t)
	item := testAsset("purge-api", StatusDeleted)
	mustCreate(t, repo, item)
	path, err := local.Resolve(item.OriginalKey)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/admin/assets/purge-api/purge", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	updated, err := repo.Get(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if updated.Status != StatusPurged {
		t.Fatalf("status = %q, want purged", updated.Status)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("original file stat error = %v, want not exist", err)
	}
}

func newTestAssetHTTP(t *testing.T) (http.Handler, *Repository, *storage.Local) {
	t.Helper()
	repo := newTestRepository(t)
	local, err := storage.NewLocal(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("storage.NewLocal() error = %v", err)
	}
	queue := &fakeQueue{}
	service := NewService(repo, local, fakeProcessor{metadata: imageproc.Metadata{}}, queue, 10<<20)
	handler := NewHTTPHandler(service, repo)
	router := chi.NewRouter()
	router.Mount("/assets", handler.PublicRoutes())
	router.Mount("/admin/assets", handler.AdminRoutes())
	return router, repo, local
}

func uploadRequest(t *testing.T, handler http.Handler, name string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/admin/assets", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request.WithContext(context.Background()))
	return response
}
