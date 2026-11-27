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

func TestGuideInteraction(t *testing.T) {
	command := discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "guide"}
	pick := discordgo.MessageComponentInteractionData{CustomID: guideTopicID, Values: []string{"grinch"}}
	for _, tt := range []struct {
		name string
		data discordgo.InteractionData
		ack  discordgo.InteractionResponseType
		text []string
	}{
		{"command", command, discordgo.InteractionResponseDeferredChannelMessageWithSource,
			[]string{"## 🎄 Milky Xmas guide", "guide:topic", "Pick a topic"}},
		{"topic", pick, discordgo.InteractionResponseDeferredMessageUpdate,
			[]string{"## 👽 Grinch fight", "guide:topic"}},
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
				Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}},
				Data:   tt.data,
			}})
			if !slices.Equal(calls, []string{"ack", "edit"}) {
				t.Errorf("calls = %v", calls)
			}
			for _, want := range tt.text {
				if !strings.Contains(shown, want) {
					t.Errorf("reply is missing %q:\n%s", want, shown)
				}
			}
		})
	}
	if registered := guideCommand(); registered.Name != "guide" || len(registered.Options) != 0 || registered.DefaultMemberPermissions != nil {
		t.Errorf("unexpected registration: %+v", registered)
	}
}
