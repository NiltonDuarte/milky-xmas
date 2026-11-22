package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/discord"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func main() {
	token := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN is required")
	}

	guildID := strings.TrimSpace(os.Getenv("DISCORD_TEST_GUILD_ID"))
	if id, err := strconv.ParseUint(guildID, 10, 64); err != nil || id == 0 {
		log.Fatal("DISCORD_TEST_GUILD_ID must be a valid Discord server ID")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := discord.Run(ctx, token, guildID, app.Service{}); err != nil {
		log.Fatal(err)
	}
}
