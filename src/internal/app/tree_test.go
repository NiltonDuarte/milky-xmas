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
	if got, err := tree.Water(t.Context(), alice); err != nil || got.Level != 12 {
		t.Fatalf("water = %+v, %v", got, err)
	}
	if _, err := tree.Water(t.Context(), alice); !errors.Is(err, core.ErrTreeCooldown) {
		t.Fatalf("early water = %v", err)
	}
	if got, _ := tree.View(t.Context(), "s2"); got.Level != core.TreeStartLevel {
		t.Fatalf("another server's tree = %+v", got)
	}
	now = now.Add(2 * time.Hour)
	if got, _ := tree.View(t.Context(), "s1"); got.Level != 10 {
		t.Fatalf("view after two hours = %+v", got)
	}
	if got, err := tree.Water(t.Context(), alice); err != nil || got.Level != 12 || got.Waters["alice"] != 2 {
		t.Fatalf("water after the cooldown = %+v, %v", got, err)
	}
	if _, err := tree.Water(t.Context(), Member{}); err == nil {
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
			if _, err := tree.Water(context.Background(), Member{ServerID: "s", UserID: string(rune('a' + i))}); err != nil {
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
