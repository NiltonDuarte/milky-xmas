package giftstore_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGrinchRewardsAreAtomicAndSurviveRestart(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	one, two := openGiftPlayers(t, path), openGiftPlayers(t, path)
	start := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	now := start
	fight := core.Grinch{ID: "fight", MessageID: "message", Phase: core.GrinchFinalAttack, HP: 500, StartedAt: start, TimedAt: start,
		FinalRound: 1, Joined: map[string]bool{"alice": true, "bob": true, "carol": true, "dave": true},
		PlayerDamage: map[string]int{"dave": 100}, Accepted: make(map[string]bool)}
	const scope = "guild:channel"
	if saved, err := one.SaveGrinch(ctx, scope, 0, fight, "guild", nil); err != nil || !saved {
		t.Fatalf("seed encounter = %v, %v", saved, err)
	}
	first := app.NewGrinchEncounter(one, "guild", scope, classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 })
	second := app.NewGrinchEncounter(two, "guild", scope, classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 })
	move := core.GrinchMove{Phase: core.GrinchFinalAttack, Slot: core.GrinchSnowballSlot, Member: "alice", At: start,
		FinalRound: 1, RequestID: "first"}
	type result struct {
		reply string
		err   error
	}
	results := make(chan result, 2)
	for _, encounter := range []*app.GrinchEncounter{first, second} {
		go func() {
			_, reply, err := encounter.Play(ctx, fight.ID, fight.MessageID, move)
			results <- result{reply, err}
		}()
	}
	credits := 0
	for range 2 {
		got := <-results
		if got.err != nil {
			t.Fatal(got.err)
		}
		if got.reply == classicFight(core.DefaultTuning().Grinch).SnowballEarned() {
			credits++
		}
	}
	if credits != 1 {
		t.Fatalf("duplicate interaction produced %d credits", credits)
	}
	// The Snowball button moved after the first pick.
	moved, _, err := one.LoadGrinch(ctx, scope)
	if err != nil || moved.SnowballSlot() == core.GrinchSnowballSlot {
		t.Fatalf("snowball button at %d, %v", moved.SnowballSlot(), err)
	}
	move.RequestID, move.FinalRound, move.Slot = "second", 2, moved.SnowballSlot()
	if _, reply, err := first.Play(ctx, fight.ID, fight.MessageID, move); err != nil || reply != classicFight(core.DefaultTuning().Grinch).SnowballEarned() {
		t.Fatalf("second click = %q, %v", reply, err)
	}
	alice := app.Member{ServerID: "guild", UserID: "alice"}
	if got, err := one.Load(ctx, alice); err != nil || got.Inventory.Snowballs != 2 {
		t.Fatalf("snowball credits = %+v, %v", got, err)
	}

	now = start.Add(core.GrinchDuration)
	execGiftSQL(t, path, `
		CREATE TRIGGER reject_grinch_reward BEFORE INSERT ON gift_players
		WHEN NEW.user_id = 'carol' BEGIN SELECT RAISE(ABORT, 'test inventory failure'); END;
	`)
	display := &grinchDisplayFixture{}
	if _, err := first.Sync(ctx, display); err == nil {
		t.Fatal("ended the encounter despite failed rewards")
	}
	persisted, _, err := one.LoadGrinch(ctx, scope)
	if err != nil || persisted.Phase != core.GrinchFinalAttack || !persisted.EndedAt.IsZero() {
		t.Fatalf("encounter after rollback = %+v, %v", persisted, err)
	}
	for user := range fight.Joined {
		got, err := one.Load(ctx, app.Member{ServerID: "guild", UserID: user})
		if err != nil || got.Inventory.Coal != 0 || got.Inventory.YellowSnowballs != 0 {
			t.Fatalf("inventory after rollback for %s = %+v, %v", user, got, err)
		}
	}

	execGiftSQL(t, path, "DROP TRIGGER reject_grinch_reward")
	if state, err := second.Sync(ctx, display); err != nil || state.Phase != core.GrinchEscaped {
		t.Fatalf("escape after retry = %+v, %v", state, err)
	}
	if err := one.Close(); err != nil {
		t.Fatal(err)
	}
	if err := two.Close(); err != nil {
		t.Fatal(err)
	}

	reopened := openGiftPlayers(t, path)
	restarted := app.NewGrinchEncounter(reopened, "guild", scope, classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 })
	if _, err := restarted.Sync(ctx, display); err != nil {
		t.Fatal(err)
	}
	for user := range fight.Joined {
		got, err := reopened.Load(ctx, app.Member{ServerID: "guild", UserID: user})
		if err != nil || got.Inventory.Coal != 1 || got.Inventory.YellowSnowballs != 4 {
			t.Errorf("inventory after restart for %s = %+v, %v", user, got, err)
		}
	}
}

