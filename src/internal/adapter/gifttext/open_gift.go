package gifttext

import (
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const (
	OpenCommand     = "open"
	OpenUnavailable = "Could not open your gift. Try again."
)

// OpenedGift tells a player what was in the gift of the rarity.
func OpenedGift(rarity core.GiftRarity, opened app.OpenedGift) string {
	var found string
	switch opened.Surprise {
	case core.SurpriseSnowball:
		found = "❄️ You found 1 Snowball."
	case core.SurpriseCoal:
		found = CoalEmoji + " You found 1 Coal."
	case core.SurpriseQuote:
		found = Quote(opened.Quote)
	case core.SurpriseSnowballs:
		found = fmt.Sprintf("❄️ You found %d Snowballs.", core.SurpriseSnowballCount)
	case core.SurpriseExtraWatering:
		found = "💧 You found an extra watering! Your next /water can skip the wait."
	case core.SurpriseShield:
		found = "🛡️ You found a shield! It blocks the Grinch's next hit."
	case core.SurpriseGifts:
		found = fmt.Sprintf("🎁 You found %d normal gifts.", core.SurpriseGiftCount)
	}
	return fmt.Sprintf("🎁 You opened a %s gift!\n%s", rarity, found)
}

// OpenRejection explains why a gift was not opened. It is empty for unexpected errors.
func OpenRejection(err error) string {
	switch {
	case errors.Is(err, core.ErrGiftRarity):
		return "Choose normal, rare, or epic."
	case errors.Is(err, core.ErrInsufficientGift):
		return "You don't have a gift of that rarity."
	default:
		return ""
	}
}
