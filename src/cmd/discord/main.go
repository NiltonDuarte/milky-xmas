package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/discord"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
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

	path := strings.TrimSpace(os.Getenv("MILKY_GIFT_DB"))
	if path == "" {
		log.Fatal("MILKY_GIFT_DB is required")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, token, guildID, path); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, token, guildID, path string) error {
	store, err := giftstore.OpenSQLite(ctx, path)
	if err != nil {
		return fmt.Errorf("open gift database: %w", err)
	}
	defer store.Close()

	return discord.Run(ctx, token, guildID, app.NewService(store, time.Now))
}
