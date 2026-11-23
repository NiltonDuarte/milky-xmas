package discord

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

type bot struct {
	guildID string
	service app.Service
	grinch  *grinchBot
}

func Run(ctx context.Context, token, guildID string, service app.Service, grinchChannel string, encounter *app.GrinchEncounter) error {
	b := bot{guildID: guildID, service: service}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("create Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentGuilds
	if grinchChannel != "" {
		b.grinch = &grinchBot{guild: guildID, channel: grinchChannel, encounter: encounter}
		session.Identify.Intents |= discordgo.IntentsGuildMessages
		session.AddHandler(b.grinch.observe)
	}
	session.AddHandler(b.handleInteraction)

	if err := session.Open(); err != nil {
		return fmt.Errorf("open Discord gateway: %w", err)
	}
	defer session.Close()

	if err := b.registerCommands(ctx, session); err != nil {
		return err
	}
	log.Printf("Discord bot ready with /hello, /daily claim, /advent, /send gift, /aura check, /socks, and /thrown snowball in server %s", guildID)

	if b.grinch != nil {
		log.Printf("Grinch encounter enabled in channel %s", grinchChannel)
		b.grinch.run(ctx, session)
	} else {
		<-ctx.Done()
	}
	return nil
}

func (b bot) registerCommands(ctx context.Context, session *discordgo.Session) error {
	if b.guildID == "" {
		return fmt.Errorf("a Discord server ID is required to register commands")
	}

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
		auraCommand(),
		{
			Type:        discordgo.ChatApplicationCommand,
			Name:        app.SocksCommand,
			Description: "View your gift and item inventory",
		},
		adventCommand(),
		grinchCommand(),
		thrownCommand(),
	}
	for _, command := range commands {
		if _, err := session.ApplicationCommandCreate(user.ID, b.guildID, command, discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("register /%s in Discord server %s: %w", command.Name, b.guildID, err)
		}
	}
	return nil
}

func (b bot) handleInteraction(session *discordgo.Session, event *discordgo.InteractionCreate) {
	if event.GuildID != b.guildID {
		return
	}
	if event.Type == discordgo.InteractionMessageComponent && b.grinch != nil {
		b.grinch.button(session, event.Interaction)
		return
	}
	if event.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data, ok := event.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok || data.CommandType != discordgo.ChatApplicationCommand {
		return
	}

	switch data.Name {
	case thrownCommandName:
		b.handleThrownSnowball(session, event.Interaction, data)
		return
	case app.SocksCommand:
		if len(data.Options) == 0 {
			b.handleSocks(session, event.Interaction)
		}
		return
	case sendCommandName:
		b.handleSendGift(session, event.Interaction, data)
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
	log.Printf("Discord interaction %s could not %s: %v", id, action, err)
}
