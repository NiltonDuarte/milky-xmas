package discord

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

type statsStore struct {
	*giftfixture.Store
	calls *[]string
	err   error
}

func (store statsStore) PlayerHistory(context.Context, string) ([]app.HistoryEntry, error) {
	*store.calls = append(*store.calls, "store")
	entry := app.HistoryEntry{UserID: "alice", Reason: app.ReasonSendCoal}
	entry.Delta.Coal = -3
	return []app.HistoryEntry{entry}, store.err
}

func TestStatsMenus(t *testing.T) {
	command := discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "stats"}
	choose := func(id, value string) discordgo.MessageComponentInteractionData {
		return discordgo.MessageComponentInteractionData{CustomID: id, Values: []string{value}}
	}
	for _, tt := range []struct {
		name  string
		data  discordgo.InteractionData
		err   error
		ack   discordgo.InteractionResponseType
		calls []string
		text  []string
		not   string
	}{
		{"command", command, nil, discordgo.InteractionResponseDeferredChannelMessageWithSource, []string{"ack", "store", "edit"},
			[]string{"🌑 Coal sent: **3**", "stats:view", "stats:board", "Pick a leaderboard", "## 🌌 Milky Way Galaxy", `"accent_color":8084694`}, "<@alice>"},
		{"server", choose(statsViewID, "server"), nil, discordgo.InteractionResponseDeferredMessageUpdate, []string{"ack", "store", "edit"},
			[]string{"🌑 Coal sent: **3**", "stats:view", "stats:board"}, "<@alice>"},
		{"leaderboards", choose(statsViewID, "boards"), nil, discordgo.InteractionResponseDeferredMessageUpdate, []string{"ack", "store", "edit"},
			[]string{"🌑 Coal sent: **3**", "stats:board"}, "<@alice>"},
		{"server in the second menu", choose(statsBoardID, "server"), nil, discordgo.InteractionResponseDeferredMessageUpdate, []string{"ack", "store", "edit"},
			[]string{"🌑 Coal sent: **3**", "stats:board"}, "<@alice>"},
		{"bad eggs", choose(statsBoardID, "bad"), nil, discordgo.InteractionResponseDeferredMessageUpdate, []string{"ack", "store", "edit"},
			[]string{"Bad eggs", "🥇 <@alice> — **15**", "stats:view", "stats:board", `"accent_color":15158332`}, "Nobody yet"},
		{"failed read", choose(statsBoardID, "bad"), errors.New("disk gone"), discordgo.InteractionResponseDeferredMessageUpdate,
			[]string{"ack", "store", "edit"}, []string{"Could not read the statistics. Try again.", "stats:board"}, "<@alice>"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var shown string
			service := app.NewService(statsStore{&giftfixture.Store{}, &calls, tt.err}, time.Now, func(int) int { return 0 })
			session, err := discordgo.New("Bot fixture")
			if err != nil {
				t.Fatal(err)
			}
			session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				// Undo the JSON escaping of the mentions.
				text := strings.NewReplacer(`\u003c`, "<", `\u003e`, ">").Replace(string(body))
				if strings.HasSuffix(r.URL.Path, "/callback") {
					calls = append(calls, "ack")
					var response struct {
						Type discordgo.InteractionResponseType
					}
					if err := json.Unmarshal(body, &response); err != nil || response.Type != tt.ack {
						t.Errorf("acknowledgement = %s, %v", body, err)
					}
					return fakeResponse(http.StatusNoContent, ""), nil
				}
				calls = append(calls, "edit")
				shown = text
				return fakeResponse(http.StatusOK, `{"id":"1"}`), nil
			})
			kind := discordgo.InteractionApplicationCommand
			if _, ok := tt.data.(discordgo.MessageComponentInteractionData); ok {
				kind = discordgo.InteractionMessageComponent
			}
			b := bot{guildID: "200", service: service, tree: app.NewTree(&app.MemoryTreeStore{}, time.Now)}
			b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "300", AppID: "100", Token: "fixture", GuildID: "200", Type: kind, Data: tt.data,
				Member: &discordgo.Member{User: &discordgo.User{ID: "viewer"}},
			}})
			if !slices.Equal(calls, tt.calls) {
				t.Errorf("calls = %v, want %v", calls, tt.calls)
			}
			for _, want := range tt.text {
				if !strings.Contains(shown, want) {
					t.Errorf("shown misses %q: %s", want, shown)
				}
			}
			// The first expected text is the information, in every case.
			if menu, heading := strings.Index(shown, "stats:view"), strings.Index(shown, tt.text[0]); menu < 0 || heading < menu {
				t.Errorf("the drop-downs must be above the text: %s", shown)
			}
			if strings.Contains(shown, tt.not) {
				t.Errorf("shown has %q: %s", tt.not, shown)
			}
		})
	}
}
