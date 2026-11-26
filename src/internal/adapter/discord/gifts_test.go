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

func TestDailyClaimInteraction(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	roll := 0
	b := bot{guildID: "200", service: app.NewService(store, func() time.Time { return now }, func(int) int { return roll })}
	for _, step := range []struct {
		user, reply string
		roll        int
	}{
		{"alice", "## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected 1 gifts 🎁.\nGift rarity · normal ·\n```", 0},
		{"alice", "🎅 You already got the gift for today, the next gift 🎁 is in 24 hours.", 5},
		{"bob", "## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected 10 gifts 🎁.\nGift rarity · normal ·\n```", 9},
	} {
		roll = step.roll
		session, err := discordgo.New("Bot fixture")
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			switch calls {
			case 1:
				var response discordgo.InteractionResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if r.Method != http.MethodPost || r.URL.Path != "/api/v9/interactions/300/fixture-token/callback" ||
					response.Type != discordgo.InteractionResponseDeferredChannelMessageWithSource ||
					response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral {
					t.Errorf("expected private deferred acknowledgement, got %s %s %+v", r.Method, r.URL.Path, response)
				}
				return fakeResponse(http.StatusNoContent, ""), nil
			case 2:
				var edit discordgo.WebhookEdit
				if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
					t.Fatal(err)
				}
				if r.Method != http.MethodPatch || r.URL.Path != "/api/v9/webhooks/100/fixture-token/messages/@original" ||
					edit.Content == nil || *edit.Content != step.reply {
					t.Errorf("expected original reply edit %q, got %s %s %+v", step.reply, r.Method, r.URL.Path, edit)
				}
				return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
			default:
				t.Fatalf("unexpected interaction request: %s %s", r.Method, r.URL.Path)
				return nil, nil
			}
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "fixture-token", GuildID: "200", Type: discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}},
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "daily", CommandType: discordgo.ChatApplicationCommand,
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "claim"},
				},
			},
		}})
		if calls != 2 {
			t.Errorf("expected acknowledgement and reply edit, got %d requests", calls)
		}
	}
	record, err := store.Load(t.Context(), app.Member{ServerID: "200", UserID: "alice"})
	if err != nil || record.Inventory.Normal != 1 {
		t.Fatalf("stored record = %+v, %v", record, err)
	}
}
