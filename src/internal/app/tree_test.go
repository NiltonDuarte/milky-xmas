package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestTreeWaterAndView(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	tree := NewTree(&MemoryTreeStore{}, func() time.Time { return now })
	alice := Member{ServerID: "s1", UserID: "alice"}
	if got, err := tree.View(t.Context(), "s1"); err != nil || got.Level != core.TreeStartLevel {
		t.Fatalf("new view = %+v, %v", got, err)
	}
	if got, _, err := tree.Water(t.Context(), alice); err != nil || got.Level != 12 {
		t.Fatalf("water = %+v, %v", got, err)
	}
	if _, _, err := tree.Water(t.Context(), alice); !errors.Is(err, core.ErrTreeCooldown) {
		t.Fatalf("early water = %v", err)
	}
	if got, _ := tree.View(t.Context(), "s2"); got.Level != core.TreeStartLevel {
		t.Fatalf("another server's tree = %+v", got)
	}
	now = now.Add(2 * time.Hour)
	if got, _ := tree.View(t.Context(), "s1"); got.Level != 10 {
		t.Fatalf("view after two hours = %+v", got)
	}
	if got, _, err := tree.Water(t.Context(), alice); err != nil || got.Level != 12 || got.Waters["alice"] != 2 {
		t.Fatalf("water after the cooldown = %+v, %v", got, err)
	}
	if _, _, err := tree.Water(t.Context(), Member{}); err == nil {
		t.Fatal("water without a member succeeded")
	}
}

func TestTreeWaterKeepsEveryConcurrentWatering(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	tree := NewTree(&MemoryTreeStore{}, func() time.Time { return now })
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := tree.Water(context.Background(), Member{ServerID: "s", UserID: string(rune('a' + i))}); err != nil {
				t.Errorf("water %d: %v", i, err)
			}
		}()
	}
	wg.Wait()
	got, _ := tree.View(t.Context(), "s")
	total := 0
	for _, count := range got.Waters {
		total += count
	}
	if total != 20 || got.Level != core.TreeStartLevel+20*core.TreeWaterPoints {
		t.Fatalf("after 20 concurrent waterings: level %d, %d waterings", got.Level, total)
	}
}

func TestTreeCannotBeWateredAfterTheEvent(t *testing.T) {
	now := core.TreeEventEnd.Add(-time.Minute)
	tree := NewTree(&MemoryTreeStore{}, func() time.Time { return now })
	if _, _, err := tree.Water(t.Context(), Member{ServerID: "s", UserID: "alice"}); err != nil {
		t.Fatalf("water before the end = %v", err)
	}
	now = core.TreeEventEnd
	got, _, err := tree.Water(t.Context(), Member{ServerID: "s", UserID: "bob"})
	if !errors.Is(err, core.ErrTreeEventOver) || got.Level != core.TreeStartLevel+core.TreeWaterPoints {
		t.Fatalf("water at the end = %+v, %v", got, err)
	}
	if view, _ := tree.View(t.Context(), "s"); view.Level != got.Level || view.Waters["bob"] != 0 {
		t.Fatalf("view after the end = %+v", view)
	}
}

func TestWateringIsAGoodDeed(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	players := &giftStoreFixture{records: make(map[Member]core.GiftRecord)}
	service := Service{gifts: players}
	tree := service.RewardTree(NewTree(&MemoryTreeStore{}, func() time.Time { return now }))
	alice := Member{ServerID: "s", UserID: "alice"}
	if _, did, err := tree.Water(t.Context(), alice); err != nil || did.GoodDeeds != 1 || players.records[alice].Aura != 1 {
		t.Fatalf("water = %d, %v, aura %d", did.GoodDeeds, err, players.records[alice].Aura)
	}
	if _, did, err := tree.Water(t.Context(), alice); !errors.Is(err, core.ErrTreeCooldown) || did.GoodDeeds != 0 || players.records[alice].Aura != 1 {
		t.Fatalf("refused water = %d, %v, aura %d", did.GoodDeeds, err, players.records[alice].Aura)
	}
	now = now.Add(time.Hour)
	players.records[alice] = core.GiftRecord{Aura: core.AuraMax}
	if _, did, err := tree.Water(t.Context(), alice); err != nil || did.GoodDeeds != 0 || players.records[alice].Aura != core.AuraMax {
		t.Fatalf("water at the top of the scale = %d, %v", did.GoodDeeds, err)
	}
	now = now.Add(time.Hour)
	players.records[alice], players.saveError = core.GiftRecord{Aura: 20}, errors.New("disk full")
	if got, did, err := tree.Water(t.Context(), alice); !errors.Is(err, ErrGoodDeedNotSaved) || did.GoodDeeds != 0 || got.Waters["alice"] != 3 {
		t.Fatalf("water with a failing aura save = %+v, %d, %v", got, did.GoodDeeds, err)
	}
}
