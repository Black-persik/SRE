package prober

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	dto "github.com/prometheus/client_model/go"
)

func checkCount(t *testing.T, c check, result string) float64 {
	t.Helper()
	var m dto.Metric
	if err := checksTotal.WithLabelValues(string(c), result).Write(&m); err != nil {
		t.Fatalf("read counter: %v", err)
	}
	return m.GetCounter().GetValue()
}

// fakeApp mimics the API: entities are encoded without json tags, like the real handlers do.
type fakeApp struct {
	failCreateUser   bool
	failStatusUpdate bool

	username        string
	status          string
	createResponses int
	deletes         int
	statusUpdates   int
}

func (f *fakeApp) server(t *testing.T) *httptest.Server {
	t.Helper()
	internal := func(w http.ResponseWriter) {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		if f.failCreateUser {
			internal(w)
			return
		}
		var req struct{ Username string }
		json.NewDecoder(r.Body).Decode(&req)
		f.username = req.Username
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"ID": "0199f0a0-0000-7000-8000-000000000001", "Username": f.username})
	})
	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ID": r.PathValue("id"), "Username": f.username})
	})
	mux.HandleFunc("POST /api/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		f.createResponses++
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"ID": "0199f0a0-0000-7000-8000-000000000002", "Status": "waiting"})
	})
	mux.HandleFunc("PATCH /api/v1/responses/{id}/status/{status}", func(w http.ResponseWriter, r *http.Request) {
		f.statusUpdates++
		if f.failStatusUpdate {
			internal(w)
			return
		}
		f.status = r.PathValue("status")
	})
	mux.HandleFunc("GET /api/v1/responses/{id}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ID": r.PathValue("id"), "Status": f.status})
	})
	mux.HandleFunc("DELETE /api/v1/responses/{id}", func(w http.ResponseWriter, r *http.Request) {
		f.deletes++
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestProber(url string) *Prober {
	return New(&Config{TargetURL: url, Timeout: time.Second}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestProbeAllChecksSucceed(t *testing.T) {
	app := &fakeApp{}
	p := newTestProber(app.server(t).URL)

	// The first run creates the fixtures, the second one runs the full checks on them.
	for range 2 {
		if err := p.checkRegistration(t.Context()); err != nil {
			t.Fatalf("checkRegistration: %v", err)
		}
		if err := p.checkResponse(t.Context()); err != nil {
			t.Fatalf("checkResponse: %v", err)
		}
		if err := p.checkStatus(t.Context()); err != nil {
			t.Fatalf("checkStatus: %v", err)
		}
	}
	if app.deletes != 1 {
		t.Errorf("deletes = %d, want 1: the fixture response must be kept, the second one deleted", app.deletes)
	}
}

func TestProbeKeepsCheckingResponsesWhenRegistrationFails(t *testing.T) {
	app := &fakeApp{}
	p := newTestProber(app.server(t).URL)
	p.probe(t.Context())

	app.failCreateUser = true
	regFails := checkCount(t, checkRegistration, resultFail)
	respOK := checkCount(t, checkResponse, resultSuccess)
	statusOK := checkCount(t, checkStatus, resultSuccess)
	creates, updates := app.createResponses, app.statusUpdates

	p.probe(t.Context())

	if got := checkCount(t, checkRegistration, resultFail); got != regFails+1 {
		t.Errorf("registration failures = %v, want %v", got, regFails+1)
	}
	if got := checkCount(t, checkResponse, resultSuccess); got != respOK+1 {
		t.Errorf("response successes = %v, want %v", got, respOK+1)
	}
	if got := checkCount(t, checkStatus, resultSuccess); got != statusOK+1 {
		t.Errorf("status successes = %v, want %v", got, statusOK+1)
	}
	if app.createResponses != creates+1 || app.statusUpdates != updates+1 {
		t.Errorf("app got %d creates and %d status updates, want one more of each",
			app.createResponses-creates, app.statusUpdates-updates)
	}
}

func TestProbeStatusFailureDoesNotAffectOtherChecks(t *testing.T) {
	app := &fakeApp{}
	p := newTestProber(app.server(t).URL)
	p.probe(t.Context())

	app.failStatusUpdate = true
	statusFails := checkCount(t, checkStatus, resultFail)
	respOK := checkCount(t, checkResponse, resultSuccess)

	p.probe(t.Context())

	if got := checkCount(t, checkStatus, resultFail); got != statusFails+1 {
		t.Errorf("status failures = %v, want %v", got, statusFails+1)
	}
	if got := checkCount(t, checkResponse, resultSuccess); got != respOK+1 {
		t.Errorf("response successes = %v, want %v", got, respOK+1)
	}
	if p.fixtureResponse == uuid.Nil {
		t.Error("fixture response dropped on a 500, want it kept")
	}
}

func TestProbeAppDown(t *testing.T) {
	srv := (&fakeApp{}).server(t)
	srv.Close()

	if err := newTestProber(srv.URL).checkRegistration(t.Context()); err == nil {
		t.Fatal("checkRegistration: want error, got nil")
	}
}
