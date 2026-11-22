package discord

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

type grinchBot struct {
	guild        string
	channel      string
	encounter    *app.GrinchEncounter
	lastActivity atomic.Int64
}

func grinchCommand() *discordgo.ApplicationCommand {
	permission := int64(discordgo.PermissionManageMessages)
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: "grinch", Description: "Manage the Grinch fight",
		DefaultMemberPermissions: &permission,
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "summon", Description: "Summon the Grinch in the encounter channel"},
		},
	}
}

func (b bot) summonGrinch(s *discordgo.Session, i *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) != 1 || options[0] == nil || options[0].Name != "summon" ||
		options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return
	}
	respondPrivately(s, i, func(ctx context.Context, _ app.Member) string {
		if i.Member.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) == 0 {
			return "You need Manage Messages permission to summon the Grinch."
		}
		if b.grinch == nil {
			return "The Grinch encounter channel has not been configured."
		}
		if i.ChannelID != b.grinch.channel {
			return "Use this command in the configured Grinch channel."
		}
		display := grinchDisplay{s, b.grinch.channel}
		if _, err := b.grinch.encounter.Sync(ctx, display); err != nil {
			logInteractionError(i.ID, "reconcile Grinch encounter", err)
		}
		spawned, err := b.grinch.encounter.Spawn(ctx, time.Time{}, true, display)
		if err != nil {
			logInteractionError(i.ID, "summon the Grinch", err)
			return "Could not confirm the Grinch announcement."
		}
		if !spawned {
			return "A Grinch encounter is already in progress."
		}
		return "The Grinch is coming! Players can register with 🎅."
	})
}

func (b *grinchBot) observe(_ *discordgo.Session, event *discordgo.MessageCreate) {
	if event.GuildID != b.guild || event.ChannelID != b.channel || event.Author == nil ||
		event.Author.Bot || event.WebhookID != "" {
		return
	}
	when := event.Timestamp
	if when.IsZero() {
		when = time.Now()
	}
	next := when.UnixNano()
	for previous := b.lastActivity.Load(); next > previous; previous = b.lastActivity.Load() {
		if b.lastActivity.CompareAndSwap(previous, next) {
			return
		}
	}
}

func (b *grinchBot) run(ctx context.Context, s *discordgo.Session) {
	refresh := time.NewTicker(time.Second)
	defer refresh.Stop()
	roll := time.NewTicker(time.Minute)
	defer roll.Stop()
	display := grinchDisplay{s, b.channel}
	for {
		select {
		case <-ctx.Done():
			return
		case <-refresh.C:
			attempt, cancel := context.WithTimeout(ctx, 3*time.Second)
			_, err := b.encounter.Sync(attempt, display)
			cancel()
			if err != nil && ctx.Err() == nil {
				log.Printf("Update Grinch fight: %v", err)
			}
		case <-roll.C:
			attempt, cancel := context.WithTimeout(ctx, 4*time.Second)
			_, err := b.encounter.Spawn(attempt, time.Unix(0, b.lastActivity.Load()), false, display)
			cancel()
			if err != nil && ctx.Err() == nil {
				log.Printf("Announce Grinch fight: %v", err)
			}
		}
	}
}

func (b *grinchBot) button(s *discordgo.Session, i *discordgo.Interaction) {
	data, ok := i.Data.(discordgo.MessageComponentInteractionData)
	if !ok || i.ChannelID != b.channel || i.Message == nil {
		return
	}
	parts := strings.Split(data.CustomID, ":")
	if len(parts) != 4 || parts[0] != "grinch" {
		return
	}
	slot, err := strconv.Atoi(parts[3])
	if err != nil {
		return
	}
	at, err := discordgo.SnowflakeTimestamp(i.ID)
	if err != nil {
		at = time.Now()
	}
	if i.Member == nil || i.Member.User == nil || i.Member.User.ID == "" {
		return
	}
	playContext, stopPlay := context.WithTimeout(context.Background(), time.Second)
	state, reply, err := b.encounter.Play(playContext, parts[1], i.Message.ID, core.GrinchMove{
		Phase: core.GrinchPhase(parts[2]), Slot: slot, Member: i.Member.User.ID, RequestID: i.ID, At: at,
	})
	stopPlay()
	response := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}
	snowflake := state.Phase == core.GrinchDefense && parts[2] == string(core.GrinchDefense) &&
		slot >= 0 && slot < 8 && state.Board&(1<<slot) == 0
	if err != nil || !state.Accepted[i.ID] || snowflake || parts[2] == string(core.GrinchLobby) {
		if err != nil {
			logInteractionError(i.ID, "play Grinch action", err)
			reply = "Could not record your action. Try again."
		}
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: reply},
		}
	}
	ackContext, stopAck := context.WithTimeout(context.Background(), time.Second)
	defer stopAck()
	if err := s.InteractionRespond(i, response, discordgo.WithContext(ackContext)); err != nil {
		logInteractionError(i.ID, "acknowledge Grinch action", err)
	}
}

