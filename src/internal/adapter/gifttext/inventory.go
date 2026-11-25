package gifttext

import (
	"fmt"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const CoalEmoji = "🌑"

const InventoryUnavailable = "Could not load your inventory. Try again."

// ANSI colors Discord renders inside an ansi code block.
const (
	green = "32"
	pink  = "35" // Discord's closest color to purple.
	white = "37"
)

// Inventory is the whole stocking as one text, zeros included. colored asks for the Discord
// word colors.
func Inventory(items core.Inventory, colored bool) string {
	return strings.Join(InventorySections(items, colored), "\n")
}

// InventorySections returns one section per kind of item, each with a large title, so an
// adapter can draw a line between them. Sections need no alignment with each other.
func InventorySections(items core.Inventory, colored bool) []string {
	language := ""
	if colored {
		language = "ansi"
	}
	paint := func(color, text string) string {
		if !colored {
			return text
		}
		return "\x1b[1;" + color + "m" + text + "\x1b[0m"
	}
	// row pads after the word, so the numbers of a section line up.
	row := func(emoji, word, color string, value int64) string {
		return emoji + paint(color, word) + fmt.Sprintf("%*d", 10-len(word), value)
	}
	section := func(title string, rows ...string) string {
		return "### " + title + "\n```" + language + "\n" + strings.Join(rows, "\n") + "\n```"
	}
	snow := SnowballEmoji(core.NormalSnowball)
	return []string{
		section("🎁 Gifts", row("", "Normal", white, items.Normal), row("", "Rare", green, items.Rare), row("", "Epic", pink, items.Epic)),
		section(snow+" Snowballs", row(snow+" ", "Normal", white, items.Snowballs),
			row(SnowballEmoji(core.YellowSnowball)+" ", "Yellow", white, items.YellowSnowballs)),
		section(CoalEmoji+" Coal", paint(white, fmt.Sprint(items.Coal))),
	}
}
