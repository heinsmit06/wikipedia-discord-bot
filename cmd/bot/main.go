package main

import (
	"discord-wiki-bot/internal/bot"
	"log"
)

func main() {
	if err := bot.DiscordBotRun(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
