package gifttext

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// Aura draws the scale from bad to good with a marker fixed on its middle and a finger pointing
// up at the player's place. Both sides carry a number, as the player's value on their side and zero on
// the other. The ends of the scale are not shown to players.
func Aura(aura core.Aura) string {
	const cells, pointsPerCell = 5, 10
	// Box-drawing cells are one column wide in code blocks, so the finger lines up.
	bar := []rune(strings.Repeat("─", 2*cells+1))
	magnitude := int(aura)
	if magnitude < 0 {
		magnitude = -magnitude
	}
	step := min(magnitude/pointsPerCell, cells-1) + 1
	bar[cells] = '●' // The marker always stands for the neutral middle.
	place, bad, good := cells, "0", "0"
	switch {
	case aura < 0:
		place, bad = cells-step, fmt.Sprintf("%d", aura)
	case aura > 0:
		place, good = cells+step, fmt.Sprintf("+%d", aura)
	}
	left := bad + " 😈 "
	// The emoji takes two columns but counts as one character, so this is the place's column.
	// The fingertip sits on the left of 👆, which therefore starts right there.
	indent := strings.Repeat(" ", utf8.RuneCountInString(left)+1+place)
	return fmt.Sprintf("```\n%s%s 😇 %s\n%s👆🏻\n```", left, string(bar), good, indent)
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
