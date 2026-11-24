package discord

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const (
	randomActiveDropMessage = "Watch out! Something is falling down!"
	giftRainMessage         = "🎁 It's raining gifts! 🎁 Hallelujah"
)

type dropBot struct{ drops *app.ActivityDrops }

func newDropBot(drops *app.ActivityDrops) *dropBot { return &dropBot{drops: drops} }

func (b bot) observeDrop(session *discordgo.Session, event *discordgo.MessageCreate) {
	if b.drops == nil || b.drops.drops == nil || event == nil || event.Author == nil || event.Author.Bot ||
		event.WebhookID != "" || !b.permitsGuild(event.GuildID) {
		return
	}
	if b.setup != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		settings, err := b.setup.Settings(ctx, event.GuildID)
		cancel()
		if err != nil {
			log.Printf("Load activity-drop channel settings: %v", err)
			return
		}
		if !settings.Channels[event.ChannelID].Commands {
			return
		}
	}
	drop := b.drops.drops.Observe(event.GuildID, event.ChannelID)
	if drop == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := session.ChannelMessageSendComplex(event.ChannelID, dropMessage(*drop), discordgo.WithContext(ctx)); err != nil {
		log.Printf("Announce activity gift drop: %v", err)
	}
}

func dropMessage(drop app.GiftDrop) *discordgo.MessageSend {
	content := randomActiveDropMessage
	if drop.Kind == core.GiftRainDrop {
		content = giftRainMessage
	}
	rows := make([]discordgo.MessageComponent, 0, 2)
	for row := range 2 {
		buttons := make([]discordgo.MessageComponent, 0, 4)
		for column := range 4 {
			slot := row*4 + column
			buttons = append(buttons, discordgo.Button{
				Label: "🎁", Style: discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("drop:%s:%d", drop.ID, slot),
			})
		}
		rows = append(rows, discordgo.ActionsRow{Components: buttons})
	}
	return &discordgo.MessageSend{Content: content, Components: rows, AllowedMentions: &discordgo.MessageAllowedMentions{}}
}

func (b *dropBot) button(session *discordgo.Session, interaction *discordgo.Interaction) {
	if b == nil || b.drops == nil || interaction == nil || interaction.Member == nil || interaction.Member.User == nil {
		return
	}
	data, ok := interaction.Data.(discordgo.MessageComponentInteractionData)
	if !ok {
		return
	}
	parts := strings.Split(data.CustomID, ":")
	if len(parts) != 3 || parts[0] != "drop" || parts[1] == "" {
		return
	}
	slot, err := strconv.Atoi(parts[2])
	if err != nil || slot < 0 || slot >= core.GiftDropSlots {
		return
	}
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		award, err := b.drops.Claim(ctx, app.GiftDropClaim{Member: member, DropID: parts[1]})
		if err == nil {
			return fmt.Sprintf("You collected a %s gift.", award.Rarity)
		}
		switch {
		case errors.Is(err, app.ErrDropCooldown):
			return "You cannot collect another gift yet."
		case errors.Is(err, app.ErrDropCollected):
			return "You already collected a gift from this drop."
		case errors.Is(err, app.ErrDropUnavailable):
			return "This gift drop is no longer available."
		default:
			logInteractionError(interaction.ID, "collect activity gift", err)
			return "Could not collect this gift. Try again."
		}
	})
}
