package app

import (
	"context"
	"fmt"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type Member struct {
	ServerID string
	UserID   string
}

type GiftStore interface {
	Load(context.Context, Member) (core.GiftRecord, error)
	// Save and SaveTransfer compare the stored records with the observed snapshots.
	// Missing records are zero values. SaveTransfer applies both updates atomically.
	Save(context.Context, GiftUpdate) (bool, error)
	SaveTransfer(context.Context, GiftUpdate, GiftUpdate) (bool, error)
}

type GiftUpdate struct {
	Member   Member
	Previous core.GiftRecord
	Next     core.GiftRecord
}

type GiftClaim struct {
	Collected bool
	Rarity    core.GiftRarity
	Amount    int64
	Total     int64
	NextClaim time.Time
	Remaining time.Duration
}

func NewService(store Store, now func() time.Time, random func(int) int) Service {
	return Service{gifts: store, advent: store, now: now, random: random}
}

func (service Service) ClaimGift(ctx context.Context, member Member) (GiftClaim, error) {
	now := service.now().UTC()
	roll := service.random(core.MaxDailyGifts)
	rarity := core.DailyGiftRarity(service.random(core.GiftRarityRolls))
	for ctx.Err() == nil {
		record, err := service.gifts.Load(ctx, member)
		if err != nil {
			return GiftClaim{}, fmt.Errorf("load gift total: %w", err)
		}
		updated, next, collected, err := record.Collect(now, roll, rarity)
		if err != nil {
			return GiftClaim{}, err
		}
		if !collected {
			return GiftClaim{Total: record.Inventory.Total(), NextClaim: next, Remaining: next.Sub(now)}, nil
		}
		saved, err := service.gifts.Save(ctx, GiftUpdate{Member: member, Previous: record, Next: updated})
		if err != nil {
			return GiftClaim{}, fmt.Errorf("save gift claim: %w", err)
		}
		if saved {
			return GiftClaim{
				Collected: true,
				Rarity:    rarity,
				Amount:    int64(roll) + 1,
				Total:     updated.Inventory.Total(),
				NextClaim: next,
			}, nil
		}
	}
	return GiftClaim{}, ctx.Err()
}
