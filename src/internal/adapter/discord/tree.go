package discord

import (
	"bytes"
	"context"
	_ "embed"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const waterCommandName = "water"

//go:embed deadtree.jpg
var deadTreeImage []byte

func waterCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: waterCommandName, Description: "Water the Christmas tree",
	}
}

// treeImage picks the picture for a water stage. Only the withered tree (0-25%) has a picture so
// far, so every stage uses it until the other pictures are added here.
func treeImage(stage int) (name string, data []byte) {
	switch stage {
	default:
		return "deadtree.jpg", deadTreeImage
	}
}

func (b bot) handleWater(session *discordgo.Session, interaction *discordgo.Interaction) {
	if b.tree == nil || interaction.Member == nil || interaction.Member.User == nil || interaction.Member.User.ID == "" {
		return
	}
	member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	mention := func(userID string) string { return "<@" + userID + ">" }
	tree, err := b.tree.Water(ctx, member)
	data := &discordgo.InteractionResponseData{AllowedMentions: &discordgo.MessageAllowedMentions{}}
	reply := gifttext.TreeRejection(err)
	switch {
	case err == nil:
		// The whole server sees a watering and the tree's new level.
		name, image := treeImage(core.TreeStage(tree.Level))
		data.Embeds = []*discordgo.MessageEmbed{
			{Image: &discordgo.MessageEmbedImage{URL: "attachment://" + name}},
			{Description: gifttext.Watered(member.UserID, tree, mention)},
		}
		data.Files = []*discordgo.File{{Name: name, ContentType: "image/jpeg", Reader: bytes.NewReader(image)}}
	case reply != "":
		data.Content, data.Flags = reply, discordgo.MessageFlagsEphemeral
	default:
		logInteractionError(interaction.ID, "water the tree", err)
		data.Content, data.Flags = gifttext.TreeUnavailable, discordgo.MessageFlagsEphemeral
	}
	respondCtx, stop := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer stop()
	if err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource, Data: data,
	}, discordgo.WithContext(respondCtx)); err != nil {
		logInteractionError(interaction.ID, "show the tree", err)
	}
}
