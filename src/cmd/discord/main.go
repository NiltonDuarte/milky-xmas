package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
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
	channel := strings.TrimSpace(os.Getenv("MILKY_GRINCH_CHANNEL_ID"))
	if channel != "" {
		if id, err := strconv.ParseUint(channel, 10, 64); err != nil || id == 0 {
			log.Fatal("MILKY_GRINCH_CHANNEL_ID must be a valid Discord channel ID")
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, token, guildID, path, channel); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, token, guildID, path, channel string) error {
	store, err := giftstore.OpenSQLite(ctx, path)
	if err != nil {
		return fmt.Errorf("open gift database: %w", err)
	}
	defer store.Close()

	encounter := app.NewGrinchEncounter(store, guildID+":"+channel, time.Now, rand.IntN)
	return discord.Run(ctx, token, guildID, app.NewService(store, time.Now, rand.IntN), channel, encounter)
}
