package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func setupLogger() {
	w := os.Stderr

	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	// Set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
		}),
	))
}
