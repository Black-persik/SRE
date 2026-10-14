package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type errKey struct{}

type errHolder struct {
	err error
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func SetError(r *http.Request, err error) {
	holder, ok := r.Context().Value(errKey{}).(*errHolder)
	if !ok {
		return
	}
	holder.err = err
}

func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			holder := &errHolder{}
			ctx := context.WithValue(r.Context(), errKey{}, holder)

			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))
			duration := time.Since(start)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration", duration,
			}
			if holder.err != nil {
				attrs = append(attrs, "error", holder.err)
				log.Error("http request", attrs...)
				return
			}

			log.Info("http request", attrs...)
		})
	}
}
