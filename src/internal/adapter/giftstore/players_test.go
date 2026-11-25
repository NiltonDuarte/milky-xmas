package giftstore_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGiftTransferAfterLegacyMigration(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	now := time.Date(2026, 12, 1, 19, 0, 0, 0, time.UTC)
	execGiftSQL(t, path, `
		CREATE TABLE gift_claims (
			server_id TEXT NOT NULL, user_id TEXT NOT NULL,
			total INTEGER NOT NULL CHECK (total > 0), last_claim TEXT NOT NULL,
			PRIMARY KEY (server_id, user_id)
		);
		INSERT INTO gift_claims VALUES ('guild', 'alice', 2, '2026-12-01T18:00:00Z');
	`)
	alice := app.Member{ServerID: "guild", UserID: "alice"}
	bob := app.Member{ServerID: "guild", UserID: "bob"}

	store := openGiftPlayers(t, path)
	service := app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	if err := service.SendGift(ctx, alice, app.Recipient{Member: bob}, core.NormalGift, 2); err != nil {
		t.Fatalf("send migrated gifts: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openGiftPlayers(t, path)
	service = app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	got, err := service.CheckAura(ctx, alice)
	if err != nil || got != 10 {
		t.Fatalf("sender aura after restart = %+v, %v", got, err)
	}
	got, err = service.CheckAura(ctx, bob)
	if err != nil || got != 0 {
		t.Fatalf("recipient aura after restart = %+v, %v", got, err)
	}
	claim, err := service.ClaimGift(ctx, alice)
	if err != nil || claim.Collected || claim.Total != 0 || claim.Remaining != 23*time.Hour {
		t.Fatalf("claim with empty migrated inventory = %+v, %v", claim, err)
	}
	claim, err = service.ClaimGift(ctx, bob)
	if err != nil || !claim.Collected || claim.Amount != 1 || claim.Total != 3 {
		t.Fatalf("first claim after receiving gifts = %+v, %v", claim, err)
	}

	for _, member := range []app.Member{{ServerID: "another-guild", UserID: "alice"}, {ServerID: "guild", UserID: "new"}} {
		aura, err := service.CheckAura(ctx, member)
		if err != nil || aura != 0 {
			t.Errorf("initial aura for %+v = %+v, %v", member, aura, err)
		}
	}
}

func TestGiftTransfersCompeteForLastGift(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	stores := []*giftstore.SQLite{openGiftPlayers(t, path), openGiftPlayers(t, path)}
	sender := app.Member{ServerID: "guild", UserID: "alice"}
	newService := func(store *giftstore.SQLite) app.Service {
		return app.NewService(store, time.Now, func(n int) int {
			if n == core.GiftRarityRolls {
				return 95
			}
			return 0
		})
	}
	claim, err := newService(stores[0]).ClaimGift(ctx, sender)
	if err != nil || claim.Amount != 1 || claim.Rarity != core.EpicGift {
		t.Fatalf("seed gift = %+v, %v", claim, err)
	}
	targets := []app.Member{{ServerID: "guild", UserID: "bob"}, {ServerID: "guild", UserID: "carol"}}
	start := make(chan struct{})
	results := make(chan error, len(targets))
	for i, target := range targets {
		go func() {
			<-start
			err := newService(stores[i]).SendGift(ctx, sender, app.Recipient{Member: target}, core.EpicGift, 1)
			results <- err
		}()
	}
	close(start)
	success := 0
	for range targets {
		switch err := <-results; {
		case err == nil:
			success++
		case !errors.Is(err, core.ErrInsufficientGift):
			t.Errorf("unexpected transfer result: %v", err)
		}
	}
	record, err := stores[0].Load(ctx, sender)
	if err != nil || success != 1 || record.Inventory.GiftTotal() != 0 || record.Aura != 30 {
		t.Errorf("sender = %+v, err = %v, successful transfers = %d", record, err, success)
	}
	var received int64
	for _, target := range targets {
		record, err := stores[0].Load(ctx, target)
		if err != nil || record.Aura != 0 {
			t.Errorf("recipient = %+v, err = %v", record, err)
		}
		received += record.Inventory.Epic
	}
	if received != 1 {
		t.Errorf("recipients gained %d epic gifts, want 1", received)
	}
}

func TestGiftTransferRollsBackSenderWhenRecipientWriteFails(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	store := openGiftPlayers(t, path)
	execGiftSQL(t, path, `
		CREATE TRIGGER reject_recipient BEFORE INSERT ON gift_players
		WHEN NEW.user_id = 'bob' BEGIN SELECT RAISE(ABORT, 'test storage error'); END;
	`)
	from := app.Member{ServerID: "guild", UserID: "alice"}
	to := app.Member{ServerID: "guild", UserID: "bob"}
	service := app.NewService(store, time.Now, func(int) int { return 0 })
	if _, err := service.ClaimGift(ctx, from); err != nil {
		t.Fatal(err)
	}
	if err := service.SendGift(ctx, from, app.Recipient{Member: to}, core.NormalGift, 1); err == nil {
		t.Fatal("send succeeded despite storage failure")
	}
	sender, err := store.Load(ctx, from)
	if err != nil || sender.Inventory.Normal != 1 || sender.Aura != 0 {
		t.Errorf("sender after rollback = %+v, %v", sender, err)
	}
	recipient, err := store.Load(ctx, to)
	if err != nil || recipient != (core.GiftRecord{}) {
		t.Errorf("recipient after rollback = %+v, %v", recipient, err)
	}
	execGiftSQL(t, path, "DROP TRIGGER reject_recipient")
	if err := service.SendGift(ctx, from, app.Recipient{Member: to}, core.NormalGift, 1); err != nil {
		t.Errorf("send after storage recovered: %v", err)
	}
	if aura, err := service.CheckAura(ctx, from); err != nil || aura != 5 {
		t.Errorf("aura after storage recovered = %d, %v", aura, err)
	}
}

func openGiftPlayers(t *testing.T, path string) *giftstore.SQLite {
	t.Helper()
	store, err := giftstore.OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func execGiftSQL(t *testing.T, path, statement string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(t.Context(), statement); err != nil {
		t.Fatal(err)
	}
}

func TestInventoryUpgradeAndTransfer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	execGiftSQL(t, path, `
		CREATE TABLE gift_players (
			server_id TEXT NOT NULL, user_id TEXT NOT NULL,
			normal INTEGER NOT NULL DEFAULT 0, rare INTEGER NOT NULL DEFAULT 0, epic INTEGER NOT NULL DEFAULT 0,
			bad INTEGER NOT NULL DEFAULT 0, good INTEGER NOT NULL DEFAULT 0, last_claim TEXT NOT NULL DEFAULT '',
			PRIMARY KEY (server_id, user_id)
		);
		INSERT INTO gift_players VALUES ('guild', 'alice', 1, 2, 3, 4, 5, '2026-12-01T18:00:00Z');
	`)
	ctx := t.Context()
	store := openGiftPlayers(t, path)
	alice := app.Member{ServerID: "guild", UserID: "alice"}
	bob := app.Member{ServerID: "guild", UserID: "bob"}
	now := time.Date(2026, 12, 1, 19, 0, 0, 0, time.UTC)
	service := app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	check := func(service app.Service, member app.Member, want core.Inventory) {
		t.Helper()
		if got, err := service.ViewInventory(ctx, member); err != nil || got != want {
			t.Fatalf("inventory for %s = %+v, %v; want %+v", member.UserID, got, err, want)
		}
	}
	check(service, alice, core.Inventory{Normal: 1, Rare: 2, Epic: 3})
	check(service, bob, core.Inventory{})

	previous, err := store.Load(ctx, alice)
	if err != nil || previous.Aura != 0 /* the old Bad and Good counters are not carried over */ || !previous.LastClaim.Equal(now.Add(-time.Hour)) {
		t.Fatalf("migrated record = %+v, %v", previous, err)
	}
	next := previous
	next.Inventory.Snowballs = 6
	next.Inventory.Coal = 7
	if saved, err := store.Save(ctx, app.GiftUpdate{Member: alice, Previous: previous, Next: next}); err != nil || !saved {
		t.Fatalf("save inventory = %v, %v", saved, err)
	}
	if err := service.SendGift(ctx, alice, app.Recipient{Member: bob}, core.RareGift, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	service = app.NewService(openGiftPlayers(t, path), func() time.Time { return now }, func(int) int { return 0 })
	check(service, alice, core.Inventory{Normal: 1, Rare: 1, Epic: 3, Snowballs: 6, Coal: 7})
	check(service, bob, core.Inventory{Rare: 1})
}
