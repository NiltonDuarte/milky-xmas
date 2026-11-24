package discord

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func TestWaterInteraction(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	b := bot{guildID: "200", tree: app.NewTree(&app.MemoryTreeStore{}, func() time.Time { return now })}
	for _, step := range []struct {
		user    string
		public  bool
		content string
	}{
		{"alice", true, "💧 <@alice> watered the tree! +2%\n## 🎄 The Christmas tree\n```\n💧 Water level: 12%  [█░░░░░░░░░]\n```\n🥀 The tree is withered."},
		{"alice", false, "💧 You watered the tree a moment ago. Try again in 10 minutes."},
		{"bob", true, "💧 <@bob> watered the tree! +2%"},
	} {
		session, err := discordgo.New("Bot fixture")
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			var response struct {
				Type discordgo.InteractionResponseType
				Data struct {
					Content string
					Flags   discordgo.MessageFlags
					Embeds  []struct {
						Description string
						Image       struct{ URL string }
					}
				}
			}
			if !strings.Contains(r.Header.Get("Content-Type"), "multipart") {
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
			} else {
				decodeMultipartPayload(t, r, &response, "deadtree.jpg")
			}
			if response.Type != discordgo.InteractionResponseChannelMessageWithSource {
				t.Errorf("response type = %v", response.Type)
			}
			if step.public {
				got := ""
				if len(response.Data.Embeds) == 2 {
					got = response.Data.Embeds[1].Description
				}
				if response.Data.Flags != 0 || len(response.Data.Embeds) != 2 || response.Data.Embeds[0].Image.URL != "attachment://deadtree.jpg" ||
					!strings.HasPrefix(got, step.content) {
					t.Errorf("public watering = %+v, want %q", response.Data, step.content)
				}
			} else if response.Data.Flags != discordgo.MessageFlagsEphemeral || response.Data.Content != step.content {
				t.Errorf("private reply = %+v, want %q", response.Data, step.content)
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "fixture", GuildID: "200", Type: discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}},
			Data:   discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "water"},
		}})
		if calls != 1 {
			t.Errorf("expected one response for %s, got %d", step.user, calls)
		}
	}
	if command := waterCommand(); command.Name != "water" || len(command.Options) != 0 {
		t.Errorf("unexpected /water registration: %+v", command)
	}
}
