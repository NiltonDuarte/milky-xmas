package discord

import (
	"context"
	"log"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const (
	statsViewID  = "stats:view"
	statsBoardID = "stats:board"
)

func statsCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: gifttext.StatsCommand, Description: "Server stats and leaderboards",
	}
}

// statsColors is the side colour of the panel for each leaderboard. Anything else is the galaxy.
var statsColors = map[string]int{
	"bad": 0xE74C3C, "good": 0x2ECC71, "throwers": 0x3498DB, "gifters": 0xE91E63, "plants": 0x1ABC9C,
}

const statsGalaxyColor = 0x7B5CD6

// statsMenu is a drop-down with the current choice selected.
func statsMenu(id, placeholder string, choices []gifttext.StatsChoice, selected string) discordgo.ActionsRow {
	options := make([]discordgo.SelectMenuOption, len(choices))
	for i, choice := range choices {
		options[i] = discordgo.SelectMenuOption{
			Label: choice.Label, Value: choice.Value, Default: choice.Value == selected,
			Emoji: &discordgo.ComponentEmoji{Name: choice.Emoji},
		}
	}
	return discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.SelectMenu{CustomID: id, Placeholder: placeholder, Options: options},
	}}
}

// statsReply is the whole private panel: both drop-downs and, below them, the chosen information.
func (b bot) statsReply(blocks []string, view, board string) privateReply {
	menus := []discordgo.MessageComponent{
		statsMenu(statsViewID, "Server stats or leaderboards?", gifttext.StatsViews, view),
		statsMenu(statsBoardID, "Pick a leaderboard", gifttext.StatsPanels, board),
	}
	color, ok := statsColors[board]
	if !ok {
		color = statsGalaxyColor
	}
	return privateReply{Blocks: blocks, Components: menus, ComponentsFirst: true, Accent: color}
}

// handleStats opens the panel on the leaderboards, with the server's totals until one is picked.
func (b bot) handleStats(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivatelyWith(session, interaction, func(context.Context, app.Member) privateReply {
		return b.statsReply(b.avatars.blocks(b.statsBlocks(interaction, "")), gifttext.StatsBoardsView, "")
	})
}

// slowStatsChoice is how long a choice may take before its timing is logged.
const slowStatsChoice = 1500 * time.Millisecond

// statsSelect redraws the panel after a choice in one of its drop-downs.
func (b bot) statsSelect(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.MessageComponentInteractionData) {
	if len(data.Values) != 1 {
		return
	}
	view, board := gifttext.StatsBoardsView, ""
	switch {
	case data.CustomID == statsBoardID:
		board = data.Values[0]
	case data.CustomID == statsViewID && data.Values[0] == gifttext.StatsServerView:
		view, board = gifttext.StatsServerView, gifttext.StatsServerView
	case data.CustomID != statsViewID || data.Values[0] != gifttext.StatsBoardsView:
		return
	}

	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}, discordgo.WithContext(ctx))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "answer stats choice", err)
		return
	}
	answered := time.Since(started)
	showResult(session, interaction, b.statsReply(b.avatars.blocks(b.statsBlocks(interaction, board)), view, board))
	if total := time.Since(started); total > slowStatsChoice {
		log.Printf("Slow stats choice %s %q: answered in %s, shown in %s", view, board, answered, total)
	}
}

// statsBlocks is the information below the drop-downs. No pick yet shows the server's totals.
func (b bot) statsBlocks(interaction *discordgo.Interaction, board string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stats, err := b.service.Stats(ctx, interaction.GuildID, b.tree)
	if err != nil {
		logInteractionError(interaction.ID, "read the statistics", err)
		return []string{gifttext.StatsUnavailable}
	}
	if board == "" || board == gifttext.StatsServerView {
		return gifttext.ServerStats(stats)
	}
	blocks, ok := gifttext.Leaderboard(board, stats, mentionUser)
	if !ok {
		return []string{gifttext.StatsUnavailable}
	}
	return blocks
}
