package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func TestRegisterCommands(t *testing.T) {
	session, err := discordgo.New("Bot fake-token")
	if err != nil {
		t.Fatal(err)
	}
	var requests []string
	var commands []string
	session.Client.Transport = transportFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request.Method+" "+request.URL.Path)
		switch request.URL.Path {
		case "/api/v9/users/@me":
			return fakeResponse(http.StatusOK, `{"id":"100"}`), nil
		case "/api/v9/applications/100/guilds/200/commands":
			if request.Method == http.MethodGet {
				return fakeResponse(http.StatusOK, `[{"id":"900","name":"socks"},{"id":"300","name":"hello"}]`), nil
			}
			var command discordgo.ApplicationCommand
			if err := json.NewDecoder(request.Body).Decode(&command); err != nil {
				t.Fatal(err)
			}
			commands = append(commands, command.Name)
			if command.Type != discordgo.ChatApplicationCommand ||
				((command.Name == "hello") && len(command.Options) != 0) ||
				(command.Name == "daily" && (len(command.Options) != 1 || command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand || command.Options[0].Name != "claim")) {
				t.Errorf("unexpected registered command: %+v", command)
			}
			if command.Name == "grinch" {
				if command.DefaultMemberPermissions == nil || *command.DefaultMemberPermissions != discordgo.PermissionManageMessages ||
					len(command.Options) != 1 || command.Options[0].Name != "summon" ||
					command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
					t.Errorf("unexpected moderator command: %+v", command)
				}
			}
			if command.Name == "advent" {
				if len(command.Options) != 2 || command.Options[0].Name != "view" || command.Options[1].Name != "open" ||
					command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand ||
					command.Options[1].Type != discordgo.ApplicationCommandOptionSubCommand ||
					len(command.Options[1].Options) != 1 {
					t.Fatalf("unexpected Advent subcommands: %+v", command.Options)
				}
				day := command.Options[1].Options[0]
				if day.Name != "day" || day.Type != discordgo.ApplicationCommandOptionInteger || !day.Required ||
					day.MinValue == nil || *day.MinValue != 1 || day.MaxValue != 24 {
					t.Errorf("unexpected day option: %+v", day)
				}
			}
			switch command.Name {
			case "send":
				checkSendRegistration(t, command)
			case "setup":
				if command.DefaultMemberPermissions == nil || *command.DefaultMemberPermissions != discordgo.PermissionManageGuild ||
					len(command.Options) != 2 || command.Options[0].Name != "view" || command.Options[1].Name != "channel" {
					t.Errorf("unexpected setup registration: %+v", command)
				}
			case "thrown":
				if len(command.Options) != 1 || command.Options[0].Name != "snowball" ||
					command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand || len(command.Options[0].Options) != 2 {
					t.Fatalf("unexpected snowball subcommand: %+v", command)
				}
				recipient := command.Options[0].Options[0]
				if recipient.Name != "recipient" || recipient.Type != discordgo.ApplicationCommandOptionUser || !recipient.Required {
					t.Errorf("unexpected snowball recipient option: %+v", recipient)
				}
				kind := command.Options[0].Options[1]
				if kind.Name != "type" || kind.Required || !kind.Autocomplete || kind.Type != discordgo.ApplicationCommandOptionString {
					t.Errorf("unexpected snowball type option: %+v", kind)
				}
			case "look":
				if len(command.Options) != 1 || command.Options[0].Name != "in-your-stocking" ||
					command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand || len(command.Options[0].Options) != 0 {
					t.Errorf("unexpected look command: %+v", command)
				}
			case "aura":
				if len(command.Options) != 1 || command.Options[0].Name != "check" ||
					command.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand || len(command.Options[0].Options) != 0 {
					t.Errorf("unexpected aura command: %+v", command)
				}
			}
			return fakeResponse(http.StatusCreated, `{"id":"300","name":"hello"}`), nil
		case "/api/v9/applications/100/guilds/200/commands/900":
			return fakeResponse(http.StatusNoContent, ""), nil
		default:
			t.Fatalf("unexpected request to %s", request.URL)
			return nil, nil
		}
	})

	if err := (bot{guildID: "200"}).registerCommands(t.Context(), session); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(commands, ","); got != "hello,daily,send,give,aura,look,advent,grinch,thrown,water,helpmemilky,setup" {
		t.Errorf("registered commands = %q", got)
	}
	want := "GET /api/v9/users/@me" + strings.Repeat("\nPOST /api/v9/applications/100/guilds/200/commands", len(commands)) +
		"\nGET /api/v9/applications/100/guilds/200/commands\nDELETE /api/v9/applications/100/guilds/200/commands/900"
	if got := strings.Join(requests, "\n"); got != want {
		t.Errorf("registration requests = %q, want %q", got, want)
	}
}

func TestHandleInteraction(t *testing.T) {
	for _, tt := range []struct {
		name, guild, command string
		interactionType      discordgo.InteractionType
		commandType          discordgo.ApplicationCommandType
		wantCalls            int
	}{
		{"hello", "200", "hello", discordgo.InteractionApplicationCommand, discordgo.ChatApplicationCommand, 1},
		{"other guild", "201", "hello", discordgo.InteractionApplicationCommand, discordgo.ChatApplicationCommand, 0},
		{"unknown command", "200", "unknown", discordgo.InteractionApplicationCommand, discordgo.ChatApplicationCommand, 0},
		{"non-slash command", "200", "hello", discordgo.InteractionApplicationCommand, discordgo.UserApplicationCommand, 0},
		{"non-command", "200", "hello", discordgo.InteractionMessageComponent, discordgo.ChatApplicationCommand, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			session, err := discordgo.New("Bot fake-token")
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			session.Client.Transport = transportFunc(func(request *http.Request) (*http.Response, error) {
				calls++
				if request.Method != http.MethodPost || request.URL.Path != "/api/v9/interactions/300/interaction-token/callback" {
					t.Errorf("unexpected acknowledgement request: %s %s", request.Method, request.URL.Path)
				}
				var response discordgo.InteractionResponse
				if err := json.NewDecoder(request.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Type != discordgo.InteractionResponseChannelMessageWithSource ||
					response.Data == nil || response.Data.Content != "Hello, world!" || response.Data.Flags != 0 {
					t.Errorf("unexpected acknowledgement type %v and data %+v", response.Type, response.Data)
				}
				return fakeResponse(http.StatusNoContent, ""), nil
			})

			b := bot{guildID: "200", service: app.Service{}}
			b.handleInteraction(session, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID:      "300",
				Token:   "interaction-token",
				GuildID: tt.guild,
				Type:    tt.interactionType,
				Data:    discordgo.ApplicationCommandInteractionData{Name: tt.command, CommandType: tt.commandType},
			}})
			if calls != tt.wantCalls {
				t.Errorf("acknowledgements = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func fakeResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
