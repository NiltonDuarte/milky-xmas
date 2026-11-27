package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestActivityDropsOnlyObserveConfiguredChannelsAndRenderEightButtons(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "drops.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	setup := app.NewGuildSetup(store)
	if saved, err := setup.Configure(t.Context(), "200", "400", true, app.ChannelChoice{Commands: true, Boss: false}, func(context.Context, string, string) error { return nil }); !saved || err != nil {
		t.Fatalf("configure = %v, %v", saved, err)
	}
	service := app.NewService(store, time.Now, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			return 0
		}
	})
	b := bot{guildID: "200", setup: setup, drops: newDropBot(app.NewActivityDrops(service))}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	posts := 0
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v9/channels/400/messages" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var message struct {
			Content    string `json:"content"`
			Components []struct {
				Components []json.RawMessage `json:"components"`
			} `json:"components"`
			Embeds []struct{ Image struct{ URL string } } `json:"embeds"`
		}
		if files := decodeMultipartPayload(t, r, &message, "giftrain.webp"); files != 1 ||
			len(message.Embeds) != 1 || message.Embeds[0].Image.URL != "attachment://giftrain.webp" {
			t.Fatalf("rain image = %+v with %d files", message.Embeds, files)
		}
		posts++
		if message.Content != giftRainMessage || len(message.Components) != 2 ||
			len(message.Components[0].Components) != 4 || len(message.Components[1].Components) != 4 {
			t.Fatalf("drop message = %+v", message)
		}
		for row, components := range message.Components {
			for column, raw := range components.Components {
				var button discordgo.Button
				if err := json.Unmarshal(raw, &button); err != nil || button.Label != "🎁" ||
					button.CustomID != "drop:1:"+string(rune('0'+row*4+column)) {
					t.Fatalf("button %d,%d = %+v, %v", row, column, button, err)
				}
			}
		}
		return fakeResponse(http.StatusOK, `{"id":"500"}`), nil
	})
	message := func(channel string) *discordgo.MessageCreate {
		return &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "200", ChannelID: channel, Author: &discordgo.User{ID: "300"}}}
	}
	for range 6 {
		b.observeDrop(session, message("401"))
	}
	if posts != 0 {
		t.Fatalf("posts in disabled channel = %d", posts)
	}
	for range core.GiftRainMinimumMessages {
		b.observeDrop(session, message("400"))
	}
	if posts != 1 {
		t.Fatalf("posts in enabled channel = %d", posts)
	}
}

func TestActivityDropButtonRejectsRepeatedAndHiddenRainCooldown(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	service := app.NewService(&giftfixture.Store{}, func() time.Time { return now }, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			return 70
		}
	})
	drops := app.NewActivityDrops(service)
	var rain *app.GiftDrop
	for range core.GiftRainMinimumMessages {
		rain = drops.Observe("200", "400")
	}
	if rain == nil || rain.Kind != core.GiftRainDrop {
		t.Fatalf("rain = %+v", rain)
	}
	b := bot{guildID: "200", drops: newDropBot(drops)}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var replies []string
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			return fakeResponse(http.StatusNoContent, ""), nil
		case strings.HasSuffix(r.URL.Path, "/messages/@original"):
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil || edit.Content == nil {
				t.Fatalf("reply edit = %+v, %v", edit, err)
			}
			replies = append(replies, *edit.Content)
			return fakeResponse(http.StatusOK, `{"id":"501"}`), nil
		default:
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
			return nil, nil
		}
	})
	click := func() {
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "300", AppID: "100", Token: "token", GuildID: "200", ChannelID: "400", Type: discordgo.InteractionMessageComponent,
			Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: "drop:" + rain.ID + ":0"},
		}})
	}
	click()
	click()
	if len(replies) != 2 || replies[0] != "You collected a rare gift." || replies[1] != "You cannot collect another gift yet." || strings.Contains(replies[1], "minute") {
		t.Fatalf("replies = %q", replies)
	}
}

