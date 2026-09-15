// Package logger configures the application-wide structured logger.
package logger

import (
	"log/slog"
	"os"
)

// New creates a JSON structured logger writing to stdout.
func New(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
