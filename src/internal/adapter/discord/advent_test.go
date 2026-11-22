package discord

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func TestAdventInteractions(t *testing.T) {
	now := time.Date(2026, 12, 4, 0, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	b := bot{guildID: "200", service: app.NewService(store, func() time.Time { return now })}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct{ options, reply string }{
		{`{"name":"open","type":1,"options":[{"name":"day","type":4,"value":1}]}`, "You opened door 1! Advent gifts: 1."},
		{`{"name":"open","type":1,"options":[{"name":"day","type":4,"value":1}]}`, "Door 1 is already open."},
		{`{"name":"view","type":1}`, "01 · opened\n02 · available"},
	} {
		calls := 0
		session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				var response discordgo.InteractionResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if r.Method != http.MethodPost || response.Type != discordgo.InteractionResponseDeferredChannelMessageWithSource ||
					response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral {
					t.Errorf("unexpected acknowledgement: %+v", response)
				}
				return fakeResponse(http.StatusNoContent, ""), nil
			}
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
				t.Fatal(err)
			}
			if r.Method != http.MethodPatch || edit.Content == nil || !strings.Contains(*edit.Content, step.reply) {
				t.Errorf("unexpected reply: %+v, want %q", edit.Content, step.reply)
			}
			return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
		})
		var interaction discordgo.Interaction
		payload := `{"id":"300","application_id":"100","guild_id":"200","token":"fixture","type":2,"member":{"user":{"id":"555"}},"data":{"name":"advent","type":1,"options":[` + step.options + `]}}`
		if err := json.Unmarshal([]byte(payload), &interaction); err != nil {
			t.Fatal(err)
		}
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &interaction})
		if calls != 2 {
			t.Errorf("expected acknowledgement and edit, got %d calls", calls)
		}
	}
	doors, err := store.AdventDoors(t.Context(), app.Member{ServerID: "200", UserID: "555"}, 2026)
	if err != nil || doors != 1 {
		t.Fatalf("stored doors = %b, %v", doors, err)
	}
}
