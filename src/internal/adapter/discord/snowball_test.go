package discord

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func TestThrownSnowballInteraction(t *testing.T) {
	for _, step := range []struct {
		name    string
		blocked bool
	}{
		{name: "delivered"},
		{name: "missing channel permissions", blocked: true},
	} {
		t.Run(step.name, func(t *testing.T) {
			store := &giftfixture.Store{}
			service := app.NewService(store, time.Now, func(int) int { return 0 })
			b := bot{guildID: "200", service: service}
			target := app.Member{ServerID: "200", UserID: "500"}
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
					if err != nil || inventory.Snowballs != 1 {
						t.Fatalf("channel post requested before credit: %+v, %v", inventory, err)
					}
					var message discordgo.MessageSend
					if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
						t.Fatal(err)
					}
					if message.Content != "<@500> Alice have threw a ❄️ snowball at you, now you have 1 ❄️ snowball. 🥶" {
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
					want := "You have thrown a ❄️ snowball at Bob. 🥶"
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
						{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "snowball", Options: []*discordgo.ApplicationCommandInteractionDataOption{
							{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Value: "500"},
						}},
					},
					Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
						Users:   map[string]*discordgo.User{"500": {ID: "500"}},
						Members: map[string]*discordgo.Member{"500": {Nick: "Bob"}},
					},
				},
			}})
			inventory, err := service.ViewInventory(t.Context(), target)
			if calls != len(paths) || err != nil || inventory.Snowballs != 1 {
				t.Fatalf("throw requests = %d, recipient inventory = %+v, %v", calls, inventory, err)
			}
		})
	}
}
