package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"main/internal/app"
	"main/internal/prober"
)

func main() {
	slog.SetDefault(app.NewLogger(os.Getenv("ENVIRONMENT")))

	cfg, err := prober.NewConfig()
	if err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := prober.New(cfg, slog.Default()).Run(ctx); err != nil {
		slog.Error("prober stopped", "error", err)
		os.Exit(1)
	}
}