type grinchDisplay struct {
	session *discordgo.Session
	channel string
}

func (display grinchDisplay) Announce(ctx context.Context, state core.Grinch) (string, time.Time, error) {
	state.Phase = core.GrinchLobby
	state.StartedAt = time.Now()
	content, components := grinchMessage(state)
	message, err := display.session.ChannelMessageSendComplex(display.channel, &discordgo.MessageSend{
		Content: content, Components: components,
	}, discordgo.WithContext(ctx), discordgo.WithRetryOnRatelimit(false), discordgo.WithRestRetries(0))
	if err != nil {
		return "", time.Time{}, err
	}
	start := message.Timestamp
	if start.IsZero() {
		start = state.StartedAt
	}
	return message.ID, start, nil
}

func (display grinchDisplay) Update(ctx context.Context, state core.Grinch) error {
	content, components := grinchMessage(state)
	_, err := display.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID: state.MessageID, Channel: display.channel, Content: &content, Components: &components,
	}, discordgo.WithContext(ctx))
	var rest *discordgo.RESTError
	if errors.As(err, &rest) && rest.Message != nil && rest.Message.Code == 10008 {
		return nil // The original message was deleted.
	}
	return err
}

func grinchMessage(state core.Grinch) (string, []discordgo.MessageComponent) {
	content := fmt.Sprintf("🎄 **Grinch** · **%.2f/100 HP**", state.Health())
	if state.Phase == core.GrinchCancelled {
		return "The Grinch encounter was cancelled because no one joined.", []discordgo.MessageComponent{}
	}
	if state.Phase == core.GrinchLobby {
		return fmt.Sprintf("🎄 **The Grinch is coming!**\nJoin with 🎅 before <t:%d:R>. **%d joined**.",
				state.StartedAt.Add(core.GrinchJoinDuration).Unix(), len(state.Joined)),
			[]discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Join", Emoji: &discordgo.ComponentEmoji{Name: "🎅"}, Style: discordgo.SuccessButton,
					CustomID: "grinch:" + state.ID + ":lobby:0"},
			}}}
	}
	if !state.EndedAt.IsZero() {
		if state.Phase == core.GrinchWon {
			return content + "\n🎉 The group defeated the Grinch!", []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.Button{Label: "Final attack", CustomID: "grinch:ended", Disabled: true, Style: discordgo.PrimaryButton,
						Emoji: &discordgo.ComponentEmoji{Name: "⚔️"}},
				}},
			}
		}
		return content + "\n💨 The Grinch escaped!", []discordgo.MessageComponent{}
	}
	if deadline := state.Deadline(); !deadline.IsZero() {
		content += fmt.Sprintf("\nEscapes <t:%d:R>", deadline.Unix())
	}
	button := func(slot int, emoji, label string) discordgo.Button {
		return discordgo.Button{
			Label: label, Emoji: &discordgo.ComponentEmoji{Name: emoji}, Style: discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("grinch:%s:%s:%d", state.ID, state.Phase, slot),
		}
	}
	var rows []discordgo.MessageComponent
	switch state.Phase {
	case core.GrinchSnowballs:
		content += "\nThrow snowballs until the Grinch reaches **75 HP**. The one-minute timer starts at 75 HP."
		rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{button(0, "☃️", "Attack")}}}
	case core.GrinchDefense, core.GrinchAttack:
		on, off := "🎄", "❄️"
		content += "\nThe Grinch attacks! Choose a tree to defend."
		if state.Phase == core.GrinchAttack {
			on, off = "🧊", "🫯"
			content = fmt.Sprintf("🎄 **Grinch** · **%.2f/100 HP**\n“Something, something… ouch!”\nEscapes <t:%d:R>",
				state.Health(), state.Deadline().Unix())
		}
		for row := range 2 {
			columns := make([]discordgo.MessageComponent, 4)
			for column := range columns {
				slot := row*4 + column
				emoji := off
				if state.Board&(1<<slot) != 0 {
					emoji = on
				}
				columns[column] = button(slot, emoji, "")
			}
			rows = append(rows, discordgo.ActionsRow{Components: columns})
		}
	case core.GrinchFinalAttack:
		rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{button(0, "⚔️", "Final attack")}}}
	default:
		content += "\nThe snowball phase ended."
		rows = []discordgo.MessageComponent{}
	}
	return content, rows
}
