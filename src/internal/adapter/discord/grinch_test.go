package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestGrinchSummon(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	channel := "400"
	worker := &grinchBot{guild: "200", channel: channel,
		encounter: app.NewGrinchEncounter(&giftfixture.Store{}, "200", channel, func() time.Time { return now }, func(n int) int { return n - 1 }),
	}
	b := bot{guildID: "200", grinches: &grinchChannels{workers: map[grinchChannel]grinchWorker{{"200", channel}: {bot: worker}}}}
	for _, step := range []struct {
		channel      string
		permission   int64
		announcement bool
		reply        string
	}{
		{channel, 0, false, "You need Manage Messages permission"},
		{"other", discordgo.PermissionManageMessages, false, "Use this command in a configured Grinch channel"},
		{channel, discordgo.PermissionManageMessages, true, "The Grinch is coming!"},
		{channel, discordgo.PermissionAdministrator, false, "A Grinch encounter is already in progress."},
	} {
		session, err := discordgo.New("Bot fixture")
		if err != nil {
			t.Fatal(err)
		}
		posted := false
		callbacks, replies := 0, 0
		session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/callback"):
				callbacks++
				var response discordgo.InteractionResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral {
					t.Error("summon acknowledgement must be private")
				}
				return fakeResponse(http.StatusNoContent, ""), nil
			case strings.HasSuffix(r.URL.Path, "/@original"):
				replies++
				var edit discordgo.WebhookEdit
				if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
					t.Fatal(err)
				}
				if edit.Content == nil || !strings.Contains(*edit.Content, step.reply) {
					t.Errorf("expected reply %q, got %+v", step.reply, edit.Content)
				}
				return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
			case r.URL.Path == "/api/v9/channels/400/messages" && r.Method == http.MethodPost:
				if posted {
					t.Error("duplicate announcement")
				}
				posted = true
				var message struct {
					Content    string
					Components []json.RawMessage
				}
				if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
					t.Fatal(err)
				}
				if len(message.Components) != 1 || !strings.Contains(message.Content, "The Grinch is coming!") || !strings.Contains(string(message.Components[0]), "🎅") {
					t.Errorf("announcement = %+v", message)
				}
				return fakeResponse(http.StatusOK, `{"id":"999","timestamp":"2026-12-01T12:00:00Z"}`), nil
			case r.Method == http.MethodPatch && r.URL.Path == "/api/v9/channels/400/messages/999":
				return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
			default:
				t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
				return nil, nil
			}
		})
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "100", AppID: "200", GuildID: "200", ChannelID: step.channel, Token: "fake",
			Type: discordgo.InteractionApplicationCommand, Member: &discordgo.Member{Permissions: step.permission, User: &discordgo.User{ID: "50"}},
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "grinch", CommandType: discordgo.ChatApplicationCommand,
				Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: "summon", Type: discordgo.ApplicationCommandOptionSubCommand}},
			},
		}})
		if posted != step.announcement || callbacks != 1 || replies != 1 {
			t.Errorf("posted %v, %d callbacks, %d replies, want post %v and one of each response", posted, callbacks, replies, step.announcement)
		}
	}
}

func TestGrinchDefenseBoard(t *testing.T) {
	_, components := grinchMessage(core.Grinch{
		ID: "fight", Phase: core.GrinchDefense, StartedAt: time.Now(), Board: 1, Participants: map[string]bool{"one": true},
	})
	if len(components) != 2 {
		t.Fatalf("rows = %d", len(components))
	}
	for row, component := range components {
		buttons := component.(discordgo.ActionsRow).Components
		if len(buttons) != 4 {
			t.Fatalf("row %d has %d buttons", row, len(buttons))
		}
		for col, item := range buttons {
			want := "❄️"
			if row == 0 && col == 0 {
				want = "🎄"
			}
			if button := item.(discordgo.Button); button.Emoji.Name != want {
				t.Errorf("unexpected tile: %+v", button)
			}
		}
	}
}

