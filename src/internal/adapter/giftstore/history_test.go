package giftstore

import (
	"path/filepath"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestHistoryKeepsTransfersAndEveryFinishedFight(t *testing.T) {
	store, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "history.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	alice, bob := app.Member{ServerID: "s", UserID: "alice"}, app.Member{ServerID: "s", UserID: "bob"}
	stocked := core.GiftRecord{Inventory: core.Inventory{Rare: 3}}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Next: stocked, Reason: app.ReasonDaily}); err != nil || !saved {
		t.Fatalf("save = %v, %v", saved, err)
	}
	send := app.GiftUpdate{Member: alice, Previous: stocked, Next: core.GiftRecord{Inventory: core.Inventory{Rare: 1}, Aura: 20}, Reason: app.ReasonSendGift}
	receive := app.GiftUpdate{Member: bob, Next: core.GiftRecord{Inventory: core.Inventory{Rare: 2}}, Reason: app.ReasonSendGift}
	if saved, err := store.SaveTransfer(t.Context(), send, receive); err != nil || !saved {
		t.Fatalf("transfer = %v, %v", saved, err)
	}
	// A save that loses the race keeps nothing.
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: alice, Previous: stocked, Next: stocked, Reason: app.ReasonDaily}); err != nil || saved {
		t.Fatalf("stale save = %v, %v", saved, err)
	}
	var rows, sent, aura int
	if err := store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM player_history").Scan(&rows); err != nil || rows != 3 {
		t.Fatalf("history rows = %d, %v, want 3", rows, err)
	}
	if err := store.db.QueryRowContext(t.Context(),
		"SELECT rare, aura FROM player_history WHERE user_id = 'alice' AND peer_id = 'bob' AND reason = 'send_gift'").Scan(&sent, &aura); err != nil || sent != -2 || aura != 20 {
		t.Fatalf("sender history = %d rare, %d aura, %v", sent, aura, err)
	}

	fight := core.Grinch{ID: "1", Phase: core.GrinchLobby}
	revision := int64(0)
	save := func(next core.Grinch, rewards map[string]core.GrinchReward) {
		t.Helper()
		if saved, err := store.SaveGrinch(t.Context(), "s:c", revision, next, "s", rewards); err != nil || !saved {
			t.Fatalf("save fight = %v, %v", saved, err)
		}
		revision++
	}
	save(fight, nil)
	fight.Phase = core.GrinchDefeat
	save(fight, map[string]core.GrinchReward{"alice": {Coal: 1}})
	save(fight, nil) // The final message was shown.
	save(core.Grinch{ID: "2", Phase: core.GrinchWon}, nil)
	save(core.Grinch{ID: "3", Phase: core.GrinchCancelled}, nil)
	if err := store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM grinch_history").Scan(&rows); err != nil || rows != 2 {
		t.Fatalf("fights kept = %d, %v, want 2", rows, err)
	}
	if err := store.db.QueryRowContext(t.Context(),
		"SELECT coal FROM player_history WHERE reason = 'grinch_reward'").Scan(&sent); err != nil || sent != 1 {
		t.Fatalf("reward history = %d coal, %v", sent, err)
	}
}
