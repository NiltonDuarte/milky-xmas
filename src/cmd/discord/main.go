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
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/logfile"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/tuningfile"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func main() {
	token := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN is required")
	}

	guildID := strings.TrimSpace(os.Getenv("DISCORD_TEST_GUILD_ID"))
	if guildID != "" {
		if id, err := strconv.ParseUint(guildID, 10, 64); err != nil || id == 0 {
			log.Fatal("DISCORD_TEST_GUILD_ID must be a valid Discord server ID")
		}
	}

	path := strings.TrimSpace(os.Getenv("MILKY_GIFT_DB"))
	if path == "" {
		log.Fatal("MILKY_GIFT_DB is required")
	}
	closeLog, err := logfile.Setup(strings.TrimSpace(os.Getenv(logfile.EnvVar)))
	if err != nil {
		log.Fatal(err)
	}
	defer closeLog()
	tuning, err := tuningfile.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, token, guildID, path, tuning); err != nil {
		log.Print(err)
		closeLog()
		os.Exit(1)
	}
}

func run(ctx context.Context, token, guildID, path string, tuning core.Tuning) error {
	store, err := giftstore.OpenSQLite(ctx, path)
	if err != nil {
		return fmt.Errorf("open gift database: %w", err)
	}
	defer store.Close()

	service := app.NewService(store, time.Now, rand.IntN).WithTuning(tuning)
	return discord.Run(ctx, token, guildID, service, app.NewGuildSetup(store),
		func(server, scope string) *app.GrinchEncounter {
			return app.NewGrinchEncounter(store, server, scope, tuning.Grinch, time.Now, rand.IntN)
		}, service.RewardTree(app.NewTree(store, time.Now).WithRules(tuning.Tree)))
}
