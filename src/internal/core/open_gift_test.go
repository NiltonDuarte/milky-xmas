package core

import (
	"errors"
	"testing"
	"time"
)

func TestOpenGiftSurprises(t *testing.T) {
	full := Inventory{Normal: 1, Rare: 1, Epic: 1}
	kept := Inventory{Normal: 1, Rare: 1, Epic: 1, Shields: 1, ExtraWaterings: 1}
	for _, tt := range []struct {
		name   string
		start  Inventory
		rarity GiftRarity
		roll   int
		want   Surprise
		after  Inventory
	}{
		{"normal snowball", full, NormalGift, 0, SurpriseSnowball, Inventory{Rare: 1, Epic: 1, Snowballs: 1}},
		{"normal coal", full, NormalGift, 1, SurpriseCoal, Inventory{Rare: 1, Epic: 1, Coal: 1}},
		{"normal quote", full, NormalGift, 5, SurpriseQuote, Inventory{Rare: 1, Epic: 1}},
		{"rare snowballs", full, RareGift, 0, SurpriseSnowballs, Inventory{Normal: 1, Epic: 1, Snowballs: 3}},
		{"rare watering", full, RareGift, 1, SurpriseExtraWatering, Inventory{Normal: 1, Epic: 1, ExtraWaterings: 1}},
		{"epic shield", full, EpicGift, 0, SurpriseShield, Inventory{Normal: 1, Rare: 1, Shields: 1}},
		{"epic gifts", full, EpicGift, 1, SurpriseGifts, Inventory{Normal: 3, Rare: 1}},
		{"a second shield", kept, EpicGift, 0, SurpriseGifts, Inventory{Normal: 3, Rare: 1, Shields: 1, ExtraWaterings: 1}},
		{"a second watering", kept, RareGift, 1, SurpriseSnowballs, Inventory{Normal: 1, Epic: 1, Snowballs: 3, Shields: 1, ExtraWaterings: 1}},
	} {
		got, surprise, err := OpenGift(GiftRecord{Inventory: tt.start, Aura: 5}, tt.rarity, tt.roll)
		if err != nil || surprise != tt.want || got.Inventory != tt.after || got.Aura != 5 {
			t.Errorf("%s = %q, %+v, %v", tt.name, surprise, got, err)
		}
	}
	if _, _, err := OpenGift(GiftRecord{}, EpicGift, 0); !errors.Is(err, ErrInsufficientGift) {
		t.Errorf("no gift = %v", err)
	}
	if _, _, err := OpenGift(GiftRecord{Inventory: full}, "coal", 0); !errors.Is(err, ErrGiftRarity) {
		t.Errorf("unknown rarity = %v", err)
	}
}

func TestGrinchShieldBlocksOneHit(t *testing.T) {
	rules := DefaultTuning().Grinch
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{Phase: GrinchSnowballs, StartedAt: now, HP: GrinchBossHP, Joined: map[string]bool{"alice": true},
		PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}, Accepted: map[string]bool{}}
	reply := fight.Move(rules, GrinchMove{Phase: GrinchSnowballs, Member: "alice", RequestID: "1", At: now, CounterRoll: 99, Shield: true})
	if fight.PlayerDamage["alice"] != 0 || fight.Blocked["alice"] != 1 || reply != "Snowball hit for 1.67 HP. Your shield blocked the Grinch's counterattack." {
		t.Fatalf("shielded counterattack = %q, damage %d, blocked %d", reply, fight.PlayerDamage["alice"], fight.Blocked["alice"])
	}
}
