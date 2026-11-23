package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

var ErrInvalidRecipient = errors.New("choose another player in this server")

type Recipient struct {
	Member Member
	IsBot  bool
}

func (recipient Recipient) canReceiveFrom(sender Member) bool {
	target := recipient.Member
	return sender.ServerID != "" && sender.UserID != "" && target.ServerID == sender.ServerID &&
		target.UserID != "" && target != sender && !recipient.IsBot
}

// SendGift returns the Good deeds increase from the committed transfer.
func (service Service) SendGift(ctx context.Context, sender Member, recipient Recipient, rarity core.GiftRarity, quantity int64) (int, error) {
	if !recipient.canReceiveFrom(sender) {
		return 0, ErrInvalidRecipient
	}
	target := recipient.Member

	for ctx.Err() == nil {
		from, err := service.gifts.Load(ctx, sender)
		if err != nil {
			return 0, fmt.Errorf("load sender inventory: %w", err)
		}
		to, err := service.gifts.Load(ctx, target)
		if err != nil {
			return 0, fmt.Errorf("load recipient inventory: %w", err)
		}
		nextFrom, nextTo, err := core.TransferGift(from, to, rarity, quantity)
		if err != nil {
			return 0, err
		}
		saved, err := service.gifts.SaveTransfer(ctx,
			GiftUpdate{Member: sender, Previous: from, Next: nextFrom},
			GiftUpdate{Member: target, Previous: to, Next: nextTo})
		if err != nil {
			return 0, fmt.Errorf("save gift transfer: %w", err)
		}
		if saved {
			return nextFrom.Aura.Good - from.Aura.Good, nil
		}
	}
	return 0, ctx.Err()
}

func (service Service) CheckAura(ctx context.Context, member Member) (core.Aura, error) {
	record, err := service.gifts.Load(ctx, member)
	return record.Aura, err
}
