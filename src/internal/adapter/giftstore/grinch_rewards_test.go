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
	first := app.NewGrinchEncounter(one, "guild", scope, func() time.Time { return now }, func(int) int { return 0 })
	second := app.NewGrinchEncounter(two, "guild", scope, func() time.Time { return now }, func(int) int { return 0 })
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
		if got.reply == core.GrinchSnowballEarned {
			credits++
		}
	}
	if credits != 1 {
		t.Fatalf("duplicate interaction produced %d credits", credits)
	}
	move.RequestID, move.FinalRound = "second", 2
	if _, reply, err := first.Play(ctx, fight.ID, fight.MessageID, move); err != nil || reply != core.GrinchSnowballEarned {
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
	restarted := app.NewGrinchEncounter(reopened, "guild", scope, func() time.Time { return now }, func(int) int { return 0 })
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
