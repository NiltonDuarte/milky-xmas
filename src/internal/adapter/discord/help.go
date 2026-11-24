package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func helpCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: gifttext.HelpCommand, Description: "List the commands you can use",
	}
}

func (b bot) handleHelp(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivately(session, interaction, func(context.Context, app.Member) string {
		permissions := interaction.Member.Permissions
		admin := permissions&discordgo.PermissionAdministrator != 0
		return gifttext.Help(gifttext.HelpAccess{
			Moderator: admin || permissions&discordgo.PermissionManageMessages != 0,
			Manager:   admin || permissions&discordgo.PermissionManageGuild != 0,
		})
	})
}
