package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const sendCommandName = "send"
const auraCommandName = "aura"

func sendCommand() *discordgo.ApplicationCommand {
	minQuantity := float64(1)
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        sendCommandName,
		Description: "Send Christmas gifts",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Description: "Send gifts to another player",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Description: "Player receiving the gifts", Required: true},
					{
						Type: discordgo.ApplicationCommandOptionString, Name: "rarity", Description: "Gift rarity", Required: true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Normal", Value: string(core.NormalGift)},
							{Name: "Rare", Value: string(core.RareGift)},
							{Name: "Epic", Value: string(core.EpicGift)},
						},
					},
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "quantity", Description: "Number of gifts", Required: true, MinValue: &minQuantity},
				},
			},
		},
	}
}

func auraCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        auraCommandName,
		Description: "View your personal aura",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "check", Description: "Check your Bad deeds and Good deeds"},
		},
	}
}

func (b bot) handleSendGift(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.ApplicationCommandInteractionData) {
	sub := giftSubcommand(data.Options, "gift")
	if sub == nil {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		recipient, rarity, quantity := giftOptions(member.ServerID, sub.Options, data.Resolved)
		aura, err := b.service.SendGift(ctx, member, recipient, rarity, quantity)
		if err != nil {
			if reply := gifttext.TransferRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "send gift", err)
			return "Could not send your gifts. Try again."
		}
		return gifttext.Transfer(rarity, quantity, aura)
	})
}

func (b bot) handleAura(session *discordgo.Session, interaction *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	sub := giftSubcommand(options, "check")
	if sub == nil || len(sub.Options) != 0 {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		aura, err := b.service.CheckAura(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "load aura", err)
			return "Could not load your aura. Try again."
		}
		return gifttext.Aura(aura)
	})
}

func giftSubcommand(options []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	if len(options) != 1 || options[0] == nil || options[0].Name != name || options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return nil
	}
	return options[0]
}

func giftOptions(serverID string, options []*discordgo.ApplicationCommandInteractionDataOption, resolved *discordgo.ApplicationCommandInteractionDataResolved) (app.GiftRecipient, core.GiftRarity, int64) {
	var recipient app.GiftRecipient
	var rarity core.GiftRarity
	var quantity int64
	for _, option := range options {
		if option == nil {
			continue
		}
		switch {
		case option.Name == "recipient" && option.Type == discordgo.ApplicationCommandOptionUser:
			id, _ := option.Value.(string)
			if resolved != nil && resolved.Users[id] != nil && resolved.Members[id] != nil {
				recipient = app.GiftRecipient{
					Member: app.Member{ServerID: serverID, UserID: id},
					IsBot:  resolved.Users[id].Bot,
				}
			}
		case option.Name == "rarity" && option.Type == discordgo.ApplicationCommandOptionString:
			value, _ := option.Value.(string)
			rarity = core.GiftRarity(value)
		case option.Name == "quantity" && option.Type == discordgo.ApplicationCommandOptionInteger:
			// DiscordGo decodes numeric options as float64.
			value, ok := option.Value.(float64)
			if ok && value >= 1 && value <= 1<<53 && value == float64(int64(value)) {
				quantity = int64(value)
			}
		}
	}
	return recipient, rarity, quantity
}
