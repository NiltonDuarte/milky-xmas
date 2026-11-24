package core

import (
	"errors"
	"math"
	"testing"
)

func TestActivityDropRules(t *testing.T) {
	for _, tt := range []struct {
		roll int
		want bool
	}{
		{0, true},
		{1, false},
		{RandomActiveDropRolls - 1, false},
	} {
		if got := RandomActiveDropEligible(tt.roll); got != tt.want {
			t.Errorf("active drop roll %d = %v, want %v", tt.roll, got, tt.want)
		}
	}
	for _, tt := range []struct {
		roll int
		want int
	}{
		{0, GiftRainMinimumMessages},
		{GiftRainMessageRange - 1, 100},
	} {
		if got := NextGiftRainThreshold(tt.roll); got != tt.want {
			t.Errorf("rain threshold roll %d = %d, want %d", tt.roll, got, tt.want)
		}
	}
}

func TestAwardGift(t *testing.T) {
	for _, rarity := range []GiftRarity{NormalGift, RareGift, EpicGift} {
		t.Run(string(rarity), func(t *testing.T) {
			record, err := AwardGift(GiftRecord{}, rarity)
			if err != nil || *record.Inventory.balance(rarity) != 1 {
				t.Fatalf("award = %+v, %v", record, err)
			}
		})
	}
	full := GiftRecord{Inventory: Inventory{Normal: math.MaxInt64}}
	updated, err := AwardGift(full, NormalGift)
	if !errors.Is(err, ErrGiftCapacity) || updated != full {
		t.Fatalf("full award = %+v, %v", updated, err)
	}
}
