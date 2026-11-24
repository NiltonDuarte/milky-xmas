package core

import "math"

const (
	GiftDropSlots           = 8
	RandomActiveDropRolls   = 20
	GiftRainMinimumMessages = 5
	GiftRainMessageRange    = 96
)

type GiftDropKind string

const (
	RandomActiveDrop GiftDropKind = "active"
	GiftRainDrop     GiftDropKind = "rain"
)

// RandomActiveDropEligible reports whether a qualifying activity message starts a random drop.
func RandomActiveDropEligible(roll int) bool {
	return roll == 0
}

// NextGiftRainThreshold converts a uniform roll in [0, GiftRainMessageRange)
// to the inclusive activity range requested for Gift Rain.
func NextGiftRainThreshold(roll int) int {
	return GiftRainMinimumMessages + roll
}

// AwardGift adds one gift without affecting the daily-claim cooldown or aura.
func AwardGift(record GiftRecord, rarity GiftRarity) (GiftRecord, error) {
	balance := record.Inventory.balance(rarity)
	if balance == nil {
		return record, ErrGiftRarity
	}
	if *balance == math.MaxInt64 || record.Inventory.GiftTotal() == math.MaxInt64 {
		return record, ErrGiftCapacity
	}
	*balance++
	return record, nil
}
