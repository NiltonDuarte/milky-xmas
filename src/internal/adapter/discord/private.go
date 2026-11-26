package discord

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

// privateReply is a private command result. With Image set, the image is shown above the text.
// With Image or Blocks set, the reply is one box.
type privateReply struct {
	Content string
	// Blocks, when set, replace Content in the box: each is shown apart, with a dividing line.
	Blocks     []string
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
	if !deferPrivately(session, interaction) {
		return
	}

	result := privateReply{Content: "Could not identify the member making this request."}
	if interaction.Member != nil && interaction.Member.User != nil && interaction.Member.User.ID != "" {
		member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result = reply(ctx, member)
		cancel()
	}

	if err := showReply(session, interaction, result); err != nil {
		logInteractionError(interaction.ID, "show command result", err)
	}
}

// deferPrivately acknowledges a command whose result only its user sees. It reports false when
// the command must not go on.
func deferPrivately(session *discordgo.Session, interaction *discordgo.Interaction) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}, discordgo.WithContext(ctx))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "acknowledge command", err)
		clearUncertainAcknowledgement(session, interaction)
		return false
	}
	return true
}

// announceBox posts a boxed result to the channel and removes the private acknowledgement. If
// the post fails, the player still sees the box.
func announceBox(session *discordgo.Session, interaction *discordgo.Interaction, result privateReply) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	message := &discordgo.MessageSend{
		Flags:           discordgo.MessageFlagsIsComponentsV2,
		Components:      boxedComponents(result),
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}
	if result.Image != nil {
		message.Files = []*discordgo.File{result.Image.file()}
	}
	if _, err := session.ChannelMessageSendComplex(interaction.ChannelID, message, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "post the result to the channel", err)
		showResult(session, interaction, result)
		return
	}
	if err := session.InteractionResponseDelete(interaction, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "remove the private acknowledgement", err)
	}
}

// showResult ends a deferred private reply with result and logs a failure.
func showResult(session *discordgo.Session, interaction *discordgo.Interaction, result privateReply) {
	if err := showReply(session, interaction, result); err != nil {
		logInteractionError(interaction.ID, "show command result", err)
	}
}

// showReply edits the deferred reply into result. A reply with an image or blocks is one box
// holding the image, the text, and the buttons. If Discord refuses the box, two embeds are sent instead so
// the player still gets the result.
func showReply(session *discordgo.Session, interaction *discordgo.Interaction, result privateReply) error {
	if result.Image != nil || len(result.Blocks) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := editBoxedReply(ctx, session, interaction, result)
		cancel()
		if err == nil {
			return nil
		}
		logInteractionError(interaction.ID, "show boxed result", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := session.InteractionResponseEdit(interaction, privateEdit(result), discordgo.WithContext(ctx))
	return err
}

// editBoxedReply puts the image above the text and buttons inside one Components V2 container.
// discordgo's WebhookEdit has no flags field, so the request body is built here. Content and
// embeds are cleared so a message sent with embeds can become a box.
func editBoxedReply(ctx context.Context, session *discordgo.Session, interaction *discordgo.Interaction, result privateReply) error {
	var files []*discordgo.File
	if result.Image != nil {
		files = []*discordgo.File{result.Image.file()}
	}
	body := struct {
		Content         *string                           `json:"content"`
		Embeds          []*discordgo.MessageEmbed         `json:"embeds"`
		Flags           discordgo.MessageFlags            `json:"flags"`
		Components      []discordgo.MessageComponent      `json:"components"`
		Attachments     *[]*discordgo.MessageAttachment   `json:"attachments"`
		AllowedMentions *discordgo.MessageAllowedMentions `json:"allowed_mentions"`
	}{
		Embeds:          []*discordgo.MessageEmbed{},
		Flags:           discordgo.MessageFlagsIsComponentsV2,
		Components:      boxedComponents(result),
		Attachments:     attachmentsFor(files),
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}
	contentType, payload, err := discordgo.MultipartBodyWithJSON(body, files)
	if err != nil {
		return err
	}
	uri := discordgo.EndpointWebhookMessage(interaction.AppID, interaction.Token, "@original")
	_, err = session.RequestRaw("PATCH", uri, contentType, payload, discordgo.EndpointWebhookToken("", ""), 0, discordgo.WithContext(ctx))
	return err
}

// boxedComponents is the one box holding the image, the text, and the buttons of a reply.
func boxedComponents(result privateReply) []discordgo.MessageComponent {
	var box []discordgo.MessageComponent
	if result.Image != nil {
		box = append(box, discordgo.MediaGallery{Items: []discordgo.MediaGalleryItem{{Media: discordgo.UnfurledMediaItem{URL: "attachment://" + result.Image.Name}}}})
	}
	blocks := result.Blocks
	if len(blocks) == 0 {
		blocks = []string{result.Content}
	}
	for i, block := range blocks {
		if i > 0 {
			box = append(box, discordgo.Separator{})
		}
		box = append(box, discordgo.TextDisplay{Content: block})
	}
	return []discordgo.MessageComponent{discordgo.Container{Components: append(box, result.Components...)}}
}

// file is the image as a new upload.
func (image privateImage) file() *discordgo.File {
	return &discordgo.File{Name: image.Name, ContentType: http.DetectContentType(image.Data), Reader: bytes.NewReader(image.Data)}
}

// privateEdit builds the message edit for a private reply. Files are listed in the attachments by
// position, so editing a message that already has them replaces them instead of adding copies.
func privateEdit(result privateReply) *discordgo.WebhookEdit {
	edit := &discordgo.WebhookEdit{AllowedMentions: &discordgo.MessageAllowedMentions{}}
	text := result.Content
	if text == "" {
		text = strings.Join(result.Blocks, "\n")
	}
	if result.Image == nil {
		edit.Content = &text
	} else {
		edit.Embeds = &[]*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: "attachment://" + result.Image.Name}},
			{Description: text},
		}
		edit.Files = []*discordgo.File{result.Image.file()}
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
