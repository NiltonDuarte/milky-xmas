package discord

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const waterCommandName = "water"

//go:embed tree1.jpg
var witheredTreeImage []byte

//go:embed tree2.jpg
var wiltingTreeImage []byte

//go:embed tree3.png
var recoveringTreeImage []byte

//go:embed tree4.jpg
var decoratedTreeImage []byte

func waterCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: waterCommandName, Description: "Water the Christmas tree",
	}
}

// treeImage picks the picture for a water stage: up to 25%, up to 50%, up to 75%, and from 76%.
func treeImage(stage int) *privateImage {
	switch stage {
	case 0:
		return &privateImage{Name: "tree1.jpg", Data: witheredTreeImage}
	case 1:
		return &privateImage{Name: "tree2.jpg", Data: wiltingTreeImage}
	case 2:
		return &privateImage{Name: "tree3.png", Data: recoveringTreeImage}
	default:
		return &privateImage{Name: "tree4.jpg", Data: decoratedTreeImage}
	}
}

// handleWater acknowledges privately before it knows the outcome, so a refusal stays private. A
// watering that counts is posted to the channel as its own message, and the private
// acknowledgement is removed.
func (b bot) handleWater(session *discordgo.Session, interaction *discordgo.Interaction) {
	if b.tree == nil || !deferPrivately(session, interaction) {
		return
	}
	if interaction.Member == nil || interaction.Member.User == nil || interaction.Member.User.ID == "" {
		showResult(session, interaction, privateReply{Content: gifttext.TreeUnavailable})
		return
	}
	member := app.Member{ServerID: interaction.GuildID, UserID: interaction.Member.User.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	tree, watering, err := b.tree.Water(ctx, member)
	cancel()
	if errors.Is(err, app.ErrGoodDeedNotSaved) {
		logInteractionError(interaction.ID, "count the watering as a good deed", err)
		err = nil
	}
	mention := func(userID string) string { return "<@" + userID + ">" }
	image := func() *privateImage { return treeImage(core.TreeStage(tree.Level)) }
	switch reply := gifttext.TreeRejection(err); {
	case err == nil:
		// The whole server sees a watering and the tree's new level.
		watered := privateReply{Image: image(), Blocks: b.avatars.blocks(gifttext.WateredBlocks(member.UserID, watering, tree, b.tree.Rules(), b.tree.Now(), mention))}
		announceBox(session, interaction, watered)
	case errors.Is(err, core.ErrTreeEventOver):
		showResult(session, interaction, privateReply{Image: image(), Blocks: gifttext.TreeEventOverBlocks(tree, b.tree.Rules(), b.tree.Now())})
	case errors.Is(err, core.ErrTreeFull):
		// Only this player sees the full tree.
		showResult(session, interaction, privateReply{Image: image(), Content: reply})
	case reply != "":
		showResult(session, interaction, privateReply{Content: reply})
	default:
		logInteractionError(interaction.ID, "water the tree", err)
		showResult(session, interaction, privateReply{Content: gifttext.TreeUnavailable})
	}
}
