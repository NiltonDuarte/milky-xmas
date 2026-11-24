package discord

import (
	"bytes"
	"context"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

// privateReply is a private command result. With Image set, the image is shown above the text.
type privateReply struct {
	Content string
	Image   *privateImage
}

type privateImage struct {
	Name string
	Data []byte
}

func respondPrivately(session *discordgo.Session, interaction *discordgo.Interaction, reply func(context.Context, app.Member) string) {
	respondPrivatelyWith(session, interaction, func(ctx context.Context, member app.Member) privateReply {
		return privateReply{Content: reply(ctx, member)}
	})
}

func respondPrivatelyWith(session *discordgo.Session, interaction *discordgo.Interaction, reply func(context.Context, app.Member) privateReply) {
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}, discordgo.WithContext(ctx))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "acknowledge command", err)
		return
	}

	result := privateReply{Content: "Could not identify the member making this request."}
	if interaction.Member != nil && interaction.Member.User != nil && interaction.Member.User.ID != "" {
		member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result = reply(ctx, member)
		cancel()
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	edit := &discordgo.WebhookEdit{AllowedMentions: &discordgo.MessageAllowedMentions{}}
	if result.Image == nil {
		edit.Content = &result.Content
	} else {
		edit.Embeds = &[]*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: "attachment://" + result.Image.Name}},
			{Description: result.Content},
		}
		edit.Files = []*discordgo.File{{Name: result.Image.Name, ContentType: "image/jpeg", Reader: bytes.NewReader(result.Image.Data)}}
	}
	if _, err := session.InteractionResponseEdit(interaction, edit, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "show command result", err)
	}
}