func TestGrinchLossRewardsUseTunedNumbers(t *testing.T) {
	ctx := t.Context()
	store := openGiftPlayers(t, filepath.Join(t.TempDir(), "gifts.sqlite"))
	defer store.Close()
	start := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	now := start
	rules := classicFight(core.DefaultTuning().Grinch)
	rules.FightDuration = 30 * time.Second
	rules.LossCoal, rules.LossYellowPerGroupMember = 2, 3
	fight := core.Grinch{ID: "fight", MessageID: "message", Phase: core.GrinchDefense, HP: 7500, StartedAt: start, TimedAt: start,
		DefensePlayers: 3, Joined: map[string]bool{"alice": true, "bob": true, "carol": true},
		PlayerDamage: map[string]int{}, Accepted: make(map[string]bool)}
	const scope = "guild:channel"
	if saved, err := store.SaveGrinch(ctx, scope, 0, fight, "guild", nil); err != nil || !saved {
		t.Fatalf("seed encounter = %v, %v", saved, err)
	}
	encounter := app.NewGrinchEncounter(store, "guild", scope, rules, func() time.Time { return now }, func(int) int { return 0 })

	now = start.Add(29 * time.Second)
	if state, err := encounter.Sync(ctx, &grinchDisplayFixture{}); err != nil || state.Phase != core.GrinchDefense {
		t.Fatalf("before the tuned deadline = %+v, %v", state, err)
	}
	now = start.Add(30 * time.Second)
	if state, err := encounter.Sync(ctx, &grinchDisplayFixture{}); err != nil || state.Phase != core.GrinchEscaped {
		t.Fatalf("at the tuned deadline = %+v, %v", state, err)
	}
	for _, user := range []string{"alice", "bob", "carol"} {
		got, err := store.Load(ctx, app.Member{ServerID: "guild", UserID: user})
		if err != nil || got.Inventory.Coal != 2 || got.Inventory.YellowSnowballs != 9 {
			t.Fatalf("%s inventory = %+v, %v; want 2 Coal and 9 Yellow Snowballs", user, got.Inventory, err)
		}
	}
}

func TestGrinchSpawnChanceAndWindowAreTuned(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := classicFight(core.DefaultTuning().Grinch)
	rules.SpawnChancePercent, rules.ActivityWindow = 50, time.Minute
	for _, tt := range []struct {
		name     string
		roll     int
		activity time.Time
		want     int
	}{
		{"roll inside the tuned chance", 49, now.Add(-time.Minute), 1},
		{"roll at the tuned chance", 50, now.Add(-time.Minute), 0},
		{"activity outside the tuned window", 0, now.Add(-time.Minute - time.Second), 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := openGiftPlayers(t, filepath.Join(t.TempDir(), "gifts.sqlite"))
			defer store.Close()
			encounter := app.NewGrinchEncounter(store, "guild", "guild:channel", rules, func() time.Time { return now }, func(int) int { return tt.roll })
			display := &grinchDisplayFixture{now: now}
			if _, err := encounter.Spawn(t.Context(), tt.activity, false, display); err != nil || display.created != tt.want {
				t.Fatalf("announcements = %d, %v; want %d", display.created, err, tt.want)
			}
		})
	}
}

func TestRevivingCostsAGift(t *testing.T) {
	store := openGiftPlayers(t, filepath.Join(t.TempDir(), "gifts.db"))
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	service := app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	alice := app.Member{ServerID: "guild", UserID: "alice"}
	scope := "guild:channel"
	fight := core.Grinch{ID: "f", MessageID: "m", Phase: core.GrinchSnowballs, StartedAt: now, HP: core.GrinchBossHP, DownAt: now,
		Joined: map[string]bool{"alice": true}, PlayerDamage: map[string]int{"alice": 100},
		DamageDealt: map[string]core.GrinchDamage{}, Accepted: map[string]bool{}}
	if saved, err := store.SaveGrinch(t.Context(), scope, 0, fight, "guild", nil); err != nil || !saved {
		t.Fatal(saved, err)
	}
	encounter := service.AuraEncounter(app.NewGrinchEncounter(store, "guild", scope, core.DefaultTuning().Grinch, func() time.Time { return now }, func(int) int { return 0 }))
	revive := func(request string) (core.Grinch, string) {
		t.Helper()
		state, reply, err := encounter.Play(t.Context(), "f", "m", core.GrinchMove{Revive: core.RareGift, Member: "alice", RequestID: request, At: now})
		if err != nil {
			t.Fatal(err)
		}
		return state, reply
	}
	if state, reply := revive("1"); reply != app.NeedGift || state.PlayerDamage["alice"] != 100 {
		t.Fatalf("without a gift = %q, alice %d", reply, state.PlayerDamage["alice"])
	}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Next: core.GiftRecord{Inventory: core.Inventory{Rare: 2}}}); err != nil || !saved {
		t.Fatal(saved, err)
	}
	state, _ := revive("2")
	record, err := store.Load(t.Context(), alice)
	if err != nil || state.PlayerDamage["alice"] != 50 || record.Inventory.Rare != 1 || record.Aura != 0 {
		t.Fatalf("revival: alice %d, record %+v, %v", state.PlayerDamage["alice"], record, err)
	}
}
