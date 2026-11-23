package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const thrownCommandName = "thrown"

func thrownCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: thrownCommandName, Description: "Throw a snowball",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "snowball", Description: "Throw a snowball at another player",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Description: "Player receiving the snowball", Required: true},
				},
			},
		},
	}
}

func (b bot) handleThrownSnowball(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.ApplicationCommandInteractionData) {
	sub := commandSubcommand(data.Options, "snowball")
	if sub == nil || len(sub.Options) != 1 {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, sender app.Member) string {
		recipient, name := resolveRecipient(sender.ServerID, sub.Options[0], data.Resolved)
		total, err := b.service.ThrowSnowball(ctx, sender, recipient)
		if err != nil {
			if reply := gifttext.SnowballRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "record snowball", err)
			return "Could not record your snowball throw. Try again."
		}

		reply := gifttext.SnowballThrown(name)
		senderName := memberDisplayName(interaction.Member, interaction.Member.User)
		if err := notifySnowball(ctx, session, interaction.ChannelID, recipient.Member.UserID, gifttext.SnowballReceived(senderName, total)); err != nil {
			logInteractionError(interaction.ID, "notify snowball recipient", err)
			reply += "\nThe snowball was added to their inventory, but I couldn't post the notification in this channel."
		}
		return reply
	})
}

func notifySnowball(ctx context.Context, session *discordgo.Session, channelID, recipientID, text string) error {
	_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: "<@" + recipientID + "> " + text,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{},
			Users: []string{recipientID},
		},
	}, discordgo.WithContext(ctx))
	return err
}
