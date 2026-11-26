package discord

import (
	"context"
	"encoding/json"
	"errors"
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

// spiritStore records when the auras are read and can fail the read.
type spiritStore struct {
	*giftfixture.Store
	calls *[]string
	err   error
}

func (store spiritStore) ServerAuras(ctx context.Context, server string) ([]core.Aura, error) {
	*store.calls = append(*store.calls, "store")
	if store.err != nil {
		return nil, store.err
	}
	return []core.Aura{30, -10}, nil
}

func TestSpiritIsRevealedByModeratorsOnly(t *testing.T) {
	for _, tt := range []struct {
		name        string
		permissions int64
		err         error
		calls       []string
		text        string
	}{
		{"moderator", discordgo.PermissionManageMessages, nil, []string{"ack", "store", "channel", "delete"},
			"😇 Good deeds: **30**\n😈 Bad deeds: **10**"},
		{"player", 0, nil, []string{"ack", "edit"}, "Only moderators can reveal the Christmas spirit."},
		{"store failure", discordgo.PermissionAdministrator, errors.New("disk gone"), []string{"ack", "store", "edit"},
			"Could not read the Christmas spirit. Try again."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var shown string
			service := app.NewService(spiritStore{&giftfixture.Store{}, &calls, tt.err}, time.Now, func(int) int { return 0 })
			session, err := discordgo.New("Bot fixture")
			if err != nil {
				t.Fatal(err)
			}
			session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/callback"):
					calls = append(calls, "ack")
					return fakeResponse(http.StatusNoContent, ""), nil
				case r.Method == http.MethodDelete:
					calls = append(calls, "delete")
					return fakeResponse(http.StatusNoContent, ""), nil
				case r.Method == http.MethodPost:
					calls = append(calls, "channel")
					var payload boxPayload
					if files := decodeMultipartPayload(t, r, &payload, "spirit.jpg"); files != 1 {
						t.Errorf("spirit pictures = %d", files)
					}
					shown, _ = payload.read()
				default:
					calls = append(calls, "edit")
					var edit struct{ Content string }
					if err := json.NewDecoder(r.Body).Decode(&edit); err != nil {
						t.Fatal(err)
					}
					shown = edit.Content
				}
				return fakeResponse(http.StatusOK, `{"id":"1"}`), nil
			})
			bot{guildID: "200", service: service}.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "300", AppID: "100", Token: "fixture", GuildID: "200", ChannelID: "400", Type: discordgo.InteractionApplicationCommand,
				Member: &discordgo.Member{User: &discordgo.User{ID: "mod"}, Permissions: tt.permissions},
				Data:   discordgo.ApplicationCommandInteractionData{CommandType: discordgo.ChatApplicationCommand, Name: "spirit"},
			}})
			if !slices.Equal(calls, tt.calls) || !strings.Contains(shown, tt.text) {
				t.Errorf("calls = %v, shown = %q; want %v with %q", calls, shown, tt.calls, tt.text)
			}
		})
	}
}