func dropEditSession(t *testing.T, edits chan<- capturedEdit, replies chan<- string) *discordgo.Session {
	t.Helper()
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			return fakeResponse(http.StatusNoContent, ""), nil
		case strings.HasSuffix(r.URL.Path, "/messages/@original"):
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil || edit.Content == nil {
				t.Errorf("reply edit = %+v, %v", edit, err)
				return fakeResponse(http.StatusOK, `{"id":"501"}`), nil
			}
			replies <- *edit.Content
			return fakeResponse(http.StatusOK, `{"id":"501"}`), nil
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v9/channels/400/messages/500":
			var edit capturedEdit
			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
				edit.Files = decodeMultipartPayload(t, r, &edit, dropEndImageName)
			} else if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
				t.Error(err)
			}
			edits <- edit
			return fakeResponse(http.StatusOK, `{"id":"500"}`), nil
		default:
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
			return fakeResponse(http.StatusNotFound, `{}`), nil
		}
	})
	return session
}

func clickDrop(b bot, session *discordgo.Session, user, dropID string, slot int) {
	b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "300", AppID: "100", Token: "token", GuildID: "200", ChannelID: "400", Type: discordgo.InteractionMessageComponent,
		Member:  &discordgo.Member{User: &discordgo.User{ID: user}},
		Message: &discordgo.Message{ID: "500", ChannelID: "400"},
		Data:    discordgo.MessageComponentInteractionData{CustomID: "drop:" + dropID + ":" + strconv.Itoa(slot)},
	}})
}

func TestActiveDropButtonIsDisabledAfterCollecting(t *testing.T) {
	service := app.NewService(&giftfixture.Store{}, time.Now, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return core.GiftRainMessageRange - 1
		case core.RandomActiveDropRolls:
			return 0
		default:
			return 70
		}
	})
	drops := app.NewActivityDrops(service)
	drop := drops.Observe("200", "400")
	if drop == nil || drop.Kind != core.RandomActiveDrop {
		t.Fatalf("drop = %+v", drop)
	}
	b := bot{guildID: "200", drops: newDropBot(drops)}
	edits, replies := make(chan capturedEdit, 4), make(chan string, 4)
	session := dropEditSession(t, edits, replies)
	clickDrop(b, session, "alice", drop.ID, 5)
	if reply := <-replies; reply != "You collected a rare gift." {
		t.Fatalf("reply = %q", reply)
	}
	edit := <-edits
	var disabled []string
	for _, row := range edit.Components {
		for _, button := range row.Components {
			if button.Disabled {
				disabled = append(disabled, button.CustomID)
			}
		}
	}
	if len(disabled) != 1 || disabled[0] != "drop:"+drop.ID+":5" || len(edit.Components) != 2 {
		t.Fatalf("disabled buttons = %v in %d rows", disabled, len(edit.Components))
	}
	clickDrop(b, session, "bob", drop.ID, 5)
	clickDrop(b, session, "alice", drop.ID, 6)
	if first, second := <-replies, <-replies; first != "Someone already opened that gift." || second != "You already collected a gift from this drop." {
		t.Fatalf("replies = %q, %q", first, second)
	}
	if len(edits) != 0 {
		t.Fatalf("rejected clicks edited the message %d times", len(edits))
	}
}

