package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
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

func TestGrinchSummon(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	channel := "400"
	worker := &grinchBot{guild: "200", channel: channel,
		encounter: app.NewGrinchEncounter(&giftfixture.Store{}, "200", channel, classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(n int) int { return n - 1 }),
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
					Flags      discordgo.MessageFlags
					Components []json.RawMessage
				}
				if files := decodeMultipartPayload(t, r, &message, "grinch.jpg"); files != 1 {
					t.Errorf("announcement carried %d images", files)
				}
				if announced := string(message.Components[0]); len(message.Components) != 1 || message.Content != "" ||
					message.Flags != discordgo.MessageFlagsIsComponentsV2 || !strings.Contains(announced, "attachment://grinch.jpg") ||
					!strings.Contains(announced, "The Grinch is coming!") || !strings.Contains(announced, "Ready") || !strings.Contains(announced, "🎅") {
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
	_, components := grinchMessage(classicFight(core.DefaultTuning().Grinch), core.Grinch{
		ID: "fight", Phase: core.GrinchDefense, StartedAt: time.Now(), Board: 1, Joined: map[string]bool{"one": true},
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
		encounter: app.NewGrinchEncounter(&giftfixture.Store{}, "200", "scope", classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 }),
	}
	display := grinchDisplay{s, b.channel, classicFight(core.DefaultTuning().Grinch), nil, nil}
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
	id := func(index int64) string {
		return fmt.Sprint(((now.UnixMilli() - 1420070400000) << 22) + index)
	}
	join := "grinch:" + fight.ID + ":start:0"
	for index, step := range []struct {
		user      string
		confirmed bool // The player joins now and gets one private message.
	}{
		{"50", true}, {"51", true}, {"50", false}, {"51", false},
	} {
		var requests []string
		s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/callback"):
				var response discordgo.InteractionResponse
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Type != discordgo.InteractionResponseDeferredMessageUpdate || response.Data != nil {
					t.Errorf("expected a silent deferred acknowledgement, got %+v", response)
				}
				requests = append(requests, "ack")
				return fakeResponse(http.StatusNoContent, ""), nil
			case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/webhooks/"):
				var followup struct {
					Content string                 `json:"content"`
					Flags   discordgo.MessageFlags `json:"flags"`
				}
				if err := json.NewDecoder(r.Body).Decode(&followup); err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(followup.Content, "✅ You pressed Start!") || followup.Flags != discordgo.MessageFlagsEphemeral {
					t.Errorf("follow-up = %+v, want a private join confirmation", followup)
				}
				requests = append(requests, "confirmation")
				return fakeResponse(http.StatusOK, `{"id":"1"}`), nil
			}
			t.Fatalf("unexpected request %s %s", r.Method, r.URL)
			return nil, nil
		})
		b.button(s, &discordgo.Interaction{
			ID: id(int64(index)), ChannelID: "400", Message: &discordgo.Message{ID: "999"},
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: join},
		})
		want := []string{"ack"}
		if step.confirmed {
			want = append(want, "confirmation")
		}
		if !slices.Equal(requests, want) {
			t.Fatalf("click %d by %s: requests = %q, want %q", index, step.user, requests, want)
		}
	}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
	})
	now = now.Add(core.GrinchJoinDuration)
	if state, err := b.encounter.Sync(t.Context(), display); err != nil || len(state.Joined) != 2 {
		t.Fatalf("expected two distinct players: %+v, %v", state.Joined, err)
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
	if err != nil || state.Health(core.DefaultTuning().Grinch) != 95.83 {
		t.Fatalf("health after click: %v, %v", state.Health(core.DefaultTuning().Grinch), err)
	}
}

func TestGrinchAttackBoardEmojis(t *testing.T) {
	_, rows := grinchMessage(classicFight(core.DefaultTuning().Grinch), core.Grinch{Phase: core.GrinchAttack, Board: 1})
	buttons := rows[0].(discordgo.ActionsRow).Components
	ice, base := buttons[0].(discordgo.Button), buttons[1].(discordgo.Button)
	if ice.Emoji == nil || ice.Emoji.Name != "🧊" || base.Emoji == nil || base.Emoji.Name != "☃️" {
		t.Fatalf("invalid attack buttons: %+v %+v", ice, base)
	}
}

