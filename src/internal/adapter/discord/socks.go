package discord

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

func (b bot) handleStocking(session *discordgo.Session, interaction *discordgo.Interaction) {
	respondPrivately(session, interaction, func(ctx context.Context, member app.Member) string {
		items, err := b.service.ViewInventory(ctx, member)
		if err != nil {
			logInteractionError(interaction.ID, "load inventory", err)
			return gifttext.InventoryUnavailable
		}
		return gifttext.Inventory(items)
	})
}
