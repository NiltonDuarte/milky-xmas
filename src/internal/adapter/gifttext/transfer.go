package gifttext

import (
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func Aura(aura core.Aura) string {
	return fmt.Sprintf("😈 Bad %d%%  |  😇 Good %d%%", aura.Bad, aura.Good)
}

func Transfer(item string, quantity int64, recipient string, increase int) string {
	if item == app.CoalItem {
		return fmt.Sprintf("Sent %d %s Coal to %s. Your 😈 bad deeds increased by %d%%.", quantity, CoalEmoji, recipient, increase)
	}
	gift := "gifts"
	if quantity == 1 {
		gift = "gift"
	}
	return fmt.Sprintf("Sent %d 🎁%s %s to %s. Your 😇 good deeds is %d%% more high.", quantity, item, gift, recipient, increase)
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
