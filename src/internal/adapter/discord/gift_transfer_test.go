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
		{"item", discordgo.ApplicationCommandOptionString},
		{"quantity", discordgo.ApplicationCommandOptionInteger},
	} {
		if got := options[i]; got.Name != want.name || got.Type != want.kind || !got.Required {
			t.Errorf("unexpected gift option: %+v", got)
		}
	}
	choices := options[1].Choices
	if len(choices) != 4 || choices[0].Value != "normal" || choices[1].Value != "rare" || choices[2].Value != "epic" || choices[3].Value != "coal" {
		t.Errorf("unexpected rarity choices: %+v", choices)
	}
	if options[2].MinValue == nil || *options[2].MinValue != 1 {
		t.Errorf("unexpected quantity minimum: %+v", options[2])
	}
}

func TestGiftTransferAuraAndSocksInteractions(t *testing.T) {
	store := &giftfixture.Store{}
	b := bot{guildID: "200", service: app.NewService(store, time.Now, func(n int) int {
		if n == core.GiftRarityRolls {
			return 70
		}
		return 1
	})}
	alice := app.Member{ServerID: b.guildID, UserID: "400"}
	bob := app.Member{ServerID: b.guildID, UserID: "500"}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Next: core.GiftRecord{Inventory: core.Inventory{Coal: 2}}}); err != nil || !saved {
		t.Fatalf("seed coal = %v, %v", saved, err)
	}
	if _, err := b.service.ClaimGift(t.Context(), alice); err != nil {
		t.Fatal(err)
	}
	send := func(id string, isBot bool, quantity float64, item string) discordgo.ApplicationCommandInteractionData {
		return discordgo.ApplicationCommandInteractionData{
			CommandType: discordgo.ChatApplicationCommand, Name: "send",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "quantity", Value: quantity},
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Value: id},
					{Type: discordgo.ApplicationCommandOptionString, Name: "item", Value: item},
				}},
			},
			Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
				Users:   map[string]*discordgo.User{id: {ID: id, Username: "bob_account", Bot: isBot}},
				Members: map[string]*discordgo.Member{id: {Nick: "Bob"}},
			},
		}
	}
	aura := discordgo.ApplicationCommandInteractionData{
		CommandType: discordgo.ChatApplicationCommand, Name: "aura",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "check"},
		},
	}
	socks := discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "socks"}
	for _, step := range []struct {
		user    string
		data    discordgo.ApplicationCommandInteractionData
		content string
	}{
		{alice.UserID, aura, "😈 Bad 0%  |  😇 Good 0%"},
		{alice.UserID, socks, "🧦 Your inventory\nNormal gifts: 0\nRare gifts: 2\nEpic gifts: 0\nSnowballs: 0\nYellow Snowballs: 0\nCoal: 2"},
		{alice.UserID, send("600", true, 1, "rare"), "Choose another player in this server."},
		{alice.UserID, send(bob.UserID, false, 1.5, "rare"), "Enter a positive whole number of items."},
		{alice.UserID, send(bob.UserID, false, 1, "rare"), "Sent 1 🎁rare gift to Bob. Your 😇 good deeds is 10% more high."},
		{alice.UserID, send(bob.UserID, false, 1, "rare"), "Sent 1 🎁rare gift to Bob. Your 😇 good deeds is 10% more high."},
		{alice.UserID, aura, "😈 Bad 0%  |  😇 Good 20%"},
		{bob.UserID, aura, "😈 Bad 0%  |  😇 Good 0%"},
		{alice.UserID, socks, "🧦 Your inventory\nNormal gifts: 0\nRare gifts: 0\nEpic gifts: 0\nSnowballs: 0\nYellow Snowballs: 0\nCoal: 2"},
		{bob.UserID, socks, "🧦 Your inventory\nNormal gifts: 0\nRare gifts: 2\nEpic gifts: 0\nSnowballs: 0\nYellow Snowballs: 0\nCoal: 0"},
		{alice.UserID, send(bob.UserID, false, 2, "coal"), "Sent 2 🪨 Coal to Bob. Your 😈 bad deeds increased by 10%."},
		{alice.UserID, aura, "😈 Bad 10%  |  😇 Good 20%"},
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
	if err != nil || record.Inventory.Rare != 2 || record.Inventory.Coal != 2 || record.Aura != (core.Aura{}) {
		t.Fatalf("recipient record = %+v, %v", record, err)
	}
}
