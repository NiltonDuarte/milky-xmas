package discord

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestQuotesArePostedOnlyInQuoteChannels(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "quotes.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	setup := app.NewGuildSetup(store)
	// A quotes-only channel and a commands channel without quotes.
	for channel, choice := range map[string]app.ChannelChoice{"400": {Quotes: new(true)}, "401": {Commands: true}} {
		if saved, err := setup.Configure(t.Context(), "200", channel, true, choice, nil); err != nil || !saved {
			t.Fatalf("configure %s: %v %v", channel, saved, err)
		}
	}
	tuning := core.DefaultTuning()
	tuning.Quotes = core.QuoteRules{EveryMessages: 2, MinInterval: time.Hour}
	service := app.NewService(store, time.Now, func(int) int { return 0 }).WithTuning(tuning)
	b := bot{guildID: "200", setup: setup, quotes: app.NewQuoteFeed(service)}

	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var posted []string
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v9/channels/400/messages" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var message struct {
			Content         string          `json:"content"`
			AllowedMentions json.RawMessage `json:"allowed_mentions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(message.AllowedMentions), `"parse":null`) {
			t.Fatalf("a quote must not ping anyone: %s", message.AllowedMentions)
		}
		posted = append(posted, message.Content)
		return fakeResponse(http.StatusOK, `{"id":"101"}`), nil
	})
	send := func(channel string, author *discordgo.User) {
		b.observeQuote(session, &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "200", ChannelID: channel, Author: author}})
	}
	player, other := &discordgo.User{ID: "300"}, &discordgo.User{ID: "301", Bot: true}
	for range 5 {
		send("401", player)
		send("400", other)
	}
	if len(posted) != 0 {
		t.Fatalf("bots and channels without quotes brought %q", posted)
	}
	send("400", player)
	send("400", player)
	send("400", player)
	send("400", player) // The count is reached again, but the interval has not passed.
	if len(posted) != 1 || !strings.HasPrefix(posted[0], "🎄\n> ") || !strings.Contains(posted[0], "-# Charles Dickens") {
		t.Fatalf("quotes posted = %q", posted)
	}
	// A restart forgets the counts but not when the channel had its quote.
	b.quotes = app.NewQuoteFeed(service)
	send("400", player)
	send("400", player)
	if len(posted) != 1 {
		t.Fatalf("a restart brought another quote inside the interval: %q", posted)
	}
}
