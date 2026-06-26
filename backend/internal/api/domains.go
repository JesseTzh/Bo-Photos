package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MountDomains(assets, albums, tags http.Handler, guides ...http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Mount("/assets", assets)
	router.Mount("/albums", albums)
	router.Mount("/tags", tags)
	if len(guides) > 0 && guides[0] != nil {
		router.Mount("/guides", guides[0])
	}
	return router
}
