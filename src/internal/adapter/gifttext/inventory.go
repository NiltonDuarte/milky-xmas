package gifttext

import (
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const InventoryUnavailable = "Could not load your inventory. Try again."

func Inventory(items core.Inventory) string {
	return fmt.Sprintf("🧦 Your inventory\nNormal gifts: %d\nRare gifts: %d\nEpic gifts: %d\nSnowballs: %d\nYellow Snowballs: %d\nCoal: %d",
		items.Normal, items.Rare, items.Epic, items.Snowballs, items.YellowSnowballs, items.Coal)
}