func TestGrinchStandings(t *testing.T) {
	rules := classicFight(core.DefaultTuning().Grinch)
	fight := core.Grinch{Phase: core.GrinchEscaped, EndedAt: time.Now(), HP: 4000,
		Joined: map[string]bool{"123": true, "456": true, "789": true}, DamageDealt: map[string]core.GrinchDamage{"123": {Hundredths: 1250}},
		Gifts: map[string]int{"123": 4}, Snowballs: map[string]int{"456": 2}}
	sections, overflow := grinchStandings(rules, fight)
	// Only players with points are listed. 789 scored nothing and 456 has no damage or gifts.
	want := []string{
		"**🏆 Damage leaderboard**\n1. <@123> — 12.50 dmg",
		"**🎁 Gift leaderboard**\n1. <@123> — 4",
		"**❄️ Snowball leaderboard**\n1. <@456> — 2",
	}
	if overflow || !slices.Equal(sections, want) {
		t.Fatalf("standings = %q, overflow %v, want %q", sections, overflow, want)
	}
	if sections, overflow := grinchStandings(rules, core.Grinch{Phase: core.GrinchEscaped, Joined: map[string]bool{"123": true}, DamageDealt: map[string]core.GrinchDamage{}}); len(sections) != 0 || overflow {
		t.Errorf("a fight where nobody scored lists %q", sections)
	}

	for i := range 100 {
		member := fmt.Sprintf("%018d", i+100)
		fight.Joined[member] = true
		fight.DamageDealt[member] = core.GrinchDamage{Hundredths: 100}
	}
	sections, overflow = grinchStandings(rules, fight)
	if !overflow || strings.Count(sections[0], "\n") != grinchStandingsLimit {
		t.Fatalf("with 101 scoring players: overflow %v, damage section has %d lines, want the top %d", overflow, strings.Count(sections[0], "\n"), grinchStandingsLimit)
	}
	text := grinchStandingsText(rules, fight)
	if !strings.Contains(text, "101. Discord user 000000000000000199 —") || strings.Contains(text, "Discord user 789") {
		t.Fatalf("standings file = %q, want every scoring player and nobody without points", text[len(text)-160:])
	}
}

