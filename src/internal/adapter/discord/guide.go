package discord

import (
	"context"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const (
	guideTopicID = "guide:topic"
	guideColor   = 0x165B33
)

func guideCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: gifttext.GuideCommand, Description: "Everything you can do in the event",
	}
}

// guideReply is the whole private guide: the menu of topics and, below it, the chosen one.
func (b bot) guideReply(topic string) (privateReply, bool) {
	blocks, ok := gifttext.Guide(topic, b.service.Tuning())
	if !ok {
		return privateReply{}, false
	}
	menu := statsMenu(guideTopicID, "Pick a topic", gifttext.GuideTopics, topic)
	return privateReply{Blocks: blocks, Components: []discordgo.MessageComponent{menu}, ComponentsFirst: true, Accent: guideColor}, true
}

func (b bot) handleGuide(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivatelyWith(session, interaction, func(context.Context, app.Member) privateReply {
		reply, _ := b.guideReply(gifttext.GuideStart)
		return reply
	})
}

// guideSelect shows the topic picked in the guide's menu.
func (b bot) guideSelect(session *discordgo.Session, interaction *discordgo.Interaction, data discordgo.MessageComponentInteractionData) {
	if data.CustomID != guideTopicID || len(data.Values) != 1 {
		return
	}
	reply, ok := b.guideReply(data.Values[0])
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}, discordgo.WithContext(ctx))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "answer guide choice", err)
		return
	}
	showResult(session, interaction, reply)
}
