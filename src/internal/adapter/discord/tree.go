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

//go:embed deadtree.jpg
var deadTreeImage []byte

//go:embed fulltree.png
var fullTreeImage []byte

func waterCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: waterCommandName, Description: "Water the Christmas tree",
	}
}

// treeImage picks the picture for a water stage: withered up to 50% and decorated when full.
// From 51% to 99% there is no picture yet, so it returns nil.
func treeImage(stage int) *privateImage {
	switch stage {
	case 4:
		return &privateImage{Name: "fulltree.png", Data: fullTreeImage}
	case 2, 3:
		return nil
	default:
		return &privateImage{Name: "deadtree.jpg", Data: deadTreeImage}
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
	tree, goodDeeds, err := b.tree.Water(ctx, member)
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
		watered := privateReply{Image: image(), Blocks: gifttext.WateredBlocks(member.UserID, goodDeeds, tree, b.tree.Rules(), b.tree.Now(), mention)}
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
