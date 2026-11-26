package discord

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

const setupCommandName = "setup"

func setupCommand() *discordgo.ApplicationCommand {
	permission := int64(discordgo.PermissionManageGuild)
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: setupCommandName, Description: "Choose where Milky Xmas can run",
		DefaultMemberPermissions: &permission,
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "view", Description: "View the configured channels"},
			{
				Type: discordgo.ApplicationCommandOptionSubCommand, Name: "channel", Description: "Set a channel's command, boss and quote choices",
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionChannel, Name: "channel", Description: "Discord text channel", Required: true,
						ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText}},
					{Type: discordgo.ApplicationCommandOptionBoolean, Name: "commands", Description: "Allow bot commands here", Required: true},
					{Type: discordgo.ApplicationCommandOptionBoolean, Name: "boss", Description: "Allow Grinch encounters here", Required: true},
					{Type: discordgo.ApplicationCommandOptionBoolean, Name: "quotes", Description: "Post Christmas quotes here when people chat"},
				},
			},
		},
	}
}

func (b bot) permitsGuild(guild string) bool {
	return guild != "" && (b.guildID == "" || b.guildID == guild)
}

func (b bot) guildAvailable(ctx context.Context, session *discordgo.Session, guild *discordgo.Guild) {
	if guild == nil || guild.Unavailable || !b.permitsGuild(guild.ID) {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	if err := b.grinches.refresh(ctx, session, guild.ID); err != nil {
		log.Printf("Restore boss channels in guild %s: %v", guild.ID, err)
		return
	}
	settings, err := b.setup.Settings(ctx, guild.ID)
	if err != nil || settings.Welcomed || len(settings.Channels) != 0 {
		return
	}
	if channel := setupPromptChannel(ctx, session, guild); channel != "" {
		if err := b.setup.Welcome(ctx, guild.ID, channel, setupPoster(session)); err != nil {
			log.Printf("Post setup prompt in guild %s: %v", guild.ID, err)
		}
	}
}

func setupPromptChannel(ctx context.Context, session *discordgo.Session, guild *discordgo.Guild) string {
	session.State.RLock()
	user := session.State.User
	session.State.RUnlock()
	if user == nil {
		return ""
	}
	channels := slices.Clone(guild.Channels)
	slices.SortStableFunc(channels, func(a, b *discordgo.Channel) int {
		if a == nil || b == nil {
			return 0
		}
		switch {
		case a.ID == guild.SystemChannelID && b.ID != guild.SystemChannelID:
			return -1
		case b.ID == guild.SystemChannelID && a.ID != guild.SystemChannelID:
			return 1
		default:
			return a.Position - b.Position
		}
	})
	const required = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages
	for _, channel := range channels {
		if channel != nil && channel.Type == discordgo.ChannelTypeGuildText {
			permissions, err := session.UserChannelPermissions(user.ID, channel.ID, discordgo.WithContext(ctx))
			if err == nil && permissions&required == required {
				return channel.ID
			}
		}
	}
	return ""
}

func setupPoster(session *discordgo.Session) app.SetupAnnouncement {
	return func(ctx context.Context, channel, content string) error {
		message := &discordgo.MessageSend{Content: content, AllowedMentions: &discordgo.MessageAllowedMentions{}}
		if content == app.ChristmasIntroduction {
			// The introduction is the last setup message: it carries the way to join the elf army.
			message.Components = elfArmyComponents()
		}
		_, err := session.ChannelMessageSendComplex(channel, message, discordgo.WithContext(ctx))
		return err
	}
}

func (b bot) handleSetup(session *discordgo.Session, interaction *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		canManage := interaction.Member.Permissions&(discordgo.PermissionManageGuild|discordgo.PermissionAdministrator) != 0
		if !canManage {
			return "You need Manage Server permission to configure Milky Xmas."
		}
		if commandSubcommand(options, "view") != nil {
			settings, err := b.setup.Settings(ctx, member.ServerID)
			if err != nil {
				logInteractionError(interaction.ID, "load channel choices", err)
				return "Could not load the channel choices."
			}
			return channelChoices(settings)
		}

		sub := commandSubcommand(options, "channel")
		if sub == nil {
			return "Use /setup channel or /setup view."
		}
		channelID, choice, ok := setupOptions(sub.Options)
		if !ok {
			return "Choose a text channel and set both commands and boss."
		}
		channel, err := session.State.Channel(channelID)
		if err != nil {
			channel, err = session.Channel(channelID, discordgo.WithContext(ctx))
		}
		if err != nil || channel == nil || channel.GuildID != member.ServerID || channel.Type != discordgo.ChannelTypeGuildText {
			return "Choose a text channel in this server."
		}

		saved, err := b.setup.Configure(ctx, member.ServerID, channelID, canManage, choice, setupPoster(session))
		if !saved {
			logInteractionError(interaction.ID, "save channel choices", err)
			return "Could not save the channel choices."
		}
		if choice.Boss {
			prepareElfRole(ctx, session, member.ServerID)
		}
		reply := fmt.Sprintf("Saved <#%s> · commands %s · boss %s", channelID, setupToggle(choice.Commands), setupToggle(choice.Boss))
		if choice.Quotes != nil {
			reply += " · quotes " + setupToggle(*choice.Quotes)
		}
		reply += "."
		// A failed Discord post may exhaust the reply context. Refresh the saved choices independently.
		refreshCtx, cancel := context.WithTimeout(b.grinches.ctx, 2*time.Second)
		defer cancel()
		if refreshErr := b.grinches.refresh(refreshCtx, session, member.ServerID); refreshErr != nil {
			logInteractionError(interaction.ID, "apply boss channels", refreshErr)
			reply += "\nI could not refresh the boss channels. Repeat the setup to retry."
		}
		if err != nil {
			logInteractionError(interaction.ID, "announce Christmas setup", err)
			reply += "\nI could not confirm the Christmas announcement. Repeat the setup to retry."
		}
		return reply
	})
}

func setupOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (string, app.ChannelChoice, bool) {
	var channel string
	var choice app.ChannelChoice
	var commandsSet, bossSet bool
	for _, option := range options {
		if option == nil {
			continue
		}
		switch {
		case option.Name == "channel" && option.Type == discordgo.ApplicationCommandOptionChannel:
			channel, _ = option.Value.(string)
		case option.Name == "commands" && option.Type == discordgo.ApplicationCommandOptionBoolean:
			choice.Commands, commandsSet = option.Value.(bool)
		case option.Name == "boss" && option.Type == discordgo.ApplicationCommandOptionBoolean:
			choice.Boss, bossSet = option.Value.(bool)
		case option.Name == "quotes" && option.Type == discordgo.ApplicationCommandOptionBoolean:
			if quotes, ok := option.Value.(bool); ok {
				choice.Quotes = &quotes
			}
		}
	}
	return channel, choice, channel != "" && commandsSet && bossSet
}

func channelChoices(settings app.GuildSettings) string {
	var commands, bosses, quotes []string
	for id, channel := range settings.Channels {
		if channel.Commands {
			commands = append(commands, "<#"+id+">")
		}
		if channel.Boss {
			bosses = append(bosses, "<#"+id+">")
		}
		if channel.Quotes {
			quotes = append(quotes, "<#"+id+">")
		}
	}
	var lines, counts []string
	for _, group := range []struct {
		name     string
		channels []string
	}{{"Commands", commands}, {"Grinch", bosses}, {"Quotes", quotes}} {
		slices.Sort(group.channels)
		list := strings.Join(group.channels, ", ")
		if list == "" {
			list = "none"
		}
		lines = append(lines, group.name+" · "+list)
		counts = append(counts, fmt.Sprintf("%s · %d channels", group.name, len(group.channels)))
	}
	reply := strings.Join(lines, "\n")
	if len(reply) > 1900 {
		return strings.Join(counts, "\n") + ". The full list exceeds Discord's message limit."
	}
	return reply
}

func setupToggle(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}
