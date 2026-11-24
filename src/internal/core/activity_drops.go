package core

import "math"

type GiftDropKind string

const (
	RandomActiveDrop GiftDropKind = "active"
	GiftRainDrop     GiftDropKind = "rain"
)

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

// AwardCoal adds one Coal without affecting gifts, aura, or cooldowns.
func AwardCoal(record GiftRecord) (GiftRecord, error) {
	if record.Inventory.Coal == math.MaxInt64 {
		return record, ErrCoalCapacity
	}
	record.Inventory.Coal++
	return record, nil
}
