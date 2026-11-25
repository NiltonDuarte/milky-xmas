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
		if got := defaults.Drops.RandomActiveDropEligible(tt.roll); got != tt.want {
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
		if got := defaults.Drops.NextGiftRainThreshold(tt.roll); got != tt.want {
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

func TestActiveDropRarityBoost(t *testing.T) {
	for _, tt := range []struct {
		roll int
		good Aura
		want GiftRarity
	}{
		{69, 24, NormalGift}, {70, 24, RareGift}, {95, 24, EpicGift},
		{49, 25, NormalGift}, {50, 25, RareGift}, {84, 25, RareGift}, {85, 25, EpicGift},
		{70, 50, RareGift}, {99, 50, EpicGift}, {69, -50, NormalGift},
	} {
		if got := defaults.Gifts.ActiveDropRarity(tt.roll, tt.good); got != tt.want {
			t.Errorf("ActiveDropRarity(%d, %d) = %s, want %s", tt.roll, tt.good, got, tt.want)
		}
	}
}

func TestRainCoalAndAwardCoal(t *testing.T) {
	if !defaults.Drops.RainPrizeIsCoal(0) || !defaults.Drops.RainPrizeIsCoal(RainCoalPercent-1) || defaults.Drops.RainPrizeIsCoal(RainCoalPercent) {
		t.Error("coal chance is not the first 25 of 100 rolls")
	}
	record, err := AwardCoal(GiftRecord{Inventory: Inventory{Coal: 2, Normal: 1}, Aura: -5})
	if err != nil || record.Inventory.Coal != 3 || record.Inventory.Normal != 1 || record.Aura != -5 {
		t.Fatalf("AwardCoal = %+v, %v", record, err)
	}
	full := GiftRecord{Inventory: Inventory{Coal: math.MaxInt64}}
	if got, err := AwardCoal(full); !errors.Is(err, ErrCoalCapacity) || got != full {
		t.Fatalf("full AwardCoal = %+v, %v", got, err)
	}
}

var defaults = DefaultTuning()
