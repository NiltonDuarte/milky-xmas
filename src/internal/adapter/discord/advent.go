package discord

import (
	"context"
	_ "embed"
	"strconv"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/adventtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const adventCommandName = "advent"

func adventCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        adventCommandName,
		Description: "Open your Christmas Advent calendar",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "view", Description: "View your Advent doors"},
		},
	}
}

func (b bot) handleAdvent(session *discordgo.Session, interaction *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) != 1 || options[0] == nil || options[0].Type != discordgo.ApplicationCommandOptionSubCommand ||
		options[0].Name != "view" {
		return
	}
	respondPrivatelyWith(session, interaction, func(ctx context.Context, member app.Member) privateReply {
		calendar, err := b.service.ViewAdvent(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "load Advent calendar", err)
			return privateReply{Content: "Could not load your Advent calendar. Try again."}
		}
		return adventReply(adventtext.Header(calendar), calendar)
	})
}

const adventImageName = "advent.jpg"

//go:embed advent.jpg
var adventImage []byte

const adventButtonPrefix = "advent:"

func adventReply(text string, calendar core.AdventCalendar) privateReply {
	return privateReply{
		Content:    text,
		Image:      &privateImage{Name: adventImageName, Data: adventImage},
		Components: adventComponents(calendar),
	}
}

// adventComponents renders one button per door in rows of five, in the player's own order. An
// opened door is green and stays clickable to show its surprise again, a door ready to open is
// blue, and the others are disabled.
func adventComponents(calendar core.AdventCalendar) []discordgo.MessageComponent {
	const perRow = 5
	var rows []discordgo.MessageComponent
	var buttons []discordgo.MessageComponent
	for position, door := range calendar.Order {
		day := door + 1
		button := discordgo.Button{
			Label: strconv.Itoa(day), Style: discordgo.SecondaryButton, CustomID: adventButtonPrefix + strconv.Itoa(day),
		}
		switch calendar.Doors[door] {
		case core.DoorOpened:
			button.Style = discordgo.SuccessButton
		case core.DoorAvailable:
			button.Style = discordgo.PrimaryButton
		default:
			button.Disabled = true
		}
		buttons = append(buttons, button)
		if len(buttons) == perRow || position == core.AdventDays-1 {
			rows = append(rows, discordgo.ActionsRow{Components: buttons})
			buttons = nil
		}
	}
	return rows
}

// adventButton shows the surprise behind the clicked door, opening it first if needed, and
// redraws the calendar so the door turns green.
func (b bot) adventButton(session *discordgo.Session, interaction *discordgo.Interaction) {
	data, ok := interaction.Data.(discordgo.MessageComponentInteractionData)
	if !ok || interaction.Member == nil || interaction.Member.User == nil || interaction.Member.User.ID == "" {
		return
	}
	day, err := strconv.Atoi(strings.TrimPrefix(data.CustomID, adventButtonPrefix))
	if err != nil || day < 1 || day > core.AdventDays {
		return
	}
	ackContext, stopAck := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err = session.InteractionRespond(interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate},
		discordgo.WithContext(ackContext))
	stopAck()
	if err != nil {
		logInteractionError(interaction.ID, "acknowledge Advent door", err)
		return
	}

	member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var calendar core.AdventCalendar
	opening, err := b.service.OpenAdvent(ctx, member, day)
	if err == nil && opening.Status == core.DoorOpened {
		calendar, err = b.service.ViewAdvent(ctx, member)
	}
	if err != nil {
		logInteractionError(interaction.ID, "open Advent door", err)
		// Leave the calendar and its buttons as they were so the player can try again.
		privateFollowup(session, interaction, "Could not open the Advent door. Try again.")
		return
	}
	if opening.Status != core.DoorOpened {
		privateFollowup(session, interaction, adventtext.Opening(day, opening))
		return
	}
	// An embed holds a long recipe; a plain message stops at 2000 characters.
	if _, err := session.FollowupMessageCreate(interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{{Description: adventtext.Opening(day, opening)}},
		Flags:  discordgo.MessageFlagsEphemeral,
	}, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "show Advent surprise", err)
	}
	if err := showReply(session, interaction, adventReply(adventtext.Header(calendar), calendar)); err != nil {
		logInteractionError(interaction.ID, "show Advent calendar", err)
	}
}
