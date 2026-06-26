package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	apiresponse "github.com/besscroft/bophotos/backend/internal/api"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	service  *Service
	sessions *scs.SessionManager
	limiter  *failureLimiter
}

type passwordRequest struct {
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func NewHTTPHandler(service *Service, sessions *scs.SessionManager) *HTTPHandler {
	return &HTTPHandler{
		service:  service,
		sessions: sessions,
		limiter:  newFailureLimiter(5, 5*time.Minute),
	}
}

func (h *HTTPHandler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Use(h.sessions.LoadAndSave)
	router.Use(sameOrigin)

	router.Post("/auth/login", h.login)
	router.Post("/auth/logout", h.logout)
	router.Get("/auth/session", h.session)

	router.Route("/admin", func(router chi.Router) {
		router.Use(h.requireAdministrator)
		router.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
		})
		router.Put("/password", h.changePassword)
	})
	return router
}

func (h *HTTPHandler) SessionMiddleware(next http.Handler) http.Handler {
	return h.sessions.LoadAndSave(next)
}

func (h *HTTPHandler) ProtectAdministrator(next http.Handler) http.Handler {
	return sameOrigin(h.requireAdministrator(next))
}

func (h *HTTPHandler) IsAdministrator(ctx context.Context) bool {
	return h.sessions.GetBool(ctx, sessionAdminKey)
}

func (h *HTTPHandler) login(w http.ResponseWriter, r *http.Request) {
	key := clientKey(r)
	if !h.limiter.Allow(key) {
		writeError(w, r, http.StatusTooManyRequests, "LOGIN_RATE_LIMITED", "登录尝试过于频繁，请稍后重试")
		return
	}

	var request passwordRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.service.VerifyPassword(r.Context(), request.Password); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.limiter.Failed(key)
			writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "密码不正确")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "登录失败")
		return
	}

	h.limiter.Reset(key)
	if err := h.sessions.RenewToken(r.Context()); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "创建登录会话失败")
		return
	}
	h.sessions.Put(r.Context(), sessionAdminKey, true)
	writeJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessions.Destroy(r.Context()); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "退出登录失败")
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) session(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.service.IsInitialized(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "无法读取认证状态")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{
		"initialized":   initialized,
		"authenticated": h.sessions.GetBool(r.Context(), sessionAdminKey),
	})
}

func (h *HTTPHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	var request changePasswordRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式不正确")
		return
	}

	switch err := h.service.ChangePassword(r.Context(), request.CurrentPassword, request.NewPassword); {
	case errors.Is(err, ErrInvalidCredentials):
		writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "当前密码不正确")
		return
	case errors.Is(err, ErrInvalidPassword):
		writeError(w, r, http.StatusUnprocessableEntity, "PASSWORD_INVALID", "新密码长度必须为 12 到 128 个字符")
		return
	case err != nil:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "修改密码失败")
		return
	}

	if err := h.sessions.Destroy(r.Context()); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "撤销当前会话失败")
		return
	}
	if err := h.service.RevokeSessions(r.Context()); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "撤销登录会话失败")
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	apiresponse.WriteJSON(w, status, value)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	apiresponse.WriteError(w, r, status, code, message)
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

type failureWindow struct {
	count   int
	expires time.Time
}

type failureLimiter struct {
	mu       sync.Mutex
	max      int
	duration time.Duration
	failures map[string]failureWindow
}

func newFailureLimiter(max int, duration time.Duration) *failureLimiter {
	return &failureLimiter{
		max:      max,
		duration: duration,
		failures: make(map[string]failureWindow),
	}
}

func (l *failureLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	window, exists := l.failures[key]
	if !exists || time.Now().After(window.expires) {
		delete(l.failures, key)
		return true
	}
	return window.count < l.max
}

func (l *failureLimiter) Failed(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	window, exists := l.failures[key]
	if !exists || now.After(window.expires) {
		l.failures[key] = failureWindow{count: 1, expires: now.Add(l.duration)}
		return
	}
	window.count++
	l.failures[key] = window
}

func (l *failureLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}
