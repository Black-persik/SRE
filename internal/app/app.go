package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/api"
	"main/config"
	"main/internal/contoller/http/middleware"
	"main/web"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	cfg       *config.Config
	log       *slog.Logger
	container *Container
	server    *http.Server
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.NewConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewConfig: %w", err)
	}

	log := NewLogger(cfg.App.Environment)
	slog.SetDefault(log)

	container, err := NewContainer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("NewContainer: %w", err)
	}

	mux := http.NewServeMux()
	api.Register(mux)
	mux.HandleFunc("POST /api/v1/users", container.Controllers.Users.CreateUser)
	mux.HandleFunc("GET /api/v1/users/{id}", container.Controllers.Users.GetUserByID)
	mux.HandleFunc("GET /api/v1/users/username/{username}", container.Controllers.Users.GetUserByUsername)
	mux.HandleFunc("PATCH /api/v1/users/{id}/email/{email}", container.Controllers.Users.UpdateUserEmail)
	mux.HandleFunc("GET /api/v1/responses/user/{user_id}", container.Controllers.Responses.GetResponsesByUserID)

	mux.HandleFunc("POST /api/v1/responses", container.Controllers.Responses.CreateResponse)
	mux.HandleFunc("GET /api/v1/responses/{id}", container.Controllers.Responses.GetResponseByID)
	mux.HandleFunc("PATCH /api/v1/responses/{id}/status/{status}", container.Controllers.Responses.UpdateResponseStatus)
	mux.HandleFunc("DELETE /api/v1/responses/{id}", container.Controllers.Responses.DeleteResponse)
	mux.Handle("GET /", http.FileServer(http.FS(web.Files)))

	return &App{
		cfg:       cfg,
		log:       log,
		container: container,
		server: &http.Server{
			Addr:     fmt.Sprintf(":%d", cfg.App.Port),
			Handler:  middleware.Logging(log)(mux),
			ErrorLog: slog.NewLogLogger(log.Handler(), slog.LevelError),
		},
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.container.Close()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		a.log.Info("server started", "server.port", a.cfg.App.Port)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		a.log.Info("shutdown")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server.Shutdown: %w", err)
	}
	return nil
}
