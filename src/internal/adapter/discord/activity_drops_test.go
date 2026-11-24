package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestActivityDropsOnlyObserveConfiguredChannelsAndRenderEightButtons(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "drops.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	setup := app.NewGuildSetup(store)
	if saved, err := setup.Configure(t.Context(), "200", "400", true, true, false, func(context.Context, string, string) error { return nil }); !saved || err != nil {
		t.Fatalf("configure = %v, %v", saved, err)
	}
	service := app.NewService(store, time.Now, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			return 0
		}
	})
	b := bot{guildID: "200", setup: setup, drops: newDropBot(app.NewActivityDrops(service))}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	posts := 0
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v9/channels/400/messages" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var message struct {
			Content    string `json:"content"`
			Components []struct {
				Components []json.RawMessage `json:"components"`
			} `json:"components"`
		}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Fatal(err)
		}
		posts++
		if message.Content != giftRainMessage || len(message.Components) != 2 ||
			len(message.Components[0].Components) != 4 || len(message.Components[1].Components) != 4 {
			t.Fatalf("drop message = %+v", message)
		}
		for row, components := range message.Components {
			for column, raw := range components.Components {
				var button discordgo.Button
				if err := json.Unmarshal(raw, &button); err != nil || button.Label != "🎁" ||
					button.CustomID != "drop:1:"+string(rune('0'+row*4+column)) {
					t.Fatalf("button %d,%d = %+v, %v", row, column, button, err)
				}
			}
		}
		return fakeResponse(http.StatusOK, `{"id":"500"}`), nil
	})
	message := func(channel string) *discordgo.MessageCreate {
		return &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "200", ChannelID: channel, Author: &discordgo.User{ID: "300"}}}
	}
	for range 6 {
		b.observeDrop(session, message("401"))
	}
	if posts != 0 {
		t.Fatalf("posts in disabled channel = %d", posts)
	}
	for range core.GiftRainMinimumMessages {
		b.observeDrop(session, message("400"))
	}
	if posts != 1 {
		t.Fatalf("posts in enabled channel = %d", posts)
	}
}

func TestActivityDropButtonRejectsRepeatedAndHiddenRainCooldown(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	service := app.NewService(&giftfixture.Store{}, func() time.Time { return now }, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			return 0
		}
	})
	drops := app.NewActivityDrops(service)
	var rain *app.GiftDrop
	for range core.GiftRainMinimumMessages {
		rain = drops.Observe("200", "400")
	}
	if rain == nil || rain.Kind != core.GiftRainDrop {
		t.Fatalf("rain = %+v", rain)
	}
	b := bot{guildID: "200", drops: newDropBot(drops)}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var replies []string
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			return fakeResponse(http.StatusNoContent, ""), nil
		case strings.HasSuffix(r.URL.Path, "/messages/@original"):
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil || edit.Content == nil {
				t.Fatalf("reply edit = %+v, %v", edit, err)
			}
			replies = append(replies, *edit.Content)
			return fakeResponse(http.StatusOK, `{"id":"501"}`), nil
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
			return nil, nil
		}
	})
	click := func() {
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "token", GuildID: "200", ChannelID: "400", Type: discordgo.InteractionMessageComponent,
			Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: "drop:" + rain.ID + ":0"},
		}})
	}
	click()
	click()
	if len(replies) != 2 || replies[0] != "You collected a normal gift." || replies[1] != "You cannot collect another gift yet." || strings.Contains(replies[1], "minute") {
		t.Fatalf("replies = %q", replies)
	}
}
