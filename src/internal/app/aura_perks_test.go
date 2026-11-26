package app

import (
	"errors"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestListPerksReachDailyWaterAndRain(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	nice, naughty := Member{ServerID: "g", UserID: "nice"}, Member{ServerID: "g", UserID: "naughty"}
	tuning := core.DefaultTuning()
	tuning.Drops.GiftRain = core.GiftRainTuning{MinMessages: 1, MaxMessages: 1, ClaimsToEnd: 5, ClaimCooldown: time.Minute}
	tuning.Drops.RandomActiveOneIn = 1000
	drops, store := tunedDrops(t, &now, tuning, func(int) int { return 0 })
	store.records[nice] = core.GiftRecord{Aura: 25}
	store.records[naughty] = core.GiftRecord{Aura: -25}
	service := drops.service

	claim, err := service.ClaimGift(t.Context(), nice)
	if err != nil || claim.Amount != 2 || claim.Bonus != (core.DailyBonus{Gifts: 1}) || claim.Total != 2 {
		t.Fatalf("nice daily = %+v, %v", claim, err)
	}
	claim, err = service.ClaimGift(t.Context(), naughty)
	if err != nil || claim.Amount != 1 || store.records[naughty].Inventory.YellowSnowballs != 1 || store.records[naughty].Inventory.Coal != 1 {
		t.Fatalf("naughty daily = %+v, %v", claim, err)
	}

	tree := service.RewardTree(NewTree(&MemoryTreeStore{}, func() time.Time { return now }))
	got, watering, err := tree.Water(t.Context(), nice)
	if err != nil || got.Level != 13 || watering.Added != 3 || watering.ListChange != 1 {
		t.Fatalf("nice water = %d, %+v, %v", got.Level, watering, err)
	}
	got, watering, err = tree.Water(t.Context(), naughty)
	if err != nil || got.Level != 14 || watering.Added != 1 || watering.ListChange != -1 {
		t.Fatalf("naughty water = %d, %+v, %v", got.Level, watering, err)
	}

	// Watering is a good deed only from zero up, so the naughty player stays on the list.
	if store.records[nice].Aura != 26 || store.records[naughty].Aura != -25 {
		t.Fatalf("auras after watering = %d, %d", store.records[nice].Aura, store.records[naughty].Aura)
	}

	rain := drops.Observe("g", "c")
	if rain == nil || rain.Kind != core.GiftRainDrop {
		t.Fatalf("rain = %+v", rain)
	}
	for _, tt := range []struct {
		member  Member
		doubled bool
		gifts   int64
	}{{naughty, true, 3}, {nice, false, 3}} {
		award, err := drops.Claim(t.Context(), GiftDropClaim{Member: tt.member, DropID: rain.ID})
		if err != nil || award.Doubled != tt.doubled || store.records[tt.member].Inventory.GiftTotal() != tt.gifts {
			t.Errorf("%s rain = %+v, %d gifts, %v", tt.member.UserID, award, store.records[tt.member].Inventory.GiftTotal(), err)
		}
	}

	// The Naughty list waits half as long between rain gifts.
	now = now.Add(30 * time.Second)
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: nice, DropID: rain.ID}); !errors.Is(err, ErrDropCooldown) {
		t.Errorf("nice rain after 30s = %v", err)
	}
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: naughty, DropID: rain.ID}); err != nil {
		t.Errorf("naughty rain after 30s = %v", err)
	}
}
