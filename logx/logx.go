// Package logx 统一结构化日志：固定字段 svc/traceId/requestId/userId（见 docs/04 §2.6）。
package logx

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyUserID
)

// Init 返回带服务名的根 logger 并设为默认。format: json|text。
func Init(svc, level, format string) *slog.Logger {
	var lv slog.Level
	switch level {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lv}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	l := slog.New(h).With("svc", svc)
	slog.SetDefault(l)
	return l
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func WithUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, keyUserID, uid)
}

func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(keyRequestID).(string)
	return v
}

func UserID(ctx context.Context) string {
	v, _ := ctx.Value(keyUserID).(string)
	return v
}

// L 返回携带 ctx 内 requestId/userId 字段的 logger。
func L(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if id := RequestID(ctx); id != "" {
		l = l.With("requestId", id)
	}
	if uid := UserID(ctx); uid != "" {
		l = l.With("userId", uid)
	}
	return l
}
