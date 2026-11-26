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
	// Reason says what changed the inventory, so stores can keep a history for scoreboards.
	Reason Reason
}

// Reason is what changed a player's inventory or aura.
type Reason string

const (
	ReasonDaily         Reason = "daily_claim"
	ReasonDrop          Reason = "drop_claim"
	ReasonSendGift      Reason = "send_gift"
	ReasonSendCoal      Reason = "send_coal"
	ReasonThrowSnowball Reason = "throw_snowball"
	ReasonModGift       Reason = "mod_gift"
	ReasonGrinchReward  Reason = "grinch_reward"
	ReasonWater         Reason = "water"
)

type GiftClaim struct {
	Collected bool
	Rarity    core.GiftRarity
	Amount    int64
	Total     int64
	NextClaim time.Time
	Remaining time.Duration
}

func NewService(store Store, now func() time.Time, random func(int) int) Service {
	return Service{gifts: store, advent: store, auras: store, stats: store, now: now, random: random, tuning: core.DefaultTuning()}
}

func (service Service) ClaimGift(ctx context.Context, member Member) (GiftClaim, error) {
	now := service.now().UTC()
	roll := service.random(core.MaxDailyGifts)
	rarity := service.tuning.Gifts.DailyGiftRarity(service.random(core.GiftRarityRolls))
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
			return GiftClaim{Total: record.Inventory.GiftTotal(), NextClaim: next, Remaining: next.Sub(now)}, nil
		}
		saved, err := service.gifts.Save(ctx, GiftUpdate{Member: member, Previous: record, Next: updated, Reason: ReasonDaily})
		if err != nil {
			return GiftClaim{}, fmt.Errorf("save gift claim: %w", err)
		}
		if saved {
			return GiftClaim{
				Collected: true,
				Rarity:    rarity,
				Amount:    int64(roll) + 1,
				Total:     updated.Inventory.GiftTotal(),
				NextClaim: next,
			}, nil
		}
	}
	return GiftClaim{}, ctx.Err()
}