func TestGiftRainEndsAtTenClicksAndListsPlayers(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	rolls := []int{0, 99}
	index := 0
	service := app.NewService(&giftfixture.Store{}, func() time.Time { return now }, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			index++
			return rolls[(index-1)%2]
		}
	})
	drops := app.NewActivityDrops(service)
	var rain *app.GiftDrop
	for range core.GiftRainMinimumMessages {
		rain = drops.Observe("200", "400")
	}
	b := bot{guildID: "200", drops: newDropBot(drops)}
	edits, replies := make(chan capturedEdit, 4), make(chan string, 16)
	session := dropEditSession(t, edits, replies)
	for i := range core.GiftRainClaimLimit {
		if len(edits) != 0 {
			t.Fatalf("message edited before the tenth click (click %d)", i+1)
		}
		clickDrop(b, session, "user"+strconv.Itoa(i), rain.ID, i%8)
		if reply := <-replies; reply != "You collected "+"🌑 Coal." {
			t.Fatalf("reply %d = %q", i+1, reply)
		}
	}
	edit := <-edits
	if edit.Content == nil || !strings.HasPrefix(*edit.Content, "🎉 All the gifts were collected!\n**Collected by**\n<@user0> — 🌑 Coal") ||
		!strings.HasSuffix(*edit.Content, "<@user9> — 🌑 Coal") || len(edit.Components) != 0 {
		t.Fatalf("final message = %+v", edit)
	}
	clickDrop(b, session, "late", rain.ID, 0)
	if reply := <-replies; reply != "This gift drop is no longer available." {
		t.Fatalf("late reply = %q", reply)
	}
}

func TestDropExpiresAfterItsLifetimeAndListsPlayers(t *testing.T) {
	service := app.NewService(&giftfixture.Store{}, time.Now, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return core.GiftRainMessageRange - 1
		case core.RandomActiveDropRolls:
			return 0
		default:
			return 70
		}
	})
	drops := app.NewActivityDrops(service)
	dropBot := newDropBot(drops)
	dropBot.lifetime = 50 * time.Millisecond
	b := bot{guildID: "200", drops: dropBot}
	edits, replies := make(chan capturedEdit, 4), make(chan string, 4)
	session := dropEditSession(t, edits, replies)
	sent := make(chan struct{}, 4)
	base := session.Client.Transport
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v9/channels/400/messages" {
			sent <- struct{}{}
			return fakeResponse(http.StatusOK, `{"id":"500"}`), nil
		}
		return base.RoundTrip(r)
	})
	b.observeDrop(session, &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "200", ChannelID: "400", Author: &discordgo.User{ID: "300"}}})
	<-sent
	clickDrop(b, session, "alice", "1", 0)
	<-replies
	<-edits // Button disabled.
	select {
	case edit := <-edits:
		if edit.Content == nil || *edit.Content != "⏰ Time's up! The Grinch stole all the presents! 😈\n**Collected by**\n<@alice> — 🎁 rare gift" || len(edit.Components) != 0 {
			t.Fatalf("expiry message = %+v", edit)
		}
		if edit.Files != 1 || len(edit.Embeds) != 1 || edit.Embeds[0].Image.URL != "attachment://dropend.gif" {
			t.Fatalf("expiry picture = %+v", edit)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the drop never expired")
	}
	if _, ok := drops.Close("1"); ok {
		t.Fatal("the expired drop was still open")
	}
}

// capturedEdit is the part of a message edit these tests check.
type capturedEdit struct {
	Files  int `json:"-"` // Copies of the end picture uploaded with the edit.
	Embeds []struct {
		Image struct {
			URL string `json:"url"`
		} `json:"image"`
	} `json:"embeds"`
	Content    *string `json:"content"`
	Components []struct {
		Components []struct {
			CustomID string `json:"custom_id"`
			Disabled bool   `json:"disabled"`
		} `json:"components"`
	} `json:"components"`
}

func TestDropMessageImages(t *testing.T) {
	for _, tt := range []struct {
		kind    core.GiftDropKind
		content string
		image   string
	}{
		{core.RandomActiveDrop, randomActiveDropMessage, "grinchmax.png"},
		{core.GiftRainDrop, giftRainMessage, "giftrain.webp"},
	} {
		message := dropMessage(app.GiftDrop{ID: "7", Kind: tt.kind})
		if message.Content != tt.content || len(message.Embeds) != 1 || message.Embeds[0].Image.URL != "attachment://"+tt.image ||
			len(message.Files) != 1 || message.Files[0].Name != tt.image || len(message.Components) != 2 {
			t.Errorf("%s message = %+v", tt.kind, message)
		}
	}
}
