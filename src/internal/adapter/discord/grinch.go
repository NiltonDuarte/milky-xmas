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

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/grinchtext"
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
		grinch := b.grinches.get(i.GuildID, i.ChannelID)
		if grinch == nil {
			return "Use this command in a configured Grinch channel."
		}
		display := grinchDisplay{s, grinch.channel}
		if _, err := grinch.encounter.Sync(ctx, display); err != nil {
			logInteractionError(i.ID, "reconcile Grinch encounter", err)
		}
		spawned, err := grinch.encounter.Spawn(ctx, time.Time{}, true, display)
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
	if (len(parts) != 4 && len(parts) != 5) || parts[0] != "grinch" {
		return
	}
	slot, err := strconv.Atoi(parts[3])
	if err != nil {
		return
	}
	var round int
	if len(parts) == 5 {
		round, err = strconv.Atoi(parts[4])
		if err != nil {
			return
		}
	}
	at, err := discordgo.SnowflakeTimestamp(i.ID)
	if err != nil {
		at = time.Now()
	}
	if i.Member == nil || i.Member.User == nil || i.Member.User.ID == "" {
		return
	}
	playContext, stopPlay := context.WithTimeout(context.Background(), time.Second)
	_, _, err = b.encounter.Play(playContext, parts[1], i.Message.ID, core.GrinchMove{
		Phase: core.GrinchPhase(parts[2]), Slot: slot, DefenseRound: round, FinalRound: round, Member: i.Member.User.ID, RequestID: i.ID, At: at,
	})
	stopPlay()
	response := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}
	if err != nil {
		logInteractionError(i.ID, "play Grinch action", err)
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: "Could not record your action. Try again."},
		}
	}
	ackContext, stopAck := context.WithTimeout(context.Background(), 1750*time.Millisecond)
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
	edit := &discordgo.MessageEdit{
		ID: state.MessageID, Channel: display.channel, Content: &content, Components: &components,
		AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}},
	}
	var overflow bool
	if !state.EndedAt.IsZero() && state.Phase != core.GrinchCancelled {
		var extra string
		content, extra = grinchLeaderboards(content, state.Leaderboard(), state.GiftLeaderboard())
		overflow = extra != ""
		if overflow {
			edit.Attachments = &[]*discordgo.MessageAttachment{}
			edit.Files = []*discordgo.File{{Name: "grinch-leaderboard.txt", Reader: strings.NewReader(extra)}}
		}
	}
	_, err := display.session.ChannelMessageEditComplex(edit, discordgo.WithContext(ctx))
	var denied *discordgo.RESTError
	if overflow && errors.As(err, &denied) && denied.Message != nil && denied.Message.Code == 50013 {
		content = strings.TrimSuffix(content, "\nFull standings are attached.") + "\nFull standings require permission to attach files."
		edit.Attachments, edit.Files = nil, nil
		_, err = display.session.ChannelMessageEditComplex(edit, discordgo.WithContext(ctx))
	}
	var rest *discordgo.RESTError
	if errors.As(err, &rest) && rest.Message != nil && rest.Message.Code == 10008 {
		return nil // The original message was deleted.
	}
	return err
}

func grinchLeaderboards(content string, damage, gifts []core.GrinchScore) (string, string) {
	if len(damage) == 0 && len(gifts) == 0 {
		return content, ""
	}
	var standings strings.Builder
	overflow := false
	for _, section := range []struct {
		title  string
		format string
		scores []core.GrinchScore
	}{
		{"Damage leaderboard", "%.2f", damage},
		{"Gift leaderboard", "%.0f", gifts},
	} {
		if len(section.scores) == 0 {
			continue
		}
		content += "\n\n**" + section.title + "**"
		fmt.Fprintf(&standings, "%s\n", section.title)
		limit := min(len(content)+820, 1900)
		for index, score := range section.scores {
			points := fmt.Sprintf(section.format, score.Points)
			line := fmt.Sprintf("\n%d. <@%s> — %s pts", index+1, score.Member, points)
			fmt.Fprintf(&standings, "%d. Discord user %s — %s pts\n", index+1, score.Member, points)
			if len(content)+len(line) <= limit {
				content += line
			} else {
				overflow = true
			}
		}
		fmt.Fprintln(&standings)
	}
	if overflow {
		return content + "\nFull standings are attached.", standings.String()
	}
	return content, ""
}

func grinchMessage(state core.Grinch) (string, []discordgo.MessageComponent) {
	content := grinchtext.Health(state.Health())
	if state.Phase == core.GrinchCancelled {
		return "The Grinch encounter was cancelled because no one joined.", []discordgo.MessageComponent{}
	}
	join := discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{Label: "Join", Emoji: &discordgo.ComponentEmoji{Name: "🎅"}, Style: discordgo.SuccessButton,
			CustomID: "grinch:" + state.ID + ":lobby:0"},
	}}
	if state.Phase == core.GrinchLobby {
		return fmt.Sprintf("%s\n🎄 **The Grinch is coming!**\nJoin with 🎅 before <t:%d:R>. **%d/3 to start early**.",
				content, state.StartedAt.Add(core.GrinchJoinDuration).Unix(), len(state.Joined)),
			[]discordgo.MessageComponent{join}
	}
	if !state.EndedAt.IsZero() {
		if state.Phase == core.GrinchWon {
			return content + "\n🎉 The group defeated the Grinch!", []discordgo.MessageComponent{}
		}
		if state.Phase == core.GrinchDefeat {
			return content + "\n💥 The Grinch defeated all registered players!", []discordgo.MessageComponent{}
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
			on, off = "🧊", "☃️"
			content = grinchtext.Health(state.Health()) + fmt.Sprintf("\n“Something, something… ouch!”\nEscapes <t:%d:R>", state.Deadline().Unix())
		}
		for row := range 2 {
			columns := make([]discordgo.MessageComponent, 4)
			for column := range columns {
				slot := row*4 + column
				emoji := off
				if state.Board&(1<<slot) != 0 {
					emoji = on
				}
				tile := button(slot, emoji, "")
				if state.Phase == core.GrinchDefense {
					tile.CustomID += ":" + strconv.Itoa(state.DefenseRound)
				}
				columns[column] = tile
			}
			rows = append(rows, discordgo.ActionsRow{Components: columns})
		}
	case core.GrinchFinalAttack:
		content += "\n⚔️ Attack for 2 damage points, pick 🎁 for 1–5 gift points, or pick ❄️ for 1 Snowball."
		if state.FinalRound == 0 {
			rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{button(0, "⚔️", "Attack"), button(core.GrinchSnowballSlot, "❄️", "Snowball")}}}
			break
		}
		tiles := make([]discordgo.MessageComponent, 0, core.GrinchSnowballSlot+1)
		for slot := range core.GrinchSnowballSlot {
			emoji := "⚔️"
			if state.FinalBoard&(1<<slot) != 0 {
				emoji = "🎁"
			}
			tile := button(slot, emoji, "")
			tile.CustomID += ":" + strconv.Itoa(state.FinalRound)
			tiles = append(tiles, tile)
		}
		snowball := button(core.GrinchSnowballSlot, "❄️", "")
		snowball.CustomID += ":" + strconv.Itoa(state.FinalRound)
		tiles = append(tiles, snowball)
		rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: tiles}}
	default:
		content += "\nThe snowball phase ended."
		rows = []discordgo.MessageComponent{}
	}
	if state.Joined != nil {
		content += fmt.Sprintf("\n🎅 **%d joined**", len(state.Joined))
		rows = append(rows, join)
	}
	return content, rows
}
