package app

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewLoggerWritesJSONToStdout(t *testing.T) {
	for _, env := range []string{"", "dev", "staging", "prod"} {
		t.Run(env, func(t *testing.T) {
			out := captureStdout(t, func() {
				NewLogger(env).Info("hello", slog.Int("answer", 42))
			})

			var entry map[string]any
			if err := json.Unmarshal(out, &entry); err != nil {
				t.Fatalf("log is not JSON: %v\n%s", err, out)
			}
			want := map[string]any{
				"level":                  "INFO",
				"message":                "hello",
				"service.name":           ServiceName,
				"deployment.environment": env,
				"answer":                 float64(42),
			}
			for k, v := range want {
				if entry[k] != v {
					t.Errorf("%s = %v, want %v", k, entry[k], v)
				}
			}
			ts, _ := entry["timestamp"].(string)
			if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
				t.Errorf("timestamp = %q: %v", ts, err)
			}
			for _, k := range []string{"time", "msg"} {
				if _, ok := entry[k]; ok {
					t.Errorf("unexpected key %q", k)
				}
			}
		})
	}
}

func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
