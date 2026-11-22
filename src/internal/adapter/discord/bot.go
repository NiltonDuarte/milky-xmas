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
}

func Run(ctx context.Context, token, guildID string, service app.Service) error {
	b := bot{guildID: guildID, service: service}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("create Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentGuilds
	session.AddHandler(b.handleInteraction)

	if err := session.Open(); err != nil {
		return fmt.Errorf("open Discord gateway: %w", err)
	}
	defer session.Close()

	if err := b.registerHello(ctx, session); err != nil {
		return err
	}
	log.Printf("Discord bot ready with /%s in server %s", app.HelloCommand, guildID)

	<-ctx.Done()
	return nil
}

func (b bot) registerHello(ctx context.Context, session *discordgo.Session) error {
	if b.guildID == "" {
		return fmt.Errorf("a Discord server ID is required to register /%s", app.HelloCommand)
	}

	user, err := session.User("@me", discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("look up Discord bot user: %w", err)
	}
	_, err = session.ApplicationCommandCreate(user.ID, b.guildID, &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        app.HelloCommand,
		Description: "Get a greeting from Milky Xmas",
	}, discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("register /%s in Discord server %s: %w", app.HelloCommand, b.guildID, err)
	}
	return nil
}

func (b bot) handleInteraction(session *discordgo.Session, event *discordgo.InteractionCreate) {
	if event.Type != discordgo.InteractionApplicationCommand || event.GuildID != b.guildID {
		return
	}
	data, ok := event.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok || data.CommandType != discordgo.ChatApplicationCommand {
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
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			// The callback URL contains the interaction token.
			err = urlErr.Err
		}
		log.Printf("Discord /%s acknowledgement failed for interaction %s: %v", data.Name, event.ID, err)
	}
}
