package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lmittmann/tint"
	"github.com/tinyhkia/bot/discord"
)

func main() {
	setupLogger()

	bot := discord.NewBot()

	log.Fatal(bot.Start())
}

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
