package main

import (
	"context"
	"log/slog"

	_ "github.com/joho/godotenv/autoload"
	"github.com/tinyhkia/bot/discord"
)

func main() {
	setupLogger() // todo: add bot Logger

	b := discord.NewBot(context.Background())

	b.HandleCommand("ping", PingHandler)
	b.HandleCommand("characters", CharactersHandler)
	slog.Error(b.Start().Error())
}

func PingHandler(ctx discord.Context) error {
	return ctx.SendContent("PONG")
}

func CharactersHandler(ctx discord.Context) error {
	return ctx.SendContent("kuromi, keropi, hello kitty")
}
