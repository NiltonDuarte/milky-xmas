package giftstore

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestStatsCountWhatPlayersDid(t *testing.T) {
	store, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "stats.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := func() time.Time { return time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC) }
	service := app.NewService(store, now, func(int) int { return 0 })
	tree := service.RewardTree(app.NewTree(store, now))
	alice, bob := app.Member{ServerID: "s", UserID: "alice"}, app.Member{ServerID: "s", UserID: "bob"}
	stock := core.Inventory{Rare: 3, Coal: 12, Snowballs: 1, YellowSnowballs: 1}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Next: core.GiftRecord{Inventory: stock}}); err != nil || !saved {
		t.Fatalf("stock = %v, %v", saved, err)
	}
	to := app.Recipient{Member: bob}
	for _, step := range []func() error{
		func() error { return service.SendGift(t.Context(), alice, to, core.RareGift, 2) },
		func() error { _, err := service.SendItem(t.Context(), alice, to, app.CoalItem, 12); return err },
		func() error { _, err := service.ThrowSnowball(t.Context(), alice, to, core.NormalSnowball); return err },
		func() error { _, err := service.ThrowSnowball(t.Context(), alice, to, core.YellowSnowball); return err },
		func() error { _, _, err := tree.Water(t.Context(), bob); return err },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	if saved, err := store.SaveGrinch(t.Context(), "s:c", 0, core.Grinch{ID: "1", Phase: core.GrinchWon}, "s", nil); err != nil || !saved {
		t.Fatalf("fight = %v, %v", saved, err)
	}
	stats, err := service.Stats(t.Context(), "s", tree)
	if err != nil {
		t.Fatal(err)
	}
	want := core.PlayerActivity{RareGiftsSent: 2, CoalSent: 12, SnowballsThrown: 1, YellowThrown: 1}
	if got := stats.Players["alice"]; got != want || got.BadPoints() != 65 || got.GoodPoints(stats.WaterPoints) != 20 {
		t.Errorf("alice = %+v", got)
	}
	if got := stats.Players["bob"]; got != (core.PlayerActivity{Waterings: 1}) {
		t.Errorf("bob = %+v; receiving must not count", got)
	}
	if stats.Fights != (core.FightTotals{Won: 1}) || stats.TreeLevel != 12 {
		t.Errorf("fights = %+v, tree = %d", stats.Fights, stats.TreeLevel)
	}
	if other, err := service.Stats(t.Context(), "other", tree); err != nil || len(other.Players) != 0 || other.Fights != (core.FightTotals{}) {
		t.Errorf("another server = %+v, %v", other, err)
	}
}
