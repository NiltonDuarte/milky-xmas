package app

import (
	"context"
	"fmt"
)

func (service Service) ThrowSnowball(ctx context.Context, sender Member, recipient Recipient) (int64, error) {
	if !recipient.canReceiveFrom(sender) {
		return 0, ErrInvalidRecipient
	}
	for ctx.Err() == nil {
		previous, err := service.gifts.Load(ctx, recipient.Member)
		if err != nil {
			return 0, fmt.Errorf("load snowball recipient inventory: %w", err)
		}
		inventory, err := previous.Inventory.ReceiveSnowball()
		if err != nil {
			return 0, err
		}
		next := previous
		next.Inventory = inventory
		saved, err := service.gifts.Save(ctx, GiftUpdate{Member: recipient.Member, Previous: previous, Next: next})
		if err != nil {
			return 0, fmt.Errorf("save snowball: %w", err)
		}
		if saved {
			return inventory.Snowballs, nil
		}
	}
	return 0, ctx.Err()
}
