package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func elfClick(b bot, session *discordgo.Session, roles []string) {
	b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "300", AppID: "100", Token: "token", GuildID: "200", ChannelID: "400", Type: discordgo.InteractionMessageComponent,
		Member:  &discordgo.Member{Roles: roles, User: &discordgo.User{ID: "alice"}},
		Message: &discordgo.Message{ID: "500", ChannelID: "400"},
		Data:    discordgo.MessageComponentInteractionData{CustomID: elfButtonID},
	}})
}

func elfSession(t *testing.T, events *[]string, replies *[]string, roleStatus int) *discordgo.Session {
	t.Helper()
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			var response discordgo.InteractionResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Type != discordgo.InteractionResponseDeferredChannelMessageWithSource ||
				response.Data == nil || response.Data.Flags != discordgo.MessageFlagsEphemeral {
				t.Errorf("expected a private deferred acknowledgement, got %+v", response)
			}
			*events = append(*events, "ack")
			return fakeResponse(http.StatusNoContent, ""), nil
		case r.Method == http.MethodGet && r.URL.Path == "/api/v9/guilds/200/roles":
			*events = append(*events, "roles")
			return fakeResponse(http.StatusOK, `[{"id":"900","name":"Elfs"}]`), nil
		case r.Method == http.MethodPut && r.URL.Path == "/api/v9/guilds/200/members/alice/roles/900":
			*events = append(*events, "add")
			return fakeResponse(roleStatus, ""), nil
		case strings.HasSuffix(r.URL.Path, "/messages/@original"):
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil || edit.Content == nil {
				t.Fatalf("reply edit = %+v, %v", edit, err)
			}
			*replies = append(*replies, *edit.Content)
			return fakeResponse(http.StatusOK, `{"id":"501"}`), nil
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		return nil, nil
	})
	return session
}

func TestElfButtonAcknowledgesFirstThenGivesTheElfsRoleOnce(t *testing.T) {
	var events, replies []string
	session := elfSession(t, &events, &replies, http.StatusNoContent)
	b := bot{guildID: "200"}
	elfClick(b, session, nil)
	if want := []string{"ack", "roles", "add"}; !slices.Equal(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if len(replies) != 1 || !strings.Contains(replies[0], "joined the elf army") || !strings.Contains(replies[0], "<@&900>") {
		t.Fatalf("replies = %q", replies)
	}

	// Clicking again does nothing but say so.
	events, replies = nil, nil
	elfClick(b, session, []string{"900"})
	if want := []string{"ack", "roles"}; !slices.Equal(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if len(replies) != 1 || !strings.Contains(replies[0], "already in the elf army") {
		t.Fatalf("replies = %q", replies)
	}
}

func TestElfButtonReportsAFailedRoleChangePrivately(t *testing.T) {
	var events, replies []string
	session := elfSession(t, &events, &replies, http.StatusForbidden)
	elfClick(bot{guildID: "200"}, session, nil)
	if want := []string{"ack", "roles", "add"}; !slices.Equal(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if len(replies) != 1 || !strings.Contains(replies[0], "Could not add you to the elf army") {
		t.Fatalf("replies = %q, want the failure to reach the player", replies)
	}
}

// When the acknowledgement call fails after Discord may have received it, the player must not be
// left with a "thinking…" message, and no game work may run.
func TestPrivateReplyClearsAnUncertainAcknowledgement(t *testing.T) {
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var events, replies []string
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/callback"):
			events = append(events, "ack")
			return nil, context.DeadlineExceeded
		case strings.HasSuffix(r.URL.Path, "/messages/@original"):
			var edit discordgo.WebhookEdit
			if err := json.NewDecoder(r.Body).Decode(&edit); err != nil || edit.Content == nil {
				t.Fatalf("reply edit = %+v, %v", edit, err)
			}
			events = append(events, "edit")
			replies = append(replies, *edit.Content)
			return fakeResponse(http.StatusOK, `{"id":"501"}`), nil
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		return nil, nil
	})
	worked := false
	respondPrivately(session, &discordgo.Interaction{ID: "300", Token: "token", AppID: "100", GuildID: "200",
		Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}}},
		func(context.Context, app.Member) string { worked = true; return "done" })
	if worked || !slices.Equal(events, []string{"ack", "edit"}) || len(replies) != 1 || !strings.Contains(replies[0], "slow to answer") {
		t.Fatalf("worked = %v, events = %q, replies = %q", worked, events, replies)
	}
}
