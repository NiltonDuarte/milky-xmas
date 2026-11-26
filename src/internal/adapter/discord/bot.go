package discord

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

type bot struct {
	guildID  string
	service  app.Service
	setup    *app.GuildSetup
	grinches *grinchChannels
	drops    *dropBot
	tree     *app.Tree
}

func Run(ctx context.Context, token, guildID string, service app.Service, setup *app.GuildSetup, encounter func(string, string) *app.GrinchEncounter, tree *app.Tree) error {
	runCtx, stop := context.WithCancel(ctx)
	channels := &grinchChannels{ctx: runCtx, setup: setup, build: encounter, workers: make(map[grinchChannel]grinchWorker)}
	defer func() {
		stop()
		channels.wait()
	}()
	b := bot{guildID: guildID, service: service, setup: setup, grinches: channels, drops: newDropBot(app.NewActivityDrops(service)), tree: tree}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("create Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentGuildMessages
	session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildCreate) {
		b.guildAvailable(runCtx, s, event.Guild)
	})
	session.AddHandler(func(_ *discordgo.Session, event *discordgo.GuildDelete) {
		if event.Guild != nil {
			channels.remove(event.ID)
		}
	})
	session.AddHandler(func(s *discordgo.Session, event *discordgo.MessageCreate) {
		b.observeDrop(s, event)
		if worker := channels.get(event.GuildID, event.ChannelID); worker != nil {
			worker.observe(s, event)
		}
	})
	session.AddHandler(b.handleInteraction)
	if err := session.Open(); err != nil {
		return fmt.Errorf("open Discord gateway: %w", err)
	}
	defer session.Close()
	if err := b.registerCommands(ctx, session); err != nil {
		return err
	}
	log.Printf("Discord bot ready with /setup and gameplay commands (configured test guild %q)", guildID)
	<-ctx.Done()
	return nil
}

func (b bot) registerCommands(ctx context.Context, session *discordgo.Session) error {
	user, err := session.User("@me", discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("look up Discord bot user: %w", err)
	}
	commands := []*discordgo.ApplicationCommand{
		{
			Type:        discordgo.ChatApplicationCommand,
			Name:        app.HelloCommand,
			Description: "Get a greeting from Milky Xmas",
		},
		dailyCommand(),
		sendCommand(),
		giveCommand(),
		auraCommand(),
		{
			Type:        discordgo.ChatApplicationCommand,
			Name:        app.LookCommand,
			Description: "Look at your items",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionSubCommand, Name: app.StockingSubcommand, Description: "View your gift and item inventory"},
			},
		},
		adventCommand(),
		grinchCommand(),
		throwCommand(),
		waterCommand(),
		spiritCommand(),
		helpCommand(),
		setupCommand(),
	}
	wanted := make(map[string]bool, len(commands))
	for _, command := range commands {
		wanted[command.Name] = true
		if b.guildID == "" {
			command.Contexts = &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild}
		}
		if _, err := session.ApplicationCommandCreate(user.ID, b.guildID, command, discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("register /%s in Discord server %s: %w", command.Name, b.guildID, err)
		}
	}
	// Remove commands this bot no longer offers, such as the old /socks.
	existing, err := session.ApplicationCommands(user.ID, b.guildID, discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("list Discord commands in server %s: %w", b.guildID, err)
	}
	for _, command := range existing {
		if !wanted[command.Name] {
			if err := session.ApplicationCommandDelete(user.ID, b.guildID, command.ID, discordgo.WithContext(ctx)); err != nil {
				return fmt.Errorf("remove /%s from Discord server %s: %w", command.Name, b.guildID, err)
			}
		}
	}
	return nil
}

