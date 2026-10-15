package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"

type errKey struct{}

type errHolder struct {
	err error
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
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
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}
			w.Header().Set(RequestIDHeader, requestID)

			holder := &errHolder{}
			req := r.WithContext(context.WithValue(r.Context(), errKey{}, holder))

			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, req)
			duration := time.Since(start)

			attrs := []slog.Attr{
				slog.String("request_id", requestID),
				slog.String("http.request.method", r.Method),
				// ServeMux writes the matched pattern into the request it was given.
				slog.String("http.route", req.Pattern),
				slog.String("url.path", r.URL.Path),
				slog.Int("http.response.status_code", sw.status),
				slog.Int("http.response.body.size", sw.bytes),
				slog.Float64("duration_ms", float64(duration.Microseconds())/1000),
				slog.String("client.address", r.RemoteAddr),
				slog.String("user_agent.original", r.UserAgent()),
			}

			level := slog.LevelInfo
			switch {
			case holder.err != nil || sw.status >= http.StatusInternalServerError:
				level = slog.LevelError
			case sw.status >= http.StatusBadRequest:
				level = slog.LevelWarn
			}
			if holder.err != nil {
				attrs = append(attrs, slog.String("error", holder.err.Error()))
			}

			log.LogAttrs(r.Context(), level, "http request", attrs...)
		})
	}
}
