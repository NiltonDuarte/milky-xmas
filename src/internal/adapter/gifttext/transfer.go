package gifttext

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// Aura draws the scale from -50 to +50 with a marker on the player's place and the number
// under it. Each cell on either side of the centre covers ten points.
func Aura(aura core.Aura) string {
	const cells, pointsPerCell = 5, 10
	bar := []rune(strings.Repeat("▱", cells) + "|" + strings.Repeat("▱", cells))
	magnitude := int(aura)
	if magnitude < 0 {
		magnitude = -magnitude
	}
	step := min(magnitude/pointsPerCell, cells-1) + 1
	marker, label := cells, "0"
	switch {
	case aura < 0:
		marker, label = cells-step, fmt.Sprintf("%d", aura)
	case aura > 0:
		marker, label = cells+step, fmt.Sprintf("+%d", aura)
	}
	bar[marker] = '●'
	left := fmt.Sprintf("😈 %d ", core.AuraMin)
	// The emoji takes two columns in a code block but counts as one character.
	column := utf8.RuneCountInString(left) + 1 + marker
	indent := strings.Repeat(" ", column-len(label)/2)
	return fmt.Sprintf("```\n%s%s +%d 😇\n%s%s\n```", left, string(bar), core.AuraMax, indent, label)
}

func Transfer(item string, quantity int64, recipient string) string {
	if item == app.CoalItem {
		return fmt.Sprintf("Sent %d %s Coal to %s. Your 😈 bad deeds are increasing. Beware!", quantity, CoalEmoji, recipient)
	}
	gift := "gifts"
	if quantity == 1 {
		gift = "gift"
	}
	return fmt.Sprintf("Sent %d 🎁 %s %s to %s. Your 😇 good deeds are going up.", quantity, item, gift, recipient)
}

func TransferRejection(err error) string {
	switch {
	case errors.Is(err, app.ErrInvalidRecipient):
		return "Choose another player in this server."
	case errors.Is(err, core.ErrGiftRarity):
		return "Choose normal, rare, epic, or coal."
	case errors.Is(err, core.ErrGiftQuantity):
		return "Enter a positive whole number of items."
	case errors.Is(err, core.ErrInsufficientCoal):
		return "You don't have enough Coal."
	case errors.Is(err, core.ErrCoalCapacity):
		return "The recipient's Coal inventory is full."
	case errors.Is(err, core.ErrInsufficientGift):
		return "You don't have enough gifts of that rarity."
	case errors.Is(err, core.ErrGiftCapacity):
		return "The recipient's gift inventory is full."
	default:
		return ""
	}
}

func GiftGiven(quantity int64, rarity, recipient string) string {
	gift := "gifts"
	if quantity == 1 {
		gift = "gift"
	}
	return fmt.Sprintf("Gave %d 🎁 %s %s to %s.", quantity, rarity, gift, recipient)
}

const GiveNeedsModerator = "You need Manage Messages permission to give gifts."
