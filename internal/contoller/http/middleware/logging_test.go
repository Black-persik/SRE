package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/error"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingWritesHandlerError(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		SetError(r, errors.New("scan uuid"))
		apperror.Write(w, apperror.ErrInternal)
	})
	handler := Logging(log)(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	req.Header.Set(RequestIDHeader, "req-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != apperror.ErrInternal.Status {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "scan uuid") {
		t.Fatal("internal error leaked to the client")
	}
	if got := rec.Header().Get(RequestIDHeader); got != "req-1" {
		t.Fatalf("response %s = %q", RequestIDHeader, got)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log is not JSON: %v\n%s", err, buf.String())
	}
	want := map[string]any{
		"level":                     "ERROR",
		"error":                     "scan uuid",
		"request_id":                "req-1",
		"http.request.method":       http.MethodPost,
		"http.route":                "POST /api/v1/users",
		"http.response.status_code": float64(apperror.ErrInternal.Status),
	}
	for k, v := range want {
		if entry[k] != v {
			t.Errorf("%s = %v, want %v", k, entry[k], v)
		}
	}
	if _, ok := entry["duration_ms"].(float64); !ok {
		t.Errorf("duration_ms = %v, want number", entry["duration_ms"])
	}
}

func TestLoggingGeneratesRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	handler := Logging(log)(http.NotFoundHandler())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log is not JSON: %v\n%s", err, buf.String())
	}
	if entry["request_id"] == "" || entry["request_id"] != rec.Header().Get(RequestIDHeader) {
		t.Errorf("request_id = %v, header = %q", entry["request_id"], rec.Header().Get(RequestIDHeader))
	}
	if entry["level"] != "WARN" {
		t.Errorf("level = %v, want WARN", entry["level"])
	}
}
