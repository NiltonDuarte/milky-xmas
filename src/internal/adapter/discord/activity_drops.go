package discord

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const (
	randomActiveDropMessage = "Watch out! Something is falling down!"
	giftRainMessage         = "🎁 It's raining gifts! 🎁 Hallelujah"
)

//go:embed grinchmax.png
var activeDropImage []byte

//go:embed giftrain.webp
var giftRainImage []byte

type dropBot struct {
	drops *app.ActivityDrops
	// lifetime is how long a drop stays open. Tests shorten it.
	lifetime time.Duration
}

func newDropBot(drops *app.ActivityDrops) *dropBot {
	return &dropBot{drops: drops, lifetime: drops.Lifetime()}
}

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
	sent, err := session.ChannelMessageSendComplex(event.ChannelID, dropMessage(*drop), discordgo.WithContext(ctx))
	if err != nil {
		log.Printf("Announce activity gift drop: %v", err)
		return
	}
	dropID, channelID, messageID := drop.ID, event.ChannelID, sent.ID
	time.AfterFunc(b.drops.lifetime, func() { b.drops.expire(session, channelID, messageID, dropID) })
}

// expire ends a drop whose time ran out and lists who collected from it.
func (b *dropBot) expire(session *discordgo.Session, channelID, messageID, dropID string) {
	summary, ok := b.drops.Close(dropID)
	if !ok {
		return // Already finished, for example when the tenth Gift Rain gift was collected.
	}
	showDropEnd(session, channelID, messageID, summary)
}

func showDropEnd(session *discordgo.Session, channelID, messageID string, summary app.DropSummary) {
	content := gifttext.DropExpired(summary.Claims, func(userID string) string { return "<@" + userID + ">" })
	components := []discordgo.MessageComponent{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID: messageID, Channel: channelID, Content: &content, Components: &components,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}, discordgo.WithContext(ctx))
	if err != nil {
		log.Printf("Show expired gift drop: %v", err)
	}
}

func dropMessage(drop app.GiftDrop) *discordgo.MessageSend {
	content, name, image, contentType := randomActiveDropMessage, "grinchmax.png", activeDropImage, "image/png"
	if drop.Kind == core.GiftRainDrop {
		content, name, image, contentType = giftRainMessage, "giftrain.webp", giftRainImage, "image/webp"
	}
	return &discordgo.MessageSend{
		Content:         content,
		Components:      dropComponents(drop.ID, 0),
		Embeds:          []*discordgo.MessageEmbed{{Image: &discordgo.MessageEmbedImage{URL: "attachment://" + name}}},
		Files:           []*discordgo.File{{Name: name, ContentType: contentType, Reader: bytes.NewReader(image)}},
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}
}

// dropComponents renders two rows of four gift buttons. Slots set in taken are disabled.
func dropComponents(dropID string, taken uint8) []discordgo.MessageComponent {
	rows := make([]discordgo.MessageComponent, 0, 2)
	for row := range 2 {
		buttons := make([]discordgo.MessageComponent, 0, 4)
		for column := range 4 {
			slot := row*4 + column
			buttons = append(buttons, discordgo.Button{
				Label: "🎁", Style: discordgo.SecondaryButton, Disabled: taken&(1<<slot) != 0,
				CustomID: fmt.Sprintf("drop:%s:%d", dropID, slot),
			})
		}
		rows = append(rows, discordgo.ActionsRow{Components: buttons})
	}
	return rows
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
		award, err := b.drops.Claim(ctx, app.GiftDropClaim{Member: member, DropID: parts[1], Slot: slot + 1})
		if err != nil {
			if reply := gifttext.DropRejection(err); reply != "" {
				return reply
			}
			logInteractionError(interaction.ID, "collect activity gift", err)
			return "Could not collect this gift. Try again."
		}
		if message := interaction.Message; message != nil {
			switch {
			case award.Finished != nil:
				showDropEnd(session, interaction.ChannelID, message.ID, *award.Finished)
			case award.Kind == core.RandomActiveDrop:
				components := dropComponents(parts[1], award.Taken)
				if _, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
					ID: message.ID, Channel: interaction.ChannelID, Components: &components,
				}, discordgo.WithContext(ctx)); err != nil {
					logInteractionError(interaction.ID, "disable the opened gift", err)
				}
			}
		}
		return gifttext.DropCollected(award)
	})
}
