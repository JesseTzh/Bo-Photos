package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerServesAssetsAndFallsBackToIndex(t *testing.T) {
	t.Parallel()

	dist := t.TempDir()
	writeTestFile(t, filepath.Join(dist, "index.html"), `<div id="root"></div>`)
	writeTestFile(t, filepath.Join(dist, "assets", "app.js"), `console.log("ok")`)
	handler := New(dist)

	t.Run("asset", func(t *testing.T) {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, `console.log`) {
			t.Errorf("body = %q, want JavaScript asset", body)
		}
	})

	t.Run("page fallback", func(t *testing.T) {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/albums/travel", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, `<div id="root"></div>`) {
			t.Errorf("body = %q, want SPA index", body)
		}
	})
}

func TestHandlerDoesNotFallbackReservedOrUnsafeRequests(t *testing.T) {
	t.Parallel()

	dist := t.TempDir()
	writeTestFile(t, filepath.Join(dist, "index.html"), `<div id="root"></div>`)
	handler := New(dist)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "api", method: http.MethodGet, path: "/api/missing"},
		{name: "media", method: http.MethodGet, path: "/media/missing"},
		{name: "health", method: http.MethodGet, path: "/health/missing"},
		{name: "unsafe method", method: http.MethodPost, path: "/albums"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if response.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404", response.Code)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
