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
	b := bot{guildID: "200", service: app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct{ options, reply string }{
		{`{"name":"open","type":1,"options":[{"name":"day","type":4,"value":1}]}`, "You opened door 1! Advent gifts: 1."},
		{`{"name":"open","type":1,"options":[{"name":"day","type":4,"value":1}]}`, "Door 1 is already open."},
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

type adventEdit struct {
	Embeds []struct {
		Description string
		Image       struct{ URL string }
	}
	Attachments []struct{ ID string }
	Components  []struct {
		Components []struct {
			Label    string
			CustomID string `json:"custom_id"`
			Disabled bool
			Emoji    struct{ Name string }
		}
	}
}

func TestAdventViewAndButtons(t *testing.T) {
	now := time.Date(2026, 12, 4, 10, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	member := app.Member{ServerID: "200", UserID: "555"}
	b := bot{guildID: "200", service: app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })}
	if _, _, err := store.OpenAdventDoor(t.Context(), member, 2026, 2); err != nil {
		t.Fatal(err)
	}
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var edits []adventEdit
	var acks []discordgo.InteractionResponseType
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/callback") {
			var response discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			acks = append(acks, response.Type)
			return fakeResponse(http.StatusNoContent, ""), nil
		}
		var edit adventEdit
		if files := decodeMultipartPayload(t, r, &edit, "advent.jpg"); files != 1 {
			t.Errorf("expected the calendar image once, got %d files", files)
		}
		edits = append(edits, edit)
		return fakeResponse(http.StatusOK, `{"id":"123"}`), nil
	})
	view := func() {
		var interaction discordgo.Interaction
		payload := `{"id":"300","application_id":"100","guild_id":"200","token":"fixture","type":2,"member":{"user":{"id":"555"}},"data":{"name":"advent","type":1,"options":[{"name":"view","type":1}]}}`
		if err := json.Unmarshal([]byte(payload), &interaction); err != nil {
			t.Fatal(err)
		}
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &interaction})
	}
	click := func(day string) {
		b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
			ID: "301", AppID: "100", Token: "fixture", GuildID: "200", Type: discordgo.InteractionMessageComponent,
			Member: &discordgo.Member{User: &discordgo.User{ID: "555"}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: "advent:" + day},
		}})
	}
	state := func(edit adventEdit) (rows []int, labels map[string]string, disabled map[string]bool) {
		labels, disabled = map[string]string{}, map[string]bool{}
		for _, row := range edit.Components {
			rows = append(rows, len(row.Components))
			for _, button := range row.Components {
				labels[button.CustomID] = button.Label + button.Emoji.Name
				disabled[button.CustomID] = button.Disabled
			}
		}
		return
	}

	view()
	if len(edits) != 1 || len(edits[0].Embeds) != 2 || edits[0].Embeds[0].Image.URL != "attachment://advent.jpg" ||
		edits[0].Embeds[1].Description != "🎄 Advent 2026 · December 1–24 UTC · Advent gifts: 1" {
		t.Fatalf("view = %+v", edits)
	}
	rows, labels, disabled := state(edits[0])
	if len(rows) != 5 || rows[0] != 5 || rows[3] != 5 || rows[4] != 4 {
		t.Fatalf("button rows = %v", rows)
	}
	// Today is December 4: doors 1, 3 and 4 can be opened, door 2 was opened, the rest are locked.
	for day, want := range map[string]struct {
		label    string
		disabled bool
	}{"1": {"1", false}, "2": {"❌", true}, "3": {"3", false}, "4": {"4", false}, "5": {"5", true}, "24": {"24", true}} {
		if labels["advent:"+day] != want.label || disabled["advent:"+day] != want.disabled {
			t.Errorf("door %s = %q disabled %v, want %q disabled %v", day, labels["advent:"+day], disabled["advent:"+day], want.label, want.disabled)
		}
	}

	click("3")
	if len(acks) != 2 || acks[1] != discordgo.InteractionResponseDeferredMessageUpdate || len(edits) != 2 {
		t.Fatalf("click acknowledgements %v, edits %d", acks, len(edits))
	}
	after := edits[1]
	_, labels, disabled = state(after)
	if labels["advent:3"] != "❌" || !disabled["advent:3"] || disabled["advent:1"] || disabled["advent:4"] ||
		after.Embeds[1].Description != "🎄 Advent 2026 · December 1–24 UTC · Advent gifts: 2\n🎁 You opened door 3! Advent gifts: 2." ||
		len(after.Attachments) != 1 || after.Attachments[0].ID != "0" {
		t.Fatalf("after clicking door 3: %+v", after)
	}
	doors, err := store.AdventDoors(t.Context(), member, 2026)
	if err != nil || doors != 0b110 {
		t.Fatalf("stored doors = %b, %v", doors, err)
	}

	click("3") // A stale button from before the first click.
	if got := edits[2].Embeds[1].Description; got != "🎄 Advent 2026 · December 1–24 UTC · Advent gifts: 2\nDoor 3 is already open. Advent gifts: 2." {
		t.Fatalf("second click = %q", got)
	}
	click("9") // A locked door cannot be opened from a stale button either.
	if got := edits[3].Embeds[1].Description; !strings.Contains(got, "Door 9 unlocks on December 9") {
		t.Fatalf("locked click = %q", got)
	}
	if doors, _ := store.AdventDoors(t.Context(), member, 2026); doors != 0b110 {
		t.Fatalf("stored doors after rejected clicks = %b", doors)
	}
	before := len(acks)
	click("0")
	click("25")
	click("x")
	if len(acks) != before {
		t.Fatalf("invalid advent buttons were answered: %v", acks[before:])
	}
}
