package discord

import (
	"context"
	"log"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/bwmarrin/discordgo"
)

// playerMessage reports whether the event is a message written by a person in a served server.
func (b bot) playerMessage(event *discordgo.MessageCreate) bool {
	return event != nil && event.Message != nil && event.Author != nil && !event.Author.Bot &&
		event.WebhookID == "" && b.permitsGuild(event.GuildID)
}

// observeQuote counts a message in a quote channel and posts a Christmas quote when one is due.
func (b bot) observeQuote(session *discordgo.Session, event *discordgo.MessageCreate) {
	if b.quotes == nil || b.setup == nil || !b.playerMessage(event) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	settings, err := b.setup.Settings(ctx, event.GuildID)
	cancel()
	if err != nil {
		log.Printf("Load quote channel settings: %v", err)
		return
	}
	if !settings.Channels[event.ChannelID].Quotes {
		return
	}
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	quote, err := b.quotes.Observe(ctx, event.GuildID, event.ChannelID)
	cancel()
	if err != nil {
		log.Printf("Draw Christmas quote: %v", err)
		return
	}
	if quote == nil {
		return
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := session.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
		Content: gifttext.Quote(*quote), AllowedMentions: &discordgo.MessageAllowedMentions{},
	}, discordgo.WithContext(ctx)); err != nil {
		log.Printf("Post Christmas quote: %v", err)
	}
}
