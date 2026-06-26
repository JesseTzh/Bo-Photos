package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/besscroft/bophotos/backend/internal/repository"
	"github.com/go-chi/chi/v5/middleware"
)

func TestAuthenticationHTTPFlow(t *testing.T) {
	handler, _ := newTestHTTPHandler(t)

	response := requestJSON(t, handler, http.MethodPost, "/auth/login", map[string]string{
		"password": "wrong password",
	}, nil)
	assertStatus(t, response, http.StatusUnauthorized)

	response = requestJSON(t, handler, http.MethodPost, "/auth/login", map[string]string{
		"password": "correct horse battery staple",
	}, nil)
	assertStatus(t, response, http.StatusOK)
	cookie := findSessionCookie(t, response)

	response = requestJSON(t, handler, http.MethodGet, "/admin/ping", nil, cookie)
	assertStatus(t, response, http.StatusOK)
}

func TestAdminRequiresSession(t *testing.T) {
	handler, _ := newTestHTTPHandler(t)
	handler = middleware.RequestID(handler)

	response := requestJSON(t, handler, http.MethodGet, "/admin/ping", nil, nil)
	assertStatus(t, response, http.StatusUnauthorized)
	assertErrorCode(t, response, "AUTH_REQUIRED")
	assertRequestID(t, response)
}

func TestMutationRejectsCrossOriginRequest(t *testing.T) {
	handler, _ := newTestHTTPHandler(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"http://example.com/auth/login",
		bytes.NewBufferString(`{"password":"anything"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	assertStatus(t, response, http.StatusForbidden)
	assertErrorCode(t, response, "ORIGIN_FORBIDDEN")
}

func TestPasswordChangeRevokesAllSessions(t *testing.T) {
	handler, db := newTestHTTPHandler(t)

	firstLogin := requestJSON(t, handler, http.MethodPost, "/auth/login", map[string]string{
		"password": "correct horse battery staple",
	}, nil)
	firstCookie := findSessionCookie(t, firstLogin)
	secondLogin := requestJSON(t, handler, http.MethodPost, "/auth/login", map[string]string{
		"password": "correct horse battery staple",
	}, nil)
	secondCookie := findSessionCookie(t, secondLogin)

	response := requestJSON(t, handler, http.MethodPut, "/admin/password", map[string]string{
		"current_password": "correct horse battery staple",
		"new_password":     "new correct horse battery staple",
	}, firstCookie)
	assertStatus(t, response, http.StatusOK)

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("session count = %d, want 0", count)
	}

	for _, cookie := range []*http.Cookie{firstCookie, secondCookie} {
		response := requestJSON(t, handler, http.MethodGet, "/admin/ping", nil, cookie)
		assertStatus(t, response, http.StatusUnauthorized)
	}
}

func TestLoginRateLimitAfterFiveFailures(t *testing.T) {
	handler, _ := newTestHTTPHandler(t)

	for attempt := 0; attempt < 5; attempt++ {
		response := requestJSON(t, handler, http.MethodPost, "/auth/login", map[string]string{
			"password": "wrong password",
		}, nil)
		assertStatus(t, response, http.StatusUnauthorized)
	}

	response := requestJSON(t, handler, http.MethodPost, "/auth/login", map[string]string{
		"password": "correct horse battery staple",
	}, nil)
	assertStatus(t, response, http.StatusTooManyRequests)
	assertErrorCode(t, response, "LOGIN_RATE_LIMITED")
}

func newTestHTTPHandler(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("repository.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatalf("repository.Migrate() error = %v", err)
	}

	service := NewService(NewRepository(db))
	if err := service.EnsureInitialized(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("EnsureInitialized() error = %v", err)
	}
	sessions, stopCleanup := NewSessionManager(db, false)
	t.Cleanup(stopCleanup)
	return NewHTTPHandler(service, sessions).Routes(), db
}

func requestJSON(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body any,
	cookie *http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}
	request := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func findSessionCookie(t *testing.T, response *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			return cookie
		}
	}
	t.Fatalf("response has no %s cookie", sessionCookieName)
	return nil
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, want, response.Body.String())
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Error.Code != want {
		t.Errorf("error code = %q, want %q", payload.Error.Code, want)
	}
}

func assertRequestID(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var payload struct {
		Error struct {
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Error.RequestID == "" {
		t.Error("request_id is empty")
	}
}
