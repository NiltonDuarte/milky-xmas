package app

import (
	"errors"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func tunedDrops(t *testing.T, now *time.Time, tuning core.Tuning, random func(int) int) (*ActivityDrops, *dropStoreFixture) {
	t.Helper()
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	service := NewService(store, func() time.Time { return *now }, random).WithTuning(tuning)
	return NewActivityDrops(service), store
}

func TestGiftRainUsesTunedNumbers(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	tuning := core.DefaultTuning()
	tuning.Drops.GiftRain = core.GiftRainTuning{MinMessages: 2, MaxMessages: 3, ClaimsToEnd: 3, ClaimCooldown: 5 * time.Minute, CoalPercent: 0}
	tuning.Drops.RandomActiveOneIn = 1
	var bounds []int
	drops, store := tunedDrops(t, &now, tuning, func(bound int) int {
		bounds = append(bounds, bound)
		switch bound {
		case 2: // the rain range: two thresholds
			return 0
		case 1: // one in one: a Random Active Drop on every message
			return 0
		default:
			return 99 // rarity: epic with the default daily odds
		}
	})

	// The rain threshold is 2. The first message rolls a Random Active Drop (one in one).
	first := drops.Observe("g", "c")
	if first == nil || first.Kind != core.RandomActiveDrop {
		t.Fatalf("first message = %+v, want a Random Active Drop", first)
	}
	rain := drops.Observe("g", "c")
	if rain == nil || rain.Kind != core.GiftRainDrop {
		t.Fatalf("second message = %+v, want the Gift Rain", rain)
	}
	if !contains(bounds, 2) {
		t.Fatalf("random bounds %v miss the tuned message range of 2", bounds)
	}

	member := func(id string) Member { return Member{ServerID: "g", UserID: id} }
	for i, id := range []string{"a", "b", "c"} {
		award, err := drops.Claim(t.Context(), GiftDropClaim{Member: member(id), DropID: rain.ID})
		if err != nil || award.Coal || award.Rarity != core.EpicGift {
			t.Fatalf("claim %d = %+v, %v", i+1, award, err)
		}
		if (award.Finished != nil) != (i == 2) {
			t.Fatalf("claim %d finished = %v, want only the third to finish", i+1, award.Finished != nil)
		}
	}
	if store.records[member("c")].Inventory.Epic != 1 {
		t.Fatalf("third collector inventory = %+v", store.records[member("c")].Inventory)
	}
}

func TestGiftRainCooldownIsTuned(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	tuning := core.DefaultTuning()
	tuning.Drops.GiftRain.MinMessages, tuning.Drops.GiftRain.MaxMessages = 1, 1
	tuning.Drops.GiftRain.ClaimCooldown = 5 * time.Minute
	drops, _ := tunedDrops(t, &now, tuning, func(int) int { return 0 })
	first := drops.Observe("g", "c")
	if first == nil || first.Kind != core.GiftRainDrop {
		t.Fatalf("rain = %+v", first)
	}
	member := Member{ServerID: "g", UserID: "a"}
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: first.ID}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(4*time.Minute + 59*time.Second)
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: first.ID}); !errors.Is(err, ErrDropCooldown) {
		t.Fatalf("claim before the tuned cooldown = %v", err)
	}
	now = now.Add(time.Second)
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: first.ID}); err != nil {
		t.Fatalf("claim after the tuned cooldown = %v", err)
	}
}

func TestDropLifetimeIsTuned(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	tuning := core.DefaultTuning()
	tuning.Drops.Lifetime = 2 * time.Minute
	tuning.Drops.RandomActiveOneIn = 1
	drops, _ := tunedDrops(t, &now, tuning, func(bound int) int {
		if bound == 1 {
			return 0 // one in one
		}
		return 50
	})
	if drops.Lifetime() != 2*time.Minute {
		t.Fatalf("Lifetime() = %s", drops.Lifetime())
	}
	drop := drops.Observe("g", "c")
	if drop == nil {
		t.Fatal("no drop")
	}
	member := Member{ServerID: "g", UserID: "a"}
	now = now.Add(2*time.Minute - time.Second)
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: drop.ID, Slot: 1}); err != nil {
		t.Fatalf("claim just before the lifetime = %v", err)
	}
	now = now.Add(time.Second)
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: Member{ServerID: "g", UserID: "b"}, DropID: drop.ID, Slot: 2}); !errors.Is(err, ErrDropUnavailable) {
		t.Fatalf("claim at the lifetime = %v", err)
	}
}

func TestRainCoalAndBoostOddsAreTuned(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	tuning := core.DefaultTuning()
	tuning.Drops.GiftRain.MinMessages, tuning.Drops.GiftRain.MaxMessages = 1, 1
	tuning.Drops.GiftRain.CoalPercent = 100
	drops, store := tunedDrops(t, &now, tuning, func(int) int { return 0 })
	rain := drops.Observe("g", "c")
	member := Member{ServerID: "g", UserID: "a"}
	award, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: rain.ID})
	if err != nil || !award.Coal || store.records[member].Inventory.Coal != 1 {
		t.Fatalf("always-coal rain = %+v, %v", award, err)
	}

	tuning = core.DefaultTuning()
	tuning.Gifts.GoodDeedsThreshold = 10
	tuning.Gifts.BoostedRarity = core.RarityOdds{Normal: 0, Rare: 0, Epic: 100}
	tuning.Drops.RandomActiveOneIn = 1
	drops, store = tunedDrops(t, &now, tuning, func(int) int { return 0 })
	boosted := Member{ServerID: "g", UserID: "good"}
	store.records[boosted] = core.GiftRecord{Aura: core.Aura{Good: 10}}
	plain := Member{ServerID: "g", UserID: "plain"}
	store.records[plain] = core.GiftRecord{Aura: core.Aura{Good: 9}}
	drop := drops.Observe("g", "c")
	for slot, tt := range []struct {
		member Member
		want   core.GiftRarity
	}{{boosted, core.EpicGift}, {plain, core.NormalGift}} {
		award, err := drops.Claim(t.Context(), GiftDropClaim{Member: tt.member, DropID: drop.ID, Slot: slot + 1})
		if err != nil || award.Rarity != tt.want {
			t.Fatalf("%s collected %+v, %v; want %s", tt.member.UserID, award, err, tt.want)
		}
	}
}

func contains(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