func (b bot) handleInteraction(session *discordgo.Session, event *discordgo.InteractionCreate) {
	if !b.permitsGuild(event.GuildID) {
		return
	}
	if event.Type == discordgo.InteractionMessageComponent {
		data, ok := event.Data.(discordgo.MessageComponentInteractionData)
		if ok && strings.HasPrefix(data.CustomID, "drop:") && b.drops != nil {
			b.drops.button(session, event.Interaction)
		} else if ok && strings.HasPrefix(data.CustomID, adventButtonPrefix) {
			b.adventButton(session, event.Interaction)
		} else if ok && data.CustomID == elfButtonID {
			b.elfButton(session, event.Interaction)
		} else if ok && strings.HasPrefix(data.CustomID, "grinch:") {
			if worker := b.grinches.get(event.GuildID, event.ChannelID); worker != nil {
				worker.button(session, event.Interaction)
			} else {
				replyChannelDenied(session, event.Interaction, "The Grinch is disabled in this channel.")
			}
		}
		return
	}

	if event.Type != discordgo.InteractionApplicationCommand && event.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}
	data, ok := event.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok || data.CommandType != discordgo.ChatApplicationCommand {
		return
	}

	if event.Type == discordgo.InteractionApplicationCommandAutocomplete {
		if data.Name == throwCommandName && b.setup != nil {
			b.autocompleteSnowball(session, event.Interaction, data)
		}
		return
	}

	if b.setup != nil {
		if data.Name == setupCommandName {
			b.handleSetup(session, event.Interaction, data.Options)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		settings, err := b.setup.Settings(ctx, event.GuildID)
		cancel()
		if err != nil {
			logInteractionError(event.ID, "load channel permissions", err)
			replyChannelDenied(session, event.Interaction, "Could not check which channels are enabled.")
			return
		}
		if !settings.Channels[event.ChannelID].Commands {
			replyChannelDenied(session, event.Interaction, "Commands are disabled in this channel. Ask a moderator to use /setup channel.")
			return
		}
	}

	switch data.Name {
	case waterCommandName:
		if len(data.Options) == 0 {
			b.handleWater(session, event.Interaction)
		}
		return
	case gifttext.SpiritCommand:
		if len(data.Options) == 0 {
			b.handleSpirit(session, event.Interaction)
		}
		return
	case gifttext.HelpCommand:
		if len(data.Options) == 0 {
			b.handleHelp(session, event.Interaction)
		}
		return
	case throwCommandName:
		b.handleThrowSnowball(session, event.Interaction, data)
		return
	case app.LookCommand:
		if commandSubcommand(data.Options, app.StockingSubcommand) != nil {
			b.handleStocking(session, event.Interaction)
		}
		return
	case sendCommandName:
		b.handleSendGift(session, event.Interaction, data)
		return
	case giveCommandName:
		b.handleGiveGift(session, event.Interaction, data)
		return
	case auraCommandName:
		b.handleAura(session, event.Interaction, data.Options)
		return
	case "grinch":
		b.summonGrinch(session, event.Interaction, data.Options)
		return
	case adventCommandName:
		b.handleAdvent(session, event.Interaction, data.Options)
		return
	case dailyCommandName:
		b.handleDailyClaim(session, event.Interaction, data.Options)
		return
	}

	reply, err := b.service.Reply(data.Name)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()
	err = session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: reply},
	}, discordgo.WithContext(ctx))
	if err != nil {
		logInteractionError(event.ID, "acknowledge /"+data.Name, err)
	}
}

func logInteractionError(id, action string, err error) {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// The callback and webhook URLs contain the interaction token.
		err = urlErr.Err
	}
	log.Printf("Discord interaction %s could not %s after %s: %v", id, action, interactionAge(id), err)
}

func replyChannelDenied(session *discordgo.Session, interaction *discordgo.Interaction, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1750*time.Millisecond)
	defer cancel()
	if err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: text, Flags: discordgo.MessageFlagsEphemeral},
	}, discordgo.WithContext(ctx)); err != nil {
		logInteractionError(interaction.ID, "show channel restriction", err)
	}
}

// interactionAge is how long ago Discord created the interaction. Discord wants the first response
// within three seconds of that moment, so the age of a failed response shows whether the time went
// into delivery, our own work, or the response itself.
func interactionAge(id string) time.Duration {
	created, err := discordgo.SnowflakeTimestamp(id)
	if err != nil {
		return 0
	}
	return time.Since(created).Round(time.Millisecond)
}
