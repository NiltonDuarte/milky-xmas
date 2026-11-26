package discord

import (
	"context"
	_ "embed"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

//go:embed spirit.jpg
var spiritGoodImage []byte

func spiritCommand() *discordgo.ApplicationCommand {
	permission := int64(discordgo.PermissionManageMessages)
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: gifttext.SpiritCommand,
		Description: "Reveal the Christmas spirit of the server", DefaultMemberPermissions: &permission,
	}
}

// spiritImage picks the picture for the winning side. There is only one picture so far, so a
// server with more bad deeds shows it too.
func spiritImage(core.Spirit) *privateImage {
	return &privateImage{Name: "spirit.jpg", Data: spiritGoodImage}
}

// handleSpirit posts the server's Christmas spirit to the channel for a moderator. Refusals and
// failures stay private, as for /water.
func (b bot) handleSpirit(session *discordgo.Session, interaction *discordgo.Interaction) {
	if !deferPrivately(session, interaction) {
		return
	}
	if interaction.Member == nil || interaction.Member.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) == 0 {
		showResult(session, interaction, privateReply{Content: gifttext.SpiritNeedsModerator})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	spirit, err := b.service.ChristmasSpirit(ctx, strings.TrimSpace(interaction.GuildID))
	cancel()
	if err != nil {
		logInteractionError(interaction.ID, "read the Christmas spirit", err)
		showResult(session, interaction, privateReply{Content: gifttext.SpiritUnavailable})
		return
	}
	announceBox(session, interaction, privateReply{Image: spiritImage(spirit), Blocks: gifttext.SpiritBlocks(spirit)})
}
