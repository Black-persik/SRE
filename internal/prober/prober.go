// Package prober walks the main user scenario of the application over its public HTTP API
// and exposes the outcome as Prometheus metrics (blackbox monitoring).
package prober

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const userAgent = "sre-tbank-prober"

// The app encodes entities without json tags; decoding is case-insensitive, so these match either way.
type user struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type response struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

type Prober struct {
	cfg    *Config
	log    *slog.Logger
	client *http.Client
}

func New(cfg *Config, log *slog.Logger) *Prober {
	return &Prober{
		cfg:    cfg,
		log:    log,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

// Run serves metrics and repeats the scenario every Interval until ctx is cancelled.
func (p *Prober) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	server := &http.Server{Addr: fmt.Sprintf(":%d", p.cfg.MetricsPort), Handler: mux}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	p.log.Info("prober started", "target", p.cfg.TargetURL, "interval", p.cfg.Interval.String(), "metrics.port", p.cfg.MetricsPort)

	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()
	for {
		p.probe(ctx)
		select {
		case err := <-errCh:
			return fmt.Errorf("metrics server: %w", err)
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		case <-ticker.C:
		}
	}
}

func (p *Prober) probe(ctx context.Context) {
	start := time.Now()
	err := p.scenario(ctx)
	if ctx.Err() != nil {
		// Shutting down: an interrupted run says nothing about the application.
		return
	}
	if err != nil {
		scenariosTotal.WithLabelValues(resultFail).Inc()
		p.log.Error("probe failed", "error", err, "duration", time.Since(start).String())
		return
	}
	scenariosTotal.WithLabelValues(resultSuccess).Inc()
	lastSuccess.SetToCurrentTime()
	p.log.Info("probe succeeded", "duration", time.Since(start).String())
}

// scenario registers a user, creates a job response, accepts it and deletes it.
// It stops at the first failed step, but always deletes a response it managed to create.
func (p *Prober) scenario(ctx context.Context) (err error) {
	// Unique per run, so registration never hits "username already exists".
	suffix := uuid.NewString()
	username := "prober-" + suffix

	var userID uuid.UUID
	if err := p.step(ctx, stepCreateUser, func(ctx context.Context) error {
		var u user
		body := map[string]string{"username": username, "email": suffix + "@prober.local"}
		if err := p.call(ctx, http.MethodPost, "/api/v1/users", body, http.StatusCreated, &u); err != nil {
			return err
		}
		if u.ID == uuid.Nil {
			return errors.New("empty user id in response")
		}
		userID = u.ID
		return nil
	}); err != nil {
		return err
	}

	if err := p.step(ctx, stepGetUser, func(ctx context.Context) error {
		var u user
		if err := p.call(ctx, http.MethodGet, "/api/v1/users/"+userID.String(), nil, http.StatusOK, &u); err != nil {
			return err
		}
		if u.Username != username {
			return fmt.Errorf("username = %q, want %q", u.Username, username)
		}
		return nil
	}); err != nil {
		return err
	}

	var responseID uuid.UUID
	if err := p.step(ctx, stepCreateResponse, func(ctx context.Context) error {
		var r response
		body := map[string]string{"user_id": userID.String(), "company": "prober", "vacancy": "probe"}
		if err := p.call(ctx, http.MethodPost, "/api/v1/responses", body, http.StatusCreated, &r); err != nil {
			return err
		}
		if r.ID == uuid.Nil {
			return errors.New("empty response id in response")
		}
		responseID = r.ID
		return nil
	}); err != nil {
		return err
	}

	defer func() {
		delErr := p.step(ctx, stepDeleteResponse, func(ctx context.Context) error {
			return p.call(ctx, http.MethodDelete, "/api/v1/responses/"+responseID.String(), nil, http.StatusOK, nil)
		})
		if err == nil {
			err = delErr
		}
	}()

	const newStatus = "accepted"
	if err := p.step(ctx, stepUpdateStatus, func(ctx context.Context) error {
		path := fmt.Sprintf("/api/v1/responses/%s/status/%s", responseID, newStatus)
		return p.call(ctx, http.MethodPatch, path, nil, http.StatusOK, nil)
	}); err != nil {
		return err
	}

	return p.step(ctx, stepGetResponse, func(ctx context.Context) error {
		var r response
		if err := p.call(ctx, http.MethodGet, "/api/v1/responses/"+responseID.String(), nil, http.StatusOK, &r); err != nil {
			return err
		}
		if r.Status != newStatus {
			return fmt.Errorf("status = %q, want %q", r.Status, newStatus)
		}
		return nil
	})
}

// step runs fn and records its duration and result.
func (p *Prober) step(ctx context.Context, s step, fn func(ctx context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	stepDuration.WithLabelValues(string(s)).Observe(time.Since(start).Seconds())

	result := resultSuccess
	if err != nil {
		result = resultFail
	}
	stepsTotal.WithLabelValues(string(s), result).Inc()

	if err != nil {
		return fmt.Errorf("%s: %w", s, err)
	}
	return nil
}

// call sends a JSON request and decodes the body into out when out is not nil.
func (p *Prober) call(ctx context.Context, method, path string, body any, wantStatus int, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(p.cfg.TargetURL, "/")+path, reqBody)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s %s: status %d, want %d: %s", method, path, resp.StatusCode, wantStatus, bytes.TrimSpace(msg))
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("%s %s: decode body: %w", method, path, err)
		}
	}
	return nil
}
