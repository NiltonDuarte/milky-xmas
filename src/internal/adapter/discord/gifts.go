package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const dailyCommandName = "daily"
const claimSubcommand = "claim"

func dailyCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        dailyCommandName,
		Description: "Collect Christmas gifts",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: claimSubcommand, Description: "Claim your daily gift"},
		},
	}
}

func (b bot) handleDailyClaim(session *discordgo.Session, interaction *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) != 1 || options[0] == nil || options[0].Name != claimSubcommand ||
		options[0].Type != discordgo.ApplicationCommandOptionSubCommand || len(options[0].Options) != 0 {
		return
	}

	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		claim, err := b.service.ClaimGift(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "record gift claim", err)
			return "Could not record your gift claim. Try again."
		}
		if claim.Collected {
			return fmt.Sprintf("🎁 You collected one gift! Total gifts: %d.", claim.Total)
		}
		// Discord timestamps use whole seconds. Round up to avoid showing an early retry time.
		next := claim.NextClaim.Add(time.Second - time.Nanosecond).Unix()
		return fmt.Sprintf("Your next gift is available <t:%d:F> (<t:%d:R>).", next, next)
	})
}