func TestGrinchFinalBoardSnowballButton(t *testing.T) {
	for _, round := range []int{0, 7} {
		content, rows := grinchMessage(classicFight(core.DefaultTuning().Grinch), core.Grinch{ID: "fight", Phase: core.GrinchFinalAttack, HP: 500, FinalRound: round})
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

func TestGrinchHealthDisplay(t *testing.T) {
	for _, state := range []core.Grinch{
		{Phase: core.GrinchLobby, HP: 10000},
		{Phase: core.GrinchAttack, HP: 4500},
		{Phase: core.GrinchWon, HP: 0, EndedAt: time.Now()},
	} {
		content, _ := grinchMessage(classicFight(core.DefaultTuning().Grinch), state)
		line, _, _ := strings.Cut(content, "\n")
		want := fmt.Sprintf("👽 Grinch · %.2f/100%% HP", state.Health(core.DefaultTuning().Grinch))
		if line != want {
			t.Errorf("health line for %s = %q, want %q", state.Phase, line, want)
		}
	}
}

func TestGrinchReshuffleEditsOriginalMessage(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	b := &grinchBot{guild: "200", channel: "400",
		encounter: app.NewGrinchEncounter(store, "200", "scope", classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 })}
	fight := core.Grinch{ID: "fight", MessageID: "999", HP: 500, Phase: core.GrinchFinalAttack, FinalRound: 1,
		StartedAt: now, TimedAt: now, Joined: map[string]bool{"50": true}, Accepted: make(map[string]bool), DamageDealt: map[string]core.GrinchDamage{}}
	if saved, err := store.SaveGrinch(t.Context(), "scope", 0, fight, "200", nil); !saved || err != nil {
		t.Fatalf("seed final phase: %v, %v", saved, err)
	}
	s, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	callbacks, edits := 0, 0
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			callbacks++
			var response discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Type != discordgo.InteractionResponseDeferredMessageUpdate || response.Data != nil {
				t.Errorf("expected silent acknowledgement: %+v", response)
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		case r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/channels/400/messages/999"):
			edits++
			var update struct {
				Content    string                 `json:"content"`
				Flags      discordgo.MessageFlags `json:"flags"`
				Components []json.RawMessage      `json:"components"`
			}
			if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
				t.Fatal(err)
			}
			if update.Content != "" || update.Flags != discordgo.MessageFlagsIsComponentsV2 || len(update.Components) != 2 ||
				!strings.Contains(string(update.Components[0]), "👽 Grinch · 5.00/100% HP") ||
				strings.Contains(string(update.Components[0]), "grinch:fight:final") ||
				!strings.Contains(string(update.Components[1]), "grinch:fight:final:4:2") ||
				!strings.Contains(string(update.Components[1]), "The Grinch") || !strings.Contains(string(update.Components[1]), "@50") {
				t.Errorf("unexpected update: %+v", update)
			}
			return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
			return nil, nil
		}
	})
	for index := range 2 {
		b.button(s, &discordgo.Interaction{
			ID:        fmt.Sprint(((now.UnixMilli() - 1420070400000) << 22) + int64(index)),
			ChannelID: "400", Message: &discordgo.Message{ID: "999"},
			Member: &discordgo.Member{User: &discordgo.User{ID: "50"}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: "grinch:fight:final:4:1"},
		})
	}
	b.shown.set("999", grinchHeartImageName) // The final attack's picture is already on the message.
	if _, err := b.encounter.Sync(t.Context(), grinchDisplay{s, b.channel, classicFight(core.DefaultTuning().Grinch), &b.shown, nil}); err != nil {
		t.Fatal(err)
	}
	player, err := store.Load(t.Context(), app.Member{ServerID: "200", UserID: "50"})
	if err != nil || callbacks != 2 || edits != 1 || player.Inventory.Snowballs != 1 {
		t.Fatalf("callbacks = %d, edits = %d, inventory = %+v, %v", callbacks, edits, player.Inventory, err)
	}
}

// decodeMultipartPayload decodes the JSON payload of a multipart request and returns how many
// non-empty files with the given name it carried.
func decodeMultipartPayload(t *testing.T, r *http.Request, payload any, fileName string) int {
	t.Helper()
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(r.Body, params["boundary"])
	files := 0
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return files
		}
		if err != nil {
			t.Fatal(err)
		}
		if part.FormName() == "payload_json" {
			if err := json.NewDecoder(part).Decode(payload); err != nil {
				t.Fatal(err)
			}
		} else if part.FileName() == fileName {
			if data, _ := io.ReadAll(part); len(data) > 0 {
				files++
			}
		}
	}
}

func TestGrinchFinalMessageImages(t *testing.T) {
	for _, tt := range []struct {
		name   string
		phase  core.GrinchPhase
		images []string
	}{
		{"won", core.GrinchWon, []string{"grinch.jpg"}},
		{"defeat", core.GrinchDefeat, []string{"maxpee.jpg"}},
		{"escaped", core.GrinchEscaped, []string{"maxpee.jpg"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			session, err := discordgo.New("Bot fixture")
			if err != nil {
				t.Fatal(err)
			}
			edits := 0
			session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				edits++
				_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil {
					t.Fatal(err)
				}
				var edit struct {
					Components  []json.RawMessage
					Attachments []json.RawMessage
				}
				var files []string
				reader := multipart.NewReader(r.Body, params["boundary"])
				for {
					part, err := reader.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if part.FormName() == "payload_json" {
						if err := json.NewDecoder(part).Decode(&edit); err != nil {
							t.Fatal(err)
						}
					} else if data, _ := io.ReadAll(part); len(data) > 0 {
						files = append(files, part.FileName())
					}
				}
				var panel struct {
					Components []struct {
						Items []struct{ Media struct{ URL string } }
					}
				}
				if err := json.Unmarshal(edit.Components[0], &panel); err != nil {
					t.Fatal(err)
				}
				var urls []string
				for _, item := range panel.Components[0].Items {
					urls = append(urls, strings.TrimPrefix(item.Media.URL, "attachment://"))
				}
				if !slices.Equal(urls, tt.images) || !slices.Equal(files, tt.images) {
					t.Errorf("final message pictures %v and files %v, want %v", urls, files, tt.images)
				}
				if len(edit.Attachments) != len(tt.images) {
					t.Errorf("expected the attachments to list the %d uploaded files, got %v", len(tt.images), edit.Attachments)
				}
				return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
			})
			state := core.Grinch{ID: "g", MessageID: "999", Phase: tt.phase, EndedAt: time.Now()}
			if err := (grinchDisplay{session, "400", classicFight(core.DefaultTuning().Grinch), nil, nil}).Update(t.Context(), state); err != nil || edits != 1 {
				t.Fatalf("Update = %v after %d edits", err, edits)
			}
		})
	}
}

