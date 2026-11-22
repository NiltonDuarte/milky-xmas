package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/adventtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const adventCommandName = "advent"

func adventCommand() *discordgo.ApplicationCommand {
	minDay := float64(1)
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        adventCommandName,
		Description: "Open your Christmas Advent calendar",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "view", Description: "View your Advent doors"},
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "open", Description: "Open an Advent door",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "day", Description: "Door number", Required: true, MinValue: &minDay, MaxValue: core.AdventDays},
				},
			},
		},
	}
}

func (b bot) handleAdvent(session *discordgo.Session, interaction *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) != 1 || options[0] == nil || options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return
	}
	sub := options[0]
	switch sub.Name {
	case "view":
		respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
			calendar, err := b.service.ViewAdvent(ctx, member)
			if err != nil {
				logInteractionError(interaction.ID, "load Advent calendar", err)
				return "Could not load your Advent calendar. Try again."
			}
			return adventtext.Calendar(calendar)
		})
	case "open":
		day := adventDay(sub.Options)
		respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
			opening, err := b.service.OpenAdvent(ctx, member, day)
			if err != nil {
				logInteractionError(interaction.ID, "open Advent door", err)
				return "Could not open the Advent door. Try again."
			}
			return adventtext.Opening(day, opening)
		})
	}
}

func adventDay(options []*discordgo.ApplicationCommandInteractionDataOption) int {
	if len(options) != 1 || options[0] == nil || options[0].Name != "day" || options[0].Type != discordgo.ApplicationCommandOptionInteger {
		return 0
	}
	// DiscordGo decodes numeric option values as float64.
	value, ok := options[0].Value.(float64)
	if !ok || value < 1 || value > core.AdventDays || value != float64(int(value)) {
		return 0
	}
	return int(value)
}
