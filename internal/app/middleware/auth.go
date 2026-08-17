package middleware

import (
	"net/http"
	"strings"
)

// APIKeyAuth protects routes with X-API-Key when a non-placeholder secret is configured.
func APIKeyAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" || secret == "CHANGE_ME" {
				next.ServeHTTP(w, r)
				return
			}
			if r.Header.Get("X-API-Key") != secret {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MetricsAuth protects the metrics endpoint; uses same API key semantics.
func MetricsAuth(secret string) func(http.Handler) http.Handler {
	return APIKeyAuth(secret)
}

// SkipPaths returns a middleware that bypasses auth for exact path prefixes.
func SkipPaths(prefixes []string, auth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	protected := auth
	return func(next http.Handler) http.Handler {
		guarded := protected(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, prefix := range prefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}
			guarded.ServeHTTP(w, r)
		})
	}
}
