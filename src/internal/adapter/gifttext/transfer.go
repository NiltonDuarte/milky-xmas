package gifttext

import (
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func Aura(aura core.Aura) string {
	return fmt.Sprintf("Bad deeds %d%%  |  Good deeds %d%%", aura.Bad, aura.Good)
}

func Transfer(rarity core.GiftRarity, quantity int64, aura core.Aura) string {
	gift := "gifts"
	if quantity == 1 {
		gift = "gift"
	}
	return fmt.Sprintf("Sent %d %s %s. Good deeds %d%%.", quantity, rarity, gift, aura.Good)
}

func TransferRejection(err error) string {
	switch {
	case errors.Is(err, app.ErrGiftRecipient):
		return "Choose another player in this server."
	case errors.Is(err, core.ErrGiftRarity):
		return "Choose a gift rarity: normal, rare, or epic."
	case errors.Is(err, core.ErrGiftQuantity):
		return "Enter a positive whole number of gifts."
	case errors.Is(err, core.ErrInsufficientGift):
		return "You don't have enough gifts of that rarity."
	case errors.Is(err, core.ErrGiftCapacity):
		return "The recipient's gift inventory is full."
	default:
		return ""
	}
}
