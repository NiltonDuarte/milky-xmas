package core

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	first := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	record, next, collected, err := (GiftRecord{}).Collect(first, 0, NormalGift)
	if err != nil || !collected || record.Inventory.Normal != 1 || next != first.Add(24*time.Hour) {
		t.Fatalf("first claim = %+v, %v, %v, %v", record, next, collected, err)
	}
	empty := GiftRecord{LastClaim: record.LastClaim}
	refused, retry, collected, err := empty.Collect(next.Add(-time.Nanosecond), 9, EpicGift)
	if err != nil || collected || refused != empty || retry != next {
		t.Fatalf("early claim after giving everything away = %+v, %v, %v, %v", refused, retry, collected, err)
	}
	second, retry, collected, err := record.Collect(next, 9, EpicGift)
	if err != nil || !collected || second.Inventory != (Inventory{Normal: 1, Epic: 10}) || second.LastClaim != next || retry != next.Add(24*time.Hour) {
		t.Fatalf("claim at boundary = %+v, %v, %v, %v", second, retry, collected, err)
	}
}

func TestDailyGiftRarity(t *testing.T) {
	counts := make(map[GiftRarity]int)
	for roll := range GiftRarityRolls {
		counts[defaults.Gifts.DailyGiftRarity(roll)]++
	}
	if counts[NormalGift] != 70 || counts[RareGift] != 25 || counts[EpicGift] != 5 {
		t.Fatalf("rarity counts = %v", counts)
	}
}

func TestTransferGift(t *testing.T) {
	sender := GiftRecord{Inventory: Inventory{Normal: 2, Rare: 2, Epic: 2}, Aura: -15}
	recipient := GiftRecord{Inventory: Inventory{Normal: 1}, Aura: 30}
	for _, tt := range []struct {
		rarity GiftRarity
		from   Inventory
		to     Inventory
		aura   Aura
	}{
		{NormalGift, Inventory{Rare: 2, Epic: 2}, Inventory{Normal: 3}, -5},
		{RareGift, Inventory{Normal: 2, Epic: 2}, Inventory{Normal: 1, Rare: 2}, 5},
		{EpicGift, Inventory{Normal: 2, Rare: 2}, Inventory{Normal: 1, Epic: 2}, 45},
	} {
		t.Run(string(tt.rarity), func(t *testing.T) {
			from, to, err := TransferGift(sender, recipient, tt.rarity, 2)
			if err != nil || from.Inventory != tt.from || to.Inventory != tt.to || from.Aura != tt.aura || to.Aura != recipient.Aura {
				t.Fatalf("transfer = %+v, %+v, %v", from, to, err)
			}
		})
	}
	sender.Aura = 40
	from, _, err := TransferGift(sender, recipient, EpicGift, 1)
	if err != nil || from.Aura != AuraMax {
		t.Fatalf("capped transfer = %+v, %v", from, err)
	}
	recipients := GiftRecord{Inventory: Inventory{Normal: math.MaxInt64}}
	for _, tt := range []struct {
		rarity   GiftRarity
		quantity int64
		target   GiftRecord
		want     error
	}{
		{"legendary", 1, recipient, ErrGiftRarity},
		{NormalGift, 0, recipient, ErrGiftQuantity},
		{RareGift, 3, recipient, ErrInsufficientGift},
		{EpicGift, 1, recipients, ErrGiftCapacity},
	} {
		from, to, err := TransferGift(sender, tt.target, tt.rarity, tt.quantity)
		if !errors.Is(err, tt.want) || from != sender || to != tt.target {
			t.Errorf("invalid transfer = %+v, %+v, %v; want %v without changes", from, to, err, tt.want)
		}
	}
}

func TestAuraMoveStaysOnTheScale(t *testing.T) {
	for _, tt := range []struct {
		from     Aura
		quantity int64
		points   int
		want     Aura
	}{
		{0, 1, 5, 5}, {0, 1, -5, -5}, {5, 3, -5, -10}, {45, 1, 30, AuraMax}, {-48, 2, -5, AuraMin},
		{AuraMin, math.MaxInt64, 30, AuraMax}, {AuraMax, math.MaxInt64, -5, AuraMin},
	} {
		if got := tt.from.Move(tt.quantity, tt.points); got != tt.want {
			t.Errorf("Aura(%d).Move(%d, %d) = %d, want %d", tt.from, tt.quantity, tt.points, got, tt.want)
		}
	}
}
