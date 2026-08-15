package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MountDomains(assets, albums, tags http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Mount("/assets", assets)
	router.Mount("/albums", albums)
	router.Mount("/tags", tags)
	return router
}
