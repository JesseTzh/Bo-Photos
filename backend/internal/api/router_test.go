package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/besscroft/bophotos/backend/internal/frontend"
	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestHealthEndpoints(t *testing.T) {
	t.Parallel()

	handler, _ := newTestRouter(t)
	for _, path := range []string{"/health/live", "/health/ready"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("%s status = %d, want 200", path, response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, `"status":"ok"`) {
			t.Errorf("%s body = %q, want status ok", path, body)
		}
	}
}

func TestUnknownAPIPathDoesNotUseSPAFallback(t *testing.T) {
	t.Parallel()

	handler, _ := newTestRouter(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/missing", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
	if body := response.Body.String(); !strings.Contains(body, `"code":"NOT_FOUND"`) {
		t.Errorf("body = %q, want NOT_FOUND error", body)
	}
}

func TestUnknownPageUsesSPAFallback(t *testing.T) {
	t.Parallel()

	handler, _ := newTestRouter(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/albums/travel", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, `<div id="root"></div>`) {
		t.Errorf("body = %q, want SPA index", body)
	}
}

func TestReadinessReportsDatabaseFailure(t *testing.T) {
	t.Parallel()

	handler, db := newTestRouter(t)
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, `"code":"DATABASE_UNAVAILABLE"`) {
		t.Errorf("body = %q, want DATABASE_UNAVAILABLE", body)
	}
}

func TestRouterMountsVersionedAPIRoutes(t *testing.T) {
	t.Parallel()

	_, db := newTestRouter(t)
	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte(`<div id="root"></div>`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	versionedAPI := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := NewRouter(db, frontend.New(dist), versionedAPI)

	response := httptest.NewRecorder()
	handler.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil),
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
}

func newTestRouter(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("repository.Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte(`<div id="root"></div>`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return NewRouter(db, frontend.New(dist)), db
}
