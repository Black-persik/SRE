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

	// Until the config is loaded, startup errors are logged in the same JSON format.
	slog.SetDefault(app.NewLogger(os.Getenv("ENVIRONMENT")))

	application, err := app.New(ctx)
	if err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}

	if err := application.Run(ctx); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