// loggingGrinchStore fails every load and records when the store was used.
type loggingGrinchStore struct {
	*giftfixture.Store
	events *[]string
}

func (store loggingGrinchStore) LoadGrinch(context.Context, string) (core.Grinch, int64, error) {
	*store.events = append(*store.events, "store")
	return core.Grinch{}, 0, errors.New("database is busy")
}

func TestGrinchActionAcknowledgesBeforeWorkingAndReportsFailuresPrivately(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	var events []string
	b := &grinchBot{guild: "200", channel: "400",
		encounter: app.NewGrinchEncounter(loggingGrinchStore{&giftfixture.Store{}, &events}, "200", "scope", classicFight(core.DefaultTuning().Grinch),
			func() time.Time { return now }, func(int) int { return 0 })}
	s, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			var response discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Type != discordgo.InteractionResponseDeferredMessageUpdate || response.Data != nil {
				t.Errorf("expected a silent deferred acknowledgement, got %+v", response)
			}
			events = append(events, "ack")
			return fakeResponse(http.StatusNoContent, ""), nil
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/webhooks/"):
			var followup struct {
				Content string                 `json:"content"`
				Flags   discordgo.MessageFlags `json:"flags"`
			}
			if err := json.NewDecoder(r.Body).Decode(&followup); err != nil {
				t.Fatal(err)
			}
			if followup.Content != "Could not record your action. Try again." || followup.Flags != discordgo.MessageFlagsEphemeral {
				t.Errorf("follow-up = %+v, want a private failure message", followup)
			}
			events = append(events, "followup")
			return fakeResponse(http.StatusOK, `{"id":"1"}`), nil
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		return nil, nil
	})
	b.button(s, &discordgo.Interaction{
		ID: fmt.Sprint((now.UnixMilli() - 1420070400000) << 22), ChannelID: "400", Message: &discordgo.Message{ID: "999"},
		Member: &discordgo.Member{User: &discordgo.User{ID: "50"}},
		Data:   discordgo.MessageComponentInteractionData{CustomID: "grinch:fight:snowballs:0"},
	})
	if want := []string{"ack", "store", "followup"}; !slices.Equal(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

// Announcements from before the Start button have nothing to answer them: their buttons do nothing.
func TestGrinchIgnoresButtonsOfOlderAnnouncements(t *testing.T) {
	s, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		return nil, nil
	})
	b := &grinchBot{guild: "200", channel: "400",
		encounter: app.NewGrinchEncounter(&giftfixture.Store{}, "200", "scope", classicFight(core.DefaultTuning().Grinch), time.Now, func(int) int { return 0 })}
	for _, customID := range []string{"grinch:fight:registration:0", "grinch:fight:lobby:0", "grinch:fight:join:999:1"} {
		b.button(s, &discordgo.Interaction{
			ID: "300", ChannelID: "400", Message: &discordgo.Message{ID: "999", Flags: discordgo.MessageFlagsEphemeral},
			Member: &discordgo.Member{User: &discordgo.User{ID: "50"}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: customID},
		})
	}
}

// classicFight keeps the numbers these tests were written for: fixed damage, no wait between
// clicks, no counterattack, no last chance, and a one-minute fight clock.
func classicFight(rules core.GrinchRules) core.GrinchRules {
	rules.HitsPerPlayer, rules.MinHits, rules.ClickCooldown, rules.CounterattackPercent = 0, 0, 0, 0
	rules.LastChance = 0
	rules.FightDuration = time.Minute
	return rules
}
