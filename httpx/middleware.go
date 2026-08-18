package httpx

import (
	"net/http"
	"strings"
	"time"

	"github.com/tommax-bai/tommax-go-kit/errs"
	"github.com/tommax-bai/tommax-go-kit/idgen"
	"github.com/tommax-bai/tommax-go-kit/logx"
)

// Chain 按 docs/03 §1.2 的统一顺序组装中间件：recovery → requestId → logging → auth。
// tracing/metrics/ratelimit 属 Phase 1 之后补齐项。
func Chain(next http.Handler, auth func(http.Handler) http.Handler) http.Handler {
	h := next
	if auth != nil {
		h = auth(h)
	}
	h = Logging(h)
	h = RequestID(h)
	h = Recovery(h)
	return h
}

// Recovery 捕获 panic，转 500 响应。
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logx.L(r.Context()).Error("panic recovered", "panic", rec)
				Fail(w, r, errs.ErrInternal)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestID 注入/透传 X-Request-Id。
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = idgen.NextString()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(logx.WithRequestID(r.Context(), id)))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Logging 访问日志（RED 指标的日志侧）。
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logx.L(r.Context()).Info("http",
			"method", r.Method, "path", r.URL.Path,
			"status", sw.status, "durMs", time.Since(start).Milliseconds())
	})
}

// DevAuth 是 Phase 1 纵向切片的临时鉴权（终态由 Casdoor JWT 替代，见 docs/07）：
// 接受 `Authorization: Bearer dev-<userId>` 或 `X-Dev-User: <userId>`。
// 仅允许在 dev 环境启用；启用时会打警告日志。
func DevAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Header.Get("X-Dev-User")
		if uid == "" {
			bearer := r.Header.Get("Authorization")
			if v, ok := strings.CutPrefix(bearer, "Bearer dev-"); ok {
				uid = v
			}
		}
		if uid == "" {
			Fail(w, r, errs.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(logx.WithUserID(r.Context(), uid)))
	})
}
