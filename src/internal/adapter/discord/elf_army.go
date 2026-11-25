package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const (
	elfRoleName   = "elfs"
	elfButtonID   = "elfs:join"
	elfButtonText = "Join the elf army 🛷"
)

// elfArmyComponents is the button row under the Christmas introduction.
func elfArmyComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{Label: elfButtonText, Style: discordgo.SuccessButton, CustomID: elfButtonID},
	}}}
}

// ensureElfRole returns the ID of the server's "elfs" role, creating it when it is missing.
func ensureElfRole(ctx context.Context, session *discordgo.Session, guild string) (string, error) {
	roles, err := session.GuildRoles(guild, discordgo.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("list server roles: %w", err)
	}
	for _, role := range roles {
		if role != nil && strings.EqualFold(role.Name, elfRoleName) {
			return role.ID, nil
		}
	}
	mentionable := true
	role, err := session.GuildRoleCreate(guild, &discordgo.RoleParams{Name: elfRoleName, Mentionable: &mentionable},
		discordgo.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("create the %s role: %w", elfRoleName, err)
	}
	return role.ID, nil
}

// prepareElfRole creates the role as soon as a server finishes setup, so the button works at once.
func prepareElfRole(ctx context.Context, session *discordgo.Session, guild string) {
	if _, err := ensureElfRole(ctx, session, guild); err != nil {
		log.Printf("Prepare the %s role in guild %s: %v", elfRoleName, guild, err)
	}
}

// elfButton gives the clicker the "elfs" role. The deferred private reply comes first, because
// the role calls go to Discord and can take longer than the three seconds Discord allows.
func (b bot) elfButton(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		role, err := ensureElfRole(ctx, session, member.ServerID)
		if err != nil {
			logInteractionError(interaction.ID, "find the elfs role", err)
			return "Could not find the elfs role. Ask a moderator to check my Manage Roles permission."
		}
		for _, owned := range interaction.Member.Roles {
			if owned == role {
				return "You are already in the elf army. 🛷"
			}
		}
		if err := session.GuildMemberRoleAdd(member.ServerID, member.UserID, role, discordgo.WithContext(ctx)); err != nil {
			logInteractionError(interaction.ID, "give the elfs role", err)
			return "Could not add you to the elf army. Ask a moderator to move my role above elfs."
		}
		return fmt.Sprintf("You joined the elf army! You now carry the <@&%s> tag. 🛷", role)
	})
}
