package app

import (
	"log/slog"
	"os"
	"time"
)

const ServiceName = "sre_tbank"

// NewLogger returns a JSON logger writing to stdout. The output format is the same
// in every environment, only the level differs.
func NewLogger(environment string) *slog.Logger {
	level := slog.LevelDebug
	if environment == "prod" {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceAttr,
	})

	hostname, _ := os.Hostname()

	return slog.New(handler).With(
		slog.String("service.name", ServiceName),
		slog.String("deployment.environment", environment),
		slog.String("host.name", hostname),
	)
}

// replaceAttr renames slog's built-in keys to the agreed field names.
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}
	switch a.Key {
	case slog.TimeKey:
		a.Key = "timestamp"
		a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339Nano))
	case slog.MessageKey:
		a.Key = "message"
	}
	return a
}
