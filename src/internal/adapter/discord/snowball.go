package discord

import (
	"bytes"
	"context"
	_ "embed"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const throwCommandName = "throw"

//go:embed snowballnormal.png
var normalSnowballImage []byte

//go:embed snowballyellow.png
var yellowSnowballImage []byte

// snowballPicture is the picture posted with the notice of a thrown snowball.
func snowballPicture(kind core.SnowballType) *discordgo.File {
	if kind == core.YellowSnowball {
		return &discordgo.File{Name: "snowballyellow.png", ContentType: "image/png", Reader: bytes.NewReader(yellowSnowballImage)}
	}
	return &discordgo.File{Name: "snowballnormal.png", ContentType: "image/png", Reader: bytes.NewReader(normalSnowballImage)}
}

func throwCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: throwCommandName, Description: "Throw a snowball",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "snowball", Description: "Throw a snowball at another player",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionUser, Name: "recipient", Description: "Player receiving the snowball", Required: true},
					{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Choose a normal or yellow snowball", Required: true, Autocomplete: true},
				},
			},
		},
	}
}

func (b bot) handleThrowSnowball(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.ApplicationCommandInteractionData) {
	sub := commandSubcommand(data.Options, "snowball")
	if sub == nil {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, sender app.Member) string {
		var recipient app.Recipient
		var name string
		kind := core.NormalSnowball
		for _, option := range sub.Options {
			if option == nil {
				continue
			}
			switch {
			case option.Name == "recipient":
				recipient, name = resolveRecipient(sender.ServerID, option, data.Resolved)
			case option.Name == "type" && option.Type == discordgo.ApplicationCommandOptionString:
				value, _ := option.Value.(string)
				kind = core.SnowballType(value)
			}
		}
		_, err := b.service.ThrowSnowball(ctx, sender, recipient, kind)
		if err != nil {
			if reply := gifttext.SnowballRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "record snowball", err)
			return "Could not record your snowball throw. Try again."
		}

		reply := gifttext.SnowballThrown(name, kind)
		senderName := memberDisplayName(interaction.Member, interaction.Member.User)
		b.avatars.see(resolvedMember(sender.ServerID, recipient.Member.UserID, data.Resolved))
		if err := b.notifySnowball(ctx, session, interaction.ChannelID, recipient.Member.UserID, kind, senderName); err != nil {
			logInteractionError(interaction.ID, "notify snowball recipient", err)
			reply += "\nThe snowball was added to their inventory, but I couldn't post the notification in this channel."
		}
		return reply
	})
}

// resolvedMember is a player named in a command's options, with their user.
func resolvedMember(guild, userID string, resolved *discordgo.ApplicationCommandInteractionDataResolved) *discordgo.Member {
	if resolved == nil || resolved.Members[userID] == nil || resolved.Users[userID] == nil {
		return nil
	}
	member := *resolved.Members[userID]
	member.User, member.GuildID = resolved.Users[userID], guild
	return &member
}

func (b bot) notifySnowball(ctx context.Context, session *discordgo.Session, channelID, recipientID string, kind core.SnowballType, senderName string) error {
	picture := snowballPicture(kind)
	_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: b.avatars.message(gifttext.SnowballReceived("<@"+recipientID+">", senderName, kind)),
		Embeds:  []*discordgo.MessageEmbed{{Image: &discordgo.MessageEmbedImage{URL: "attachment://" + picture.Name}}},
		Files:   []*discordgo.File{picture},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{},
			Users: []string{recipientID},
		},
	}, discordgo.WithContext(ctx))
	return err
}

func (b bot) autocompleteSnowball(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.ApplicationCommandInteractionData) {
	sub := commandSubcommand(data.Options, "snowball")
	if sub == nil {
		return
	}
	query := ""
	for _, option := range sub.Options {
		if option != nil && option.Name == "type" && option.Focused {
			query, _ = option.Value.(string)
			break
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	if interaction.Member != nil && interaction.Member.User != nil && b.snowballChannelAllowed(ctx, interaction) {
		member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
		inventory, err := b.service.ViewInventory(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "load snowball choices", err)
		}
		query = strings.ToLower(strings.TrimSpace(query))
		if err == nil && inventory.Snowballs > 0 && strings.Contains("normal snowball", query) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: gifttext.SnowballEmoji(core.NormalSnowball) + " Normal Snowball", Value: string(core.NormalSnowball)})
		}
		if err == nil && inventory.YellowSnowballs > 0 && strings.Contains("yellow snowball", query) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: gifttext.SnowballEmoji(core.YellowSnowball) + " Yellow Snowball", Value: string(core.YellowSnowball)})
		}
	}
	replyCtx, stop := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer stop()
	// DiscordGo omits an empty Choices slice, but an empty result needs choices: [].
	endpoint := discordgo.EndpointInteractionResponse(interaction.ID, interaction.Token)
	_, err := session.RequestWithBucketID("POST", endpoint, map[string]any{
		"type": discordgo.InteractionApplicationCommandAutocompleteResult,
		"data": map[string]any{"choices": choices},
	}, endpoint, discordgo.WithContext(replyCtx))
	if err != nil {
		logInteractionError(interaction.ID, "show snowball choices", err)
	}
}

func (b bot) snowballChannelAllowed(ctx context.Context, interaction *discordgo.Interaction) bool {
	settings, err := b.setup.Settings(ctx, interaction.GuildID)
	return err == nil && settings.Channels[interaction.ChannelID].Commands
}
