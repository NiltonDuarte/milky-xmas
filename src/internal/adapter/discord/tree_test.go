package discord

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

// waterReply is what the player and the channel got from one /water.
type waterReply struct {
	calls   []string
	private string // The deferred reply's text, or the text of its box.
	channel string // The text of the box posted to the channel.
	image   string
}

type boxPayload struct {
	Type discordgo.InteractionResponseType
	Data struct{ Flags discordgo.MessageFlags }

	Content    *string
	Flags      discordgo.MessageFlags
	Components []struct {
		Components []struct {
			Content string
			Items   []struct{ Media struct{ URL string } }
		}
	}
}

func (payload boxPayload) read() (text, image string) {
	if len(payload.Components) != 1 {
		if payload.Content != nil {
			return *payload.Content, ""
		}
		return "", ""
	}
	var texts []string
	for _, part := range payload.Components[0].Components {
		if part.Content != "" {
			texts = append(texts, part.Content)
		}
		if len(part.Items) == 1 {
			image = strings.TrimPrefix(part.Items[0].Media.URL, "attachment://")
		}
	}
	return strings.Join(texts, "\n"), image
}

func water(t *testing.T, b bot, user string, channelStatus int) waterReply {
	t.Helper()
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var got waterReply
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		kind := "edit"
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			kind = "ack"
		case r.Method == http.MethodDelete:
			kind = "delete"
		case r.Method == http.MethodPost:
			kind = "channel"
		}
		got.calls = append(got.calls, kind)
		if kind == "delete" {
			return fakeResponse(http.StatusNoContent, ""), nil
		}
		var payload boxPayload
		if strings.Contains(r.Header.Get("Content-Type"), "multipart") {
			name := treeImage(core.TreeStage(levelOf(t, b))).Name
			if files := decodeMultipartPayload(t, r, &payload, name); files != 1 {
				t.Errorf("%s: %d files named %s", kind, files, name)
			}
		} else if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		text, image := payload.read()
		switch kind {
		case "ack":
			if payload.Type != discordgo.InteractionResponseDeferredChannelMessageWithSource || payload.Data.Flags != discordgo.MessageFlagsEphemeral {
				t.Errorf("acknowledgement = %+v", payload)
			}
		case "channel":
			if payload.Flags != discordgo.MessageFlagsIsComponentsV2 {
				t.Errorf("channel post flags = %v", payload.Flags)
			}
			got.channel, got.image = text, image
			return fakeResponse(channelStatus, `{"id":"1"}`), nil
		default:
			got.private = text
			if image != "" {
				got.image = image
			}
		}
		return fakeResponse(http.StatusOK, `{"id":"1"}`), nil
	})
	b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "300", AppID: "100", Token: "fixture", GuildID: "200", ChannelID: "400", Type: discordgo.InteractionApplicationCommand,
		Member: &discordgo.Member{User: &discordgo.User{ID: user}},
		Data:   discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "water"},
	}})
	return got
}

func levelOf(t *testing.T, b bot) int {
	t.Helper()
	tree, err := b.tree.View(t.Context(), "200")
	if err != nil {
		t.Fatal(err)
	}
	return tree.Level
}

func waterBot(now time.Time, startLevel int) bot {
	service := app.NewService(&giftfixture.Store{}, func() time.Time { return now }, func(int) int { return 0 })
	rules := core.DefaultTreeRules()
	rules.StartLevel = startLevel
	tree := app.NewTree(&app.MemoryTreeStore{}, func() time.Time { return now }).WithRules(rules)
	return bot{guildID: "200", tree: service.RewardTree(tree)}
}

func TestWaterInteraction(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	b := waterBot(now, core.TreeStartLevel)
	got := water(t, b, "alice", http.StatusOK)
	want := "💧 <@alice> watered the tree! +2%\n😇 Good deeds +1\n## 🎄 The Christmas tree\n🟦⬜⬜⬜⬜⬜⬜⬜⬜⬜  **12%** 💧\n🥀 The tree is withered."
	if !slices.Equal(got.calls, []string{"ack", "channel", "delete"}) || !strings.HasPrefix(got.channel, want) ||
		strings.Contains(got.channel, "Top waterers") || got.image != "deadtree.jpg" {
		t.Errorf("watering = %+v", got)
	}
	got = water(t, b, "alice", http.StatusOK)
	if !slices.Equal(got.calls, []string{"ack", "edit"}) || got.private != "💧 You already watered the tree. Try again in 1 hour." {
		t.Errorf("watering too soon = %+v", got)
	}
	got = water(t, b, "bob", http.StatusForbidden)
	if !slices.Equal(got.calls, []string{"ack", "channel", "edit"}) || !strings.HasPrefix(got.private, "💧 <@bob> watered the tree!") {
		t.Errorf("watering when the channel post fails = %+v", got)
	}

	for _, tt := range []struct {
		start int
		image string
	}{{48, "deadtree.jpg"}, {49, ""}, {97, ""}, {98, "fulltree.png"}} {
		if got := water(t, waterBot(now, tt.start), "alice", http.StatusOK); got.image != tt.image {
			t.Errorf("watering from %d%% shows %q, want %q", tt.start, got.image, tt.image)
		}
	}

	got = water(t, waterBot(now, core.TreeMaxLevel), "alice", http.StatusOK)
	if !slices.Equal(got.calls, []string{"ack", "edit"}) || got.image != "fulltree.png" || got.channel != "" ||
		got.private != "🎄 The tree is fully watered. Try again later!" {
		t.Errorf("watering a full tree = %+v", got)
	}

	got = water(t, waterBot(core.TreeEventEnd, core.TreeStartLevel), "alice", http.StatusOK)
	if !slices.Equal(got.calls, []string{"ack", "edit"}) || got.image != "deadtree.jpg" ||
		!strings.HasPrefix(got.private, "🎄 The event is over, so the tree can't be watered any more.\n## 🎄 The Christmas tree") ||
		!strings.Contains(got.private, "The event has ended. This is the tree's final state.") {
		t.Errorf("event-over reply = %+v", got)
	}
	if command := waterCommand(); command.Name != "water" || len(command.Options) != 0 {
		t.Errorf("unexpected /water registration: %+v", command)
	}
}
