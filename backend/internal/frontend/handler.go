package frontend

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	distDir    string
	fileServer http.Handler
}

func New(distDir string) http.Handler {
	return &Handler{
		distDir:    distDir,
		fileServer: http.FileServer(http.Dir(distDir)),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !isPageMethod(r.Method) || isReservedPath(r.URL.Path) {
		http.NotFound(w, r)
		return
	}

	requestedPath := filepath.Join(h.distDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
	if info, err := os.Stat(requestedPath); err == nil && !info.IsDir() {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	http.ServeFile(w, r, filepath.Join(h.distDir, "index.html"))
}

func isPageMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func isReservedPath(path string) bool {
	for _, prefix := range []string{"/api", "/media", "/health"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
