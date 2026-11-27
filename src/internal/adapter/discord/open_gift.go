package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func openCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: gifttext.OpenCommand, Description: "Open a gift",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Description: "Open one of your gifts for a surprise",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type: discordgo.ApplicationCommandOptionString, Name: "rarity", Description: "Gift rarity", Required: true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Normal", Value: string(core.NormalGift)},
							{Name: "Rare", Value: string(core.RareGift)},
							{Name: "Epic", Value: string(core.EpicGift)},
						},
					},
				},
			},
		},
	}
}

// handleOpenGift acknowledges privately first, so only the player sees the surprise.
func (b bot) handleOpenGift(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.ApplicationCommandInteractionData) {
	sub := commandSubcommand(data.Options, "gift")
	if sub == nil {
		return
	}
	var rarity core.GiftRarity
	for _, option := range sub.Options {
		if option != nil && option.Name == "rarity" && option.Type == discordgo.ApplicationCommandOptionString {
			value, _ := option.Value.(string)
			rarity = core.GiftRarity(value)
		}
	}
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		opened, err := b.service.OpenGift(ctx, member, rarity)
		if err != nil {
			if reply := gifttext.OpenRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "open a gift", err)
			return gifttext.OpenUnavailable
		}
		return gifttext.OpenedGift(rarity, opened)
	})
}
