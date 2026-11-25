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

func TestGiveGiftInteraction(t *testing.T) {
	store := &giftfixture.Store{}
	b := bot{guildID: "200", service: app.NewService(store, time.Now, func(int) int { return 0 })}
	give := func(id string, isBot bool, amount float64, rarity string) discordgo.ApplicationCommandInteractionData {
		return discordgo.ApplicationCommandInteractionData{
			CommandType: discordgo.ChatApplicationCommand, Name: "give",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Value: id},
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "amount", Value: amount},
					{Type: discordgo.ApplicationCommandOptionString, Name: "rarity", Value: rarity},
				}},
			},
			Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
				Users:   map[string]*discordgo.User{id: {ID: id, Username: "bob_account", Bot: isBot}},
				Members: map[string]*discordgo.Member{id: {Nick: "Bob"}},
			},
		}
	}
	const moderator = discordgo.PermissionManageMessages
	for _, step := range []struct {
		perms   int64
		data    discordgo.ApplicationCommandInteractionData
		content string
	}{
		{0, give("500", false, 3, "rare"), "You need Manage Messages permission to give gifts."},
		{moderator, give("500", false, 3, "rare"), "Gave 3 🎁 rare gifts to Bob."},
		{discordgo.PermissionAdministrator, give("500", false, 1, "epic"), "Gave 1 🎁 epic gift to Bob."},
		{moderator, give("500", false, 0, "rare"), "Enter a positive whole number of items."},
		{moderator, give("500", false, 1, "mythic"), "Choose normal, rare, epic, or coal."},
		{moderator, give("600", true, 1, "rare"), "Choose another player in this server."},
		{moderator, give("mod", false, 1, "rare"), "Choose another player in this server."},
	} {
		session, err := discordgo.New("Bot fixture")
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				var response discordgo.InteractionResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral {
					t.Errorf("expected private acknowledgement, got %+v", response)
				}
				return fakeResponse(http.StatusNoContent, ""), nil
			}
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
				t.Fatal(err)
			}
			if edit.Content == nil || *edit.Content != step.content {
				t.Errorf("reply = %+v, want %q", edit, step.content)
			}
			return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "fixture-token", GuildID: b.guildID, Type: discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{User: &discordgo.User{ID: "mod"}, Permissions: step.perms}, Data: step.data,
		}})
		if calls != 2 {
			t.Errorf("expected two requests, got %d", calls)
		}
	}
	record, err := store.Load(t.Context(), app.Member{ServerID: "200", UserID: "500"})
	if err != nil || record.Inventory.Rare != 3 || record.Inventory.Epic != 1 || record.Aura != 0 {
		t.Fatalf("recipient record = %+v, %v", record, err)
	}
	if mod, _ := store.Load(t.Context(), app.Member{ServerID: "200", UserID: "mod"}); mod != (core.GiftRecord{}) {
		t.Fatalf("moderator record changed: %+v", mod)
	}
	if command := giveCommand(); command.DefaultMemberPermissions == nil || *command.DefaultMemberPermissions != discordgo.PermissionManageMessages {
		t.Errorf("give command permissions = %v", command.DefaultMemberPermissions)
	}
}
