package main

import (
	"log"
	"log/slog"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/tinyhkia/bot/discord"
)

func main() {
	setupLogger()

	token := os.Getenv("BOT_TOKEN")
	bot := discord.NewBot(token)

	log.Fatal(bot.Start())
}

func setupLogger() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}
