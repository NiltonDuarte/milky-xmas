package discord

import (
	"context"
	_ "embed"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

//go:embed stocking.jpg
var stockingImage []byte

func (b bot) handleStocking(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivatelyWith(session, interaction, func(ctx context.Context, member app.Member) privateReply {
		items, err := b.service.ViewInventory(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "load inventory", err)
			return privateReply{Content: gifttext.InventoryUnavailable}
		}
		return privateReply{Content: gifttext.Inventory(items, true), Blocks: gifttext.InventorySections(items, true), Image: &privateImage{Name: "stocking.jpg", Data: stockingImage}}
	})
}
