package discord

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

// privateReply is a private command result. With Image set, the image is shown above the text.
type privateReply struct {
	Content    string
	Image      *privateImage
	Components []discordgo.MessageComponent
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
		clearUncertainAcknowledgement(session, interaction)
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
	edit := privateEdit(result)
	if _, err := session.InteractionResponseEdit(interaction, edit, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "show command result", err)
	}
}

// privateEdit builds the message edit for a private reply. Files are listed in the attachments by
// position, so editing a message that already has them replaces them instead of adding copies.
func privateEdit(result privateReply) *discordgo.WebhookEdit {
	edit := &discordgo.WebhookEdit{AllowedMentions: &discordgo.MessageAllowedMentions{}}
	if result.Image == nil {
		edit.Content = &result.Content
	} else {
		edit.Embeds = &[]*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: "attachment://" + result.Image.Name}},
			{Description: result.Content},
		}
		edit.Files = []*discordgo.File{{Name: result.Image.Name, ContentType: "image/jpeg", Reader: bytes.NewReader(result.Image.Data)}}
		edit.Attachments = attachmentsFor(edit.Files)
	}
	if result.Components != nil {
		edit.Components = &result.Components
	}
	return edit
}

// attachmentsFor lists newly uploaded files by their position, as Discord expects when an edit
// replaces a message's attachments.
func attachmentsFor(files []*discordgo.File) *[]*discordgo.MessageAttachment {
	attachments := make([]*discordgo.MessageAttachment, len(files))
	for i := range files {
		attachments[i] = &discordgo.MessageAttachment{ID: strconv.Itoa(i)}
	}
	return &attachments
}

// clearUncertainAcknowledgement ends a "thinking…" reply that may exist even though the
// acknowledgement call failed: when our request timed out, Discord can still have received it.
// The edit does no game work, so it cannot repeat an action. If the acknowledgement never arrived,
// the edit fails and the player sees Discord's own "did not respond" notice.
func clearUncertainAcknowledgement(session *discordgo.Session, interaction *discordgo.Interaction) {
	text := "Discord was slow to answer. Please try the command again."
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{Content: &text}, discordgo.WithContext(ctx))
		cancel()
		if err == nil {
			return
		}
	}
}
