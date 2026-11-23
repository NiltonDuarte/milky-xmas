package app

import (
	"context"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func (service Service) ThrowSnowball(ctx context.Context, sender Member, recipient Recipient, kind core.SnowballType) (int64, error) {
	switch kind {
	case core.YellowSnowball:
		transfer, err := service.transfer(ctx, sender, recipient, core.TransferYellowSnowball)
		return transfer.recipient.YellowSnowballs, err
	case core.NormalSnowball:
		// Regular throws credit a Snowball without debiting the sender.
	default:
		return 0, core.ErrSnowballType
	}
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
