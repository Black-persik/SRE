package prober

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func stepCount(t *testing.T, s step, result string) float64 {
	t.Helper()
	var m dto.Metric
	if err := stepsTotal.WithLabelValues(string(s), result).Write(&m); err != nil {
		t.Fatalf("read counter: %v", err)
	}
	return m.GetCounter().GetValue()
}

// fakeApp mimics the API: entities are encoded without json tags, like the real handlers do.
func fakeApp(t *testing.T, failStatusUpdate bool) (*httptest.Server, *int) {
	t.Helper()
	deletes := 0
	username := ""
	status := "waiting"

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Username string }
		json.NewDecoder(r.Body).Decode(&req)
		username = req.Username
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"ID": "0199f0a0-0000-7000-8000-000000000001", "Username": username})
	})
	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ID": r.PathValue("id"), "Username": username})
	})
	mux.HandleFunc("POST /api/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"ID": "0199f0a0-0000-7000-8000-000000000002", "Status": status})
	})
	mux.HandleFunc("PATCH /api/v1/responses/{id}/status/{status}", func(w http.ResponseWriter, r *http.Request) {
		if failStatusUpdate {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		status = r.PathValue("status")
	})
	mux.HandleFunc("GET /api/v1/responses/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ID": r.PathValue("id"), "Status": status})
	})
	mux.HandleFunc("DELETE /api/v1/responses/{id}", func(w http.ResponseWriter, r *http.Request) {
		deletes++
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &deletes
}

func newTestProber(url string) *Prober {
	return New(&Config{TargetURL: url, Timeout: time.Second}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestScenarioSuccess(t *testing.T) {
	srv, deletes := fakeApp(t, false)
	before := stepCount(t, stepGetResponse, resultSuccess)

	if err := newTestProber(srv.URL).scenario(t.Context()); err != nil {
		t.Fatalf("scenario: %v", err)
	}
	if *deletes != 1 {
		t.Errorf("deletes = %d, want 1", *deletes)
	}
	if got := stepCount(t, stepGetResponse, resultSuccess); got != before+1 {
		t.Errorf("get_response successes = %v, want %v", got, before+1)
	}
}

func TestScenarioFailedStepStillDeletesResponse(t *testing.T) {
	srv, deletes := fakeApp(t, true)
	before := stepCount(t, stepUpdateStatus, resultFail)

	if err := newTestProber(srv.URL).scenario(t.Context()); err == nil {
		t.Fatal("scenario: want error, got nil")
	}
	if *deletes != 1 {
		t.Errorf("deletes = %d, want 1", *deletes)
	}
	if got := stepCount(t, stepUpdateStatus, resultFail); got != before+1 {
		t.Errorf("update_status failures = %v, want %v", got, before+1)
	}
}

func TestScenarioAppDown(t *testing.T) {
	srv, _ := fakeApp(t, false)
	srv.Close()

	if err := newTestProber(srv.URL).scenario(t.Context()); err == nil {
		t.Fatal("scenario: want error, got nil")
	}
}
