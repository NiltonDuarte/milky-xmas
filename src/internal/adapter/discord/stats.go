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

// statsReply is the whole private panel for the chosen view and leaderboard.
func (b bot) statsReply(blocks []string, view, board string) privateReply {
	menus := []discordgo.MessageComponent{statsMenu(statsViewID, "Server stats or leaderboards?", gifttext.StatsViews, view)}
	if view == gifttext.StatsBoardsView {
		menus = append(menus, statsMenu(statsBoardID, "Pick a leaderboard", gifttext.StatsBoards, board))
	}
	return privateReply{Blocks: blocks, Components: menus, ComponentsFirst: true}
}

func (b bot) handleStats(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivatelyWith(session, interaction, func(context.Context, app.Member) privateReply {
		return b.statsReply([]string{"## 📊 Event stats\nChoose what you want to see."}, "", "")
	})
}

// slowStatsChoice is how long a choice may take before its timing is logged.
const slowStatsChoice = 1500 * time.Millisecond

// statsSelect redraws the panel after a choice in one of its drop-downs. Opening the leaderboards
// reads nothing, so it answers in one request. The other choices are acknowledged first.
func (b bot) statsSelect(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.MessageComponentInteractionData) {
	if len(data.Values) != 1 {
		return
	}
	view, board := data.Values[0], ""
	switch {
	case data.CustomID == statsBoardID:
		view, board = gifttext.StatsBoardsView, data.Values[0]
	case data.CustomID != statsViewID || (view != gifttext.StatsServerView && view != gifttext.StatsBoardsView):
		return
	}

	started := time.Now()
	response := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}
	prompt := view == gifttext.StatsBoardsView && board == ""
	if prompt {
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags:           discordgo.MessageFlagsIsComponentsV2,
				Components:      boxedComponents(b.statsReply([]string{"## 🏆 Leaderboards\nPick a leaderboard above."}, view, "")),
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err := session.InteractionRespond(interaction, response, discordgo.WithContext(ctx))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "answer stats choice", err)
		return
	}
	answered := time.Since(started)
	if !prompt {
		showResult(session, interaction, b.statsReply(b.statsBlocks(interaction, view, board), view, board))
	}
	if total := time.Since(started); total > slowStatsChoice {
		log.Printf("Slow stats choice %s %q: answered in %s, shown in %s", view, board, answered, total)
	}
}

func (b bot) statsBlocks(interaction *discordgo.Interaction, view, board string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stats, err := b.service.Stats(ctx, interaction.GuildID, b.tree)
	if err != nil {
		logInteractionError(interaction.ID, "read the statistics", err)
		return []string{gifttext.StatsUnavailable}
	}
	if view == gifttext.StatsServerView {
		return []string{gifttext.ServerStats(stats)}
	}
	blocks, ok := gifttext.Leaderboard(board, stats, mentionUser)
	if !ok {
		return []string{gifttext.StatsUnavailable}
	}
	return blocks
}
