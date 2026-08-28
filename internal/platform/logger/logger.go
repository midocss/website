package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New builds the process-wide structured logger: human readable text in
// development, JSON in every other environment.
func New(environment string, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if strings.EqualFold(environment, "development") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With("env", environment)
}
