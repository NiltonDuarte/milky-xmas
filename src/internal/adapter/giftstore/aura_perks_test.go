package giftstore_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestAuraCountsOneSendPerPlayerADay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gifts.db")
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	open := func() (*giftstore.SQLite, app.Service) {
		store, err := giftstore.OpenSQLite(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		// The daily limit has its own test.
		tuning := core.DefaultTuning()
		tuning.Aura.DailyLimit = 0
		return store, app.NewService(store, func() time.Time { return now }, func(int) int { return 0 }).WithTuning(tuning)
	}
	store, service := open()
	alice := app.Member{ServerID: "guild", UserID: "alice"}
	bob := app.Member{ServerID: "guild", UserID: "bob"}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Next: core.GiftRecord{Inventory: core.Inventory{Rare: 10, Coal: 10}}}); err != nil || !saved {
		t.Fatal(saved, err)
	}
	send := func(to, item string, repeated bool, aura core.Aura) {
		t.Helper()
		effect, err := service.SendItem(t.Context(), alice, app.Recipient{Member: app.Member{ServerID: "guild", UserID: to}}, item, 1)
		if err != nil {
			t.Fatal(err)
		}
		current, err := service.CheckAura(t.Context(), alice)
		if err != nil || (effect == app.AuraRepeated) != repeated || current != aura {
			t.Fatalf("send %s to %s = effect %d, aura %d, %v; want repeated %v, %d", item, to, effect, current, err, repeated, aura)
		}
	}
	send(bob.UserID, "rare", false, 10)
	send(bob.UserID, "rare", true, 10)
	send(bob.UserID, app.CoalItem, false, 5) // The other direction counts on its own.
	send(bob.UserID, app.CoalItem, true, 5)
	send("carol", "rare", false, 15)

	// The limit survives a restart and lifts after a day.
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, service = open()
	defer store.Close()
	now = now.Add(24*time.Hour - time.Second)
	send(bob.UserID, "rare", true, 15)
	now = now.Add(time.Second)
	send(bob.UserID, "rare", false, 25)
}

func TestEncounterReadsThePlayersList(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "gifts.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	service := app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	nice := app.Member{ServerID: "guild", UserID: "nice"}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: nice, Next: core.GiftRecord{Aura: 30}}); err != nil || !saved {
		t.Fatal(saved, err)
	}
	rules := classicFight(core.DefaultTuning().Grinch)
	rules.EarlyStartPlayers = 1
	encounter := service.AuraEncounter(app.NewGrinchEncounter(store, "guild", "guild:channel", rules, func() time.Time { return now }, func(int) int { return 0 }))
	if spawned, err := encounter.Spawn(t.Context(), time.Time{}, true, &announcedDisplay{}); err != nil || !spawned {
		t.Fatal(spawned, err)
	}
	fight, err := encounter.Sync(t.Context(), &announcedDisplay{})
	if err != nil {
		t.Fatal(err)
	}
	for i, phase := range []core.GrinchPhase{core.GrinchLobby, core.GrinchSnowballs} {
		fight, _, err = encounter.Play(t.Context(), fight.ID, fight.MessageID,
			core.GrinchMove{Phase: phase, Member: nice.UserID, RequestID: string(phase), At: now.Add(time.Duration(i) * time.Second)})
		if err != nil {
			t.Fatal(err)
		}
	}
	if fight.HP != core.GrinchBossHP-500 {
		t.Errorf("boss HP = %d after a Nice list snowball", fight.HP)
	}
}

type announcedDisplay struct{}

func (*announcedDisplay) Announce(context.Context, core.Grinch) (string, time.Time, error) {
	return "message", time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC), nil
}
func (*announcedDisplay) Update(context.Context, core.Grinch) error { return nil }

func TestDailyAuraLimitIsStored(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "gifts.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	service := app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	alice := app.Member{ServerID: "guild", UserID: "alice"}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Next: core.GiftRecord{Inventory: core.Inventory{Epic: 3}}}); err != nil || !saved {
		t.Fatal(saved, err)
	}
	for _, step := range []struct {
		to     string
		effect app.AuraEffect
		aura   core.Aura
	}{{"bob", app.AuraMoved, 15}, {"carol", app.AuraLimited, 15}} {
		effect, err := service.SendItem(t.Context(), alice, app.Recipient{Member: app.Member{ServerID: "guild", UserID: step.to}}, "epic", 1)
		current, auraErr := service.CheckAura(t.Context(), alice)
		if err != nil || auraErr != nil || effect != step.effect || current != step.aura {
			t.Fatalf("send to %s = effect %d, aura %d, %v %v", step.to, effect, current, err, auraErr)
		}
	}
	// A send that the limit stopped can still count the next day.
	now = now.Add(24 * time.Hour)
	effect, err := service.SendItem(t.Context(), alice, app.Recipient{Member: app.Member{ServerID: "guild", UserID: "carol"}}, "epic", 1)
	current, auraErr := service.CheckAura(t.Context(), alice)
	if err != nil || auraErr != nil || effect != app.AuraMoved || current != 30 {
		t.Fatalf("next day = effect %d, aura %d, %v %v", effect, current, err, auraErr)
	}
}
