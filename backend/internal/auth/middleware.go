package auth

import (
	"net/http"
	"net/url"
)

func (h *HTTPHandler) requireAdministrator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.sessions.GetBool(r.Context(), sessionAdminKey) {
			writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "需要管理员登录")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHeader := r.Header.Get("Origin")
		if isSafeMethod(r.Method) || originHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		origin, err := url.Parse(originHeader)
		if err != nil || origin.Scheme != requestScheme(r) || origin.Host != r.Host {
			writeError(w, r, http.StatusForbidden, "ORIGIN_FORBIDDEN", "请求来源不被允许")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestScheme(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded == "http" || forwarded == "https" {
		return forwarded
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
