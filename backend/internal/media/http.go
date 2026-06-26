package media

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/besscroft/bophotos/backend/internal/asset"
	"github.com/besscroft/bophotos/backend/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repository             *asset.Repository
	storage                *storage.Local
	isAdministrator        func(context.Context) bool
	publicOriginalDownload func(context.Context) bool
}

func NewHandler(
	repository *asset.Repository,
	localStorage *storage.Local,
	isAdministrator func(context.Context) bool,
	publicOriginalDownload any,
) http.Handler {
	allow := func(context.Context) bool { return true }
	switch value := publicOriginalDownload.(type) {
	case bool:
		allow = func(context.Context) bool { return value }
	case func(context.Context) bool:
		allow = value
	}
	handler := &Handler{
		repository: repository, storage: localStorage,
		isAdministrator: isAdministrator, publicOriginalDownload: allow,
	}
	router := chi.NewRouter()
	router.Get("/assets/{id}/{variant}", handler.serve)
	return router
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request) {
	item, err := h.repository.Get(r.Context(), chi.URLParam(r, "id"))
	admin := h.isAdministrator != nil && h.isAdministrator(r.Context())
	if errors.Is(err, asset.ErrNotFound) || item.Status == asset.StatusDeleted || item.Status == asset.StatusPurged {
		http.NotFound(w, r)
		return
	}
	if err != nil || item.Status != asset.StatusReady || ((!item.Visible || item.Private) && !admin) {
		http.NotFound(w, r)
		return
	}

	variant := chi.URLParam(r, "variant")
	var key string
	switch variant {
	case "thumbnail":
		key = item.ThumbnailKey
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	case "preview":
		key = item.PreviewKey
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	case "original":
		if !h.publicOriginalDownload(r.Context()) && !admin {
			http.NotFound(w, r)
			return
		}
		key = item.OriginalKey
		w.Header().Set("Content-Disposition", contentDisposition(item.OriginalName))
	default:
		http.NotFound(w, r)
		return
	}
	if key == "" {
		http.NotFound(w, r)
		return
	}

	path, err := h.storage.Resolve(key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if variant != "original" {
		contentType = "image/webp"
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(w, r, item.OriginalName, info.ModTime(), file)
}

func contentDisposition(name string) string {
	safe := strings.NewReplacer("\r", "", "\n", "", `"`, "'").Replace(filepath.Base(name))
	return fmt.Sprintf(`attachment; filename=%q; filename*=UTF-8''%s`, safe, url.QueryEscape(safe))
}
