package httpx

import (
	"net/http"
	"strings"
)

// CORS 允许本地开发来源跨域（127.0.0.1/localhost 任意端口）。
// 线上由网关统一处理 CORS，本中间件仅用于本地直连开发。
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isLocalOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Dev-User, X-Request-Id")
			w.Header().Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalOrigin(origin string) bool {
	for _, prefix := range []string{"http://127.0.0.1:", "http://localhost:"} {
		if strings.HasPrefix(origin, prefix) {
			return true
		}
	}
	return origin == "http://127.0.0.1" || origin == "http://localhost"
}
