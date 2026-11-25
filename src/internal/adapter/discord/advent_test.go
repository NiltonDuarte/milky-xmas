package discord

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/adventtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

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
	var edits []replyBox
	var surprises []string
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
		if r.Method == http.MethodPost {
			var followup discordgo.WebhookParams
			if err := json.NewDecoder(r.Body).Decode(&followup); err != nil {
				t.Fatal(err)
			}
			if followup.Flags != discordgo.MessageFlagsEphemeral {
				t.Errorf("surprise is not private: %+v", followup)
			}
			text := followup.Content
			if len(followup.Embeds) == 1 {
				text = followup.Embeds[0].Description
			}
			surprises = append(surprises, text)
			return fakeResponse(http.StatusOK, `{"id":"124"}`), nil
		}
		edits = append(edits, decodeBox(t, r, "advent.jpg"))
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
	state := func(edit replyBox) (rows []int, order []string, styles map[string]discordgo.ButtonStyle, disabled map[string]bool) {
		styles, disabled = map[string]discordgo.ButtonStyle{}, map[string]bool{}
		for _, row := range edit.rows {
			rows = append(rows, len(row.Components))
			for _, button := range row.Components {
				if button.CustomID != "advent:"+button.Label {
					t.Errorf("button %q has id %q", button.Label, button.CustomID)
				}
				order = append(order, button.Label)
				styles[button.Label], disabled[button.Label] = button.Style, button.Disabled
			}
		}
		return
	}
	surprise := func(day int) string {
		opening, err := b.service.OpenAdvent(t.Context(), member, day)
		if err != nil || opening.Status != core.DoorOpened {
			t.Fatalf("door %d = %+v, %v", day, opening, err)
		}
		return adventtext.Opening(day, opening)
	}

	view()
	if len(edits) != 1 || edits[0].text != "🎄 Advent 2026 · December 1–24 UTC" {
		t.Fatalf("view = %+v", edits)
	}
	rows, order, styles, disabled := state(edits[0])
	if len(rows) != 5 || rows[0] != 5 || rows[3] != 5 || rows[4] != 4 || len(styles) != 24 {
		t.Fatalf("button rows = %v", rows)
	}
	if slices.IsSortedFunc(order, func(a, b string) int { x, _ := strconv.Atoi(a); y, _ := strconv.Atoi(b); return x - y }) {
		t.Errorf("buttons are not shuffled: %v", order)
	}
	// Today is December 4: door 2 was opened, doors 1, 3 and 4 are ready, the rest are locked.
	for day, want := range map[string]struct {
		style    discordgo.ButtonStyle
		disabled bool
	}{
		"1": {discordgo.PrimaryButton, false}, "2": {discordgo.SuccessButton, false}, "3": {discordgo.PrimaryButton, false},
		"4": {discordgo.PrimaryButton, false}, "5": {discordgo.SecondaryButton, true}, "24": {discordgo.SecondaryButton, true},
	} {
		if styles[day] != want.style || disabled[day] != want.disabled {
			t.Errorf("door %s = style %d disabled %v, want %d %v", day, styles[day], disabled[day], want.style, want.disabled)
		}
	}

	click("3")
	if len(acks) != 2 || acks[1] != discordgo.InteractionResponseDeferredMessageUpdate || len(edits) != 2 || len(surprises) != 1 {
		t.Fatalf("click acknowledgements %v, edits %d, surprises %d", acks, len(edits), len(surprises))
	}
	_, after, styles, disabled := state(edits[1])
	if styles["3"] != discordgo.SuccessButton || disabled["3"] || !slices.Equal(after, order) || edits[1].attachments != 1 {
		t.Fatalf("after clicking door 3: %+v", edits[1])
	}
	doors, err := store.AdventDoors(t.Context(), member, 2026)
	if err != nil || doors != 0b110 {
		t.Fatalf("stored doors = %b, %v", doors, err)
	}

	click("3") // An opened door shows the same surprise again.
	click("2")
	if len(surprises) != 3 || surprises[1] != surprises[0] || surprises[0] != surprise(3) || surprises[2] != surprise(2) ||
		surprises[2] == surprises[0] {
		t.Fatalf("surprises = %q", surprises)
	}
	edited := len(edits)
	click("9") // A locked door cannot be opened from a stale button either.
	if len(surprises) != 4 || !strings.Contains(surprises[3], "Door 9 unlocks on December 9") || len(edits) != edited {
		t.Fatalf("locked click = %q, edits %d", surprises, len(edits))
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
