package main

import (
	"log/slog"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/tinyhkia/bot/discord"
)

func main() {
	setupLogger()
	bot := discord.NewBot()

	bot.Handle("ping", HandlePING)
	
	slog.Error(bot.Start().Error())
	os.Exit(1)
}

func HandlePING(ctx discord.Context) error {
	response := discord.InteractionResponse{
		Type: discord.Message,
		Data: &discord.Data{
			Content: "ayeeeee",
		},
	}
	return ctx.Respond(response)
}
