package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

var ErrGiftRecipient = errors.New("choose another player in this server")

type GiftRecipient struct {
	Member Member
	IsBot  bool
}

func (service Service) SendGift(ctx context.Context, sender Member, recipient GiftRecipient, rarity core.GiftRarity, quantity int64) (core.Aura, error) {
	target := recipient.Member
	if sender.ServerID == "" || sender.UserID == "" || target.ServerID != sender.ServerID ||
		target.UserID == "" || target == sender || recipient.IsBot {
		return core.Aura{}, ErrGiftRecipient
	}

	for ctx.Err() == nil {
		from, err := service.gifts.Load(ctx, sender)
		if err != nil {
			return core.Aura{}, fmt.Errorf("load sender inventory: %w", err)
		}
		to, err := service.gifts.Load(ctx, target)
		if err != nil {
			return core.Aura{}, fmt.Errorf("load recipient inventory: %w", err)
		}
		nextFrom, nextTo, err := core.TransferGift(from, to, rarity, quantity)
		if err != nil {
			return core.Aura{}, err
		}
		saved, err := service.gifts.SaveTransfer(ctx,
			GiftUpdate{Member: sender, Previous: from, Next: nextFrom},
			GiftUpdate{Member: target, Previous: to, Next: nextTo})
		if err != nil {
			return core.Aura{}, fmt.Errorf("save gift transfer: %w", err)
		}
		if saved {
			return nextFrom.Aura, nil
		}
	}
	return core.Aura{}, ctx.Err()
}

func (service Service) CheckAura(ctx context.Context, member Member) (core.Aura, error) {
	record, err := service.gifts.Load(ctx, member)
	return record.Aura, err
}
