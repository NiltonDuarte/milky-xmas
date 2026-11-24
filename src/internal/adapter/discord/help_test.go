package discord

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func TestHelpInteraction(t *testing.T) {
	for _, tt := range []struct {
		name    string
		perms   int64
		listed  []string
		missing []string
	}{
		{"member", 0, []string{"/daily claim", "/helpmemilky"}, []string{"/give gift", "/grinch summon", "/setup"}},
		{"moderator", discordgo.PermissionManageMessages, []string{"/give gift", "/grinch summon"}, []string{"/setup"}},
		{"server manager", discordgo.PermissionManageGuild, []string{"/setup view"}, []string{"/give gift"}},
		{"administrator", discordgo.PermissionAdministrator, []string{"/give gift", "/setup channel"}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			session, err := discordgo.New("Bot fixture")
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return fakeResponse(http.StatusNoContent, ""), nil
				}
				var edit discordgo.WebhookEdit
				if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
					t.Fatal(err)
				}
				if edit.Content == nil {
					t.Fatal("no help text")
				}
				for _, want := range tt.listed {
					if !strings.Contains(*edit.Content, want) {
						t.Errorf("help is missing %q:\n%s", want, *edit.Content)
					}
				}
				for _, absent := range tt.missing {
					if strings.Contains(*edit.Content, absent) {
						t.Errorf("help should not list %q", absent)
					}
				}
				return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
			})
			b := bot{guildID: "200", service: app.Service{}}
			b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "300", AppID: "100", Token: "fixture-token", GuildID: "200", Type: discordgo.InteractionApplicationCommand,
				Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}, Permissions: tt.perms},
				Data:   discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "helpmemilky"},
			}})
			if calls != 2 {
				t.Errorf("expected two requests, got %d", calls)
			}
		})
	}
	if command := helpCommand(); command.Name != "helpmemilky" || len(command.Options) != 0 || command.DefaultMemberPermissions != nil {
		t.Errorf("unexpected help registration: %+v", command)
	}
}
