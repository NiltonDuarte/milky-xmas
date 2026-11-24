package discord

import (
	"encoding/json"
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
					Embeds     []struct{ Image struct{ URL string } }
				}
				if files := decodeMultipartPayload(t, r, &message, "grinch.jpg"); files != 1 ||
					len(message.Embeds) != 1 || message.Embeds[0].Image.URL != "attachment://grinch.jpg" {
					t.Errorf("announcement image = %+v with %d files", message.Embeds, files)
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
	id := func(index int64) string {
		return fmt.Sprint(((now.UnixMilli() - 1420070400000) << 22) + index)
	}
	opener := "grinch:" + fight.ID + ":registration:0"
	personal := "grinch:" + fight.ID + ":join:999:"
	for index, step := range []struct {
		user, customID string
		privatePanel   bool
		want           discordgo.InteractionResponseType
	}{
		{"50", opener, false, discordgo.InteractionResponseChannelMessageWithSource},
		{"51", opener, false, discordgo.InteractionResponseChannelMessageWithSource},
		{"50", personal + id(0), true, discordgo.InteractionResponseUpdateMessage},
		{"50", opener, false, discordgo.InteractionResponseDeferredMessageUpdate},
		{"51", personal + id(1), true, discordgo.InteractionResponseUpdateMessage},
		{"51", opener, false, discordgo.InteractionResponseDeferredMessageUpdate},
	} {
		calls := 0
		s.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			var response struct {
				Type discordgo.InteractionResponseType `json:"type"`
				Data *struct {
					Content    string                 `json:"content"`
					Flags      discordgo.MessageFlags `json:"flags"`
					Components []json.RawMessage      `json:"components"`
				} `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if !strings.HasSuffix(r.URL.Path, "/callback") || response.Type != step.want {
				t.Fatalf("expected callback type %d, got %+v", step.want, response)
			}
			switch step.want {
			case discordgo.InteractionResponseChannelMessageWithSource:
				if response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral ||
					len(response.Data.Components) != 1 || !strings.Contains(string(response.Data.Components[0]), personal+id(int64(index))) {
					t.Errorf("expected private Join panel, got %+v", response.Data)
				}
			case discordgo.InteractionResponseUpdateMessage:
				if response.Data == nil || response.Data.Content != grinchRegistrationTitle ||
					response.Data.Components == nil || len(response.Data.Components) != 0 {
					t.Errorf("expected the private Join button removed, got %+v", response.Data)
				}
			case discordgo.InteractionResponseDeferredMessageUpdate:
				if response.Data != nil {
					t.Errorf("expected silent reopen for a registered player, got %+v", response)
				}
			}
			return fakeResponse(http.StatusNoContent, ""), nil
		})
		message := &discordgo.Message{ID: "999"}
		if step.privatePanel {
			message = &discordgo.Message{ID: "panel", Flags: discordgo.MessageFlagsEphemeral}
		}
		b.button(s, &discordgo.Interaction{
			ID: id(int64(index)), ChannelID: "400", Message: message,
			Member: &discordgo.Member{User: &discordgo.User{ID: step.user}},
			Data:   discordgo.MessageComponentInteractionData{CustomID: step.customID},
		})
		if calls != 1 {
			t.Fatalf("registration step %d sent %d responses", index, calls)
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

func TestGrinchHealthDisplay(t *testing.T) {
	for _, state := range []core.Grinch{
		{Phase: core.GrinchLobby, HP: 10000},
		{Phase: core.GrinchAttack, HP: 4500},
		{Phase: core.GrinchWon, HP: 0, EndedAt: time.Now()},
	} {
		content, _ := grinchMessage(state)
		line, _, _ := strings.Cut(content, "\n")
		want := fmt.Sprintf("👽 Grinch · %.2f/100%% HP", state.Health())
		if line != want {
			t.Errorf("health line for %s = %q, want %q", state.Phase, line, want)
		}
	}
}

func TestGrinchReshuffleEditsOriginalMessage(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	b := &grinchBot{guild: "200", channel: "400",
		encounter: app.NewGrinchEncounter(store, "200", "scope", func() time.Time { return now }, func(int) int { return 0 })}
	fight := core.Grinch{ID: "fight", MessageID: "999", HP: 500, Phase: core.GrinchFinalAttack, FinalRound: 1,
		StartedAt: now, TimedAt: now, Joined: map[string]bool{"50": true}, Accepted: make(map[string]bool)}
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
				Content    string            `json:"content"`
				Components []json.RawMessage `json:"components"`
			}
			if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(update.Content, "👽 Grinch · 5.00/100% HP\n") ||
				len(update.Components) != 2 || !strings.Contains(string(update.Components[0]), "grinch:fight:final:4:2") {
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
	if _, err := b.encounter.Sync(t.Context(), grinchDisplay{s, b.channel}); err != nil {
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
		{"defeat", core.GrinchDefeat, []string{"grinch.jpg", "maxpee.jpg"}},
		{"escaped", core.GrinchEscaped, []string{"grinch.jpg", "maxpee.jpg"}},
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
					Embeds      []struct{ Image struct{ URL string } }
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
				var urls []string
				for _, embed := range edit.Embeds {
					urls = append(urls, strings.TrimPrefix(embed.Image.URL, "attachment://"))
				}
				if !slices.Equal(urls, tt.images) || !slices.Equal(files, tt.images) {
					t.Errorf("final message embeds %v and files %v, want %v", urls, files, tt.images)
				}
				if len(edit.Attachments) != len(tt.images) {
					t.Errorf("expected the attachments to list the %d uploaded files, got %v", len(tt.images), edit.Attachments)
				}
				return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
			})
			state := core.Grinch{ID: "g", MessageID: "999", Phase: tt.phase, EndedAt: time.Now()}
			if err := (grinchDisplay{session, "400"}).Update(t.Context(), state); err != nil || edits != 1 {
				t.Fatalf("Update = %v after %d edits", err, edits)
			}
		})
	}
}
