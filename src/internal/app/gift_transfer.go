package app

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// AuraEffect says whether a send moved the sender's aura, and if not, why.
type AuraEffect int

const (
	AuraMoved    AuraEffect = iota // Moved, or already at the end of the scale.
	AuraRepeated                   // A send to this player already counted.
	AuraLimited                    // The aura already moved as far as it can today.
)

// SendItem sends gifts, which move the sender's aura toward good, or Coal, which moves it
// toward bad.
func (service Service) SendItem(ctx context.Context, sender Member, recipient Recipient, item string, quantity int64) (AuraEffect, error) {
	reason, decide := ReasonSendCoal, func(from, to core.GiftRecord) (core.GiftRecord, core.GiftRecord, error) {
		return core.TransferCoal(from, to, quantity)
	}
	if item != CoalItem {
		reason, decide = ReasonSendGift, func(from, to core.GiftRecord) (core.GiftRecord, core.GiftRecord, error) {
			return core.TransferGift(from, to, core.GiftRarity(item), quantity)
		}
	}
	transfer, err := service.transfer(ctx, sender, recipient, reason, decide)
	return transfer.aura, err
}

func (service Service) SendGift(ctx context.Context, sender Member, recipient Recipient, rarity core.GiftRarity, quantity int64) error {
	_, err := service.SendItem(ctx, sender, recipient, string(rarity), quantity)
	return err
}

type inventoryTransfer struct {
	recipient core.Inventory
	aura      AuraEffect
}

func (service Service) transfer(ctx context.Context, sender Member, recipient Recipient, reason Reason, decide func(core.GiftRecord, core.GiftRecord) (core.GiftRecord, core.GiftRecord, error)) (inventoryTransfer, error) {
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
		var send *AuraSend
		effect := AuraMoved
		if nextFrom.Aura != from.Aura {
			now, good := service.now().UTC(), nextFrom.Aura > from.Aura
			cooldown := service.tuning.Aura.SendCooldown
			var last time.Time
			if cooldown > 0 {
				if last, err = service.gifts.LastAuraSend(ctx, sender, target.UserID, good); err != nil {
					return inventoryTransfer{}, fmt.Errorf("load aura sends: %w", err)
				}
			}
			if !last.IsZero() && now.Sub(last) < cooldown {
				nextFrom.Aura, effect = from.Aura, AuraRepeated
			} else if nextFrom = service.tuning.Aura.LimitDaily(from, nextFrom, now); nextFrom.Aura == from.Aura {
				effect = AuraLimited
			} else if cooldown > 0 {
				send = &AuraSend{RecipientID: target.UserID, Good: good, At: now}
			}
		}
		saved, err := service.gifts.SaveTransfer(ctx,
			GiftUpdate{Member: sender, Previous: from, Next: nextFrom, Reason: reason, AuraSend: send},
			GiftUpdate{Member: target, Previous: to, Next: nextTo, Reason: reason})
		if err != nil {
			return inventoryTransfer{}, fmt.Errorf("save inventory transfer: %w", err)
		}
		if saved {
			return inventoryTransfer{recipient: nextTo.Inventory, aura: effect}, nil
		}
	}
	return inventoryTransfer{}, ctx.Err()
}

func (service Service) CheckAura(ctx context.Context, member Member) (core.Aura, error) {
	record, err := service.gifts.Load(ctx, member)
	return record.Aura, err
}
