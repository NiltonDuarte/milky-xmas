package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestModGuideInteraction(t *testing.T) {
	command := discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "mod-guide"}
	pick := discordgo.MessageComponentInteractionData{CustomID: modGuideTopicID, Values: []string{"fight"}}
	deferred, update := discordgo.InteractionResponseDeferredChannelMessageWithSource, discordgo.InteractionResponseDeferredMessageUpdate
	for _, tt := range []struct {
		name  string
		data  discordgo.InteractionData
		perms int64
		ack   discordgo.InteractionResponseType
		calls []string
		text  string
		not   string
	}{
		{"moderator", command, discordgo.PermissionManageMessages, deferred, []string{"ack", "edit"}, "## 🛠️ Moderator guide", "permission to read"},
		{"manager", command, discordgo.PermissionManageGuild, deferred, []string{"ack", "edit"}, "## 🛠️ Moderator guide", "permission to read"},
		{"player", command, 0, deferred, []string{"ack", "edit"}, "permission to read", "## 🛠️"},
		{"topic", pick, discordgo.PermissionAdministrator, update, []string{"ack", "edit"}, "## ⚔️ Grinch numbers", ""},
		{"player picking a topic", pick, 0, update, nil, "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			session, err := discordgo.New("Bot fixture")
			if err != nil {
				t.Fatal(err)
			}
			var calls []string
			var shown string
			session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(r.URL.Path, "/callback") {
					calls = append(calls, "ack")
					var response discordgo.InteractionResponse
					if err := json.Unmarshal(body, &response); err != nil || response.Type != tt.ack {
						t.Errorf("acknowledgement = %s, %v", body, err)
					}
					return fakeResponse(http.StatusNoContent, ""), nil
				}
				calls, shown = append(calls, "edit"), string(body)
				return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
			})
			kind := discordgo.InteractionApplicationCommand
			if _, ok := tt.data.(discordgo.MessageComponentInteractionData); ok {
				kind = discordgo.InteractionMessageComponent
			}
			b := bot{guildID: "200", service: app.Service{}.WithTuning(core.DefaultTuning())}
			b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "300", AppID: "100", Token: "fixture-token", GuildID: "200", Type: kind,
				Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}, Permissions: tt.perms},
				Data:   tt.data,
			}})
			if !slices.Equal(calls, tt.calls) {
				t.Errorf("calls = %v, want %v", calls, tt.calls)
			}
			if !strings.Contains(shown, tt.text) || (tt.not != "" && strings.Contains(shown, tt.not)) {
				t.Errorf("unexpected reply:\n%s", shown)
			}
		})
	}
	registered := modGuideCommand()
	if registered.Name != "mod-guide" || registered.DefaultMemberPermissions == nil || *registered.DefaultMemberPermissions != discordgo.PermissionManageMessages {
		t.Errorf("unexpected registration: %+v", registered)
	}
}
