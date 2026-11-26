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

type Inventory struct {
	Normal          int64
	Rare            int64
	Epic            int64
	Snowballs       int64
	YellowSnowballs int64
	Coal            int64
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

// Aura is where a player stands on one scale: negative values are bad deeds and positive
// values are good deeds. A new player is neutral at zero.
type Aura int

const (
	AuraMin Aura = -50
	AuraMax Aura = 50
)

// Move shifts the aura by points for each of quantity items, toward good when points is
// positive and toward bad when it is negative, without leaving the scale.
func (aura Aura) Move(quantity int64, points int) Aura {
	// Twenty items at five points or more each already cross the whole scale.
	shift := Aura(min(quantity, 20)) * Aura(points)
	return max(AuraMin, min(AuraMax, aura+shift))
}

type GiftRecord struct {
	Inventory Inventory
	LastClaim time.Time
	Aura      Aura
	// AuraDay is the UTC date AuraRaised and AuraLowered count for.
	AuraDay     string
	AuraRaised  int // Points the aura went up on AuraDay.
	AuraLowered int // Points the aura went down on AuraDay.
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
