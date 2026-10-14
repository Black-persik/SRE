package main

import (
	"context"
	"log/slog"
	"os"

	"main/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, err := app.New(ctx)
	if err != nil {
		slog.Error("startup failed", "err", err)
		os.Exit(1)
	}

	if err := application.Run(ctx); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
