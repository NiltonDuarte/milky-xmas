package core

import (
	"errors"
	"math"
	"time"
)

const GiftCooldown = 24 * time.Hour
const MaxDailyGifts = 10

type GiftRarity string

const (
	NormalGift GiftRarity = "normal"
	RareGift   GiftRarity = "rare"
	EpicGift   GiftRarity = "epic"
)

const GiftRarityRolls = 100

// DailyGiftRarity takes a uniform roll in [0, GiftRarityRolls).
func DailyGiftRarity(roll int) GiftRarity {
	switch {
	case roll < 70:
		return NormalGift
	case roll < 95:
		return RareGift
	default:
		return EpicGift
	}
}

type Inventory struct {
	Normal    int64
	Rare      int64
	Epic      int64
	Snowballs int64
	Coal      int64
}

func (inventory Inventory) GiftTotal() int64 {
	return inventory.Normal + inventory.Rare + inventory.Epic
}

func (inventory Inventory) remainingGiftCapacity() int64 {
	return math.MaxInt64 - inventory.GiftTotal()
}

func (inventory *Inventory) balance(rarity GiftRarity) *int64 {
	switch rarity {
	case NormalGift:
		return &inventory.Normal
	case RareGift:
		return &inventory.Rare
	case EpicGift:
		return &inventory.Epic
	default:
		return nil
	}
}

type Aura struct {
	Bad  int
	Good int
}

type GiftRecord struct {
	Inventory Inventory
	LastClaim time.Time
	Aura      Aura
}

var (
	ErrGiftRarity       = errors.New("unknown gift rarity")
	ErrGiftQuantity     = errors.New("gift quantity must be positive")
	ErrInsufficientGift = errors.New("insufficient gifts of that rarity")
	ErrGiftCapacity     = errors.New("gift inventory is full")
)

// Collect takes a uniform quantity roll in [0, MaxDailyGifts).
func (record GiftRecord) Collect(now time.Time, roll int, rarity GiftRarity) (GiftRecord, time.Time, bool, error) {
	if !record.LastClaim.IsZero() {
		next := record.LastClaim.Add(GiftCooldown)
		if now.Before(next) {
			return record, next, false, nil
		}
	}
	balance := record.Inventory.balance(rarity)
	if balance == nil {
		return record, time.Time{}, false, ErrGiftRarity
	}
	amount := int64(roll) + 1
	if amount > record.Inventory.remainingGiftCapacity() {
		return record, time.Time{}, false, ErrGiftCapacity
	}
	*balance += amount
	record.LastClaim = now
	return record, now.Add(GiftCooldown), true, nil
}
