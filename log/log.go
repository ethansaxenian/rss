package log

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type canonicalKeyCtxKeyType struct{}

var canonicalKeyCtxKey = canonicalKeyCtxKeyType{}

type canonicalAttrs struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

func (ca *canonicalAttrs) append(attrs []slog.Attr) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	ca.attrs = append(ca.attrs, attrs...)
}

func Middleware(log *slog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()

			ctx := context.WithValue(r.Context(), canonicalKeyCtxKey, &canonicalAttrs{})
			r = r.WithContext(ctx)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				end := time.Now().UTC()

				status := ww.Status()

				level := slog.LevelInfo
				if status >= http.StatusBadRequest {
					level = slog.LevelError
				}

				attrs := []slog.Attr{
					slog.GroupAttrs(
						"request",
						slog.String("method", r.Method),
						slog.String("host", r.Host),
						slog.String("path", r.URL.Path),
						slog.String("pattern", r.Pattern),
						slog.String("query", r.URL.RawQuery),
						slog.Time("time", start),
						slog.Duration("duration", time.Since(start)),
						slog.String("remote_addr", r.RemoteAddr),
					),
					slog.GroupAttrs(
						"response",
						slog.Int("status_code", status),
						slog.String("status_text", http.StatusText(status)),
						slog.Time("time", end),
					),
				}
				attrs = append(attrs, getCanonicalAttrs(ctx)...)

				log.LogAttrs(ctx, level, "canonical log", attrs...)
			}()

			h.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func Add(ctx context.Context, attrs ...slog.Attr) {
	ca, ok := ctx.Value(canonicalKeyCtxKey).(*canonicalAttrs)
	if !ok || ca == nil {
		return
	}

	ca.append(attrs)
}

func getCanonicalAttrs(ctx context.Context) []slog.Attr {
	ca, ok := ctx.Value(canonicalKeyCtxKey).(*canonicalAttrs)
	if !ok || ca == nil {
		return []slog.Attr{}
	}

	return ca.attrs
}
