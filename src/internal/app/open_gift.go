package app

import (
	"context"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// OpenedGift is what a player found in a gift.
type OpenedGift struct {
	Surprise core.Surprise
	// Quote is set when the surprise is a Christmas quote.
	Quote core.Quote
}

// OpenGift opens one of the member's gifts of the rarity. Opening does not move the aura.
func (service Service) OpenGift(ctx context.Context, member Member, rarity core.GiftRarity) (OpenedGift, error) {
	if member.ServerID == "" || member.UserID == "" {
		return OpenedGift{}, ErrInvalidRecipient
	}
	quotes := core.ChristmasQuotes()
	roll, quote := service.random(6), quotes[service.random(len(quotes))] // 6 fits two and three choices evenly.
	for ctx.Err() == nil {
		record, err := service.gifts.Load(ctx, member)
		if err != nil {
			return OpenedGift{}, fmt.Errorf("load inventory: %w", err)
		}
		next, surprise, err := core.OpenGift(record, rarity, roll)
		if err != nil {
			return OpenedGift{}, err
		}
		saved, err := service.gifts.Save(ctx, GiftUpdate{Member: member, Previous: record, Next: next, Reason: ReasonOpenGift})
		if err != nil {
			return OpenedGift{}, fmt.Errorf("save opened gift: %w", err)
		}
		if saved {
			opened := OpenedGift{Surprise: surprise}
			if surprise == core.SurpriseQuote {
				opened.Quote = quote
			}
			return opened, nil
		}
	}
	return OpenedGift{}, ctx.Err()
}