func TestGrinchButtonAcknowledgement(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	s, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	b := &grinchBot{guild: "200", channel: "400",
		encounter: app.NewGrinchEncounter(&giftfixture.Store{}, "200", "scope", func() time.Time { return now }, func(int) int { return 0 }),
	}
	display := grinchDisplay{s, b.channel}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(r.URL.Path, "/400/messages") {
			t.Fatalf("unexpected spawn request: %s", r.URL)
		}
		return fakeResponse(http.StatusOK, `{"id":"999","timestamp":"2026-12-01T12:00:00Z"}`), nil
	})
	ok, err := b.encounter.Spawn(t.Context(), time.Time{}, true, display)
	if err != nil || !ok {
		t.Fatalf("spawn = %v, %v", ok, err)
	}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
	})
	fight, err := b.encounter.Sync(t.Context(), display)
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range []string{"✅ You joined the Grinch fight!", "✅ You already joined this fight."} {
		s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			var response discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if !strings.HasSuffix(r.URL.Path, "/callback") || response.Type != discordgo.InteractionResponseChannelMessageWithSource ||
				response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral || !strings.HasPrefix(response.Data.Content, want) {
				t.Errorf("expected private confirmation %q, got %+v", want, response)
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		})
		eventID := fmt.Sprint(((now.UnixMilli() - 1420070400000) << 22) + int64(index))
		b.button(s, &discordgo.Interaction{
			ID: eventID, ChannelID: "400", Message: &discordgo.Message{ID: "999"},
			Member: &discordgo.Member{User: &discordgo.User{ID: "50"}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: "grinch:" + fight.ID + ":lobby:0"},
		})
	}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
	})
	now = now.Add(core.GrinchJoinDuration)
	if _, err := b.encounter.Sync(t.Context(), display); err != nil {
		t.Fatal(err)
	}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		var response discordgo.InteractionResponse
		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(r.URL.Path, "/callback") || response.Type != discordgo.InteractionResponseDeferredMessageUpdate {
			t.Errorf("unexpected button acknowledgement: %+v", response)
		}
		return fakeResponse(http.StatusNoContent, ""), nil
	})
	now = now.Add(time.Second)
	eventID := fmt.Sprint((now.UnixMilli() - 1420070400000) << 22)
	b.button(s, &discordgo.Interaction{
		ID: eventID, ChannelID: "400", Message: &discordgo.Message{ID: "999"},
		Member: &discordgo.Member{User: &discordgo.User{ID: "50"}},
		Data:   discordgo.MessageComponentInteractionData{CustomID: "grinch:" + fight.ID + ":snowballs:0"},
	})
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
	})
	state, err := b.encounter.Sync(t.Context(), display)
	if err != nil || state.Health() != 99 {
		t.Fatalf("health after click: %v, %v", state.Health(), err)
	}
}

func TestGrinchAttackBoardEmojis(t *testing.T) {
	_, rows := grinchMessage(core.Grinch{Phase: core.GrinchAttack, Board: 1})
	buttons := rows[0].(discordgo.ActionsRow).Components
	ice, base := buttons[0].(discordgo.Button), buttons[1].(discordgo.Button)
	if ice.Emoji == nil || ice.Emoji.Name != "🧊" || base.Emoji == nil || base.Emoji.Name != "☃️" {
		t.Fatalf("invalid attack buttons: %+v %+v", ice, base)
	}
}

func TestGrinchLeaderboardRendering(t *testing.T) {
	scores := []core.GrinchScore{{Member: "123", Points: 12.5}, {Member: "456"}}
	content, extra := grinchLeaderboards("The Grinch escaped!", scores, []core.GrinchScore{{Member: "123", Points: 4}})
	if extra != "" || !strings.Contains(content, "1. <@123> — 12.50 pts\n2. <@456> — 0.00 pts") {
		t.Fatalf("unexpected leaderboard: %q, %q", content, extra)
	}
	for i := range 100 {
		scores = append(scores, core.GrinchScore{Member: fmt.Sprintf("%018d", i+100), Points: 1})
	}
	content, extra = grinchLeaderboards("The Grinch escaped!", scores, []core.GrinchScore{{Member: "123", Points: 4}})
	if len(content) > 2000 || !strings.Contains(content, "Full standings are attached.") ||
		!strings.Contains(extra, "102. Discord user 000000000000000199 — 1.00 pts") {
		t.Fatalf("unexpected leaderboard overflow: %d bytes, %d-byte attachment", len(content), len(extra))
	}
}

func TestGrinchFinalBoardSnowballButton(t *testing.T) {
	for _, round := range []int{0, 7} {
		content, rows := grinchMessage(core.Grinch{ID: "fight", Phase: core.GrinchFinalAttack, HP: 500, FinalRound: round})
		buttons := rows[0].(discordgo.ActionsRow).Components
		button := buttons[len(buttons)-1].(discordgo.Button)
		want := "grinch:fight:final:4"
		if round != 0 {
			want += ":7"
		}
		if button.CustomID != want || button.Emoji == nil || button.Emoji.Name != "❄️" || len(buttons) > 5 || !strings.Contains(content, "1 Snowball") {
			t.Errorf("round %d: snowball button = %+v; items = %d; content = %q", round, button, len(buttons), content)
		}
	}
}
