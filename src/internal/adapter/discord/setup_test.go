package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestSetupChannels(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "settings.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx, stop := context.WithCancel(t.Context())
	setup := app.NewGuildSetup(store)
	channels := &grinchChannels{ctx: ctx, setup: setup, workers: make(map[grinchChannel]grinchWorker),
		build: func(server, scope string) *app.GrinchEncounter {
			return app.NewGrinchEncounter(store, server, scope, core.DefaultTuning().Grinch, time.Now, func(n int) int { return n - 1 })
		}}
	defer func() {
		stop()
		channels.wait()
	}()
	b := bot{setup: setup, grinches: channels}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}

	session.State.User = &discordgo.User{ID: "100"}
	guild := &discordgo.Guild{
		ID: "200", SystemChannelID: "400",
		Roles:   []*discordgo.Role{{ID: "200", Permissions: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages}},
		Members: []*discordgo.Member{{GuildID: "200", User: session.State.User}},
		Channels: []*discordgo.Channel{
			{ID: "400", GuildID: "200", Type: discordgo.ChannelTypeGuildText},
			{ID: "401", GuildID: "200", Type: discordgo.ChannelTypeGuildText},
			{ID: "402", GuildID: "200", Type: discordgo.ChannelTypeGuildText},
		},
	}
	if err := session.State.GuildAdd(guild); err != nil {
		t.Fatal(err)
	}
	var reply string
	var private bool
	var prompts, introductions, roleCreations int
	var introductionButtons []string
	roleCreated := false
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			var response discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Data != nil {
				reply = response.Data.Content
				private = response.Data.Flags == discordgo.MessageFlagsEphemeral
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		case strings.HasSuffix(r.URL.Path, "/@original"):
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil || edit.Content == nil {
				t.Fatalf("invalid reply edit %+v: %v", edit, err)
			}
			reply = *edit.Content
			return fakeResponse(http.StatusOK, `{"id":"101"}`), nil
		case r.URL.Path == "/api/v9/channels/400/messages" || r.URL.Path == "/api/v9/channels/401/messages":
			var message struct {
				Content    string `json:"content"`
				Components []struct {
					Components []struct {
						Label    string `json:"label"`
						CustomID string `json:"custom_id"`
					} `json:"components"`
				} `json:"components"`
			}
			if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
				t.Fatal(err)
			}
			switch {
			case r.URL.Path == "/api/v9/channels/400/messages" && message.Content == app.GuildSetupPrompt:
				prompts++
			case r.URL.Path == "/api/v9/channels/401/messages" && message.Content == app.ChristmasIntroduction:
				introductions++
				for _, row := range message.Components {
					for _, button := range row.Components {
						introductionButtons = append(introductionButtons, button.Label+"|"+button.CustomID)
					}
				}
			default:
				t.Fatalf("unexpected public message at %s: %q", r.URL.Path, message.Content)
			}
			return fakeResponse(http.StatusOK, `{"id":"102"}`), nil
		case r.Method == http.MethodGet && r.URL.Path == "/api/v9/guilds/200/roles":
			if roleCreated {
				return fakeResponse(http.StatusOK, `[{"id":"900","name":"elfs"}]`), nil
			}
			return fakeResponse(http.StatusOK, `[{"id":"200","name":"@everyone"}]`), nil
		case r.Method == http.MethodPost && r.URL.Path == "/api/v9/guilds/200/roles":
			roleCreations++
			roleCreated = true
			return fakeResponse(http.StatusOK, `{"id":"900","name":"elfs"}`), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
			return nil, nil
		}
	})

	b.guildAvailable(ctx, session, guild)
	b.guildAvailable(ctx, session, guild)
	if prompts != 1 {
		t.Fatalf("onboarding prompts = %d, want 1", prompts)
	}

	invoke := func(channel, name string, permissions int64, options []*discordgo.ApplicationCommandInteractionDataOption, want string, wantPrivate bool) {
		t.Helper()
		reply, private = "", false
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", GuildID: "200", ChannelID: channel, Token: "fixture",
			Type:   discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{Permissions: permissions, User: &discordgo.User{ID: "500"}},
			Data:   discordgo.ApplicationCommandInteractionData{Name: name, CommandType: discordgo.ChatApplicationCommand, Options: options},
		}})
		if !strings.Contains(reply, want) || private != wantPrivate {
			t.Fatalf("/%s in %s = %q (private %v), want %q (private %v)", name, channel, reply, private, want, wantPrivate)
		}
	}

	set := func(channel string, commands, boss bool) []*discordgo.ApplicationCommandInteractionDataOption {
		return []*discordgo.ApplicationCommandInteractionDataOption{
			{Name: "channel", Type: discordgo.ApplicationCommandOptionSubCommand, Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "channel", Type: discordgo.ApplicationCommandOptionChannel, Value: channel},
				{Name: "commands", Type: discordgo.ApplicationCommandOptionBoolean, Value: commands},
				{Name: "boss", Type: discordgo.ApplicationCommandOptionBoolean, Value: boss},
			}},
		}
	}
	mod := int64(discordgo.PermissionManageGuild)
	invoke("400", "hello", 0, nil, "Commands are disabled", true)
	invoke("402", "setup", 0, set("400", true, true), "You need Manage Server", true)
	invoke("402", "setup", mod, set("400", true, false), "commands on · boss off", true)
	invoke("402", "setup", mod, set("401", false, true), "commands off · boss on", true)
	invoke("402", "setup", mod, set("401", false, true), "commands off · boss on", true)
	if roleCreations != 1 || !slices.Equal(introductionButtons, []string{"Join the elf army 🛷|elfs:join"}) {
		t.Fatalf("elfs roles created = %d, introduction buttons = %q", roleCreations, introductionButtons)
	}
	if introductions != 1 || channels.get("200", "400") != nil || channels.get("200", "401") == nil || channels.get("201", "401") != nil {
		t.Fatalf("unexpected worker selection or repeated introduction: %d introductions", introductions)
	}
	channels.remove("200")
	b.guildAvailable(ctx, session, guild)
	if channels.get("200", "401") == nil || prompts != 1 || introductions != 1 {
		t.Fatal("restoring saved boss channels should not repost onboarding or the introduction")
	}
	view := []*discordgo.ApplicationCommandInteractionDataOption{{Name: "view", Type: discordgo.ApplicationCommandOptionSubCommand}}
	invoke("402", "setup", mod, view, "Commands · <#400>\nGrinch · <#401>", true)
	invoke("400", "hello", 0, nil, "Hello, world!", false)
	invoke("401", "hello", 0, nil, "Commands are disabled", true)
	summon := []*discordgo.ApplicationCommandInteractionDataOption{{Name: "summon", Type: discordgo.ApplicationCommandOptionSubCommand}}
	invoke("400", "grinch", discordgo.PermissionManageMessages, summon, "Use this command in a configured Grinch channel.", true)
	invoke("401", "grinch", discordgo.PermissionManageMessages, summon, "Commands are disabled", true)
	invoke("402", "setup", mod, set("401", false, false), "commands off · boss off", true)
	if channels.get("200", "401") != nil {
		t.Fatal("disabled boss channel still has a worker")
	}
	invoke("402", "setup", mod, set("400", false, false), "commands off · boss off", true)
	invoke("400", "hello", 0, nil, "Commands are disabled", true)

	withQuotes := set("402", false, false)
	withQuotes[0].Options = append(withQuotes[0].Options,
		&discordgo.ApplicationCommandInteractionDataOption{Name: "quotes", Type: discordgo.ApplicationCommandOptionBoolean, Value: true})
	invoke("402", "setup", mod, withQuotes, "commands off · boss off · quotes on.", true)
	invoke("402", "setup", mod, view, "Quotes · <#402>", true)
	// Leaving quotes out keeps the choice.
	invoke("402", "setup", mod, set("402", true, false), "commands on · boss off.", true)
	invoke("402", "setup", mod, view, "Quotes · <#402>", true)
}
