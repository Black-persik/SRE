package app

import (
	"log/slog"
	"os"
)

func NewLogger(environment string) *slog.Logger {
	opts := &slog.HandlerOptions{}
	var handler slog.Handler
	switch environment {
	case "prod":
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "staging":
		opts.Level = slog.LevelDebug
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
