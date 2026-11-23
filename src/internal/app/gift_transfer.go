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

const CoalItem = "coal"

// SendItem returns the increase to the sender's Good deeds for gifts or Bad deeds for Coal.
func (service Service) SendItem(ctx context.Context, sender Member, recipient Recipient, item string, quantity int64) (int, error) {
	if item != CoalItem {
		return service.SendGift(ctx, sender, recipient, core.GiftRarity(item), quantity)
	}
	transfer, err := service.transfer(ctx, sender, recipient, func(from, to core.GiftRecord) (core.GiftRecord, core.GiftRecord, error) {
		return core.TransferCoal(from, to, quantity)
	})
	return transfer.auraIncrease.Bad, err
}

// SendGift returns the Good deeds increase from the committed transfer.
func (service Service) SendGift(ctx context.Context, sender Member, recipient Recipient, rarity core.GiftRarity, quantity int64) (int, error) {
	transfer, err := service.transfer(ctx, sender, recipient, func(from, to core.GiftRecord) (core.GiftRecord, core.GiftRecord, error) {
		return core.TransferGift(from, to, rarity, quantity)
	})
	return transfer.auraIncrease.Good, err
}

type inventoryTransfer struct {
	auraIncrease core.Aura
	recipient    core.Inventory
}

func (service Service) transfer(ctx context.Context, sender Member, recipient Recipient, decide func(core.GiftRecord, core.GiftRecord) (core.GiftRecord, core.GiftRecord, error)) (inventoryTransfer, error) {
	if !recipient.canReceiveFrom(sender) {
		return inventoryTransfer{}, ErrInvalidRecipient
	}
	target := recipient.Member
	for ctx.Err() == nil {
		from, err := service.gifts.Load(ctx, sender)
		if err != nil {
			return inventoryTransfer{}, fmt.Errorf("load sender inventory: %w", err)
		}
		to, err := service.gifts.Load(ctx, target)
		if err != nil {
			return inventoryTransfer{}, fmt.Errorf("load recipient inventory: %w", err)
		}
		nextFrom, nextTo, err := decide(from, to)
		if err != nil {
			return inventoryTransfer{}, err
		}
		saved, err := service.gifts.SaveTransfer(ctx,
			GiftUpdate{Member: sender, Previous: from, Next: nextFrom},
			GiftUpdate{Member: target, Previous: to, Next: nextTo})
		if err != nil {
			return inventoryTransfer{}, fmt.Errorf("save inventory transfer: %w", err)
		}
		if saved {
			return inventoryTransfer{
				auraIncrease: core.Aura{Good: nextFrom.Aura.Good - from.Aura.Good, Bad: nextFrom.Aura.Bad - from.Aura.Bad},
				recipient:    nextTo.Inventory,
			}, nil
		}
	}
	return inventoryTransfer{}, ctx.Err()
}

func (service Service) CheckAura(ctx context.Context, member Member) (core.Aura, error) {
	record, err := service.gifts.Load(ctx, member)
	return record.Aura, err
}
