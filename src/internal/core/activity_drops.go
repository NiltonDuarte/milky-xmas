package core

import "math"

const (
	GiftDropSlots           = 8
	RandomActiveDropRolls   = 20
	GiftRainMinimumMessages = 5
	GiftRainMessageRange    = 96
	// GiftRainClaimLimit is how many collections end a Gift Rain.
	GiftRainClaimLimit = 10
	// RainCoalPercent is the chance that a Gift Rain collection gives Coal.
	RainCoalPercent = 25
	// GoodDeedsBoostThreshold is the Good deeds percentage that improves Random Active Drop rarity.
	GoodDeedsBoostThreshold = 50
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

// ActiveDropRarity takes a uniform roll in [0, GiftRarityRolls). Players at or above the
// Good deeds threshold have a better chance of rare and epic gifts.
func ActiveDropRarity(roll, goodDeeds int) GiftRarity {
	if goodDeeds < GoodDeedsBoostThreshold {
		return DailyGiftRarity(roll)
	}
	switch {
	case roll < 50:
		return NormalGift
	case roll < 85:
		return RareGift
	default:
		return EpicGift
	}
}

// RainPrizeIsCoal takes a uniform roll in [0, GiftRarityRolls).
func RainPrizeIsCoal(roll int) bool {
	return roll < RainCoalPercent
}

// AwardCoal adds one Coal without affecting gifts, aura, or cooldowns.
func AwardCoal(record GiftRecord) (GiftRecord, error) {
	if record.Inventory.Coal == math.MaxInt64 {
		return record, ErrCoalCapacity
	}
	record.Inventory.Coal++
	return record, nil
}
