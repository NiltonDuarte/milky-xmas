package discord

import (
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

func TestThrownSnowballInteraction(t *testing.T) {
	for _, step := range []struct {
		name    string
		blocked bool
		yellow  bool
	}{
		{name: "delivered"},
		{name: "yellow delivered", yellow: true},
		{name: "missing channel permissions", blocked: true},
	} {
		t.Run(step.name, func(t *testing.T) {
			store := &giftfixture.Store{}
			service := app.NewService(store, time.Now, func(int) int { return 0 })
			b := bot{guildID: "200", service: service}
			target := app.Member{ServerID: "200", UserID: "500"}
			sender := app.Member{ServerID: "200", UserID: "400"}
			item := "snowball"
			wantInventory := core.Inventory{Snowballs: 1}
			options := []*discordgo.ApplicationCommandInteractionDataOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Value: target.UserID},
			}
			if step.yellow {
				item = "yellow snowball"
				wantInventory = core.Inventory{YellowSnowballs: 1}
				options = append(options, &discordgo.ApplicationCommandInteractionDataOption{
					Type: discordgo.ApplicationCommandOptionString, Name: "type", Value: "yellow",
				})
				saved, err := store.Save(t.Context(), app.GiftUpdate{Member: sender, Next: core.GiftRecord{Inventory: wantInventory}})
				if err != nil || !saved {
					t.Fatalf("seed yellow snowball = %v, %v", saved, err)
				}
			}
			session, err := discordgo.New("Bot fixture")
			if err != nil {
				t.Fatal(err)
			}
			paths := []string{
				"/api/v9/interactions/300/fixture-token/callback",
				"/api/v9/channels/700/messages",
				"/api/v9/webhooks/100/fixture-token/messages/@original",
			}
			calls := 0
			session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				if calls >= len(paths) || r.URL.Path != paths[calls] {
					t.Fatalf("unexpected request %s %s at call %d", r.Method, r.URL.Path, calls)
				}
				calls++
				switch calls {
				case 1:
					var response discordgo.InteractionResponse
					if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
						t.Fatal(err)
					}
					if response.Type != discordgo.InteractionResponseDeferredChannelMessageWithSource || response.Data == nil ||
						response.Data.Flags != discordgo.MessageFlagsEphemeral {
						t.Errorf("expected private acknowledgement, got %+v", response)
					}
					return fakeResponse(http.StatusNoContent, ""), nil
				case 2:
					inventory, err := service.ViewInventory(t.Context(), target)
					if err != nil || inventory != wantInventory {
						t.Fatalf("channel post requested before credit: %+v, %v", inventory, err)
					}
					var message discordgo.MessageSend
					if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
						t.Fatal(err)
					}
					if message.Content != "<@500> Alice have threw a ❄️ "+item+" at you, now you have 1 ❄️ "+item+". 🥶" {
						t.Errorf("recipient notification = %q", message.Content)
					}
					if mentions := message.AllowedMentions; mentions == nil || len(mentions.Parse) != 0 || len(mentions.Users) != 1 || mentions.Users[0] != target.UserID {
						t.Errorf("unexpected allowed mentions: %+v", mentions)
					}
					if step.blocked {
						return fakeResponse(http.StatusForbidden, `{"message":"Missing Permissions","code":50013}`), nil
					}
					return fakeResponse(http.StatusOK, `{"id":"800"}`), nil
				default:
					var edit discordgo.WebhookEdit
					if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
						t.Fatal(err)
					}
					want := "You have thrown a ❄️ " + item + " at Bob. 🥶"
					if step.blocked {
						want += "\nThe snowball was added to their inventory, but I couldn't post the notification in this channel."
					}
					if edit.Content == nil || *edit.Content != want {
						t.Errorf("expected reply %q, got %+v", want, edit)
					}
					return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
				}
			})
			b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "300", AppID: "100", Token: "fixture-token", GuildID: "200", ChannelID: "700", Type: discordgo.InteractionApplicationCommand,
				Member: &discordgo.Member{Nick: "Alice", User: &discordgo.User{ID: "400"}},
				Data: discordgo.ApplicationCommandInteractionData{
					CommandType: discordgo.ChatApplicationCommand, Name: "thrown",
					Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "snowball", Options: options},
					},
					Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
						Users:   map[string]*discordgo.User{"500": {ID: "500"}},
						Members: map[string]*discordgo.Member{"500": {Nick: "Bob"}},
					},
				},
			}})
			inventory, err := service.ViewInventory(t.Context(), target)
			if calls != len(paths) || err != nil || inventory != wantInventory {
				t.Fatalf("throw requests = %d, recipient inventory = %+v, %v", calls, inventory, err)
			}
			if inventory, err := service.ViewInventory(t.Context(), sender); err != nil || inventory.YellowSnowballs != 0 {
				t.Fatalf("sender inventory = %+v, %v", inventory, err)
			}
		})
	}
}

func TestSnowballAutocomplete(t *testing.T) {
	ctx := t.Context()
	store, err := giftstore.OpenSQLite(ctx, filepath.Join(t.TempDir(), "items.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	setup := app.NewGuildSetup(store)
	if saved, err := setup.Configure(ctx, "200", "700", true, true, false, nil); err != nil || !saved {
		t.Fatalf("configure channel = %v, %v", saved, err)
	}
	member := app.Member{ServerID: "200", UserID: "400"}
	if saved, err := store.Save(ctx, app.GiftUpdate{Member: member, Next: core.GiftRecord{Inventory: core.Inventory{YellowSnowballs: 1}}}); err != nil || !saved {
		t.Fatalf("seed yellow snowball = %v, %v", saved, err)
	}
	b := bot{guildID: "200", service: app.NewService(store, time.Now, func(int) int { return 0 }), setup: setup}
	for _, step := range []struct {
		user, channel, query string
		wants                string
	}{
		{"400", "700", "", "normal,yellow"},
		{"400", "700", "yel", "yellow"},
		{"500", "700", "", "normal"},
		{"500", "700", "yel", ""},
		{"400", "701", "", ""},
	} {
		session, err := discordgo.New("Bot fixture")
		if err != nil {
			t.Fatal(err)
		}
		called := false
		session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			called = true
			var payload discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.Type != discordgo.InteractionApplicationCommandAutocompleteResult || payload.Data == nil || payload.Data.Choices == nil {
				t.Fatalf("invalid autocomplete response: %+v", payload)
			}
			var values []string
			for _, choice := range payload.Data.Choices {
				values = append(values, choice.Value.(string))
			}
			if got := strings.Join(values, ","); got != step.wants {
				t.Errorf("choices for %+v = %q", step, got)
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "fixture-token", GuildID: "200", ChannelID: step.channel,
			Type:   discordgo.InteractionApplicationCommandAutocomplete,
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}},
			Data: discordgo.ApplicationCommandInteractionData{
				CommandType: discordgo.ChatApplicationCommand, Name: "thrown",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "snowball", Options: []*discordgo.ApplicationCommandInteractionDataOption{
						{Type: discordgo.ApplicationCommandOptionString, Name: "type", Value: step.query, Focused: true},
					}},
				},
			},
		}})
		if !called {
			t.Fatal("autocomplete interaction was not acknowledged")
		}
	}
}
