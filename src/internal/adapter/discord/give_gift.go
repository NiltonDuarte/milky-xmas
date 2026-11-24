package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const giveCommandName = "give"

func giveCommand() *discordgo.ApplicationCommand {
	minAmount := float64(1)
	permission := int64(discordgo.PermissionManageMessages)
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: giveCommandName, Description: "Moderator tools for giving items",
		DefaultMemberPermissions: &permission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Description: "Give gifts to a player",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Description: "Player receiving the gifts", Required: true},
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "amount", Description: "Number of gifts", Required: true, MinValue: &minAmount},
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

func (b bot) handleGiveGift(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.ApplicationCommandInteractionData) {
	sub := commandSubcommand(data.Options, "gift")
	if sub == nil {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		if interaction.Member.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) == 0 {
			return gifttext.GiveNeedsModerator
		}
		recipient, name, rarity, quantity := giftOptions(member.ServerID, sub.Options, data.Resolved, "amount")
		if err := b.service.GiveGift(ctx, member, recipient, core.GiftRarity(rarity), quantity); err != nil {
			if reply := gifttext.TransferRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "give gift", err)
			return "Could not give the gifts. Try again."
		}
		return gifttext.GiftGiven(quantity, rarity, name)
	})
}
