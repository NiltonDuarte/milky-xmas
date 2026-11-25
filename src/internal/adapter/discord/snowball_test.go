package discord

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
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
			item, emoji := "snowball", "❄️"
			wantInventory := core.Inventory{Snowballs: 1}
			options := []*discordgo.ApplicationCommandInteractionDataOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Value: target.UserID},
			}
			if step.yellow {
				item, emoji = "yellow snowball", "🟡"
				wantInventory = core.Inventory{YellowSnowballs: 1}
				options = append(options, &discordgo.ApplicationCommandInteractionDataOption{
					Type: discordgo.ApplicationCommandOptionString, Name: "type", Value: "yellow",
				})
			} else {
				options = append(options, &discordgo.ApplicationCommandInteractionDataOption{
					Type: discordgo.ApplicationCommandOptionString, Name: "type", Value: "normal",
				})
			}
			// The sender owns exactly the snowball that is thrown.
			if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: sender, Next: core.GiftRecord{Inventory: wantInventory}}); err != nil || !saved {
				t.Fatalf("seed snowball = %v, %v", saved, err)
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
					message, files := decodeMultipartSend(t, r)
					if message.Content != "<@500>❕  Alice has thrown a "+emoji+" "+item+" at you. 🥶" {
						t.Errorf("recipient notification = %q", message.Content)
					}
					if len(message.Embeds) != 1 || message.Embeds[0].Image == nil || message.Embeds[0].Image.URL != "attachment://snowball.jpg" || files != 1 {
						t.Errorf("expected the snowball image attached, got %+v with %d files", message.Embeds, files)
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
					want := "You have thrown a " + emoji + " " + item + " at Bob. 🥶"
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
					CommandType: discordgo.ChatApplicationCommand, Name: "throw",
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
			wantBad := 0
			if step.yellow {
				wantBad = 5
			}
			if aura, err := service.CheckAura(t.Context(), sender); err != nil || aura.Bad != wantBad {
				t.Fatalf("sender aura = %+v, %v", aura, err)
			}
			if inventory, err := service.ViewInventory(t.Context(), sender); err != nil || inventory != (core.Inventory{}) {
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
	for user, inventory := range map[string]core.Inventory{
		"400": {Snowballs: 1, YellowSnowballs: 1}, "500": {Snowballs: 1},
	} {
		member := app.Member{ServerID: "200", UserID: user}
		if saved, err := store.Save(ctx, app.GiftUpdate{Member: member, Next: core.GiftRecord{Inventory: inventory}}); err != nil || !saved {
			t.Fatalf("seed snowballs for %s = %v, %v", user, saved, err)
		}
	}
	b := bot{guildID: "200", service: app.NewService(store, time.Now, func(int) int { return 0 }), setup: setup}
	for _, step := range []struct {
		user, channel, query string
		wants, labels        string
	}{
		{"400", "700", "", "normal,yellow", "❄️ Normal Snowball|🟡 Yellow Snowball"},
		{"400", "700", "yel", "yellow", "🟡 Yellow Snowball"},
		{"500", "700", "", "normal", "❄️ Normal Snowball"},
		{"500", "700", "yel", "", ""},
		{"600", "700", "", "", ""},
		{"400", "701", "", "", ""},
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
			var values, labels []string
			for _, choice := range payload.Data.Choices {
				values = append(values, choice.Value.(string))
				labels = append(labels, choice.Name)
			}
			if got := strings.Join(values, ","); got != step.wants {
				t.Errorf("choices for %+v = %q", step, got)
			}
			if got := strings.Join(labels, "|"); got != step.labels {
				t.Errorf("choice labels for %+v = %q", step, got)
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "fixture-token", GuildID: "200", ChannelID: step.channel,
			Type:   discordgo.InteractionApplicationCommandAutocomplete,
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}},
			Data: discordgo.ApplicationCommandInteractionData{
				CommandType: discordgo.ChatApplicationCommand, Name: "throw",
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

func decodeMultipartSend(t *testing.T, r *http.Request) (discordgo.MessageSend, int) {
	t.Helper()
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(r.Body, params["boundary"])
	var message discordgo.MessageSend
	files := 0
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return message, files
		}
		if err != nil {
			t.Fatal(err)
		}
		if part.FormName() == "payload_json" {
			if err := json.NewDecoder(part).Decode(&message); err != nil {
				t.Fatal(err)
			}
		} else if part.FileName() == "snowball.jpg" {
			if data, _ := io.ReadAll(part); len(data) > 0 {
				files++
			}
		}
	}
}
