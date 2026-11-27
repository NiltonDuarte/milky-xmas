package core

import "math"

// Surprise is what comes out of an opened gift.
type Surprise string

const (
	SurpriseSnowball      Surprise = "snowball"
	SurpriseCoal          Surprise = "coal"
	SurpriseQuote         Surprise = "quote"
	SurpriseSnowballs     Surprise = "snowballs"
	SurpriseExtraWatering Surprise = "extra_watering"
	SurpriseShield        Surprise = "shield"
	SurpriseGifts         Surprise = "gifts"
)

// How much the surprises that give items hold.
const (
	SurpriseSnowballCount = 3 // Snowballs in a rare gift.
	SurpriseGiftCount     = 2 // Normal gifts in an epic gift.
)

// surprises are what each rarity can hold, each with the same chance.
var surprises = map[GiftRarity][]Surprise{
	NormalGift: {SurpriseSnowball, SurpriseCoal, SurpriseQuote},
	RareGift:   {SurpriseSnowballs, SurpriseExtraWatering},
	EpicGift:   {SurpriseShield, SurpriseGifts},
}

// OpenGift takes one gift of the rarity out of the record and puts its surprise in. The roll is
// any number from zero up. A shield or an extra watering is kept one at a time, so a player who
// already has it gets the other surprise of the rarity.
func OpenGift(record GiftRecord, rarity GiftRarity, roll int) (GiftRecord, Surprise, error) {
	choices := surprises[rarity]
	balance := record.Inventory.balance(rarity)
	switch {
	case len(choices) == 0 || balance == nil || roll < 0:
		return record, "", ErrGiftRarity
	case *balance <= 0:
		return record, "", ErrInsufficientGift
	}
	*balance--
	surprise := choices[roll%len(choices)]
	switch {
	case surprise == SurpriseShield && record.Inventory.Shields > 0:
		surprise = SurpriseGifts
	case surprise == SurpriseExtraWatering && record.Inventory.ExtraWaterings > 0:
		surprise = SurpriseSnowballs
	}
	items := &record.Inventory
	add := func(count *int64, amount int64) { *count += min(amount, math.MaxInt64-*count) }
	switch surprise {
	case SurpriseSnowball:
		add(&items.Snowballs, 1)
	case SurpriseCoal:
		add(&items.Coal, 1)
	case SurpriseSnowballs:
		add(&items.Snowballs, SurpriseSnowballCount)
	case SurpriseExtraWatering:
		items.ExtraWaterings = 1
	case SurpriseShield:
		items.Shields = 1
	case SurpriseGifts:
		items.Normal += min(SurpriseGiftCount, items.remainingGiftCapacity())
	}
	return record, surprise, nil
}
