package discord

import (
	"context"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const (
	modGuideTopicID = "modguide:topic"
	modGuideColor   = 0xB03A2E
)

func modGuideCommand() *discordgo.ApplicationCommand {
	permission := int64(discordgo.PermissionManageMessages)
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: gifttext.ModGuideCommand,
		Description: "Everything about the bot, for moderators", DefaultMemberPermissions: &permission,
	}
}

// moderates reports whether the member may read the moderator guide.
func moderates(member *discordgo.Member) bool {
	const allowed = discordgo.PermissionManageMessages | discordgo.PermissionManageGuild | discordgo.PermissionAdministrator
	return member != nil && member.Permissions&allowed != 0
}

// modGuideReply is the private moderator guide: the menu of topics and, below it, the chosen one.
func (b bot) modGuideReply(topic string) (privateReply, bool) {
	blocks, ok := gifttext.ModGuide(topic, b.service.Tuning())
	if !ok {
		return privateReply{}, false
	}
	menu := statsMenu(modGuideTopicID, "Pick a topic", gifttext.ModGuideTopics, topic)
	return privateReply{Blocks: blocks, Components: []discordgo.MessageComponent{menu}, ComponentsFirst: true, Accent: modGuideColor}, true
}

func (b bot) handleModGuide(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivatelyWith(session, interaction, func(context.Context, app.Member) privateReply {
		if !moderates(interaction.Member) {
			return privateReply{Content: gifttext.ModGuideDenied}
		}
		reply, _ := b.modGuideReply(gifttext.GuideStart)
		return reply
	})
}

// modGuideSelect shows the topic picked in the moderator guide's menu.
func (b bot) modGuideSelect(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.MessageComponentInteractionData) {
	if data.CustomID != modGuideTopicID || len(data.Values) != 1 || !moderates(interaction.Member) {
		return
	}
	reply, ok := b.modGuideReply(data.Values[0])
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}, discordgo.WithContext(ctx))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "answer moderator guide choice", err)
		return
	}
	showResult(session, interaction, reply)
}
