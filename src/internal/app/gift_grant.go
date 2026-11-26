package app

import (
	"context"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// GiveGift adds gifts to the recipient on behalf of a moderator. Callers check moderator permission.
func (service Service) GiveGift(ctx context.Context, moderator Member, recipient Recipient, rarity core.GiftRarity, quantity int64) error {
	if !recipient.canReceiveFrom(moderator) {
		return ErrInvalidRecipient
	}
	target := recipient.Member
	for ctx.Err() == nil {
		record, err := service.gifts.Load(ctx, target)
		if err != nil {
			return fmt.Errorf("load recipient inventory: %w", err)
		}
		next, err := core.GrantGift(record, rarity, quantity)
		if err != nil {
			return err
		}
		saved, err := service.gifts.Save(ctx, GiftUpdate{Member: target, Previous: record, Next: next, Reason: ReasonModGift})
		if err != nil {
			return fmt.Errorf("save gift grant: %w", err)
		}
		if saved {
			return nil
		}
	}
	return ctx.Err()
}
