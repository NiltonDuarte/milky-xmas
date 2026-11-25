package core

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestGrantGift(t *testing.T) {
	base := GiftRecord{Inventory: Inventory{Normal: 1, Coal: 4}, Aura: 2, LastClaim: time.Unix(10, 0)}
	full := GiftRecord{Inventory: Inventory{Epic: math.MaxInt64}}
	for _, test := range []struct {
		name     string
		record   GiftRecord
		rarity   GiftRarity
		quantity int64
		want     GiftRecord
		err      error
	}{
		{"rare", base, RareGift, 3, GiftRecord{Inventory: Inventory{Normal: 1, Rare: 3, Coal: 4}, Aura: base.Aura, LastClaim: base.LastClaim}, nil},
		{"unknown rarity", base, "mythic", 1, base, ErrGiftRarity},
		{"zero", base, NormalGift, 0, base, ErrGiftQuantity},
		{"negative", base, NormalGift, -2, base, ErrGiftQuantity},
		{"full", full, NormalGift, 1, full, ErrGiftCapacity},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := GrantGift(test.record, test.rarity, test.quantity)
			if !errors.Is(err, test.err) || got != test.want {
				t.Fatalf("GrantGift = %+v, %v; want %+v, %v", got, err, test.want, test.err)
			}
		})
	}
}
