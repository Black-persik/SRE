package middleware

import (
	"bytes"
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
	log := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetError(r, errors.New("scan uuid"))
		apperror.Write(w, apperror.ErrInternal)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != apperror.ErrInternal.Status {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "scan uuid") {
		t.Fatal("internal error leaked to the client")
	}
	if !strings.Contains(buf.String(), "scan uuid") {
		t.Fatalf("log = %s", buf.String())
	}
}
