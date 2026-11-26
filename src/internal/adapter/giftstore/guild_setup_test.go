package giftstore_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func TestGuildSettingsAndAnnouncements(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.sqlite")
	ctx := t.Context()
	store := openGiftPlayers(t, path)
	setup := app.NewGuildSetup(store)
	setup.Introduce = true // Off by default for now.
	var prompts, intros int
	offline := true
	failed := errors.New("missing channel permission")
	post := func(_ context.Context, channel, text string) error {
		switch text {
		case app.GuildSetupPrompt:
			if channel != "general" {
				t.Errorf("onboarding channel = %q", channel)
			}
			prompts++
		case app.ChristmasIntroduction:
			if channel != "event" {
				t.Errorf("introduction channel = %q", channel)
			}
			settings, err := setup.Settings(ctx, "guild")
			if err != nil || !settings.Channels[channel].Boss {
				t.Fatalf("boss setting not saved before introduction: %+v, %v", settings, err)
			}
			intros++
			if offline {
				return failed
			}
		default:
			t.Errorf("unexpected message: %q", text)
		}
		return nil
	}
	if err := setup.Welcome(ctx, "guild", "general", post); err != nil {
		t.Fatal(err)
	}
	if saved, err := setup.Configure(ctx, "guild", "event", false, app.ChannelChoice{Commands: true, Boss: true}, post); saved || !errors.Is(err, app.ErrManageGuild) || intros != 0 {
		t.Fatalf("unauthorized configure = %v, %v; intros %d", saved, err, intros)
	}
	if saved, err := setup.Configure(ctx, "guild", "general", true, app.ChannelChoice{Commands: true, Boss: false}, post); !saved || err != nil {
		t.Fatalf("configure general = %v, %v", saved, err)
	}
	if saved, err := setup.Configure(ctx, "guild", "event", true, app.ChannelChoice{Commands: false, Boss: true}, post); !saved || !errors.Is(err, failed) {
		t.Fatalf("configure event with failed post = %v, %v", saved, err)
	}
	offline = false
	if saved, err := setup.Configure(ctx, "guild", "event", true, app.ChannelChoice{Commands: false, Boss: true}, post); !saved || err != nil {
		t.Fatalf("retry introduction = %v, %v", saved, err)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	setup = app.NewGuildSetup(openGiftPlayers(t, path))
	if err := setup.Welcome(ctx, "guild", "general", post); err != nil {
		t.Fatal(err)
	}
	if saved, err := setup.Configure(ctx, "guild", "event", true, app.ChannelChoice{Commands: true, Boss: true}, post); !saved || err != nil {
		t.Fatalf("reconfigure event = %v, %v", saved, err)
	}
	settings, err := setup.Settings(ctx, "guild")
	if err != nil || !settings.Welcomed || settings.Channels["general"] != (app.ChannelSettings{Commands: true}) ||
		settings.Channels["event"] != (app.ChannelSettings{Commands: true, Boss: true, Introduced: true}) || prompts != 1 || intros != 2 {
		t.Fatalf("settings = %+v, %v; prompts %d, intros %d", settings, err, prompts, intros)
	}
	other, err := setup.Settings(ctx, "another")
	if err != nil || len(other.Channels) != 0 || other.Welcomed {
		t.Fatalf("other guild settings = %+v, %v", other, err)
	}
}

func TestChristmasIntroductionIsOffByDefault(t *testing.T) {
	store := openGiftPlayers(t, filepath.Join(t.TempDir(), "settings.sqlite"))
	post := func(_ context.Context, _, content string) error {
		t.Errorf("posted %q", content)
		return nil
	}
	setup := app.NewGuildSetup(store)
	if saved, err := setup.Configure(t.Context(), "guild", "event", true, app.ChannelChoice{Boss: true}, post); !saved || err != nil {
		t.Fatalf("configure = %v, %v", saved, err)
	}
	settings, err := setup.Settings(t.Context(), "guild")
	if err != nil || settings.Channels["event"] != (app.ChannelSettings{Boss: true}) {
		t.Fatalf("settings = %+v, %v", settings, err)
	}
}
