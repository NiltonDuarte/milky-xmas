package discord

import (
	"context"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func respondPrivately(session *discordgo.Session, interaction *discordgo.Interaction, reply func(context.Context, app.Member) string) {
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

	content := "Could not identify the member making this request."
	if interaction.Member != nil && interaction.Member.User != nil && interaction.Member.User.ID != "" {
		member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		content = reply(ctx, member)
		cancel()
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
		Content:         &content,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "show command result", err)
	}
}
