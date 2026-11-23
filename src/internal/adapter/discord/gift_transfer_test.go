package discord

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func checkSendRegistration(t *testing.T, command discordgo.ApplicationCommand) {
	t.Helper()
	if len(command.Options) != 1 || command.Options[0].Name != "gift" ||
		command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand || len(command.Options[0].Options) != 3 {
		t.Fatalf("unexpected send command: %+v", command)
	}
	options := command.Options[0].Options
	for i, want := range []struct {
		name string
		kind discordgo.ApplicationCommandOptionType
	}{
		{"recipient", discordgo.ApplicationCommandOptionUser},
		{"rarity", discordgo.ApplicationCommandOptionString},
		{"quantity", discordgo.ApplicationCommandOptionInteger},
	} {
		if got := options[i]; got.Name != want.name || got.Type != want.kind || !got.Required {
			t.Errorf("unexpected gift option: %+v", got)
		}
	}
	choices := options[1].Choices
	if len(choices) != 3 || choices[0].Value != "normal" || choices[1].Value != "rare" || choices[2].Value != "epic" {
		t.Errorf("unexpected rarity choices: %+v", choices)
	}
	if options[2].MinValue == nil || *options[2].MinValue != 1 {
		t.Errorf("unexpected quantity minimum: %+v", options[2])
	}
}

func TestGiftTransferAndAuraInteractions(t *testing.T) {
	store := &giftfixture.Store{}
	b := bot{guildID: "200", service: app.NewService(store, time.Now, func(n int) int {
		if n == core.GiftRarityRolls {
			return 70
		}
		return 1
	})}
	alice := app.Member{ServerID: b.guildID, UserID: "400"}
	bob := app.Member{ServerID: b.guildID, UserID: "500"}
	if _, err := b.service.ClaimGift(t.Context(), alice); err != nil {
		t.Fatal(err)
	}
	send := func(id string, isBot bool, quantity float64) discordgo.ApplicationCommandInteractionData {
		return discordgo.ApplicationCommandInteractionData{
			CommandType: discordgo.ChatApplicationCommand, Name: "send",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "quantity", Value: quantity},
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Value: id},
					{Type: discordgo.ApplicationCommandOptionString, Name: "rarity", Value: "rare"},
				}},
			},
			Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
				Users:   map[string]*discordgo.User{id: {ID: id, Bot: isBot}},
				Members: map[string]*discordgo.Member{id: {}},
			},
		}
	}
	aura := discordgo.ApplicationCommandInteractionData{
		CommandType: discordgo.ChatApplicationCommand, Name: "aura",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "check"},
		},
	}
	for _, step := range []struct {
		user    string
		data    discordgo.ApplicationCommandInteractionData
		content string
	}{
		{alice.UserID, aura, "Bad deeds 0%  |  Good deeds 0%"},
		{alice.UserID, send("600", true, 1), "Choose another player in this server."},
		{alice.UserID, send(bob.UserID, false, 1.5), "Enter a positive whole number of gifts."},
		{alice.UserID, send(bob.UserID, false, 2), "Sent 2 rare gifts. Good deeds 20%."},
		{alice.UserID, aura, "Bad deeds 0%  |  Good deeds 20%"},
		{bob.UserID, aura, "Bad deeds 0%  |  Good deeds 0%"},
	} {
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
				if r.Method != http.MethodPost || response.Type != discordgo.InteractionResponseDeferredChannelMessageWithSource ||
					response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral {
					t.Errorf("expected private deferred acknowledgement, got %s %+v", r.Method, response)
				}
				return fakeResponse(http.StatusNoContent, ""), nil
			case 2:
				var edit discordgo.WebhookEdit
				if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
					t.Fatal(err)
				}
				if r.Method != http.MethodPatch || edit.Content == nil || *edit.Content != step.content {
					t.Errorf("expected private reply %q, got %s %+v", step.content, r.Method, edit)
				}
				return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
			default:
				t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
				return nil, nil
			}
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "fixture-token", GuildID: b.guildID, Type: discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}}, Data: step.data,
		}})
		if calls != 2 {
			t.Errorf("expected two requests for /%s, got %d", step.data.Name, calls)
		}
	}
	record, err := store.Load(t.Context(), bob)
	if err != nil || record.Inventory.Rare != 2 || record.Aura != (core.Aura{}) {
		t.Fatalf("recipient record = %+v, %v", record, err)
	}
}
