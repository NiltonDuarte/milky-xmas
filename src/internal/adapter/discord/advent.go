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
		respondPrivatelyWith(session, interaction, func(ctx context.Context, member app.Member) privateReply {
			calendar, err := b.service.ViewAdvent(ctx, member)
			if err != nil {
				logInteractionError(interaction.ID, "load Advent calendar", err)
				return privateReply{Content: "Could not load your Advent calendar. Try again."}
			}
			return adventReply(adventtext.Header(calendar), calendar)
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

// adventComponents renders one button per door in rows of five. A door that was opened shows
// ❌ and is disabled, and doors that cannot be opened today are disabled too.
func adventComponents(calendar core.AdventCalendar) []discordgo.MessageComponent {
	const perRow = 5
	var rows []discordgo.MessageComponent
	var buttons []discordgo.MessageComponent
	for i, status := range calendar.Doors {
		day := i + 1
		button := discordgo.Button{
			Style: discordgo.SecondaryButton, CustomID: adventButtonPrefix + strconv.Itoa(day),
			Disabled: status != core.DoorAvailable,
		}
		if status == core.DoorOpened {
			button.Emoji = &discordgo.ComponentEmoji{Name: "❌"}
		} else {
			button.Label = strconv.Itoa(day)
			if status == core.DoorAvailable {
				button.Style = discordgo.PrimaryButton
			}
		}
		buttons = append(buttons, button)
		if len(buttons) == perRow || day == core.AdventDays {
			rows = append(rows, discordgo.ActionsRow{Components: buttons})
			buttons = nil
		}
	}
	return rows
}

// adventButton opens the door a player clicked and redraws their calendar in place.
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
	var text string
	var calendar core.AdventCalendar
	opening, err := b.service.OpenAdvent(ctx, member, day)
	if err == nil {
		calendar, err = b.service.ViewAdvent(ctx, member)
	}
	if err != nil {
		logInteractionError(interaction.ID, "open Advent door", err)
		text = "Could not open the Advent door. Try again."
		// Keep the buttons as they were so the player can try again.
		if _, err := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{Content: &text}, discordgo.WithContext(ctx)); err != nil {
			logInteractionError(interaction.ID, "show Advent error", err)
		}
		return
	}
	reply := adventReply(adventtext.Header(calendar)+"\n"+adventtext.Opening(day, opening), calendar)
	edit := privateEdit(reply)
	if _, err := session.InteractionResponseEdit(interaction, edit, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "show Advent calendar", err)
	}
}
