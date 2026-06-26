package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestDomainRoutersCanBeMountedWithoutConflicts(t *testing.T) {
	root := MountDomains(
		statusHandler(http.StatusOK),
		statusHandler(http.StatusCreated),
		statusHandler(http.StatusAccepted),
	)

	for path, want := range map[string]int{
		"/assets": http.StatusOK,
		"/albums": http.StatusCreated,
		"/tags":   http.StatusAccepted,
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		root.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s status = %d, want %d", path, response.Code, want)
		}
	}
}

func statusHandler(status int) http.Handler {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	return router
}
