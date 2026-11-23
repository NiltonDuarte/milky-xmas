package gifttext

import (
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const CoalEmoji = "🌑"

const InventoryUnavailable = "Could not load your inventory. Try again."

func Inventory(items core.Inventory) string {
	return fmt.Sprintf("🧦 Your inventory\n🎁 Gifts: %d - normal | %d rare | %d epic\n%s Snowballs: %d | %s Yellow Snowballs: %d\n%s Coal: %d",
		items.Normal, items.Rare, items.Epic, SnowballEmoji(core.NormalSnowball), items.Snowballs, SnowballEmoji(core.YellowSnowball), items.YellowSnowballs, CoalEmoji, items.Coal)
}
