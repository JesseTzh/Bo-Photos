package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(db *sql.DB, spa http.Handler, versionedAPI ...http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			writeError(
				w,
				r,
				http.StatusServiceUnavailable,
				"DATABASE_UNAVAILABLE",
				"数据库暂时不可用",
			)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	if len(versionedAPI) > 0 && versionedAPI[0] != nil {
		router.Mount("/api/v1", versionedAPI[0])
	}
	if len(versionedAPI) > 1 && versionedAPI[1] != nil {
		router.Mount("/media", versionedAPI[1])
	}

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if isReservedPath(r.URL.Path) {
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "请求的资源不存在")
			return
		}
		spa.ServeHTTP(w, r)
	})

	return router
}

func isReservedPath(path string) bool {
	return path == "/api" || path == "/media" || path == "/health" ||
		hasPathPrefix(path, "/api/") ||
		hasPathPrefix(path, "/media/") ||
		hasPathPrefix(path, "/health/")
}

func hasPathPrefix(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}
