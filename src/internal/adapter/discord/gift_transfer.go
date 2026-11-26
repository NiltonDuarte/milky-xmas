package discord

import (
	"context"
	_ "embed"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

//go:embed badgood.jpg
var badGoodImage []byte

const sendCommandName = "send"
const auraCommandName = "aura"

func sendCommand() *discordgo.ApplicationCommand {
	minQuantity := float64(1)
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        sendCommandName,
		Description: "Send gifts or Coal",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "gift", Description: "Send a gift or Coal to another player",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Description: "Player receiving the items", Required: true},
					{
						Type: discordgo.ApplicationCommandOptionString, Name: "item", Description: "Gift rarity or Coal", Required: true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Normal", Value: "normal"},
							{Name: "Rare", Value: "rare"},
							{Name: "Epic", Value: "epic"},
							{Name: "Coal", Value: app.CoalItem},
						},
					},
					{Type: discordgo.ApplicationCommandOptionInteger, Name: "quantity", Description: "Number of items", Required: true, MinValue: &minQuantity},
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
	sub := commandSubcommand(data.Options, "gift")
	if sub == nil {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		recipient, name, item, quantity := giftOptions(member.ServerID, sub.Options, data.Resolved, "quantity")
		aura, err := b.service.SendItem(ctx, member, recipient, item, quantity)
		if err != nil {
			if reply := gifttext.TransferRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "send gift", err)
			return "Could not send your gifts. Try again."
		}
		return gifttext.Transfer(item, quantity, name, aura)
	})
}

func (b bot) handleAura(session *discordgo.Session, interaction *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	sub := commandSubcommand(options, "check")
	if sub == nil || len(sub.Options) != 0 {
		return
	}
	respondPrivatelyWith(session, interaction, func(ctx context.Context, member app.Member) privateReply {
		aura, err := b.service.CheckAura(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "load aura", err)
			return privateReply{Content: "Could not load your aura. Try again."}
		}
		return privateReply{Content: gifttext.Aura(aura), Image: &privateImage{Name: "badgood.jpg", Data: badGoodImage}}
	})
}

func commandSubcommand(options []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	if len(options) != 1 || options[0] == nil || options[0].Name != name || options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return nil
	}
	return options[0]
}

func giftOptions(serverID string, options []*discordgo.ApplicationCommandInteractionDataOption, resolved *discordgo.ApplicationCommandInteractionDataResolved, quantityName string) (app.Recipient, string, string, int64) {
	var recipient app.Recipient
	var name string
	var item string
	var quantity int64
	for _, option := range options {
		if option == nil {
			continue
		}
		switch {
		case option.Name == "recipient" && option.Type == discordgo.ApplicationCommandOptionUser:
			recipient, name = resolveRecipient(serverID, option, resolved)
		case (option.Name == "item" || option.Name == "rarity") && option.Type == discordgo.ApplicationCommandOptionString:
			item, _ = option.Value.(string)
		case option.Name == quantityName && option.Type == discordgo.ApplicationCommandOptionInteger:
			// DiscordGo decodes numeric options as float64.
			value, ok := option.Value.(float64)
			if ok && value >= 1 && value <= 1<<53 && value == float64(int64(value)) {
				quantity = int64(value)
			}
		}
	}
	return recipient, name, item, quantity
}

func resolveRecipient(serverID string, option *discordgo.ApplicationCommandInteractionDataOption, resolved *discordgo.ApplicationCommandInteractionDataResolved) (app.Recipient, string) {
	if option == nil || option.Name != "recipient" || option.Type != discordgo.ApplicationCommandOptionUser || resolved == nil {
		return app.Recipient{}, ""
	}
	id, _ := option.Value.(string)
	user, member := resolved.Users[id], resolved.Members[id]
	if id == "" || user == nil || member == nil {
		return app.Recipient{}, ""
	}
	return app.Recipient{Member: app.Member{ServerID: serverID, UserID: id}, IsBot: user.Bot}, memberDisplayName(member, user)
}

var escapeDisplayName = strings.NewReplacer(
	"\\", "\\\\", "*", "\\*", "_", "\\_", "~", "\\~", "`", "\\`", "|", "\\|",
	"<", "\\<", "[", "\\[", "]", "\\]", "\r", " ", "\n", " ",
)

func memberDisplayName(member *discordgo.Member, user *discordgo.User) string {
	name := member.Nick
	if name == "" {
		name = user.DisplayName()
	}
	if name == "" {
		name = user.ID
	}
	return escapeDisplayName.Replace(name)
}
